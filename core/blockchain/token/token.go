// Copyright (c) 2020 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package token

import (
	"encoding/binary"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/crypto/ecc/schnorr"
	"github.com/Qitmeer/qitmeer/crypto/ecc/secp256k1"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"math"
)

// the token mint/unmint transactions format definition.
//
// 1. mint token by lock meer
// TxIn[0] <sigature> <ctr pubkey> <token_id> OP_TOKEN_MINT      // schnorr-sign for better multi-sign
// TxIn[1:N] Normal meer TxIn signature scripts
// TxOut[0] OP_MEER_LOCK (meer value)                            // meer value locked  (meer locked +)
// TxOut[1] OP_TOKEN_RELEASE <p2pkh||p2sh>   (token value)       // token can be spent (coin balance +)
// TxOut[2] optional OP_MEERCHANGE <p2pkh>||p2sh> (meer value)   // optional meer changes
//
// 2. unmint coin to unlock meer.
// TxIn[0] <sigature> <ctr pubkey> <token_id> OP_TOKEN_UNMINT
// TxIn[1:N]: Normal token Txin signature script
// TxOut[0] OP_TOKEN_DESTROY (token value)                         // token value destroyed (coin balance -)
// TxOut[1] OP_MEER_RELEASE <p2pkh||p2sh>   (meer value)           // meer value released ( meer locked -)
// TxOut[2] optional OP_TOKENCHANGE <p2pkh>||p2sh> (token value)   // optional token changes

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

