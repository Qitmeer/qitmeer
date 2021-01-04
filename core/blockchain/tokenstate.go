// Copyright (c) 2020 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockchain/token"
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/database"
	"github.com/Qitmeer/qitmeer/params"
	"strings"
)

// balanceUpdateType specifies the possible types of updates that might
// change the token balance
type balanceUpdateType byte

// The following constants define the known type of balanceUpdateType
const (
	tokenMint   balanceUpdateType = 0x01
	tokenUnMint balanceUpdateType = 0x02
)

// balanceUpdate specifies the type and update record of the values that change a token
// balance.
// for TOKON_MINT, the values should add on the meerlock and token balance
// for TOKEN_UNMINT, the values should subtract from the meerlock and token balance
type balanceUpdate struct {
	typ         balanceUpdateType
	meerAmount  int64
	tokenAmount types.Amount
}

// tokenBalance specifies the token balance and the locked meer amount
type tokenBalance struct {
	balance    int64
	lockedMeer int64
}

// tokenState specifies the token balance of the current block.
// the updates are written in the same order as the tx in the block, which is
// used to verify the correctness of the token balance
type tokenState struct {
	balances tokenBalances
	updates  []balanceUpdate
}

type tokenBalances map[types.CoinID]tokenBalance

func (tbs *tokenBalances) UpdateBalance(update *balanceUpdate) error {
	tokenId := update.tokenAmount.Id
	tb := (*tbs)[tokenId]
	switch update.typ {
	case tokenMint:
		tb.balance += update.tokenAmount.Value
		tb.lockedMeer += update.meerAmount
	case tokenUnMint:
		if tb.balance-update.tokenAmount.Value < 0 {
			return fmt.Errorf("can't unmint token %v more than token balance %v", update.tokenAmount, tb)
		}
		tb.balance -= update.tokenAmount.Value
		if tb.lockedMeer-update.meerAmount < 0 {
			return fmt.Errorf("can't unlock %v meer more than locked meer %v", update.meerAmount, tb)
		}
		tb.lockedMeer -= update.meerAmount
	default:
		return fmt.Errorf("unknown balance update type %v", update.typ)
	}
	(*tbs)[tokenId] = tb
	return nil
}

func (tb *tokenBalances) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "[")
	for k, v := range *tb {
		b.WriteString(fmt.Sprintf("%v:{balance:%v,locked-meer:%v},", k.Name(), v.balance, v.lockedMeer))
	}
	fmt.Fprintf(&b, "]")
	return b.String()
}
func (tb *tokenBalances) Copy() *tokenBalances {
	newTb := tokenBalances{}
	for k, v := range *tb {
		newTb[k] = v
	}
	return &newTb
}

// serializeTokeState function will serialize the token state into byte slice
func serializeTokeState(ts tokenState) ([]byte, error) {
	// total number of bytes to serialize
	serializeSize := serializeSizeVLQ(uint64(len(ts.balances)))
	for id, b := range ts.balances {
		// sanity check
		if id == types.MEERID || b.balance < 0 || b.lockedMeer < 0 {
			return nil, fmt.Errorf("invalid token balance {%v, %v}", id, b)
		}
		serializeSize += serializeSizeVLQ(uint64(id))
		serializeSize += serializeSizeVLQ(uint64(b.balance))
		serializeSize += serializeSizeVLQ(uint64(b.lockedMeer))
	}
	serializeSize += serializeSizeVLQ(uint64(len(ts.updates)))
	for _, v := range ts.updates {
		if v.typ != tokenMint && v.typ != tokenUnMint {
			return nil, fmt.Errorf("invalid token balance update type %v", v.typ)
		}
		if v.meerAmount < 0 || v.tokenAmount.Value < 0 || !types.IsKnownCoinID(v.tokenAmount.Id) {
			return nil, fmt.Errorf("invalid token balance update %v", v)
		}
		serializeSize += 1 // balanceUpdateType takes 1 byte
		serializeSize += serializeSizeVLQ(uint64(v.meerAmount))
		serializeSize += serializeSizeVLQ(uint64(v.tokenAmount.Id))
		serializeSize += serializeSizeVLQ(uint64(v.tokenAmount.Value))
	}
	serialized := make([]byte, serializeSize)
	offset := 0
	offset = putVLQ(serialized, uint64(len(ts.balances)))
	for id, b := range ts.balances {
		offset += putVLQ(serialized[offset:], uint64(id))
		offset += putVLQ(serialized[offset:], uint64(b.balance))
		offset += putVLQ(serialized[offset:], uint64(b.lockedMeer))
	}

	offset += putVLQ(serialized[offset:], uint64(len(ts.updates)))
	for _, v := range ts.updates {
		offset += putVLQ(serialized[offset:], uint64(v.typ))
		offset += putVLQ(serialized[offset:], uint64(v.meerAmount))
		offset += putVLQ(serialized[offset:], uint64(v.tokenAmount.Id))
		offset += putVLQ(serialized[offset:], uint64(v.tokenAmount.Value))
	}
	return serialized, nil
}

