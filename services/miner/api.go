// Copyright (c) 2017-2018 The nox developers

package miner

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/noxproject/nox/common/hash"
	"github.com/noxproject/nox/core/blockchain"
	"github.com/noxproject/nox/core/json"
	"github.com/noxproject/nox/core/types"
	wire "github.com/noxproject/nox/params/dcr/types"
	"github.com/noxproject/nox/rpc"
	er "github.com/noxproject/nox/services/common/error"
	"github.com/noxproject/nox/services/mining"
	"github.com/noxproject/nox/core/blockdag"
)

//LL
// gbtNonceRange is two 32-bit big-endian hexadecimal integers which
// represent the valid ranges of nonces returned by the getblocktemplate
// RPC.
const gbtNonceRange = "00000000ffffffff"

// MaxBlockWeight defines the maximum block weight, where "block
// weight" is interpreted as defined in BIP0141. A block's weight is
// calculated as the sum of the of bytes in the existing transactions
// and header, plus the weight of each byte within a transaction. The
// weight of a "base" byte is 4, while the weight of a witness byte is
// 1. As a result, for a block to be valid, the BlockWeight MUST be
// less than, or equal to MaxBlockWeight.
// TODO, will be moved
const MaxBlockWeight = 4000000

// MaxBlockSigOpsCost is the maximum number of signature operations
// allowed for a block. It is calculated via a weighted algorithm which
// weights segregated witness sig ops lower than regular sig ops.
// TODO. will be moved
const MaxBlockSigOpsCost = 80000

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
}

func NewPublicMinerAPI(c *CPUMiner) *PublicMinerAPI {
	return &PublicMinerAPI{c}
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
	capabilities := []string{
		"", "workid", "coinbase/append",
	}
	mode := "template"

	switch mode {
	case "template":
		return handleGetBlockTemplateRequest(api, capabilities)
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

	if len(hexBlock)%2 != 0 {
		hexBlock = "0" + hexBlock
	}
	serializedBlock, err := hex.DecodeString(hexBlock)

	if err != nil {
		return nil, er.RpcDecodeHexError(hexBlock)
	}
	block, err := types.NewBlockFromBytes(serializedBlock)
	if err != nil {
		return nil, er.RpcDeserializationError("Block decode failed: ", err.Error())
	}

	// Because it's asynchronous, so you must ensure that all tips are referenced
	tips := api.miner.blockManager.GetChain().BlockDAG().GetTips()
	parents := blockdag.NewBlockSet()
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
			"on parent: %s", err.Error()), nil
	}

	// The block was accepted.
	coinbaseTxOuts := block.Block().Transactions[0].TxOut
	coinbaseTxGenerated := uint64(0)
	for _, out := range coinbaseTxOuts {
		coinbaseTxGenerated += out.Amount
	}
	return fmt.Sprintf("Block submitted accepted", "hash", block.Hash(),
		"height", block.Height(), "amount", coinbaseTxGenerated), nil

	/*
		if !api.miner.submitBlock(block) {
			return fmt.Sprintf("rejected: %s", err.Error()), nil
		}
	*/

	return nil, nil
}

