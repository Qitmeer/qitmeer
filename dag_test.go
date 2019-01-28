package main

import (
	"fmt"
	"github.com/noxproject/nox/node"
	"github.com/noxproject/nox/database"
	"github.com/noxproject/nox/log"
	"testing"
	"os"
	"runtime"
	"runtime/debug"
	"github.com/noxproject/nox/common/hash"
	"encoding/hex"
	"github.com/noxproject/nox/core/types"
	"github.com/noxproject/nox/engine/txscript"
	"github.com/noxproject/nox/core/address"
	"github.com/noxproject/nox/crypto/ecc"
	"github.com/noxproject/nox/params"
	"github.com/noxproject/nox/core/blockchain"
	"time"
	"math/rand"
	"github.com/noxproject/nox/services/mining"
	"github.com/noxproject/nox/config"
	"github.com/noxproject/nox/services/blkmgr"
	"github.com/noxproject/nox/services/mempool"
	"container/heap"
	"github.com/noxproject/nox/core/merkle"
	"github.com/noxproject/nox/core/serialization"
	"encoding/binary"
)


func TestDoubleSpent(t *testing.T) {
	addr0:="RmR7D8F2is5jWd5oXaxYyxL5NxwJPeWYXsH"
	addr1:="RmRusZsU1Ui6VZTdG5NiJbjoprhH7uRbzvf"
	privateKey:="ba2d5f8105cc3856d41b48bca01560e112dd0da52056e080ba064d1efc588e8c"


	Gen:=initDag()

	B:=buildBlock("B",[]*hash.Hash{Gen.Hash})
	C:=buildBlock("C",[]*hash.Hash{Gen.Hash})
	D:= buildBlock("D",[]*hash.Hash{Gen.Hash})
	E:= buildBlock("E",[]*hash.Hash{Gen.Hash})

	blockB,_:=ser.GetNoxFull().GetBlockManager().GetChain().BlockByHash(B.GetHash())
	buildDoubleSpentTx(blockB.Transactions()[0].Hash(),addr0,449.9,privateKey)

	F := buildBlock("F", []*hash.Hash{B.GetHash(), C.GetHash()})

	G := buildBlock("G", []*hash.Hash{C.GetHash(), D.GetHash()})

	buildDoubleSpentTx(blockB.Transactions()[0].Hash(),addr1,449.9,privateKey)
	H := buildBlock("H", []*hash.Hash{E.GetHash()})

	//I :=
	buildBlock("I", []*hash.Hash{F.GetHash(), D.GetHash()})

	//J :=
	buildBlock("J", []*hash.Hash{B.GetHash(), G.GetHash(), E.GetHash()})

	//K :=
	buildBlock("K", []*hash.Hash{D.GetHash(), H.GetHash()})
	//



	//
	fmt.Println()
	for _,v:=range barray{
		fmt.Printf("%s = %s\n",v.Name,v.Hash.String())
	}
	//
	fmt.Println()
	lastBlock:=ser.GetNoxFull().GetBlockManager().GetChain().BlockDAG().GetLastBlock()
	fmt.Printf("The Fig.2 Order: %d (%s)\n",lastBlock.GetOrder()+1,
		ser.GetNoxFull().GetBlockManager().GetChain().BlockDAG().GetName())

	var i uint
	for i=0;i<lastBlock.GetOrder()+1;i++ {
		hash:=ser.GetNoxFull().GetBlockManager().GetChain().BlockDAG().GetBlockByOrder(i)
		hb:=bmap[*hash]
		if hb==nil {
			continue
		}
		if hb.Hash.IsEqual(Gen.Hash) {
			fmt.Printf("%s",hb.Name)
		}else{
			fmt.Printf(" -> %s",hb.Name)
		}

	}
	//
	fmt.Println("\nSuppose there are two double spent transactions in F and H.")
	orderF,err:=ser.GetNoxFull().GetBlockManager().GetChain().BlockHeightByHash(F.Hash)
	if err!=nil {
		t.Error(err)
	}
	orderD,err:=ser.GetNoxFull().GetBlockManager().GetChain().BlockHeightByHash(H.Hash)
	if err!=nil {
		t.Error(err)
	}
	if orderF<orderD {
		badTxH:=ser.GetNoxFull().GetBlockManager().GetChain().GetBadTxFromBlock(H.Hash)
		if len(badTxH)>0 {
			fmt.Printf("The result is correct!\nThe badTx(%s) is in the H\n",badTxH[0].String())
		}else{
			fmt.Println("Fail1!")
		}

	}else if orderF>orderD {
		badTxF:=ser.GetNoxFull().GetBlockManager().GetChain().GetBadTxFromBlock(F.Hash)
		if len(badTxF)>0 {
			fmt.Printf("The result is correct!\nThe badTx(%s) is in the C\n",badTxF[0].String())
		}else{
			fmt.Println("Fail2!")
		}
	}else{
		fmt.Println("Fail3!")
	}

	fmt.Println()
	fmt.Println()
	//

	//time.Sleep(5*time.Second)
	End()
}
func buildDoubleSpentTx(txHash *hash.Hash,addrStr string,amount float64,privkeyStr string){

	mtx := types.NewTransaction()
	prevOut := types.NewOutPoint(txHash,2)
	txIn := types.NewTxInput(prevOut, types.NullValueIn, []byte{})
	mtx.AddTxIn(txIn)
	addr, err := address.DecodeAddress(addrStr)
	if err != nil {
		return
	}
	pkScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return
	}

	atomic, err := types.NewAmount(amount)
	if err != nil {
		return
	}
	txOut := types.NewTxOutput(uint64(atomic), pkScript)
	mtx.AddTxOut(txOut)
	//
	privkeyByte, err := hex.DecodeString(privkeyStr)
	if err!=nil {
		return
	}
	privateKey, pubKey := ecc.Secp256k1.PrivKeyFromBytes(privkeyByte)
	h160 := hash.Hash160(pubKey.SerializeCompressed())
	addrP,err := address.NewPubKeyHashAddress(h160,&params.PrivNetParams,ecc.ECDSA_Secp256k1)
	if err!=nil {
		return
	}
	pkScriptP, err := txscript.PayToAddrScript(addrP)
	if err!=nil {
		return
	}
	var kdb txscript.KeyClosure
	kdb = func(types.Address) (ecc.PrivateKey, bool, error){
		return privateKey,true,nil // compressed is true
	}
	var sigScripts [][]byte
	for i,_:= range mtx.TxIn {
		sigScript,err := txscript.SignTxOutput(&params.PrivNetParams,mtx,i,pkScriptP,txscript.SigHashAll,kdb,nil,nil,ecc.ECDSA_Secp256k1)
		if err != nil {
			return
		}
		sigScripts= append(sigScripts,sigScript)
	}

	for i2,_:=range sigScripts {
		mtx.TxIn[i2].SignScript = sigScripts[i2]
	}