// deserializeTokenState function will deserializes token state from the byte slice
func deserializeTokenState(data []byte) (*tokenState, error) {
	// Deserialize the balance.
	var balances map[types.CoinID]tokenBalance
	numOfBalances, offset := deserializeVLQ(data)
	if offset == 0 {
		return nil, errDeserialize("unexpected end of data while reading number of balances")
	}
	if numOfBalances > 0 {
		balances = make(map[types.CoinID]tokenBalance, numOfBalances)
		for i := uint64(0); i < numOfBalances; i++ {
			// token id
			derId, bytesRead := deserializeVLQ(data[offset:])
			if bytesRead == 0 {
				return nil, fmt.Errorf("unexpected end of data while reading token id at balances{%d}", i)
			}
			offset += bytesRead

			// token balance
			balance, bytesRead := deserializeVLQ(data[offset:])
			if bytesRead == 0 {
				return nil, fmt.Errorf("unexpected end of data while reading balance at balances{%d}", i)
			}
			offset += bytesRead

			// locked meer
			lockedMeer, bytesRead := deserializeVLQ(data[offset:])
			if bytesRead == 0 {
				return nil, fmt.Errorf("unexpected end of data while reading balance at balances{%d}", i)
			}
			offset += bytesRead

			id := types.CoinID(uint16(derId))
			balances[id] = tokenBalance{int64(balance), int64(lockedMeer)}
		}
	}
	updates := []balanceUpdate{}
	numOfUpdates, bytesRead := deserializeVLQ(data[offset:])
	if bytesRead == 0 {
		return nil, errDeserialize("unexpected end of data while reading number of balances")
	}
	offset += bytesRead
	if numOfUpdates > 0 {
		updates = make([]balanceUpdate, numOfUpdates)
		for i := uint64(0); i < numOfUpdates; i++ {
			//type
			updateType, bytesRead := deserializeVLQ(data[offset:])
			if bytesRead == 0 {
				return nil, fmt.Errorf("unexpected end of data while reading balance update type at update[%d]", i)
			}
			offset += bytesRead
			//meerAmount
			meerAmount, bytesRead := deserializeVLQ(data[offset:])
			if bytesRead == 0 {
				return nil, fmt.Errorf("unexpected end of data while reading meer amount at update[%d]", i)
			}
			offset += bytesRead
			//tokenId
			tokenId, bytesRead := deserializeVLQ(data[offset:])
			if bytesRead == 0 {
				return nil, fmt.Errorf("unexpected end of data while reading token id at update[%d]", i)
			}
			offset += bytesRead
			//tokenAmount
			tokenAmount, bytesRead := deserializeVLQ(data[offset:])
			if bytesRead == 0 {
				return nil, fmt.Errorf("unexpected end of data while reading token amount at update[%d]", i)
			}
			offset += bytesRead

			updates[i] = balanceUpdate{
				typ:         balanceUpdateType(updateType),
				meerAmount:  int64(meerAmount),
				tokenAmount: types.Amount{Value:int64(tokenAmount), Id:types.CoinID(uint16(tokenId))},
			}
		}
	}
	return &tokenState{balances: balances, updates: updates}, nil
}

// dbPutTokenState put a token balance record into the token state database.
// the key is the provided block hash
func dbPutTokenState(dbTx database.Tx, hash *hash.Hash, ts tokenState) error {
	// Serialize the current token state.
	serializedData, err := serializeTokeState(ts)
	if err != nil {
		return err
	}
	// Store the current token balance record into the token state database.
	meta := dbTx.Metadata()
	bucket := meta.Bucket(dbnamespace.TokenBucketName)
	return bucket.Put(hash[:], serializedData)
}

// dbFetchTokenState fetch the token balance record from the token state database.
// the key is the input block hash.
func dbFetchTokenState(dbTx database.Tx, hash hash.Hash) (*tokenState, error) {
	// if it is genesis hash, return empty tokenState directly
	if *params.ActiveNetParams.GenesisHash == hash {
		return &tokenState{}, nil
	}
	// Fetch record from the token state database by block hash
	meta := dbTx.Metadata()
	bucket := meta.Bucket(dbnamespace.TokenBucketName)
	v := bucket.Get(hash[:])
	if v == nil {
		return nil, fmt.Errorf("tokenstate db can't find record from block hash : %v", hash)
	}
	// deserialize the fetched token state record
	return deserializeTokenState(v)
}

const dagConfirmFactor = 10