var (
	zeroHash = &hash.Hash{}
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
		return nil, nil, nil, fmt.Errorf( "invalid TOKEN_MINT input[0]: none-zero outpoint")
	}
	// TxIn[0] must contains a signature, a public key and and token id and an OP_TOKEN_MINT opcode.
	txIn := tx.TxIn[0].SignScript
	if !(len(txIn) == TokenMintScriptLen &&
		txIn[0] == txscript.OP_DATA_64 &&
		txIn[65] == txscript.OP_DATA_33 &&
		txIn[99] == txscript.OP_DATA_2 &&
		txIn[102] == txscript.OP_TOKEN_MINT) {
		return nil, nil, nil, fmt.Errorf( "invalid TOKEN_MINT input[0]: incorrect signScript format")
	}


	// Pull out signature, pubkey, and tokenId
	signature = txIn[1 : 1+schnorr.SignatureSize]
	pubKey = txIn[66 : 66+secp256k1.PubKeyBytesLenCompressed]
	if !txscript.IsStrictCompressedPubKeyEncoding(pubKey) {
		return nil, nil,nil, fmt.Errorf("invalid TOKEN_MINT input[0]: wrong public key encoding")
	}
	tokenId = txIn[100:100+TokenIdSize]

	// tokenId must not meer itself
	if types.CoinID(binary.LittleEndian.Uint16(tokenId)) == types.MEERID {
		return nil, nil,nil, fmt.Errorf("invalid TOKEN_MINT input[0], invalid tokenId 0x%x",tokenId)
	}

	inputMeer := types.Amount{0,types.MEERID}
	// TxIn[1..N] must normal meer signature script
	for i, txIn := range tx.TxIn[1:] {
		// Make sure there is a script.
		if len(txIn.SignScript) == 0 {
			return nil, nil, nil, fmt.Errorf("invalid TOKEN_MINT input[%d], " +
				"script length %v", i+1, len(txIn.SignScript))
		}
		// Make sure the input value should meer
		if types.MEERID != txIn.AmountIn.Id {
			return nil, nil, nil, fmt.Errorf("invalid TOKEN_MINT input[%d]," +
				" must from meer, but from %v", i+1, txIn.AmountIn.Id )
		}
		if txIn.AmountIn.Value <= 0 {
			return nil, nil, nil, fmt.Errorf("invalid TOKEN_MINT input[%d]," +
				" invalid value %v", i+1, txIn.AmountIn )
		}
		inputMeer.Value += txIn.AmountIn.Value
	}

	// Outpus :

	// all TxOut script length should not zero.
	for i := range tx.TxOut{
		if len(tx.TxOut[i].PkScript) == 0 {
			return nil, nil, nil , fmt.Errorf("invalid TOKEN_MINT, output[%d] script is empty",i)
		}
	}
	// TxOut[0] must be an OP_MEER_LOCK
	// Txout[1]:must be an OP_TOKEN_RELEASE tagged P2SH or P2PKH script. and released token id must much with TxIn[0]

	// output[0]
	if len(tx.TxOut[0].PkScript) != 1 {
		return nil,nil,nil,fmt.Errorf("invalid TOKEN_MINT, output[0] script length is not 1 byte, got %v",
			len(tx.TxOut[0].PkScript))
	}
	if tx.TxOut[0].PkScript[0] != txscript.OP_MEER_LOCK {
		return nil,nil,nil,fmt.Errorf("invalid TOKEN_MINT, output[0] must be a MEER_LOCK, got 0x%x",
			tx.TxOut[0].PkScript[0])
	}
	if tx.TxOut[0].Amount.Id != types.MEERID {
		return nil,nil,nil,fmt.Errorf("invalid TOKEN_MINT, output[0] must be a MEER value, got %v",
			tx.TxOut[0].Amount.Id)
	}
	if tx.TxOut[0].Amount.Value <= 0 {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_MINT output[0]," +
			" invalid value %v",tx.TxOut[0].Amount )
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
	id := binary.LittleEndian.Uint16(tokenId)
	if id != uint16(tx.TxOut[1].Amount.Id) {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_MINT, " +
			"token id not match, mint token %v but release %v", types.CoinID(id), tx.TxOut[1].Amount.Id)
	}

	// check optional output[2]
	changeMeer := types.Amount{0,types.MEERID}
	if len(tx.TxOut) == 3 {
		if tx.TxOut[2].PkScript[0] != txscript.OP_MEER_CHANGE {
			return nil, nil, nil, fmt.Errorf("invalid TOKEN_MINT, output[2] is not OP_MEER_CHANGE")
		}
		if !(txscript.IsPubKeyHashScript(tx.TxOut[2].PkScript[1:]) ||
			txscript.IsPayToScriptHash(tx.TxOut[2].PkScript[1:])) {
			return nil, nil, nil, fmt.Errorf("invalid TOKEN_MINT, output[2] is not P2SH or P2PKH")
		}
		if tx.TxOut[2].Amount.Id != types.MEERID {
			return nil,nil,nil,fmt.Errorf("invalid TOKEN_MINT, output[2] must be a MEER value, got %v",
				tx.TxOut[2].Amount.Id)
		}
		changeMeer = tx.TxOut[2].Amount
	}

	// make sure the input meer value > locked meer + change meer
	if inputMeer.Value <= lockedMeer.Value + changeMeer.Value {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_MINT, " +
			"input %s <= locked %s + change %s", inputMeer.String(),lockedMeer.String(),changeMeer.String())
	}

	return signature,pubKey,tokenId,nil
}

