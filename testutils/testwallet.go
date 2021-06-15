// Copyright (c) 2020 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package testutils

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/address"
	j "github.com/Qitmeer/qitmeer/core/json"
	s "github.com/Qitmeer/qitmeer/core/serialization"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/crypto/bip32"
	"github.com/Qitmeer/qitmeer/crypto/ecc/secp256k1"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/rpc/client"
	"github.com/Qitmeer/qitmeer/rpc/client/cmds"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

var (
	// the default seed used in the testWallet
	defaultSeed = [hash.HashSize]byte{
		0x7e, 0x44, 0x5a, 0xa5, 0xff, 0xd8, 0x34, 0xcb,
		0x2d, 0x3b, 0x2d, 0xb5, 0x0f, 0x89, 0x97, 0xdd,
		0x21, 0xaf, 0x29, 0xbe, 0xc3, 0xd2, 0x96, 0xaa,
		0xa0, 0x66, 0xd9, 0x02, 0xb9, 0x3f, 0x48, 0x4b,
	}
)

type utxo struct {
	pkScript []byte
	value    types.Amount
	//the current wallet synced maturity order + coinbase maturity
	maturity int64
	//which hd index of private key/address hold the utxo
	keyIndex uint32
	//the wallet side marker to mark the utxo is being spent
	isSpent bool
}

func (u *utxo) isMature(currentOrder int64) bool {
	return currentOrder >= u.maturity
}

type update struct {
	order int64
	hash  *hash.Hash
	txs   []*types.Transaction
}

//
type undo struct {
	utxosDestroyed map[types.TxOutPoint]*utxo
	utxosCreated   []types.TxOutPoint
}

// testWallet is a simple in-memory wallet works for a test harness instance's
// node. the purpose of testWallet is to provide basic wallet functionality for
// the integrated-test, such as send tx & verify balance etc.
// testWallet works as a HD (BIP-32) wallet
type testWallet struct {
	// the harness node id which wallet is targeted for
	nodeId uint32
	// the bip32 master extended private key from a seed
	hdMaster *bip32.Key
	// the next hd child number from the master
	hdChildNumer uint32
	// addrs are all addresses which belong to the master private key.
	// the keys of address map are their hd child numbers.
	addrs map[uint32]types.Address
	// privkeys cached all private keys which derived from the master private key.
	// the keys of the private key map are their hd child number.
	privkeys map[uint32][]byte

	// the utxos set of the wallet.
	utxos map[types.TxOutPoint]*utxo
	// the updates log, it appends itself when a block is connected to the block dag.
	updates       []*update
	updateArrived chan struct{}
	updateMtx     sync.Mutex
	// the map for rewinding the utxo set when a block is disconnected from the block dag.
	undoes     map[*hash.Hash]*undo
	mempoolTx  map[string]string
	confirmTxs map[string]uint64

	// the current synced order the wallet is known
	currentOrder int64
	sync.RWMutex

	netParams        *params.Params
	t                *testing.T
	client           *Client
	maxRescanOrder   uint64
	ScanCount        uint64
	OnRescanComplete func()
}

func (w *testWallet) setRpcClient(client *Client) {
	w.client = client
}

func newTestWallet(t *testing.T, params *params.Params, nodeId uint32) (*testWallet, error) {
	return newTestWalletWithSeed(t, params, &defaultSeed, nodeId)
}

func newTestWalletWithSeed(t *testing.T, params *params.Params, seed *[hash.HashSize]byte, nodeId uint32) (*testWallet, error) {
	// The final seed is seed || nodeId, the purpose to make sure that each harness
	// node use a deterministic private key based on the its node id.
	var finalSeed [hash.HashSize + 4]byte
	// t.Logf("seed is %v",hexutil.Encode(seed[:]))
	copy(finalSeed[:], seed[:])
	// t.Logf("finalseed is %v",hexutil.Encode(finalSeed[:]))
	binary.LittleEndian.PutUint32(finalSeed[hash.HashSize:], nodeId)
	version := bip32.Bip32Version{
		PrivKeyVersion: params.HDPrivateKeyID[:],
		PubKeyVersion:  params.HDPublicKeyID[:],
	}
	// t.Logf("finalseed is %v",hexutil.Encode(finalSeed[:]))
	hdMaster, err := bip32.NewMasterKey2(finalSeed[:], version)
	if err != nil {
		return nil, err
	}
	child0, err := hdMaster.NewChildKey(0)
	if err != nil {
		return nil, err
	}
	key0 := child0.Key
	privkeys := make(map[uint32][]byte)
	privkeys[0] = key0
	addr0, err := privKeyToAddr(key0, params)
	if err != nil {
		return nil, err
	}
	addrs := make(map[uint32]types.Address)
	addrs[0] = addr0
	return &testWallet{
		nodeId:        nodeId,
		hdMaster:      hdMaster,
		hdChildNumer:  1,
		privkeys:      privkeys,
		addrs:         addrs,
		utxos:         make(map[types.TxOutPoint]*utxo),
		undoes:        make(map[*hash.Hash]*undo),
		updateArrived: make(chan struct{}),
		netParams:     params,
		t:             t,
	}, nil
}

