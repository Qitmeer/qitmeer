// Copyright (c) 2020 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package token

import (
	"encoding/binary"
	"fmt"
	"github.com/Qitmeer/qng-core/common/hash"
	"github.com/Qitmeer/qng-core/core/types"
	"github.com/Qitmeer/qng-core/crypto/ecc/schnorr"
	"github.com/Qitmeer/qng-core/crypto/ecc/secp256k1"
	"github.com/Qitmeer/qng-core/engine/txscript"
	"math"
)

// the token mint/unmint transactions format definition.
//
// 1. mint token by lock meer (fee is meer)
// TxIn[0] <signature> <ctr pubkey> <token_id> OP_TOKEN_MINT (nullp, token value)  // schnorr-sign for better multi-sign
// TxIn[1:N] Normal meer TxIn signature scripts (meer value)
// TxOut[0] OP_MEER_LOCK (meer value)                            // meer value locked  (meer locked +)
// TxOut[1] OP_TOKEN_RELEASE <p2pkh||p2sh>   (token value)       // token can be spent (coin balance +)
// TxOut[2] optional OP_MEERCHANGE <p2pkh||p2sh> (meer value)    // optional meer changes
//
// 2. unmint coin to unlock meer. (fee is meer)
// TxIn[0] <signature> <ctr pubkey> <token_id> OP_TOKEN_UNMINT (nullp, meer value)
// TxIn[1:N]: Normal token TxIn signature script (token value)
// TxOut[0] OP_TOKEN_DESTROY (token value)                        // token value destroyed (coin balance -)
// TxOut[1] OP_MEER_RELEASE <p2pkh||p2sh>   (meer value)          // meer value released ( meer locked -)
// TxOut[2] optional OP_TOKENCHANGE <p2pkh||p2sh> (token value)   // optional token changes
//

const (
	// TokenMintScriptLen is the length of a TokenMint script
	// <OP_DATA_64> <signature> <OP_DATA_33> <public key>  <OP_DATA_2> <token_id> <OP_TOKEN_MINT>
	// 1 + 64 + 1 + 33 + 1 + 2 + 1 = 103
	TokenMintScriptLen = 103
	// TokenUnMintScriptLen is the length of a TokenUnMint script
	// <OP_DATA_64> <signature> <OP_DATA_33> <public key>  <OP_DATA_2> <token_id> <OP_TOKEN_UNMINT>
	// 1 + 64 + 1 + 33 + 1 + 2 + 1 = 103
	TokenUnMintScriptLen = 103

	// TokenIdSize is the length of a TokenId (two bytes uint16)
	TokenIdSize = 2
)