// IsTokenMint returns true if the input transaction is a valid TOKEN_MINT
// NOTICE: the function DOES NOT check the signature and pubKey and tokenId.
func IsTokenMint(tx *types.Transaction) bool{
	 _,_,_,err := CheckTokenMint(tx)
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
		return nil, nil, nil, fmt.Errorf( "invalid TOKEN_UNMINT input[0]: none-zero outpoint")
	}
	// TxIn[0] must contains a signature, a public key and and token id and an OP_TOKEN_MINT opcode.
	txIn := tx.TxIn[0].SignScript
	if !(len(txIn) == TokenUnMintScriptLen &&
		txIn[0] == txscript.OP_DATA_64 &&
		txIn[65] == txscript.OP_DATA_33 &&
		txIn[99] == txscript.OP_DATA_2 &&
		txIn[102] == txscript.OP_TOKEN_UNMINT) {
		return nil, nil, nil, fmt.Errorf( "invalid TOKEN_UNMINT input[0]: incorrect signScript format")
	}

	// Pull out signature, pubkey, and tokenId
	signature = txIn[1 : 1+schnorr.SignatureSize]
	pubKey = txIn[66 : 66+secp256k1.PubKeyBytesLenCompressed]
	if !txscript.IsStrictCompressedPubKeyEncoding(pubKey) {
		return nil, nil,nil, fmt.Errorf("invalid TOKEN_UNMINT input[0]: wrong public key encoding")
	}
	tokenId = txIn[100:100+TokenIdSize]
	id := types.CoinID(binary.LittleEndian.Uint16(tokenId))
	// tokenId must not meer itself
	if id == types.MEERID {
		return nil, nil,nil, fmt.Errorf("invalid TOKEN_UNMINT input[0],  token id, can't unmint %v ",id)
	}

	// TxIn[1..N] must normal token signature script, and the input value should match with token id
	for i, txIn := range tx.TxIn[1:] {
		// Make sure there is a script.
		if len(txIn.SignScript) == 0 {
			return nil, nil, nil, fmt.Errorf("invalid TOKEN_UNMINT input[%d], " +
				"script length %v", i+1, len(txIn.SignScript))
		}
		// Make sure the input value should token
		if id != txIn.AmountIn.Id {
			return nil, nil, nil, fmt.Errorf("invalid TOKEN_UNMINT input[%d]," +
				" must from %v, but from %v", i+1, id, txIn.AmountIn.Id )
		}
	}

	// Outpus :
	// TxOut[0] must be an OP_TOKEN_DESTORY, and value must match with the unmint token id
	// Txout[1]:must be an OP_MEER_RELEASE tagged P2SH or P2PKH script. and released value must meer

	// output[0] len must 1
	if len(tx.TxOut[0].PkScript) != 1 {
		return nil,nil,nil,fmt.Errorf("invalid TOKEN_UNMINT, output[0] script length is not 1 byte, got %v",
			len(tx.TxOut[0].PkScript))
	}
	if tx.TxOut[0].PkScript[0] != txscript.OP_TOKEN_DESTORY {
		return nil,nil,nil,fmt.Errorf("invalid TOKEN_UNMINT, output[0] must be a TOKEN_DESTORY, got 0x%x",
			tx.TxOut[0].PkScript[0])
	}
	if tx.TxOut[0].Amount.Id != id {
		return nil,nil,nil,fmt.Errorf("invalid TOKEN_UNMINT, output[0] must be a %v value, got %v", id,
			tx.TxOut[0].Amount.Id)
	}
	// output[1]
	if len(tx.TxOut[1].PkScript) == 0 {
		return nil, nil, nil , fmt.Errorf("invalid TOKEN_UNMINT, output[1] script is empty")
	}
	if tx.TxOut[1].PkScript[0] != txscript.OP_MEER_RELEASE {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_UNMINT, output[1] is not a MEER_RELEASE")
	}
	if !(txscript.IsPubKeyHashScript(tx.TxOut[1].PkScript[1:]) ||
		txscript.IsPayToScriptHash(tx.TxOut[1].PkScript[1:])) {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_UNMINT, output[1] is not P2SH or P2PKH")
	}
	// check output[1] value must meer
	if types.MEERID != tx.TxOut[1].Amount.Id {
		return nil, nil, nil, fmt.Errorf("invalid TOKEN_UNMINT, " +
			"token id %v not match, token-unmint must release MEER", tx.TxOut[1].Amount.Id)
	}
	return signature,pubKey,tokenId,nil
}

// IsTokenUnMint returns true if the input transaction is a valid TOKEN_UlMINT
// NOTICE: the function DOES NOT check the signature and pubKey and tokenId.
func IsTokenUnMint(tx *types.Transaction) bool{
	_,_,_,err := CheckTokenUnMint(tx)
	return err == nil
}

func isNullOutPoint(tx *types.Transaction) bool {
	op := &tx.TxIn[0].PreviousOut
	if op.OutIndex == math.MaxUint32 && op.Hash.IsEqual(zeroHash) {
		return true
	}
	return false
}