// newAddress create a new address from the wallet's key chain.
func (w *testWallet) newAddress() (types.Address, error) {
	num := w.hdChildNumer
	childx, err := w.hdMaster.NewChildKey(num)
	if err != nil {
		return nil, err
	}
	w.privkeys[num] = childx.Key
	addrx, err := privKeyToAddr(childx.Key, w.netParams)
	if err != nil {
		return nil, err
	}
	w.addrs[num] = addrx
	w.hdChildNumer++
	return addrx, nil
}

// NewAddress return a new address from the wallet's key chain
// which is safe for concurrent access
func (m *testWallet) NewAddress() (types.Address, error) {
	m.Lock()
	defer m.Unlock()
	return m.newAddress()
}

// convert the serialized private key into the p2pkh address
func privKeyToAddr(privKey []byte, params *params.Params) (types.Address, error) {
	_, pubKey := secp256k1.PrivKeyFromBytes(privKey)
	serializedKey := pubKey.SerializeCompressed()
	addr, err := address.NewSecpPubKeyAddress(serializedKey, params)
	if err != nil {
		return nil, err
	}
	return addr.PKHAddress(), nil
}

func (w *testWallet) coinBaseAddr() types.Address {
	return w.addrs[0]
}

func (w *testWallet) coinBasePrivKey() []byte {
	return w.privkeys[0]
}

// Start will start a internal goroutine to listen the block dog update notifications
// from the target test harness node with which the wallet can be synced.
func (w *testWallet) Start() {
	go func() {
		var update *update
		// receives update signal from the channel repeatedly until it is closed.
		for range w.updateArrived {
			// pop the new update from the update queue.
			w.updateMtx.Lock()
			update = w.updates[0]
			w.updates[0] = nil // prevent GC leak.
			w.updates = w.updates[1:]
			w.updateMtx.Unlock()
			w.t.Logf("node [%v] update arrvied hash=%v,order=%v", w.nodeId, update.hash, update.order)
			w.Lock()
			if update.order != 0 {
				w.currentOrder++
				if w.currentOrder != update.order {
					w.t.Fatalf("the order not match, expect current is %v but update got %v", w.currentOrder, update.order)
				}
			}
			undo := &undo{
				utxosDestroyed: make(map[types.TxOutPoint]*utxo),
			}
			for _, tx := range update.txs {
				txHash := tx.TxHash()
				isCoinbase := tx.IsCoinBase()
				w.doInputs(tx.TxIn, undo)
				w.doOutputs(tx.TxOut, &txHash, isCoinbase, undo)
			}
			w.undoes[update.hash] = undo
			w.Unlock()
		}
	}()
	gensis, err := w.client.GetSerializedBlock(w.netParams.GenesisHash)
	if err != nil {
		w.t.Fatalf("failed to get gensis block")
	}
	txs := make([]*types.Transaction, 0)
	for _, tx := range gensis.Transactions() {
		txs = append(txs, tx.Tx)
	}
	w.blockConnected(gensis.Hash(), 0, 0, time.Now(), txs)
}

// doOutputs scan each of the passed outputs, creating utxos.
func (w *testWallet) doOutputs(outputs []*types.TxOutput, txHash *hash.Hash, isCoinbase bool, undo *undo) {

	for i, output := range outputs {
		pkScript := output.PkScript

		// Scan all the wallet controlled addresses check if
		// the output is paying to the wallet.
		for keyIndex, addr := range w.addrs {
			pkHash := addr.Script()
			if !bytes.Contains(pkScript, pkHash) {
				continue
			}

			// If a coinbase output mark the maturity
			var maturity int64
			if isCoinbase {
				maturity = w.currentOrder + int64(w.netParams.CoinbaseMaturity)
			}

			op := types.TxOutPoint{Hash: *txHash, OutIndex: uint32(i)}
			w.utxos[op] = &utxo{
				value:    output.Amount,
				keyIndex: keyIndex,
				maturity: maturity,
				pkScript: pkScript,
			}
			undo.utxosCreated = append(undo.utxosCreated, op)
		}
	}
}

