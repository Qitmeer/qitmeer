// Copyright (c) 2020 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package testutils

import (
	"bytes"
	"encoding/binary"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/address"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/crypto/bip32"
	"github.com/Qitmeer/qitmeer/crypto/ecc/secp256k1"
	"github.com/Qitmeer/qitmeer/params"
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
	isLocked bool
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
	undoes map[*hash.Hash]*undo

	// the current synced order the wallet is known
	currentOrder int64
	sync.RWMutex

	netParams *params.Params
	t         *testing.T
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

func (w *testWallet) coinBaseAddr() types.Address {
	return w.addrs[0]
}

func (w *testWallet) coinBasePrivKey() []byte {
	return w.privkeys[0]
}

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
			w.currentOrder++
			if w.currentOrder != update.order {
				w.t.Fatalf("the order not match, expect current is %v but update got %v", w.currentOrder, update.order)
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

func (w *testWallet) blockConnected(hash *hash.Hash, order int64, t time.Time, txs []*types.Transaction) {
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

func (w *testWallet) blockDisconnected(hash *hash.Hash, order int64, t time.Time, txs []*types.Transaction) {
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