func (b *BlockChain) calculateTokenBalance(dbTx database.Tx, node *blockNode) (tokenBalances, error) {

	result := tokenBalances{}

	// the passed-in node should always has already been ordered at this level
	if !node.IsOrdered() {
		panic(fmt.Sprintf("node (hash=%v,order=%v,height=%v,layer=%v) is not ordered",
			node.hash, node.order, node.height, node.layer)) //should never happen here
	}

	// if no block is mature, return empty
	if uint16(node.order) < b.params.CoinbaseMaturity {
		return result, nil
	}

	// Calculate the token balance by using the mature block and latest stable block.
	//   result balance = stable balance + mature updates
	// The latest mature block node and latest stable node should be :
	//   mature_node = get_block_by_order(current.order - COINBASE_MATURITY)
	//   stable_node = get_block_by_order(current.order - DAG_CONFIRM_FACTOR)  /* use 10 here */

	// get the latest mature node hash
	mNodeHash, err := dbFetchHashByOrder(dbTx, node.order-uint64(b.params.CoinbaseMaturity))
	if err != nil {
		return nil, err
	}
	// get the latest stable node hash
	sNodeHash, err := dbFetchHashByOrder(dbTx, node.order-dagConfirmFactor)
	if err != nil {
		return nil, err
	}

	// current balance in the stable node
	sTs, err := dbFetchTokenState(dbTx, *sNodeHash)
	if err != nil {
		return nil, err
	}
	// mature balance in the mature node
	mTs, err := dbFetchTokenState(dbTx, *mNodeHash)
	if err != nil {
		return nil, err
	}

	// result balance = stable balance + mature updates added
	// only matured updates can be added into balance
	result = sTs.balances
	for _, update := range mTs.updates {
		// 1.) the updates list order has already been step up as exactly same as the order of transactions
		//     in the block.
		// 2.) the additional checking must has already done so that the updates has already checked its
		//     legality and removed duplicated tx, we can increase/decrase the token balance and locked meer
		//     balance safely. such as the balances can not be over minted; balances can not be negative.
		//     etc..
		err = result.UpdateBalance(&update)
		if err != nil {
			// should never happen at this level
			log.Error("calculateTokenBalance internal error when update balance %v from update %v", result, update)
		}
	}
	return result, nil
}

func (b *BlockChain) dbPutTokenBalance(dbTx database.Tx, node *blockNode, block *types.SerializedBlock) error {
	balances, err := b.calculateTokenBalance(dbTx, node)
	if err != nil {
		return err
	}
	ts := tokenState{
		balances: balances,
		updates:  []balanceUpdate{},
	}
	log.Trace(fmt.Sprintf("dbPutTokenBalance: %v start token balance %s", block.Hash(), balances.String()))

	checkB := balances.Copy()
	for _, tx := range block.Transactions() {
		if tx.IsDuplicate {
			log.Trace(fmt.Sprintf("dbPutTokenBalance skip duplicate tx %v", tx.Hash()))
			continue
		}
		if token.IsTokenMint(tx.Tx) {
			// TOKEN_MINT: input[0] token output[0] meer
			update := balanceUpdate{
				typ:         tokenMint,
				tokenAmount: tx.Tx.TxIn[0].AmountIn,
				meerAmount:  tx.Tx.TxOut[0].Amount.Value}

			// check the legality of update values.
			if err := checkMintUpdate(checkB, &update); err != nil {
				return err
			}
			// try update balance
			if err := checkB.UpdateBalance(&update); err != nil {
				return err
			}
			// append to update only when check & try has done with no err
			ts.updates = append(ts.updates, update)
		}
		if token.IsTokenUnMint(tx.Tx) {
			// TOKEN_UNMINT: input[0] meer output[0] token
			// the previous logic must make sure the legality of values, here only append.
			update := balanceUpdate{
				typ:         tokenUnMint,
				meerAmount:  tx.Tx.TxIn[0].AmountIn.Value,
				tokenAmount: tx.Tx.TxOut[0].Amount}
			// check the legality of update values.
			if err := checkUnMintUpdate(checkB, &update); err != nil {
				return err
			}
			// try update balance
			if err := checkB.UpdateBalance(&update); err != nil {
				return err
			}
			// append to update only when check & try has done with no err
			ts.updates = append(ts.updates, update)
		}
	}
	return dbPutTokenState(dbTx, block.Hash(), ts)
}

func checkUnMintUpdate(b *tokenBalances, update *balanceUpdate) error {
	if update.typ != tokenUnMint {
		return fmt.Errorf("checkUnMintUpdate : wrong update type %v", update.typ)
	}
	if err := checkUpdateCommon(update); err != nil {
		return err
	}
	return nil
}

func checkMintUpdate(b *tokenBalances, update *balanceUpdate) error {
	if update.typ != tokenMint {
		return fmt.Errorf("checkUnMintUpdate : wrong update type %v", update.typ)
	}
	if err := checkUpdateCommon(update); err != nil {
		return err
	}
	return nil
}

func checkUpdateCommon(update *balanceUpdate) error {
	if !types.IsKnownCoinID(update.tokenAmount.Id) {
		return fmt.Errorf("checkUpdateCommon : unknown token id %v", update.tokenAmount.Id.Name())
	}
	if update.tokenAmount.Value <= 0 {
		return fmt.Errorf("checkUpdateCommon : wrong token amount : %v", update.tokenAmount.Value)
	}
	if update.meerAmount <= 0 {
		return fmt.Errorf("checkUpdateCommon : wrong meer amount : %v", update.meerAmount)
	}
	return nil
}