// doInputs scans all the passed inputs, destroying utxos
func (w *testWallet) doInputs(inputs []*types.TxInput, undo *undo) {

	for _, txIn := range inputs {
		op := txIn.PreviousOut
		oldUtxo, ok := w.utxos[op]
		if !ok {
			continue
		}

		undo.utxosDestroyed[op] = oldUtxo
		delete(w.utxos, op)
	}
}

// SpendOutputsAndSend will create tx to pay the specified tx outputs
// and send the tx to the test harness node.
func (w *testWallet) PayAndSend(outputs []*types.TxOutput, feePerByte types.Amount, preOutpoint *types.TxOutPoint, lockTime *int64) (*hash.Hash, error) {
	if tx, err := w.createTx(outputs, feePerByte, preOutpoint, lockTime); err != nil {
		return nil, err
	} else {
		txByte, err := tx.Serialize()
		if err != nil {
			return nil, err
		}
		txHex := hex.EncodeToString(txByte[:])
		w.t.Logf("node [%v] has been sent rawtx=%s\n", w.nodeId, txHex)
		return w.client.SendRawTx(txHex, true)
	}
}
func (w *testWallet) createTx(outputs []*types.TxOutput, feePerByte types.Amount, preOutpoint *types.TxOutPoint, lockTime *int64) (*types.Transaction, error) {
	w.Lock()
	defer w.Unlock()
	const (
		// signScriptSize is the largest possible size bytes of a signScript
		// sig may 71, 72, 73, pub may 33 or 32
		// OP_DATA_73 <sig> OP_DATA_33 <pubkey>
		maxSignScriptSize = 1 + 73 + 1 + 33
		// a possible change output
		// <coin_id> <value> <len> OP_DUP OP_HASH160 OP_DATA_20 <pk_hash> OP_EQUALVERIFY OP_CHECKSIG
		changeOutPutSize = 2 + 8 + 1 + 1 + 1 + 1 + 20 + 1 + 1
	)

	if lockTime != nil &&
		(*lockTime < 0 || *lockTime > int64(types.MaxTxInSequenceNum)) {
		return nil, fmt.Errorf("Locktime out of range")
	}

	tx := types.NewTransaction()
	txSize := int64(0)

	totalOutAmt := make(map[types.CoinID]int64)
	totalInAmt := make(map[types.CoinID]int64)
	feeCoinId := feePerByte.Id
	requiredFee := types.Amount{Value: 0, Id: feeCoinId}

	// calculate the total amount need to pay && add output into tx
	for _, o := range outputs {
		totalOutAmt[o.Amount.Id] += o.Amount.Value
		tx.AddTxOut(o)
	}

	// Set the Locktime, if given.
	if lockTime != nil {
		tx.LockTime = uint32(*lockTime)
	}
	useCLTVPubKeyHashTy := false
	enoughFund := false
	// select inputs from utxo set of the wallet && add them into tx
	for txOutPoint, utxo := range w.utxos {
		// skip immature or spent utxo at first
		if !utxo.isMature(w.currentOrder) || utxo.isSpent {
			continue
		}
		// skip the utxo if not in known spent output coin id
		if _, ok := totalOutAmt[utxo.value.Id]; !ok {
			continue
		}
		if txscript.GetScriptClass(0, utxo.pkScript) == txscript.CLTVPubKeyHashTy {
			scripts, _ := txscript.DisasmString(utxo.pkScript)
			arr := strings.Split(scripts, " ")
			needHeight, _ := strconv.ParseInt(arr[0], 16, 32)
			if tx.LockTime < uint32(needHeight) {
				continue
			}
			useCLTVPubKeyHashTy = true
		}
		if preOutpoint != nil {
			if *preOutpoint != txOutPoint {
				continue
			}
		}
		totalInAmt[utxo.value.Id] += utxo.value.Value
		// add selected input into tx
		txIn := types.NewTxInput(&txOutPoint, nil)
		if useCLTVPubKeyHashTy {
			txIn.Sequence = types.MaxTxInSequenceNum - 1
		}
		tx.AddTxIn(txIn)
		// calculate required fee
		txSize = int64(tx.SerializeSize() + maxSignScriptSize*len(tx.TxIn) + changeOutPutSize)
		//fmt.Printf("createTx: txSerSize=%v, txSize=(%v+%v*%v)=%v\n",tx.SerializeSize(), tx.SerializeSize(), maxSignScriptSize, len(tx.TxIn), txSize)
		//w.debugTxSize(tx)
		requiredFee = types.Amount{Value: txSize * feePerByte.Value, Id: feeCoinId}
		// check if enough fund
		checkNext := false
		for id, outAmt := range totalOutAmt {
			if id == feeCoinId && totalInAmt[id]-outAmt-requiredFee.Value < 0 {
				checkNext = true
				break
			} else if totalInAmt[id]-outAmt < 0 {
				checkNext = true
				break
			}
		}
		if !checkNext {
			enoughFund = true
			break
		}
	}
	if !enoughFund {
		return nil, fmt.Errorf("not engouh funds from the wallet to pay the specified outputs")
	}
	// add change if need
	changeValue := totalInAmt[feeCoinId] - totalOutAmt[feeCoinId] - requiredFee.Value
	if changeValue > 0 {
		addr, err := w.newAddress()
		if err != nil {
			return nil, err
		}
		pkScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, err
		}
		tx.AddTxOut(&types.TxOutput{
			Amount:   types.Amount{Value: changeValue, Id: feeCoinId},
			PkScript: pkScript,
		})
	}

	stxos := make([]*utxo, 0, len(tx.TxIn))

	// sign and add signature script to inputs
	for i, txIn := range tx.TxIn {
		outPoint := txIn.PreviousOut
		utxo := w.utxos[outPoint]
		// get priv key for the utxo's key index
		privkey := w.privkeys[utxo.keyIndex]
		//fmt.Printf("keyindex=%v, privkey=%x\n",utxo.keyIndex, privkey)
		key, _ := secp256k1.PrivKeyFromBytes(privkey)
		// sign
		sigScript, err := txscript.SignatureScript(tx, i, utxo.pkScript, txscript.SigHashAll, key, true)
		if err != nil {
			return nil, err
		}
		//fmt.Printf("signScript=%x, len=%v\n",sigScript, len(sigScript))
		txIn.SignScript = sigScript
		// save the utxo which will mark spent later
		stxos = append(stxos, utxo)
	}

	// all spend output need to mark spent for the wallet utxo
	for _, utxo := range stxos {
		utxo.isSpent = true
	}
	return tx, nil
}