/*	mtxHex, err := marshal.MessageToHex(&message.MsgTx{mtx})
	if err != nil {
		return
	}
	fmt.Printf("%s\n",mtxHex)*/

	//
	mp:=ser.GetNoxFull().GetBlockManager().GetMemPool()
	tx := types.NewTx(mtx)
	bestHeight := ser.GetNoxFull().GetBlockManager().GetChain().BestSnapshot().Height
	txType := types.DetermineTxType(mtx)

	utxoView, err := ser.GetNoxFull().GetBlockManager().GetChain().FetchUtxoView(tx)
	if err != nil {
		return
	}
	utxoView.AddTxOuts(tx,int64(bestHeight), types.NullTxIndex)
	//
	var fee int64
	if blockchain.IsCoinBaseTx(mtx) {
		fee=0
	}else{
		for _, txIn := range mtx.TxIn {

			txInHash := &txIn.PreviousOut.Hash
			originTxIndex := txIn.PreviousOut.OutIndex
			utxoEntry := utxoView.LookupEntry(txInHash)
			originTxAtom := utxoEntry.AmountByIndex(originTxIndex)
			fee += int64(originTxAtom)
		}
		for _, txOut := range tx.Transaction().TxOut {
			fee -= int64(txOut.Amount)
		}
	}
	//
	mp.AddTransaction(utxoView, tx, txType, bestHeight, fee)



	/*acceptedTxs, err := ser.GetNoxFull().GetBlockManager().ProcessTransaction(tx, false,
		false,false)
	if err != nil {
		fmt.Println(err,acceptedTxs)
		return
	}*/
	//time.Sleep(time.Second)
}
//some base function
var ser *node.Node=nil
var db database.DB
/////////////////////
var bmap map[hash.Hash]*HelpBlock
var barray []*HelpBlock

