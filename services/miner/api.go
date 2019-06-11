// Copyright (c) 2017-2018 The nox developers

package miner

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"qitmeer/common/hash"
	"qitmeer/core/blockchain"
	"qitmeer/core/blockdag"
	"qitmeer/core/json"
	"qitmeer/core/types"
	"qitmeer/params/dcr/types"
	"qitmeer/rpc"
	"qitmeer/services/common/error"
	"qitmeer/services/mining"
)

//LL
// gbtNonceRange is two 32-bit big-endian hexadecimal integers which
// represent the valid ranges of nonces returned by the getblocktemplate
// RPC.
const gbtNonceRange = "00000000ffffffff"

// gbtRegenerateSeconds is the number of seconds that must pass before
// a new template is generated when the previous block hash has not
// changed and there have been changes to the available transactions
// in the memory pool.
const gbtRegenerateSeconds = 60

func (c *CPUMiner) APIs() []rpc.API {
	return []rpc.API{
		{
			NameSpace: rpc.DefaultServiceNameSpace,
			Service:   NewPublicMinerAPI(c),
		},
	}
}

type PublicMinerAPI struct {
	miner *CPUMiner
	gbtWorkState *gbtWorkState
}

func NewPublicMinerAPI(c *CPUMiner) *PublicMinerAPI {
	pmAPI:=&PublicMinerAPI{miner:c}
	pmAPI.gbtWorkState=&gbtWorkState{timeSource:c.timeSource}

	return pmAPI
}

func (api *PublicMinerAPI) Generate(numBlocks uint32) ([]string, error) {
	// Respond with an error if there are no addresses to pay the
	// created blocks to.
	if len(api.miner.config.GetMinningAddrs()) == 0 {
		return nil, er.RpcInternalError("No payment addresses specified "+
			"via --miningaddr", "Configuration")
	}

	// Respond with an error if the client is requesting 0 blocks to be generated.
	if numBlocks == 0 {
		return nil, er.RpcInternalError("Invalid number of blocks",
			"Configuration")
	}
	if numBlocks > 3000 {
		return nil, fmt.Errorf("error, more than 1000")
	}
	blockHashes, err := api.miner.GenerateNBlocks(numBlocks)
	if err != nil {
		return nil, er.RpcInternalError("Could not generate blocks,"+err.Error(),
			"miner")
	}
	// Create a reply
	reply := make([]string, numBlocks)

	// Mine the correct number of blocks, assigning the hex representation of the
	// hash of each one to its place in the reply.
	for i, hash := range blockHashes {
		reply[i] = hash.String()
	}

	return reply, nil
}

//func (api *PublicMinerAPI) GetBlockTemplate(request *mining.TemplateRequest) (interface{}, error){
func (api *PublicMinerAPI) GetBlockTemplate() (interface{}, error) {
	// Set the default mode and override it if supplied.
	mode := "template"

	switch mode {
	case "template":
		return handleGetBlockTemplateRequest(api,nil)
	case "proposal":
		//TODO LL, will be added
		//return handleGetBlockTemplateProposal(s, request)
	}
	return nil, er.RpcInvalidError("Invalid mode")
}

//LL
//Attempts to submit new block to network.
//See https://en.bitcoin.it/wiki/BIP_0022 for full specification
func (api *PublicMinerAPI) SubmitBlock(hexBlock string) (interface{}, error) {
	// Deserialize the hexBlock.
	m := api.miner
	m.submitBlockLock.Lock()
	defer m.submitBlockLock.Unlock()

	if len(hexBlock)%2 != 0 {
		hexBlock = "0" + hexBlock
	}
	serializedBlock, err := hex.DecodeString(hexBlock)

	if err != nil {
		return nil, er.RpcDecodeHexError(hexBlock)
	}
	block, err := types.NewBlockFromBytes(serializedBlock)
	if err != nil {
		return nil, er.RpcDeserializationError("Block decode failed: %s", err.Error())
	}

	// Because it's asynchronous, so you must ensure that all tips are referenced
	tips := api.miner.blockManager.GetChain().BlockDAG().GetTips()
	parents := blockdag.NewHashSet()
	parents.AddList(block.Block().Parents)
	if !parents.IsEqual(tips) {
		return fmt.Sprintf("The tips of block is expired."), nil
	}

	// Process this block using the same rules as blocks coming from other
	// nodes.  This will in turn relay it to the network like normal.
	isOrphan, err := api.miner.blockManager.ProcessBlock(block, blockchain.BFNone)
	if err != nil {
		// Anything other than a rule violation is an unexpected error,
		// so log that error as an internal error.
		rErr, ok := err.(blockchain.RuleError)
		if !ok {
			return fmt.Sprintf("Unexpected error while processing "+
				"block submitted via miner: %s", err.Error()), nil
		}
		// Occasionally errors are given out for timing errors with
		// ReduceMinDifficulty and high block works that is above
		// the target. Feed these to debug.
		if api.miner.params.ReduceMinDifficulty &&
			rErr.ErrorCode == blockchain.ErrHighHash {
			return fmt.Sprintf("Block submitted via miner rejected "+
				"because of ReduceMinDifficulty time sync failure: %s", err.Error()), nil
		}

		if rErr.ErrorCode == blockchain.ErrDuplicateBlock {
			return fmt.Sprintf(rErr.Description, err.Error()), nil
		}
		// Other rule errors should be reported.
		return fmt.Sprintf("Block submitted via miner rejected: %s", err.Error()), nil
	}

	if isOrphan {
		return fmt.Sprintf("Block submitted via miner is an orphan building "+
			"on parent"), nil
	}

	// The block was accepted.
	coinbaseTxOuts := block.Block().Transactions[0].TxOut
	coinbaseTxGenerated := uint64(0)
	for _, out := range coinbaseTxOuts {
		coinbaseTxGenerated += out.Amount
	}
	return fmt.Sprintf("Block submitted accepted  hash %s, height %d, amount %d", block.Hash().String(),
		 block.Order(), coinbaseTxGenerated), nil

}