func (w *testWallet) debugTxSize(tx *types.Transaction) {

	n := 0

	// 16 = Version 4 bytes + LockTime 4 bytes + Expire 4 bytes + Timestamp 4 bytes
	// the number of inputs for prefix
	// The number of outputs
	// the number of inputs for witness
	n = 16 + s.VarIntSerializeSize(uint64(len(tx.TxIn))) +
		s.VarIntSerializeSize(uint64(len(tx.TxOut))) +
		s.VarIntSerializeSize(uint64(len(tx.TxIn)))

	w.t.Logf("debugTxSize: add ver, lock, exp, timestamp, var int size %v = 16 + %v + %v + %v", n,
		s.VarIntSerializeSize(uint64(len(tx.TxIn))),
		s.VarIntSerializeSize(uint64(len(tx.TxOut))),
		s.VarIntSerializeSize(uint64(len(tx.TxIn))))

	for i, txIn := range tx.TxIn {
		w.t.Logf("debugTxSize: add input prefix[%v] %v", i, txIn.SerializeSizePrefix())
		n += txIn.SerializeSizePrefix()
	}
	w.t.Logf("debugTxSize: add input prefix %v", n)
	for i, txOut := range tx.TxOut {
		w.t.Logf("debugTxSize: add output[%v] %v = (2 + 8 + %v + %v)", i, txOut.SerializeSize(),
			s.VarIntSerializeSize(uint64(len(txOut.PkScript))),
			len(txOut.PkScript))
		id := make([]byte, 2)
		binary.LittleEndian.PutUint16(id, uint16(txOut.Amount.Id))
		w.t.Logf("debugTxSize: add output[%v] %v = 2 -> %x", i, txOut.SerializeSize(), id)
		value := make([]byte, 8)
		binary.LittleEndian.PutUint64(value, uint64(txOut.Amount.Value))
		w.t.Logf("debugTxSize: add output[%v] %v = 8 -> %x", i, txOut.SerializeSize(), value)
		n += txOut.SerializeSize()
	}
	w.t.Logf("debugTxSize: add outputs %v", n)
	for i, txIn := range tx.TxIn {
		w.t.Logf("debugTxSize: add input witness[%v] %v = %v + %v", i, txIn.SerializeSizeWitness(),
			s.VarIntSerializeSize(uint64(len(txIn.SignScript))),
			len(txIn.SignScript))
		n += txIn.SerializeSizeWitness()
	}
	w.t.Logf("debugTxSize: add input witness %v", n)
	w.t.Logf("debugTxSize: final size %v = %v", n, tx.SerializeSize())
}