// CheckTokenMint verifies the input if is a valid TOKEN_MINT transaction.
// The function return the signature, public key, tokenId and an error.
// The function ONLY check if the format is correct. for the returned signature,
// public key and token id. The callee MUST to do additional check for the
// returned values.
func CheckTokenMint(tx *types.Transaction) (signature []byte, pubKey []byte, tokenId []byte, err error) {

	// A valid TOKEN_MINT tx
	// There must be at least 2 inputs and must be 2 or 3 outputs.
	if len(tx.TxIn) < 2 || len(tx.TxOut) < 2 || len(tx.TxOut) > 3 {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_MINT input/output size (in: %v out: %v)",
			len(tx.TxIn), len(tx.TxOut))
	}

	// Inputs :
	// TxIn[0] must point to zero (previous outpoint is zero hash and max index)
	if isNullOutPoint(tx) == false {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_MINT input[0]: none-zero outpoint")
	}
	// TxIn[0] must contains a signature, a public key and and token id and an OP_TOKEN_MINT opcode.
	txIn := tx.TxIn[0].SignScript
	if !(len(txIn) == TokenMintScriptLen &&
		txIn[0] == txscript.OP_DATA_64 &&
		txIn[65] == txscript.OP_DATA_33 &&
		txIn[99] == txscript.OP_DATA_2 &&
		txIn[102] == txscript.OP_TOKEN_MINT) {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_MINT input[0]: incorrect signScript format")
	}
	// Pull out signature, pubkey, and tokenId
	signature = txIn[1 : 1+schnorr.SignatureSize]
	pubKey = txIn[66 : 66+secp256k1.PubKeyBytesLenCompressed]
	if !txscript.IsStrictCompressedPubKeyEncoding(pubKey) {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_MINT input[0]: wrong public key encoding")
	}
	tokenId = txIn[100 : 100+TokenIdSize]
	id := types.CoinID(binary.LittleEndian.Uint16(tokenId))

	// tokenId must not meer itself
	if id == types.MEERID {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_MINT input[0], can not mint %s", id.Name())
	}
	// TxIn[0] value should match tokenId
	if id != tx.TxIn[0].AmountIn.Id {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_MINT input[0], "+
			"mint %s but value is %s", id.Name(), tx.TxIn[0].AmountIn.Id.Name())
	}
	mintAmount := tx.TxIn[0].AmountIn

	inputMeer := types.Amount{Value: 0, Id: types.MEERID}
	// TxIn[1..N] must normal meer signature script
	for i, txIn := range tx.TxIn[1:] {
		// Make sure there is a script.
		if len(txIn.SignScript) == 0 {
			return nil, nil, nil, fmt.Errorf("invalid TOKEN_MINT input[%d], "+
				"script length %v", i+1, len(txIn.SignScript))
		}
		// Make sure the input value should meer
		if types.MEERID != txIn.AmountIn.Id {
			return nil, nil, nil, fmt.Errorf("invalid TOKEN_MINT input[%d],"+
				" must from meer, but from %v", i+1, txIn.AmountIn.Id)
		}
		if txIn.AmountIn.Value <= 0 {
			return nil, nil, nil, fmt.Errorf("invalid TOKEN_MINT input[%d],"+
				" invalid value %v", i+1, txIn.AmountIn)
		}
		inputMeer.Value += txIn.AmountIn.Value
	}

	// Outputs :

	// all TxOut script length should not zero.
	for i := range tx.TxOut {
		if len(tx.TxOut[i].PkScript) == 0 {
			return nil, nil, nil, fmt.Errorf("invalid TOKEN_MINT, output[%d] script is empty", i)
		}
	}
	// TxOut[0] must be an OP_MEER_LOCK
	// Txout[1] must be an OP_TOKEN_RELEASE tagged P2SH or P2PKH script. and released token id must match with TxIn[0]
	//          and value must be equal with TxIn[0]

	// output[0]
	if len(tx.TxOut[0].PkScript) != 1 {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_MINT, output[0] script length is not 1 byte, got %v",
			len(tx.TxOut[0].PkScript))
	}
	if tx.TxOut[0].PkScript[0] != txscript.OP_MEER_LOCK {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_MINT, output[0] must be a MEER_LOCK, got 0x%x",
			tx.TxOut[0].PkScript[0])
	}
	if tx.TxOut[0].Amount.Id != types.MEERID {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_MINT, output[0] must be a MEER value, got %v",
			tx.TxOut[0].Amount.Id)
	}
	if tx.TxOut[0].Amount.Value <= 0 {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_MINT output[0],"+
			" invalid value %v", tx.TxOut[0].Amount)
	}
	lockedMeer := tx.TxOut[0].Amount

	// output[1]
	if tx.TxOut[1].PkScript[0] != txscript.OP_TOKEN_RELEASE {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_MINT, output[1] is not OP_TOKEN_RELEASE")
	}
	if !(txscript.IsPubKeyHashScript(tx.TxOut[1].PkScript[1:]) ||
		txscript.IsPayToScriptHash(tx.TxOut[1].PkScript[1:])) {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_MINT, output[1] is not P2SH or P2PKH")
	}
	// check output[1] match with token id
	if id != tx.TxOut[1].Amount.Id {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_MINT, output[1] "+
			"token id not match, mint token %v but release %v", id, tx.TxOut[1].Amount.Id)
	}
	// check mint value match
	if mintAmount.Value != tx.TxOut[1].Amount.Value {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_MINT, output[1] "+
			"mint %s but release %s", mintAmount.String(), tx.TxOut[1].Amount.String())
	}

	// check optional output[2]
	changeMeer := types.Amount{Value: 0, Id: types.MEERID}
	if len(tx.TxOut) == 3 {
		if tx.TxOut[2].PkScript[0] != txscript.OP_MEER_CHANGE {
			return nil, nil, nil, fmt.Errorf("invalid TOKEN_MINT, output[2] is not OP_MEER_CHANGE")
		}
		if !(txscript.IsPubKeyHashScript(tx.TxOut[2].PkScript[1:]) ||
			txscript.IsPayToScriptHash(tx.TxOut[2].PkScript[1:])) {
			return nil, nil, nil, fmt.Errorf("invalid TOKEN_MINT, output[2] is not P2SH or P2PKH")
		}
		if tx.TxOut[2].Amount.Id != types.MEERID {
			return nil, nil, nil, fmt.Errorf("invalid TOKEN_MINT, output[2] must be a MEER value, got %v",
				tx.TxOut[2].Amount.Id)
		}
		changeMeer = tx.TxOut[2].Amount
	}

	// make sure the input meer value > locked meer + change meer
	if inputMeer.Value <= lockedMeer.Value+changeMeer.Value {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_MINT, "+
			"input %s <= locked %s + change %s", inputMeer.String(), lockedMeer.String(), changeMeer.String())
	}

	return signature, pubKey, tokenId, nil
}