//LL
// handleGetBlockTemplateRequest is a helper for handleGetBlockTemplate which
// deals with generating and returning block templates to the caller. In addition,
// it detects the capabilities reported by the caller
// in regards to whether or not it supports creating its own coinbase (the
// coinbasetxn and coinbasevalue capabilities) and modifies the returned block
// template accordingly.
func handleGetBlockTemplateRequest(api *PublicMinerAPI,request *json.TemplateRequest) (interface{}, error) {
	// Extract the relevant passed capabilities and restrict the result to
	// either a coinbase value or a coinbase transaction object depending on
	// the request.  Default to only providing a coinbase value.
	useCoinbaseValue := true
	if request != nil {
		var hasCoinbaseValue, hasCoinbaseTxn bool
		for _, capability := range request.Capabilities {
			switch capability {
			case "coinbasetxn":
				hasCoinbaseTxn = true
			case "coinbasevalue":
				hasCoinbaseValue = true
			}
		}

		if hasCoinbaseTxn && !hasCoinbaseValue {
			useCoinbaseValue = false
		}
	}

	// When a coinbase transaction has been requested, respond with an error
	// if there are no addresses to pay the created block template to.
	if !useCoinbaseValue && len(api.miner.config.GetMinningAddrs()) == 0 {
		return nil, er.RpcInternalError("No payment addresses specified ",
			"A coinbase transaction has been requested, "+
				"but the server has not been configured with "+
				"any payment addresses via --miningaddr")
	}

	// No point in generating or accepting work before the chain is synced.
	currentOrder := api.miner.blockManager.GetChain().BestSnapshot().Order
	if currentOrder != 0 && !api.miner.blockManager.IsCurrent() {
		return nil, er.RPCClientInInitialDownloadError("Client in initial download ",
			"NOX is downloading blocks...")
	}

	// Protect concurrent access when updating block templates.
	state := api.gbtWorkState
	state.Lock()
	defer state.Unlock()

	// Get and return a block template.  A new block template will be
	// generated when the current best block has changed or the transactions
	// in the memory pool have been updated and it has been at least five
	// seconds since the last template was generated.  Otherwise, the
	// timestamp for the existing block template is updated .
	if err := state.updateBlockTemplate(api, useCoinbaseValue); err != nil {
		return nil, err
	}
	return state.blockTemplateResult(api,useCoinbaseValue, nil)
}

//LL
// encodeTemplateID encodes the passed details into an ID that can be used to
// uniquely identify a block template.
func encodeTemplateID(prevHash hash.Hash, lastGenerated time.Time) string {
	return fmt.Sprintf("%s-%d", prevHash.String(), lastGenerated.Unix())
}


// gbtWorkState houses state that is used in between multiple RPC invocations to
// getblocktemplate.
type gbtWorkState struct {
	sync.Mutex
	lastTxUpdate  time.Time
	lastGenerated time.Time
	prevHash      *hash.Hash
	minTimestamp  time.Time
	template      *types.BlockTemplate
	timeSource    blockchain.MedianTimeSource
}