func (w *testWallet) Addresses() []string {
	addrs := make([]string, 0)
	for _, a := range w.addrs {
		addrs = append(addrs, a.String())
	}
	return addrs
}

func (w *testWallet) blockConnected(hash *hash.Hash, height, order int64, t time.Time, txs []*types.Transaction) {
	w.t.Logf("node [%v] OnBlockConnected hash=%v,order=%v", w.nodeId, hash, order)
	for _, tx := range txs {
		w.t.Logf("node [%v] OnBlockConnected tx=%v", w.nodeId, tx.TxHash())
	}
	// Append the new update to the end of the queue of block dag updates.
	w.updateMtx.Lock()
	w.updates = append(w.updates, &update{order, hash, txs})
	w.updateMtx.Unlock()

	// signal the update watcher that a new update is arrived . use a goroutine
	// in order to avoid blocking this callback itself from the websocket client.
	go func() {
		w.updateArrived <- struct{}{}
	}()
}

func (w *testWallet) blockDisconnected(hash *hash.Hash, height, order int64, t time.Time, txs []*types.Transaction) {
	w.t.Logf("node [%v] OnBlockDisconnected hash=%v,order=%v", w.nodeId, hash, order)
	w.Lock()
	defer w.Unlock()

	undo, ok := w.undoes[hash]
	if !ok {
		w.t.Fatalf("the disconnected a unknown block, hash=%v, order=%v", hash, order)
	}

	for _, utxo := range undo.utxosCreated {
		delete(w.utxos, utxo)
	}

	for outPoint, utxo := range undo.utxosDestroyed {
		w.utxos[outPoint] = utxo
	}

	delete(w.undoes, hash)
}

func (w *testWallet) OnTxConfirm(txConfirm *cmds.TxConfirmResult) {
	w.Lock()
	defer w.Unlock()

	w.t.Log("OnTxConfirm", txConfirm.Tx, txConfirm.Confirms, txConfirm.Order)
	if w.confirmTxs == nil {
		w.confirmTxs = map[string]uint64{}
	}
	w.confirmTxs[txConfirm.Tx] = txConfirm.Confirms
}
func (w *testWallet) OnTxAcceptedVerbose(c *client.Client, tx *j.DecodeRawTransactionResult) {
	w.t.Log("OnTxAcceptedVerbose", tx.Order, tx.Txid, tx.Confirms, tx.Txvalid, tx.IsBlue, tx.Duplicate)
	if tx.Order <= 0 {
		// mempool tx
		w.Lock()
		defer w.Unlock()
		if w.mempoolTx == nil {
			w.mempoolTx = map[string]string{}
		}
		w.mempoolTx[tx.Txid] = tx.Vout[0].ScriptPubKey.Addresses[0]
	}
}
func (w *testWallet) OnRescanProgress(rescanPro *cmds.RescanProgressNtfn) {
	w.t.Log("OnRescanProgress", rescanPro.Order, rescanPro.Hash)
	if w.maxRescanOrder < rescanPro.Order {
		w.maxRescanOrder = rescanPro.Order
	}
	w.ScanCount++
}
func (w *testWallet) OnRescanFinish(rescanFinish *cmds.RescanFinishedNtfn) {
	w.t.Log("OnRescanFinish", rescanFinish.Order, rescanFinish.Hash)
	if w.OnRescanComplete != nil {
		w.OnRescanComplete()
	}
}