//LL
// handleGetBlockTemplateRequest is a helper for handleGetBlockTemplate which
// deals with generating and returning block templates to the caller.  It
// handles both long poll requests as specified by BIP 0022 as well as regular
// requests.  In addition, it detects the capabilities reported by the caller
// in regards to whether or not it supports creating its own coinbase (the
// coinbasetxn and coinbasevalue capabilities) and modifies the returned block
// template accordingly.
//func handleGetBlockTemplateRequest(api *PublicMinerAPI, request *mining.TemplateRequest) (interface{}, error) {
func handleGetBlockTemplateRequest(api *PublicMinerAPI, capabilities []string) (interface{}, error) {
	// Extract the relevant passed capabilities and restrict the result to
	// either a coinbase value or a coinbase transaction object depending on
	// the request.  Default to only providing a coinbase value.
	useCoinbaseValue := true
	if capabilities != nil && len(capabilities) > 0 {
		var hasCoinbaseValue, hasCoinbaseTxn bool
		for _, capability := range capabilities {
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

	// Return an error if there are no peers connected since there is no
	// way to relay a found block or receive transactions to work on.
	// However, allow this state when running in the private net mode.
	//TODO LL, will be added
	/*
		if !(api.miner.config.PrivNet) &&
			s.cfg.ConnMgr.ConnectedCount() == 0 {

			return nil, &btcjson.RPCError{
				Code:    btcjson.ErrRPCClientNotConnected,
				Message: "NOX is not connected",
			}
		}
	*/
	m := api.miner
	m.submitBlockLock.Lock()

	// No point in generating or accepting work before the chain is synced.
	currentHeight := api.miner.blockManager.GetChain().BestSnapshot().Height
	if currentHeight != 0 && !api.miner.blockManager.GetChain().IsCurrent() {
		return nil, er.RPCClientInInitialDownloadError("Client in initial download ",
			"NOX is downloading blocks...")
	}

	// When a long poll ID was provided, this is a long poll request by the
	// client to be notified when block template referenced by the ID should
	// be replaced with a new one.
	//TODO LL, will be added
	/*
		if request != nil && request.LongPollID != "" {
			return handleGetBlockTemplateLongPoll(s, request.LongPollID,
				useCoinbaseValue, closeChan)
		}
	*/

	// Protect concurrent access when updating block templates.
	//TODO, LL ???
	/*
		state := s.gbtWorkState
		state.Lock()
		defer state.Unlock()
	*/

	//TODO LL, will be added
	// Get and return a block template.  A new block template will be
	// generated when the current best block has changed or the transactions
	// in the memory pool have been updated and it has been at least five
	// seconds since the last template was generated.  Otherwise, the
	// timestamp for the existing block template is updated (and possibly
	// the difficulty on testnet per the consesus rules).
	/*
		if err := state.updateBlockTemplate(s, useCoinbaseValue); err != nil {
			return nil, err
		}
	*/
	//TODO LL,
	//return state.blockTemplateResult(useCoinbaseValue, nil)
	m.Lock()
	m.started = true
	m.discreteMining = true
	m.speedMonitorQuit = make(chan struct{})
	m.wg.Add(1)
	go m.speedMonitor()
	m.Unlock()

	log.Trace("Generating blocks", "num", 1)

	// Choose a payment address at random.
	rand.Seed(time.Now().UnixNano())
	payToAddr := m.config.GetMinningAddrs()[rand.Intn(len(m.config.GetMinningAddrs()))]

	template, err := mining.NewBlockTemplate(m.policy, m.config, m.params, m.sigCache, m.txSource, m.timeSource, m.blockManager, payToAddr, nil)

	m.submitBlockLock.Unlock()

	if err != nil {
		errStr := fmt.Sprintf("template: %v", err)
		log.Error("Failed to create new block ", "err", errStr)
		//TODO refactor the quit logic
		m.Lock()
		close(m.speedMonitorQuit)
		m.wg.Wait()
		m.started = false
		m.discreteMining = false
		m.Unlock()
		return nil, err //should miner if error
	}
	if template == nil { // should not go here
		log.Debug("Failed to create new block template", "err", "but error=nil")
		return nil, er.RpcInvalidError("Failed to create new block template")
	}

	longPollID := encodeTemplateID(template.Block.Header.ParentRoot, time.Now())
	targetDifficulty := fmt.Sprintf("%064x", blockchain.CompactToBig(template.Block.Header.Difficulty))
	pastMedianTime := api.miner.blockManager.GetChainState().GetPastMedianTime()
	minTime := pastMedianTime.Add(time.Second)
	maxTime := api.miner.timeSource.AdjustedTime().Add(time.Second * blockchain.MaxTimeOffsetSeconds)

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
	var submitOld *bool
	// gbtMutableFields are the manipulations the server allows to be made
	// to block templates generated by the getblocktemplate RPC.  It is
	// declared here to avoid the overhead of creating the slice on every
	// invocation for constant data.
	gbtMutableFields := []string{
		"time", "transactions/add", "prevblock", "coinbase/append",
	}
	gbtCapabilities := []string{"proposal"}

	reply := json.GetBlockTemplateResult{
		Bits:         strconv.FormatInt(int64(template.Block.Header.Difficulty), 16),
		StateRoot:    template.Block.Header.StateRoot.String(),
		CurTime:      template.Block.Header.Timestamp.Unix(),
		Height:       int64(template.Height),
		PreviousHash: template.Block.Header.ParentRoot.String(),
		WeightLimit:  MaxBlockWeight,
		SigOpLimit:   MaxBlockSigOpsCost,
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
		MinTime:   minTime.Unix(),
		MaxTime:   maxTime.Unix(),
		// gbtMutableFields
		Mutable:    gbtMutableFields,
		NonceRange: gbtNonceRange,
		// TODO, Capabilities
		Capabilities: gbtCapabilities,
	}

	if useCoinbaseValue {
		//reply.CoinbaseValue = &template.Block.Transactions[0].TxOut[0].Amount

		var coinbaseValue uint64 // coinbaseValue = all coinbase value
		coinbaseValue = uint64(api.miner.blockManager.GetChain().FetchSubsidyCache().CalcBlockSubsidy(int64(template.Height)))
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

//LL
// encodeTemplateID encodes the passed details into an ID that can be used to
// uniquely identify a block template.
func encodeTemplateID(prevHash hash.Hash, lastGenerated time.Time) string {
	return fmt.Sprintf("%s-%d", prevHash.String(), lastGenerated.Unix())
}