type HelpBlock struct {
	Name string
	Hash *hash.Hash
}
func (hb *HelpBlock) GetHash() *hash.Hash{
	return hb.Hash
}
func initDag() *HelpBlock{
	err:=Run()
	if err!=nil {
		return nil
	}
	//time.Sleep(time.Second*5)
	//
	bmap=make(map[hash.Hash]*HelpBlock)

	Gen:=&HelpBlock{"Gen",ser.Params.GenesisHash}
	bmap[*Gen.Hash]=Gen
	barray=append(barray,Gen)

	return Gen
}
func buildBlock(name string,parents []*hash.Hash) *HelpBlock{
	hb:=&HelpBlock{Name:name}

	hash,err:=GenerateBlockByParents(parents)
	if err!=nil {
		fmt.Printf("ERROR:%s [%s]\n",hb.Name,err.Error())
		return nil
	}
	if hash==nil {
		fmt.Printf("%s = nil\n",hb.Name)
		return nil
	}
	hb.Hash=hash

	//fmt.Printf("%s = %s\n",hb.Name,hb.Hash.String())

	bmap[*hash]=hb
	barray=append(barray,hb)

	return hb
}
////////////////////
func Run() error {
	os.Args=[]string{}
	os.Args=append(os.Args,"dag_test")
	os.Args=append(os.Args,"-A=./bin")
	os.Args=append(os.Args,"--privnet")
	os.Args=append(os.Args,"--miningaddr=RmPuiebMrf8mZmt4UAYuhD9PkK9hq6HraBa")
	os.Args=append(os.Args,"--listen=127.0.0.1:1234")
	os.Args=append(os.Args,"--rpcuser=test")
	os.Args=append(os.Args,"--rpcpass=test")
	os.Args=append(os.Args,"--txindex")
	//
	runtime.GOMAXPROCS(runtime.NumCPU())
	debug.SetGCPercent(20)
	//glogger.Verbosity(log.LvlCrit)
	glogger.Verbosity(log.LvlWarn)
	params.PrivNetParams.CoinbaseMaturity=0

	tcfg, _, err := loadConfig()
	if err != nil {
		return err
	}
	//glogger.Verbosity(log.LvlCrit)
	glogger.Verbosity(log.LvlWarn)
	cfg = tcfg

	interrupt := interruptListener()
	defer log.Info("Shutdown complete")

	log.Info("System info", "Nox Version", version(), "Go version",runtime.Version())
	log.Info("System info", "Home dir", cfg.HomeDir)
	if cfg.NoFileLogging {
		log.Info("File logging disabled")
	}
	dbPath := blockDbPath(cfg.DbType)

	fi, err := os.Stat(dbPath)
	if err == nil {
		log.Info(fmt.Sprintf("Removing test database from '%s'", dbPath))
		if fi.IsDir() {
			err := os.RemoveAll(dbPath)
			if err != nil {
				return err
			}
		} else {
			err := os.Remove(dbPath)
			if err != nil {
				return err
			}
		}
	}
	// Load the block database.
	db, err = loadBlockDB()
	if err != nil {
		log.Error("load block database","error", err)
		return err
	}

	// Return now if an interrupt signal was triggered.
	if interruptRequested(interrupt) {
		return nil
	}

	// Create node and start it.
	ser, err = node.NewNode(cfg,db,activeNetParams.Params)
	if err != nil {
		log.Error("Unable to start server","listeners",cfg.Listeners,"error", err)
		return err
	}
	err = ser.RegisterService()
	if err != nil {
		return err
	}
	err = ser.Start()
	if err != nil {
		log.Error("Uable to start server", "error",err)
		return err
	}

	return nil
}
func End(){
	//time.Sleep(time.Hour)
	db.Close()

	ser.Stop()
	ser.WaitForShutdown()
}