// IsTokenMint returns true if the input transaction is a valid TOKEN_MINT
// NOTICE: the function DOES NOT check the signature and pubKey and tokenId.
func IsTokenMint(tx *types.Transaction) bool {
	_, _, _, err := CheckTokenMint(tx)
	return err == nil
}

// CheckTokenUnMint verifies the input if is a valid TOKEN_UNMINT transaction.
// The function return the signature, public key, tokenId and an error.
// The function ONLY check if the format is correct. for the returned signature,
// public key and token id. The callee MUST to do additional check for the
// returned values.
func CheckTokenUnMint(tx *types.Transaction) (signature []byte, pubKey []byte, tokenId []byte, err error) {
	// A valid TOKEN_UNMINT tx
	// There must be at least 2 inputs and must be 2 or 3 outputs.
	if len(tx.TxIn) < 2 || len(tx.TxOut) < 2 || len(tx.TxOut) > 3 {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_UNMINT input/output size (in: %v out: %v)",
			len(tx.TxIn), len(tx.TxOut))
	}
	// Inputs :
	// TxIn[0] must point to zero (previous outpoint is zero hash and max index)
	if isNullOutPoint(tx) == false {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_UNMINT input[0]: none-zero outpoint")
	}
	// TxIn[0] must contains a signature, a public key and and token id and an OP_TOKEN_MINT opcode.
	txIn := tx.TxIn[0].SignScript
	if !(len(txIn) == TokenUnMintScriptLen &&
		txIn[0] == txscript.OP_DATA_64 &&
		txIn[65] == txscript.OP_DATA_33 &&
		txIn[99] == txscript.OP_DATA_2 &&
		txIn[102] == txscript.OP_TOKEN_UNMINT) {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_UNMINT input[0]: incorrect signScript format")
	}

	// Pull out signature, pubkey, and tokenId
	signature = txIn[1 : 1+schnorr.SignatureSize]
	pubKey = txIn[66 : 66+secp256k1.PubKeyBytesLenCompressed]
	if !txscript.IsStrictCompressedPubKeyEncoding(pubKey) {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_UNMINT input[0]: wrong public key encoding")
	}
	tokenId = txIn[100 : 100+TokenIdSize]
	id := types.CoinID(binary.LittleEndian.Uint16(tokenId))
	// tokenId must not meer itself
	if id == types.MEERID {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_UNMINT input[0],  token id, can't unmint %v ", id)
	}

	// check TxIn[0] id and value , TxIn[0] must meer
	if tx.TxIn[0].AmountIn.Id != types.MEERID {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_UNMINT input[0], value must be %s but %s",
			types.MEERID.Name(), tx.TxIn[0].AmountIn.Id.Name())
	}
	if tx.TxIn[0].AmountIn.Value <= 0 {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_UNMINT input[0], value %s is invalid",
			tx.TxIn[0].AmountIn.String())
	}
	inputMeer := tx.TxIn[0].AmountIn

	inputToken := types.Amount{Value: 0, Id: id}
	// TxIn[1..N] must normal token signature script, and the input value should match with token id
	for i, txIn := range tx.TxIn[1:] {
		// Make sure there is a script.
		if len(txIn.SignScript) == 0 {
			return nil, nil, nil, fmt.Errorf("invalid TOKEN_UNMINT input[%d], "+
				"script length %v", i+1, len(txIn.SignScript))
		}
		// Make sure the input value should same token id
		if id != txIn.AmountIn.Id {
			return nil, nil, nil, fmt.Errorf("invalid TOKEN_UNMINT input[%d],"+
				" must from %v, but from %v", i+1, id, txIn.AmountIn.Id)
		}
		// check value valid
		if txIn.AmountIn.Value <= 0 {
			return nil, nil, nil, fmt.Errorf("invalid TOKEN_UNMINT input[%d],"+
				" invalid value %v", i+1, txIn.AmountIn)
		}
		inputToken.Value += txIn.AmountIn.Value
	}

	// Outputs :
	// TxOut[0] must be an OP_TOKEN_DESTORY, and value must match with the unmint token id
	// Txout[1]:must be an OP_MEER_RELEASE tagged P2SH or P2PKH script. and released value must meer

	// all TxOut script length should not zero.
	for i := range tx.TxOut {
		if len(tx.TxOut[i].PkScript) == 0 {
			return nil, nil, nil, fmt.Errorf("invalid TOKEN_UNMINT, output[%d] script is empty", i)
		}
	}

	// output[0] len must 1
	if len(tx.TxOut[0].PkScript) != 1 {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_UNMINT, output[0] script length is not 1 byte, got %v",
			len(tx.TxOut[0].PkScript))
	}
	if tx.TxOut[0].PkScript[0] != txscript.OP_TOKEN_DESTORY {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_UNMINT, output[0] must be a TOKEN_DESTORY, got 0x%x",
			tx.TxOut[0].PkScript[0])
	}
	if tx.TxOut[0].Amount.Id != id {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_UNMINT, output[0] must be a %s, got %s", id.Name(),
			tx.TxOut[0].Amount.Id.Name())
	}

	// check the destroyed toke value
	if tx.TxOut[0].Amount.Value <= 0 {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_UNMINT, invalid output[0] value %v",
			tx.TxOut[0].Amount)
	}
	destoryed := tx.TxOut[0].Amount
	// output[1]
	if tx.TxOut[1].PkScript[0] != txscript.OP_MEER_RELEASE {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_UNMINT, output[1] is not a MEER_RELEASE")
	}
	if !(txscript.IsPubKeyHashScript(tx.TxOut[1].PkScript[1:]) ||
		txscript.IsPayToScriptHash(tx.TxOut[1].PkScript[1:])) {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_UNMINT, output[1] is not P2SH or P2PKH")
	}
	// check output[1] value must meer
	if types.MEERID != tx.TxOut[1].Amount.Id {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_UNMINT, "+
			"token id %v not match, token-unmint must release MEER", tx.TxOut[1].Amount.Id)
	}
	// check released meer value, make sure inputMeer > relased meer
	if tx.TxOut[1].Amount.Value <= 0 || inputMeer.Value <= tx.TxOut[1].Amount.Value {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_UNMINT, "+
			"release value not valid,  release %s should less than input %s", inputMeer.String(), tx.TxOut[1].Amount.String())
	}

	change := types.Amount{Value: 0, Id: id}
	// optional output[2] , token change
	if len(tx.TxOut) == 3 {
		if tx.TxOut[2].PkScript[0] != txscript.OP_TOKEN_CHANGE {
			return nil, nil, nil, fmt.Errorf("invalid TOKEN_UNMINT, output[2] is not OP_TOKEN_CHANGE")
		}
		if !(txscript.IsPubKeyHashScript(tx.TxOut[2].PkScript[1:]) ||
			txscript.IsPayToScriptHash(tx.TxOut[2].PkScript[1:])) {
			return nil, nil, nil, fmt.Errorf("invalid TOKEN_UNMINT, output[2] is not P2SH or P2PKH")
		}
		if tx.TxOut[2].Amount.Id != id {
			return nil, nil, nil, fmt.Errorf("invalid TOKEN_UNMINT, output[2] must be a %s, got %s", id.Name(),
				tx.TxOut[2].Amount.Id.Name())
		}
		change = tx.TxOut[2].Amount
	}

	// make sure input token value > destroyed + change
	if inputToken.Value != destoryed.Value+change.Value {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_UNMINT, "+
			"input %s != destoryed %s + change %s", inputToken.String(), destoryed.String(), change.String())
	}
	return signature, pubKey, tokenId, nil
}

// IsTokenUnMint returns true if the input transaction is a valid TOKEN_UlMINT
// NOTICE: the function DOES NOT check the signature and pubKey and tokenId.
func IsTokenUnMint(tx *types.Transaction) bool {
	_, _, _, err := CheckTokenUnMint(tx)
	return err == nil
}

func isNullOutPoint(tx *types.Transaction) bool {
	op := &tx.TxIn[0].PreviousOut
	if op.OutIndex == math.MaxUint32 && op.Hash.IsEqual(&hash.ZeroHash) {
		return true
	}
	return false
}

func NewUpdateFromTx(tx *types.Transaction) (ITokenUpdate, error) {
	if types.IsTokenMintTx(tx) ||
		types.IsTokenUnmintTx(tx) {
		return NewBalanceUpdate(tx)
	} else if types.IsTokenNewTx(tx) ||
		types.IsTokenRenewTx(tx) ||
		types.IsTokenValidateTx(tx) ||
		types.IsTokenInvalidateTx(tx) {
		return NewTypeUpdateFromTx(tx)
	}
	return nil, fmt.Errorf("Not supported:%s\n", types.DetermineTxType(tx))
}