// updateBlockTemplate creates or updates a block template for the work state.
// A new block template will be generated when the current best block has
// changed or the transactions in the memory pool have been updated and it has
// been long enough since the last template was generated.  Otherwise, the
// timestamp for the existing block template is updated (and possibly the
// difficulty on testnet per the consesus rules).  Finally, if the
// useCoinbaseValue flag is false and the existing block template does not
// already contain a valid payment address, the block template will be updated
// with a randomly selected payment address from the list of configured
// addresses.
//
// This function MUST be called with the state locked.
func (state *gbtWorkState) updateBlockTemplate(api *PublicMinerAPI, useCoinbaseValue bool) error {
	m := api.miner
	lastTxUpdate := m.txSource.LastUpdated()
	if lastTxUpdate.IsZero() {
		lastTxUpdate = time.Now()
	}

	// Generate a new block template when the current best block has
	// changed or the transactions in the memory pool have been updated and
	// it has been at least gbtRegenerateSecond since the last template was
	// generated.
	var targetDifficulty string
	rand.Seed(time.Now().UnixNano())
	merkles:=m.blockManager.GetChain().BlockDAG().BuildMerkleTreeStoreFromTips()
	parentRoot:=merkles[len(merkles)-1]
	template := state.template
	if template == nil || state.prevHash == nil ||
		!state.prevHash.IsEqual(parentRoot) ||
		(state.lastTxUpdate != lastTxUpdate &&
			time.Now().After(state.lastGenerated.Add(time.Second*
				gbtRegenerateSeconds))) {

		// Reset the previous best hash the block template was generated
		// against so any errors below cause the next invocation to try
		// again.
		state.prevHash = nil

		// Choose a payment address at random if the caller requests a
		// full coinbase as opposed to only the pertinent details needed
		// to create their own coinbase.
		var payToAddr types.Address
		if !useCoinbaseValue {
			// Choose a payment address at random.
			payToAddr = m.config.GetMinningAddrs()[rand.Intn(len(m.config.GetMinningAddrs()))]
		}

		// Create a new block template that has a coinbase which anyone
		// can redeem.  This is only acceptable because the returned
		// block template doesn't include the coinbase, so the caller
		// will ultimately create their own coinbase which pays to the
		// appropriate address(es).
		template, err := mining.NewBlockTemplate(m.policy, m.config, m.params, m.sigCache, m.txSource, m.timeSource, m.blockManager, payToAddr, nil)
		if err != nil {
			return er.RpcInvalidError("Failed to create new block template: %s",err.Error())
		}
		msgBlock := template.Block
		targetDifficulty = fmt.Sprintf("%064x",
			blockchain.CompactToBig(msgBlock.Header.Difficulty))

		// Get the minimum allowed timestamp for the block based on the
		// median timestamp of the last several blocks per the chain
		// consensus rules.
		best := m.blockManager.GetChain().BestSnapshot()
		minTimestamp := mining.MinimumMedianTime(best)

		// Update work state to ensure another block template isn't
		// generated until needed.
		state.template = template
		state.lastGenerated = time.Now()
		state.lastTxUpdate = lastTxUpdate
		state.prevHash = &msgBlock.Header.ParentRoot
		state.minTimestamp = minTimestamp

		log.Debug(fmt.Sprintf("Generated block template (timestamp %v, "+
			"target %s, merkle root %s)",
			msgBlock.Header.Timestamp, targetDifficulty,
			msgBlock.Header.ParentRoot))

	} else {
		// Set locals for convenience.
		msgBlock := template.Block
		targetDifficulty = fmt.Sprintf("%064x",
			blockchain.CompactToBig(msgBlock.Header.Difficulty))

		// Update the time of the block template to the current time
		// while accounting for the median time of the past several
		// blocks per the chain consensus rules.
		mining.UpdateBlockTime(msgBlock,m.blockManager,m.blockManager.GetChain(),m.timeSource,m.params,m.config)
		msgBlock.Header.Nonce = 0

		log.Debug(fmt.Sprintf("Updated block template (timestamp %v, "+
			"target %s)", msgBlock.Header.Timestamp,
			targetDifficulty))
	}

	return nil
}