func GenerateBlockByParents(parents []*hash.Hash) (*hash.Hash, error) {
	if parents==nil||len(parents)==0 {
		return nil,fmt.Errorf("Parents is invalid")
	}
	m:=ser.GetNoxFull().GetCpuMiner()
	for {
		rand.Seed(time.Now().UnixNano())
		payToAddr := ser.Config.GetMinningAddrs()[rand.Intn(len(ser.Config.GetMinningAddrs()))]
		template, _ := NewBlockTemplate(m.GetPolicy(),ser.Config,ser.Params,
			m.GetSigCache(),ser.GetNoxFull().GetBlockManager().GetMemPool(),m.GetTimeSource(),ser.GetNoxFull().GetBlockManager(),payToAddr,parents)
		//create template

		//solve block
		isSolve:=false
		header := &template.Block.Header
		targetDifficulty := blockchain.CompactToBig(header.Difficulty)
		hashesCompleted := uint64(0)
		maxNonce := ^uint64(0)
		for i := uint64(0); i <= maxNonce; i++ {

			err := mining.UpdateBlockTime(template.Block, ser.GetNoxFull().GetBlockManager(),
				ser.GetNoxFull().GetBlockManager().GetChain(), m.GetTimeSource(),
				ser.Params, ser.Config)
			if err != nil {
				log.Warn("CPU miner unable to update block template time: %v", err)
				return nil,err
			}

			header.Nonce = i
			h := header.BlockHash()
			hashesCompleted += 2

			// The block is solved when the new block hash is less
			// than the target difficulty.  Yay!
			if blockchain.HashToBig(&h).Cmp(targetDifficulty) <= 0 {
				isSolve=true
				break
			}
		}

		//
		if isSolve {

			block := types.NewBlock(template.Block)
			block.SetHeight(template.Height)
			//
			_, err := ser.GetNoxFull().GetBlockManager().ProcessBlock(block, blockchain.BFNone)
			if err != nil {
				return nil,err
			}

			return block.Hash(), nil
		}
	}
}
func NewBlockTemplate(policy *mining.Policy,config *config.Config, params *params.Params,
	sigCache *txscript.SigCache, source mining.TxSource, tsource blockchain.MedianTimeSource,
	blkMgr *blkmgr.BlockManager,  payToAddress types.Address,parents []*hash.Hash) (*types.BlockTemplate, error) {
	txSource := source
	blockManager := blkMgr
	timeSource := tsource
	chainState := blockManager.GetChainState()
	subsidyCache := blockManager.GetChain().FetchSubsidyCache()
	prevHash,nextBlockHeight,_,_ := chainState.GetNextHeightWithState()

	chainBest := blockManager.GetChain().BestSnapshot()
	if *prevHash != chainBest.Hash ||
		nextBlockHeight-1 != chainBest.Height {
		return nil, fmt.Errorf("chain state is not syncronized to the "+
			"blockchain (got %v:%v, want %v,%v",
			prevHash, nextBlockHeight-1, chainBest.Hash, chainBest.Height)
	}
	sourceTxns := txSource.MiningDescs()
	sortedByFee := policy.BlockPrioritySize == 0
	// TODO, impl more general priority func
	lessFunc := txPQByFee
	if sortedByFee {
		lessFunc = txPQByFee
	}
	priorityQueue := newTxPriorityQueue(len(sourceTxns), lessFunc)
	blockTxns := make([]*types.Tx, 0, len(sourceTxns))
	blockUtxos := blockchain.NewUtxoViewpoint()
	dependers := make(map[hash.Hash]map[hash.Hash]*txPrioItem)

	txFees := make([]int64, 0, len(sourceTxns))
	txFeesMap := make(map[hash.Hash]int64)
	txSigOpCounts := make([]int64, 0, len(sourceTxns))
	txSigOpCountsMap := make(map[hash.Hash]int64)
	txFees = append(txFees, -1) // Updated once known

	kilobyte:= 1000
	blockHeaderOverhead := types.MaxBlockHeaderPayload + serialization.MaxVarIntPayload
	coinbaseFlags := "/nox/"
	for _, txDesc := range sourceTxns {
		tx := txDesc.Tx
		msgTx := tx.Transaction()
		if blockchain.IsCoinBaseTx(msgTx) {
			log.Trace("Skipping coinbase tx %s", tx.Hash())
			continue
		}
		utxos, err := blockManager.GetChain().FetchUtxoView(tx)
		if err != nil {
			log.Warn("Unable to fetch utxo view for tx %s: "+
				"%v", tx.Hash(), err)
			continue
		}
		prioItem := &txPrioItem{tx: txDesc.Tx, txType: txDesc.Type}
		prioItem.priority = mempool.CalcPriority(tx.Transaction(), utxos,
			nextBlockHeight)

		txSize := tx.Transaction().SerializeSize()
		prioItem.feePerKB = (float64(txDesc.Fee) * float64(kilobyte)) /
			float64(txSize)
		prioItem.fee = txDesc.Fee

		if prioItem.dependsOn == nil {
			heap.Push(priorityQueue, prioItem)
		}

		viewAEntries := blockUtxos.Entries()
		for h, entryB := range utxos.Entries() {
			if entryA, exists := viewAEntries[h]; !exists ||
				entryA == nil || entryA.IsFullySpent() {
				viewAEntries[h] = entryB
			}
		}
	}
	blockSize := uint32(blockHeaderOverhead)

	blockSigOps := int64(0)
	totalFees := int64(0)
	for priorityQueue.Len() > 0 {
		prioItem := heap.Pop(priorityQueue).(*txPrioItem)
		tx := prioItem.tx
		deps := dependers[*tx.Hash()]
		txSize := uint32(tx.Transaction().SerializeSize())
		numSigOps := int64(blockchain.CountSigOps(tx, false))
		numP2SHSigOps, _ := CountP2SHSigOps(tx, false, blockUtxos)
		numSigOps += int64(numP2SHSigOps)

		for _, txIn := range tx.Transaction().TxIn {
			originHash := &txIn.PreviousOut.Hash
			originIndex := txIn.PreviousOut.OutIndex
			entry := blockUtxos.LookupEntry(originHash)
			if entry != nil {
				entry.SpendOutput(originIndex)
			}

		}

		blockUtxos.AddTxOuts(tx, int64(nextBlockHeight), types.NullTxIndex)


		blockTxns = append(blockTxns, tx)
		blockSize += txSize
		blockSigOps += numSigOps

		txFeesMap[*tx.Hash()] = prioItem.fee
		txSigOpCountsMap[*tx.Hash()] = numSigOps

		for _, item := range deps {
			delete(item.dependsOn, *tx.Hash())
			if len(item.dependsOn) == 0 {
				heap.Push(priorityQueue, item)
			}
		}
	}
	coinbaseScript := []byte{0x00, 0x00}
	coinbaseScript = append(coinbaseScript, []byte(coinbaseFlags)...)

	rand, err := serialization.RandomUint64()
	if err != nil {
		return nil, err
	}

	enData := make([]byte, 12)
	binary.LittleEndian.PutUint32(enData[0:4], uint32(nextBlockHeight))
	binary.LittleEndian.PutUint64(enData[4:12], rand)
	opReturnPkScript, err := txscript.GenerateProvablyPruneableOut(enData)
	if err != nil {
		return nil, err
	}


	voters := 0  //TODO remove voters
	coinbaseTx, err := createCoinbaseTx(subsidyCache,
		coinbaseScript,
		opReturnPkScript,
		int64(nextBlockHeight),    //TODO remove type conversion
		payToAddress,
		uint16(voters),
		params)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	numCoinbaseSigOps := int64(blockchain.CountSigOps(coinbaseTx, true))
	blockSize += uint32(coinbaseTx.Transaction().SerializeSize())
	blockSigOps += numCoinbaseSigOps
	txFeesMap[*coinbaseTx.Hash()] = 0
	txSigOpCountsMap[*coinbaseTx.Hash()] = numCoinbaseSigOps

	// Build tx lists for regular tx.
	blockTxnsRegular := make([]*types.Tx, 0, len(blockTxns)+1)

	// Append coinbase.
	blockTxnsRegular = append(blockTxnsRegular, coinbaseTx)

	// Append regular tx
	for _, tx := range blockTxns {
		blockTxnsRegular = append(blockTxnsRegular, tx)
	}

	for _, tx := range blockTxnsRegular {
		fee, ok := txFeesMap[*tx.Hash()]
		if !ok {
			return nil, fmt.Errorf("couldn't find fee for tx %v",
				*tx.Hash())
		}
		totalFees += fee
		txFees = append(txFees, fee)

		tsos, ok := txSigOpCountsMap[*tx.Hash()]
		if !ok {
			return nil, fmt.Errorf("couldn't find sig ops count for tx %v",
				*tx.Hash())
		}
		txSigOpCounts = append(txSigOpCounts, tsos)
	}


	txSigOpCounts = append(txSigOpCounts, numCoinbaseSigOps)

	if nextBlockHeight > 1 {
		blockSize -= serialization.MaxVarIntPayload -
			uint32(serialization.VarIntSerializeSize(uint64(len(blockTxnsRegular))))
		coinbaseTx.Transaction().TxOut[2].Amount += uint64(totalFees)
		txFees[0] = -totalFees
	}

	ts, err := chainState.MedianAdjustedTime(timeSource,config)
	reqDifficulty, err := blockManager.GetChain().CalcNextRequiredDifficulty(ts)
	generatedBlockVersionTest := 1
	blockVersion := generatedBlockVersionTest

	merkles := merkle.BuildMerkleTreeStore(blockTxnsRegular)
	paMerkles :=merkle.BuildParentsMerkleTreeStore(parents)
	var block types.Block
	block.Header = types.BlockHeader{
		Version:      uint32(blockVersion),
		ParentRoot:   *paMerkles[len(paMerkles)-1],
		TxRoot:       *merkles[len(merkles)-1],
		StateRoot:    hash.Hash{}, //TODO, state root
		Timestamp:    ts,
		Difficulty:   reqDifficulty,
	}
	for _,pb:=range parents{
		if err := block.AddParent(pb); err != nil {
			return nil, err
		}
	}
	for _, tx := range blockTxnsRegular {
		block.AddTransaction(tx.Transaction())
	}
	blockTemplate := &types.BlockTemplate{
		Block:           &block,
		Fees:            txFees,
		SigOpCounts:     txSigOpCounts,
		Height:          nextBlockHeight,
		ValidPayAddress: payToAddress != nil,
	}
	blockManager.SetCurrentTemplate(blockTemplate)
	return blockTemplate,nil
}
func createCoinbaseTx(subsidyCache *blockchain.SubsidyCache, coinbaseScript []byte, opReturnPkScript []byte, nextBlockHeight int64, addr types.Address, voters uint16, params *params.Params) (*types.Tx, error) {
	tx := types.NewTransaction()
	tx.AddTxIn(&types.TxInput{
		// Coinbase transactions have no inputs, so previous outpoint is
		// zero hash and max index.
		PreviousOut: *types.NewOutPoint(&hash.Hash{},
			types.MaxPrevOutIndex ),
		Sequence:        types.MaxTxInSequenceNum,
		BlockHeight:     types.NullBlockHeight,
		TxIndex:         types.NullTxIndex,
		SignScript:      coinbaseScript,
	})

	// Block one is a special block that might pay out tokens to a ledger.
	if nextBlockHeight == 1 && len(params.BlockOneLedger) != 0 {
		// Convert the addresses in the ledger into useable format.
		addrs := make([]types.Address, len(params.BlockOneLedger))
		for i, payout := range params.BlockOneLedger {
			addr, err := address.DecodeAddress(payout.Address)
			if err != nil {
				return nil, err
			}
			addrs[i] = addr
		}

		for i, payout := range params.BlockOneLedger {
			// Make payout to this address.
			pks, err := txscript.PayToAddrScript(addrs[i])
			if err != nil {
				return nil, err
			}
			tx.AddTxOut(&types.TxOutput{
				Amount:   payout.Amount,
				PkScript: pks,
			})
		}

		tx.TxIn[0].AmountIn = params.BlockOneSubsidy()

		return types.NewTx(tx), nil
	}

	// Create a coinbase with correct block subsidy and extranonce.
	subsidy := blockchain.CalcBlockWorkSubsidy(subsidyCache,
		nextBlockHeight,
		voters,
		params)
	tax := blockchain.CalcBlockTaxSubsidy(subsidyCache,
		nextBlockHeight,
		voters,
		params)

	// Tax output.
	if params.BlockTaxProportion > 0 {
		tx.AddTxOut(&types.TxOutput{
			Amount:    uint64(tax),
			PkScript: params.OrganizationPkScript,
		})
	} else {
		// Tax disabled.
		scriptBuilder := txscript.NewScriptBuilder()
		trueScript, err := scriptBuilder.AddOp(txscript.OP_TRUE).Script()
		if err != nil {
			return nil, err
		}
		tx.AddTxOut(&types.TxOutput{
			Amount:    uint64(tax),
			PkScript: trueScript,
		})
	}
	// Extranonce.
	tx.AddTxOut(&types.TxOutput{
		Amount:    0,
		PkScript: opReturnPkScript,
	})
	// AmountIn.
	tx.TxIn[0].AmountIn = subsidy + uint64(tax)  //TODO, remove type conversion

	// Create the script to pay to the provided payment address if one was
	// specified.  Otherwise create a script that allows the coinbase to be
	// redeemable by anyone.
	var pksSubsidy []byte
	if addr != nil {
		var err error
		pksSubsidy, err = txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		scriptBuilder := txscript.NewScriptBuilder()
		pksSubsidy, err = scriptBuilder.AddOp(txscript.OP_TRUE).Script()
		if err != nil {
			return nil, err
		}
	}
	// Subsidy paid to miner.
	tx.AddTxOut(&types.TxOutput{
		Amount:    subsidy,
		PkScript: pksSubsidy,
	})

	return types.NewTx(tx), nil
}
type txPrioItem struct {
	tx       *types.Tx
	txType   types.TxType
	fee      int64
	priority float64
	feePerKB float64
	dependsOn map[hash.Hash]struct{}
}
type txPriorityQueueLessFunc func(*txPriorityQueue, int, int) bool
type txPriorityQueue struct {
	lessFunc txPriorityQueueLessFunc
	items    []*txPrioItem
}
func (pq *txPriorityQueue) Len() int {
	return len(pq.items)
}
func (pq *txPriorityQueue) Less(i, j int) bool {
	return pq.lessFunc(pq, i, j)
}
func (pq *txPriorityQueue) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
}
func (pq *txPriorityQueue) Push(x interface{}) {
	pq.items = append(pq.items, x.(*txPrioItem))
}
func (pq *txPriorityQueue) Pop() interface{} {
	n := len(pq.items)
	item := pq.items[n-1]
	pq.items[n-1] = nil
	pq.items = pq.items[0 : n-1]
	return item
}
func (pq *txPriorityQueue) SetLessFunc(lessFunc txPriorityQueueLessFunc) {
	pq.lessFunc = lessFunc
	heap.Init(pq)
}
func newTxPriorityQueue(reserve int, lessFunc func(*txPriorityQueue, int, int) bool) *txPriorityQueue {
	pq := &txPriorityQueue{
		items: make([]*txPrioItem, 0, reserve),
	}
	pq.SetLessFunc(lessFunc)
	return pq
}
func txPQByFee(pq *txPriorityQueue, i, j int) bool {
	if pq.items[i].feePerKB == pq.items[j].feePerKB {
		return pq.items[i].priority > pq.items[j].priority
	}
	return pq.items[i].feePerKB > pq.items[j].feePerKB
}

func CountP2SHSigOps(tx *types.Tx, isCoinBaseTx bool,utxoView *blockchain.UtxoViewpoint) (int, error) {
	if isCoinBaseTx {
		return 0, nil
	}

	msgTx := tx.Transaction()
	totalSigOps := 0
	for _, txIn := range msgTx.TxIn {
		// Ensure the referenced input transaction is available.
		originTxHash := &txIn.PreviousOut.Hash
		originTxIndex := txIn.PreviousOut.OutIndex
		utxoEntry := utxoView.LookupEntry(originTxHash)
		pkScript := utxoEntry.PkScriptByIndex(originTxIndex)
		if !txscript.IsPayToScriptHash(pkScript) {
			continue
		}

		sigScript := txIn.SignScript
		numSigOps := txscript.GetPreciseSigOpCount(sigScript, pkScript,
			true)

		totalSigOps += numSigOps
	}

	return totalSigOps, nil
}
