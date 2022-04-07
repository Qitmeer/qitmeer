package miner

import (
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/json"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/rpc"
	"github.com/Qitmeer/qitmeer/services/mining"
	"github.com/Qitmeer/qitmeer/version"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

//LL
// gbtNonceRange is two 32-bit big-endian hexadecimal integers which
// represent the valid ranges of nonces returned by the getblocktemplate
// RPC.
const gbtNonceRange = "00000000ffffffff"

type GBTWorker struct {
	started  int32
	shutdown int32

	miner *Miner
	sync.Mutex

	coinbaseAux *json.GetBlockTemplateResultAux
}

func (w *GBTWorker) GetType() string {
	return GBTWorkerType
}

func (w *GBTWorker) Start() error {
	// Already started?
	if atomic.AddInt32(&w.started, 1) != 1 {
		return nil
	}

	log.Info("Start GBT Worker...")
	err := w.miner.initCoinbase()
	if err != nil {
		log.Warn(fmt.Sprintf("You will not be allowed to use <coinbasetxn> :%s", err.Error()))
	}
	w.miner.updateBlockTemplate(false)
	return nil
}

func (w *GBTWorker) Stop() {
	if atomic.AddInt32(&w.shutdown, 1) != 1 {
		log.Warn(fmt.Sprintf("GBT Worker is already in the process of shutting down"))
		return
	}
	log.Info("Stop GBT Worker...")
}

func (w *GBTWorker) IsRunning() bool {
	return atomic.LoadInt32(&w.started) != 0
}

func (w *GBTWorker) Update() {
	if atomic.LoadInt32(&w.shutdown) != 0 {
		return
	}
}

func (w *GBTWorker) GetRequest(request *json.TemplateRequest, reply chan *gbtResponse) {
	if atomic.LoadInt32(&w.shutdown) != 0 {
		reply <- &gbtResponse{nil, fmt.Errorf("GBTWorker is not running ")}
		return
	}

	w.Lock()
	defer w.Unlock()

	// Extract the relevant passed capabilities and restrict the result to
	// either a coinbase value or a coinbase transaction object depending on
	// the request.  Default to only providing a coinbase value.
	useCoinbaseValue := true
	var powtyp byte
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
		powtyp = request.PowType
	}

	// When a coinbase transaction has been requested, respond with an error
	// if there are no addresses to pay the created block template to.
	if !useCoinbaseValue && w.miner.coinbaseAddress == nil {
		reply <- &gbtResponse{nil, rpc.RpcInternalError("No payment addresses specified ",
			"A coinbase transaction has been requested, "+
				"but the server has not been configured with "+
				"any payment addresses via --miningaddr")}
		return
	}

	// Get and return a block template.  A new block template will be
	// generated when the current best block has changed or the transactions
	// in the memory pool have been updated and it has been at least five
	// seconds since the last template was generated.  Otherwise, the
	// timestamp for the existing block template is updated .
	if w.miner.powType != pow.PowType(powtyp) {
		w.miner.powType = pow.PowType(powtyp)
		if err := w.miner.updateBlockTemplate(true); err != nil {
			reply <- &gbtResponse{nil, err}
			return
		}
	}
	result, err := w.getResult(useCoinbaseValue, nil)
	reply <- &gbtResponse{result, err}
}