// blockTemplateResult returns the current block template associated with the
// state as a GetBlockTemplateResult that is ready to be encoded to JSON
// and returned to the caller.
//
// This function MUST be called with the state locked.
func (state *gbtWorkState) blockTemplateResult(api *PublicMinerAPI,useCoinbaseValue bool, submitOld *bool) (*json.GetBlockTemplateResult, error) {
	// Ensure the timestamps are still in valid range for the template.
	// This should really only ever happen if the local clock is changed
	// after the template is generated, but it's important to avoid serving
	// invalid block templates.
	m := api.miner
	template := state.template
	msgBlock := template.Block
	header := &msgBlock.Header
	adjustedTime := state.timeSource.AdjustedTime()
	maxTime := adjustedTime.Add(time.Second * blockchain.MaxTimeOffsetSeconds)
	if header.Timestamp.After(maxTime) {
		return nil, er.RpcInvalidError("The template time is after the maximum allowed time for a block - template time %v, maximum time %v", adjustedTime, maxTime)
	}
	// Convert each transaction in the block template to a template result
	// transaction.  The result does not include the coinbase, so notice
	// the adjustments to the various lengths and indices.
	numTx := len(template.Block.Transactions)
	transactions := make([]json.GetBlockTemplateResultTx, 0, numTx-1)
	txIndex := make(map[hash.Hash]int64, numTx)
	for i, tx := range template.Block.Transactions {
		//txHash := tx.TxHash()
		txHash := tx.TxHashFull()
		txIndex[txHash] = int64(i)

		// Skip the coinbase transaction.
		if i == 0 {
			continue
		}

		// Create an array of 1-based indices to transactions that come
		// before this one in the transactions list which this one
		// depends on.  This is necessary since the created block must
		// ensure proper ordering of the dependencies.  A map is used
		// before creating the final array to prevent duplicate entries
		// when multiple inputs reference the same transaction.
		dependsMap := make(map[int64]struct{})
		for _, txIn := range tx.TxIn {
			if idx, ok := txIndex[txIn.PreviousOut.Hash]; ok {
				dependsMap[idx] = struct{}{}
			}
		}
		depends := make([]int64, 0, len(dependsMap))
		for idx := range dependsMap {
			depends = append(depends, idx)
		}

		// Serialize the transaction for later conversion to hex.
		txBuf, err := tx.Serialize(types.TxSerializeFull)
		if err != nil {
			context := "Failed to serialize transaction"
			m.Lock()
			m.started = false
			m.Unlock()
			return nil, er.RpcInvalidError(err.Error(), context)

		}

		//TODO, bTx := btcutil.NewTx(tx)
		resultTx := json.GetBlockTemplateResultTx{
			Data:    hex.EncodeToString(txBuf),
			Hash:    txHash.String(),
			Depends: depends,
			Fee:     template.Fees[i],
			SigOps:  template.SigOpCounts[i],
			//TODO, blockchain.GetTransactionWeight(bTx)
			Weight: 2000000,
		}
		transactions = append(transactions, resultTx)
	}


	//parents
	parents := []json.GetBlockTemplateResultPt{}
	for _, v := range template.Block.Parents {
		resultPt := json.GetBlockTemplateResultPt{
			Data: hex.EncodeToString(v.Bytes()),
			Hash: v.String(),
		}
		parents = append(parents, resultPt)
	}
	//TODO,submitOld

	// gbtMutableFields are the manipulations the server allows to be made
	// to block templates generated by the getblocktemplate RPC.  It is
	// declared here to avoid the overhead of creating the slice on every
	// invocation for constant data.
	gbtMutableFields := []string{
		"time", "transactions/add", "prevblock", "coinbase/append",
	}
	gbtCapabilities := []string{"proposal"}

	targetDifficulty := fmt.Sprintf("%064x", blockchain.CompactToBig(header.Difficulty))
	longPollID := encodeTemplateID(template.Block.Header.ParentRoot, state.lastGenerated)
	reply := json.GetBlockTemplateResult{
		Bits:         strconv.FormatInt(int64(template.Block.Header.Difficulty), 16),
		StateRoot:    template.Block.Header.StateRoot.String(),
		CurTime:      template.Block.Header.Timestamp.Unix(),
		Height:       int64(template.Height),
		PreviousHash: template.Block.Header.ParentRoot.String(),
		WeightLimit:  blockchain.MaxBlockWeight,
		SigOpLimit:   blockchain.MaxBlockSigOpsCost,
		SizeLimit:    wire.MaxBlockPayload,
		//TODO，transactions
		// make([]json.GetBlockTemplateResultTx, 0, 1)
		Parents:      parents,
		Transactions: transactions,
		Version:      template.Block.Header.Version,
		LongPollID:   longPollID,
		//TODO, submitOld
		SubmitOld: submitOld,
		Target:    targetDifficulty,
		MinTime:   state.minTimestamp.Unix(),
		MaxTime:   maxTime.Unix(),
		// gbtMutableFields
		Mutable:    gbtMutableFields,
		NonceRange: gbtNonceRange,
		// TODO, Capabilities
		Capabilities: gbtCapabilities,
	}

	if useCoinbaseValue {
		//reply.CoinbaseValue = &template.Block.Transactions[0].TxOut[0].Amount

		// coinbaseValue = all coinbase value
		var coinbaseValue uint64= uint64(api.miner.blockManager.GetChain().FetchSubsidyCache().CalcBlockSubsidy(int64(template.Height)))
		reply.CoinbaseValue = &coinbaseValue
	} else {
		// Ensure the template has a valid payment address associated
		// with it when a full coinbase is requested.
		if !template.ValidPayAddress {
			return nil, er.RpcInvalidError("A coinbase transaction has been " +
				"requested, but the server has not " +
				"been configured with any payment " +
				"addresses via --miningaddr")
		}
	}
	return &reply, nil
}