// blockTemplateResult returns the current block template associated with the
// state as a GetBlockTemplateResult that is ready to be encoded to JSON
// and returned to the caller.
//
// This function MUST be called with the state locked.
func (w *GBTWorker) getResult(useCoinbaseValue bool, submitOld *bool) (*json.GetBlockTemplateResult, error) {
	// Ensure the timestamps are still in valid range for the template.
	// This should really only ever happen if the local clock is changed
	// after the template is generated, but it's important to avoid serving
	// invalid block templates.
	template := w.miner.template
	if template == nil {
		return nil, fmt.Errorf("No template")
	}
	msgBlock := template.Block
	header := &msgBlock.Header
	adjustedTime := w.miner.timeSource.AdjustedTime()
	maxTime := adjustedTime.Add(time.Second * blockchain.MaxTimeOffsetSeconds)
	if header.Timestamp.After(maxTime) {
		return nil, rpc.RpcInvalidError("The template time is after the maximum allowed time for a block - template time %v, maximum time %v", adjustedTime, maxTime)
	}
	// Convert each transaction in the block template to a template result
	// transaction.  The result does not include the coinbase, so notice
	// the adjustments to the various lengths and indices.
	numTx := len(template.Block.Transactions)
	transactions := make([]json.GetBlockTemplateResultTx, 0, numTx-1)
	txIndex := make(map[hash.Hash]int64, numTx)
	for i, tx := range template.Block.Transactions {
		txHash := tx.TxHash()
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
		txBuf, err := tx.Serialize()
		if err != nil {
			context := "Failed to serialize transaction"
			return nil, rpc.RpcInvalidError(err.Error(), context)

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
	diffBig := pow.CompactToBig(template.Difficulty)
	target := fmt.Sprintf("%064x", diffBig)
	longPollID := encodeTemplateID(template.Block.Header.ParentRoot, w.miner.lastTemplate)

	blockFeeMap := map[int]int64{}
	for coinid, val := range template.BlockFeesMap {
		blockFeeMap[int(coinid)] = val
	}
	reply := json.GetBlockTemplateResult{
		StateRoot:    template.Block.Header.StateRoot.String(),
		CurTime:      template.Block.Header.Timestamp.Unix(),
		Height:       int64(template.Height),
		NodeInfo:     version.String() + ":" + w.miner.policy.CoinbaseGenerator.PeerID(),
		Blues:        template.Blues,
		PreviousHash: template.Block.Header.ParentRoot.String(),
		WeightLimit:  types.MaxBlockWeight,
		SigOpLimit:   types.MaxBlockSigOpsCost,
		SizeLimit:    types.MaxBlockPayload,
		//TODO，transactions
		// make([]json.GetBlockTemplateResultTx, 0, 1)
		Parents:      parents,
		Transactions: transactions,
		Version:      template.Block.Header.Version,
		LongPollID:   longPollID,
		//TODO, submitOld
		SubmitOld: submitOld,
		PowDiffReference: json.PowDiffReference{
			Target: target,
			NBits:  strconv.FormatInt(int64(template.Difficulty), 16),
		},
		MinTime: w.miner.minTimestamp.Unix(),
		MaxTime: maxTime.Unix(),
		// gbtMutableFields
		Mutable:    gbtMutableFields,
		NonceRange: gbtNonceRange,
		// TODO, Capabilities
		Capabilities:    gbtCapabilities,
		BlockFeesMap:    blockFeeMap,
		CoinbaseVersion: params.ActiveNetParams.Params.CoinbaseConfig.GetCurrentVersion(int64(template.Height)),
	}

	if useCoinbaseValue {
		reply.CoinbaseAux = w.coinbaseAux
		v := uint64(msgBlock.Transactions[0].TxOut[0].Amount.Value)
		reply.CoinbaseValue = &v
	} else {
		// Ensure the template has a valid payment address associated
		// with it when a full coinbase is requested.
		if !template.ValidPayAddress {
			return nil, rpc.RpcInvalidError("A coinbase transaction has been " +
				"requested, but the server has not " +
				"been configured with any payment " +
				"addresses via --miningaddr")
		}
		// Serialize the transaction for conversion to hex.
		tx := msgBlock.Transactions[0]
		txBuf, err := tx.Serialize()
		if err != nil {
			context := "Failed to serialize transaction"
			return nil, rpc.RpcInvalidError("%s %s", err.Error(), context)
		}

		resultTx := json.GetBlockTemplateResultTx{
			Data:    hex.EncodeToString(txBuf),
			Hash:    tx.TxHash().String(),
			Depends: []int64{},
			Fee:     template.Fees[0],
			SigOps:  template.SigOpCounts[0],
		}

		reply.CoinbaseTxn = &resultTx
	}

	return &reply, nil
}

func NewGBTWorker(miner *Miner) *GBTWorker {
	w := GBTWorker{
		miner: miner,
	}
	w.coinbaseAux = &json.GetBlockTemplateResultAux{
		Flags: hex.EncodeToString(builderScript(txscript.NewScriptBuilder().
			AddData([]byte(mining.CoinbaseFlags)))),
	}
	return &w
}

func builderScript(builder *txscript.ScriptBuilder) []byte {
	script, err := builder.Script()
	if err != nil {
		panic(err)
	}
	return script
}

//LL
// encodeTemplateID encodes the passed details into an ID that can be used to
// uniquely identify a block template.
func encodeTemplateID(prevHash hash.Hash, lastGenerated time.Time) string {
	return fmt.Sprintf("%s-%d", prevHash.String(), lastGenerated.Unix())
}
