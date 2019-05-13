// Copyright (c) 2017-2018 The nox developers
// Copyright (c) 2014-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btc_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"qitmeer/engine/txscript"
	"qitmeer/crypto/ecc"
	btcparams "qitmeer/params/btc"
	btchash "qitmeer/common/hash/btc"
	"qitmeer/params/btc/addr"
	"qitmeer/params/btc/types"
	_ "qitmeer/params/btc/txscript"
	"qitmeer/common/hash"
	"qitmeer/core/types"
	"testing"
)

// This example demonstrates creating a script which pays to a bitcoin address.
// It also prints the created script hex and uses the DisasmString function to
// display the disassembled script.
func Test_ExamplePayToAddrScript(t *testing.T) {
	// Parse the address to send the coins to into a btcutil.Address
	// which is useful to ensure the accuracy of the address and determine
	// the address type.  It is also required for the upcoming call to
	// PayToAddrScript.
	addressStr := "12gpXQVcCL2qhTNQgyLVdCFG2Qs2px98nV"
	address, err := btcaddr.DecodeAddress(addressStr, &btcparams.MainNetParams)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Create a public key script that pays to the address.
	script, err := txscript.PayToAddrScript(address)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Script Hex: %x\n", script)

	disasm, err := txscript.DisasmString(script)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Script Disassembly:", disasm)

	// Output:
	// Script Hex: 76a914128004ff2fcaf13b2b91eb654b1dc2b674f7ec6188ac
	// Script Disassembly: OP_DUP OP_HASH160 128004ff2fcaf13b2b91eb654b1dc2b674f7ec61 OP_EQUALVERIFY OP_CHECKSIG
}


// This example demonstrates extracting information from a standard public key
// script.
func Test_ExampleExtractPkScriptAddrs(t *testing.T) {
	// Start with a standard pay-to-pubkey-hash script.
	scriptHex := "76a914128004ff2fcaf13b2b91eb654b1dc2b674f7ec6188ac"
	script, err := hex.DecodeString(scriptHex)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Extract and print details from the script.
	s,err := txscript.ParsePkScript(script)
	if err != nil {
		fmt.Println(err)
		return
	}
	scriptClass := s.GetClass()
	addresses := s.GetAddresses()
	reqSigs := s.RequiredSigs()

	fmt.Println("Script Class:", scriptClass)
	fmt.Println("Addresses:", addresses)
	fmt.Println("Required Signatures:", reqSigs)

	// Output:
	// Script Class: pubkeyhash
	// Addresses: [12gpXQVcCL2qhTNQgyLVdCFG2Qs2px98nV]
	// Required Signatures: true
}

//This example demonstrates manually creating and signing a redeem transaction.

func Test_ExampleSignTxOutput(t *testing.T) {

	// Ordinarily the private key would come from whatever storage mechanism
	// is being used, but for this example just hard code it.
	privKeyBytes, err := hex.DecodeString("22a47fa09a223f2aa079edf85a7c2" +
		"d4f8720ee63e502ee2869afab7de234b80c")
	if err != nil {
		fmt.Println(err)
		return
	}
	privKey, pubKey := ecc.Secp256k1.PrivKeyFromBytes(privKeyBytes)
	pubKeyHash := btchash.Hash160(pubKey.SerializeCompressed())
	addr, err := btcaddr.NewAddressPubKeyHash(pubKeyHash,
		&btcparams.MainNetParams)
	if err != nil {
		fmt.Println(err)
		return
	}
	/*
	$ echo 22a47fa09a223f2aa079edf85a7c2d4f8720ee63e502ee2869afab7de234b80c|bx ec-to-public
	02a673638cb9587cb68ea08dbef685c6f2d2a751a8b3c6f2a7e9a4999e6e4bfaf5

	$ echo 22a47fa09a223f2aa079edf85a7c2d4f8720ee63e502ee2869afab7de234b80c|bx ec-to-public|bx bitcoin160
3dee47716e3cfa57df45113473a6312ebeaef311

	$ echo 22a47fa09a223f2aa079edf85a7c2d4f8720ee63e502ee2869afab7de234b80c|bx ec-to-public|bx ec-to-address
16eTfd5Qsh3CRjW2bPKAF3iQqmYs1MJcZR
	 */
	fmt.Printf("pk=%x\n",pubKey.SerializeCompressed())   //pk=02a673638cb9587cb68ea08dbef685c6f2d2a751a8b3c6f2a7e9a4999e6e4bfaf5
	fmt.Printf("pkhash=%x\n",pubKeyHash)                 //pkhash=3dee47716e3cfa57df45113473a6312ebeaef311
	fmt.Printf("addr=%s\n",addr)                         //addr=16eTfd5Qsh3CRjW2bPKAF3iQqmYs1MJcZR

	// For this example, create a fake transaction that represents what
	// would ordinarily be the real transaction that is being spent.  It
	// contains a single output that pays to address in the amount of 1 BTC.
	//originTx :=  types.NewTransaction()
	originTx :=  btctypes.NewBtcTx(btctypes.TxVersion)

	//prevOut := types.NewOutPoint(&hash.Hash{}, ^uint32(0))   //coin-base 0x0 & 0xffffffff
	prevOut := btctypes.NewOutPoint(&hash.Hash{}, ^uint32(0)) //coin-base 0x0 & 0xffffffff
	fmt.Printf("prevOut=%v\n",prevOut)
	//txIn := types.NewTxInput(prevOut, 100000000, []byte{txscript.OP_0, txscript.OP_0})
	txIn := btctypes.NewTxIn(prevOut, []byte{txscript.OP_0, txscript.OP_0}, nil)
	originTx.AddTxIn(txIn)
	pkScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("pkScript=%x\n",pkScript)   //pkScript=76a9143dee47716e3cfa57df45113473a6312ebeaef31188ac
	/*
	  $ echo 76a9143dee47716e3cfa57df45113473a6312ebeaef31188ac|bx script-decode
	  dup hash160 [3dee47716e3cfa57df45113473a6312ebeaef311] equalverify checksig
	 */
	//txOut := types.NewTxOutput(100000000, pkScript)
	txOut := btctypes.NewTxOut(100000000, pkScript)
	fmt.Printf("txOut=%x\n",txOut)   //txOut=&{5f5e100 76a9143dee47716e3cfa57df45113473a6312ebeaef31188ac}
	originTx.AddTxOut(txOut)

	originTxHash := originTx.TxHash()
	fmt.Printf("originTxHash=%s\n",originTxHash.String()) //originTxHash=349f60d2bbddfb156536de22c5cd8c4c5a14baa0fadc65dc4e0a9f02e5207b3e
	originTxHashWit := originTx.WitnessHash()   //originTxHashWit=349f60d2bbddfb156536de22c5cd8c4c5a14baa0fadc65dc4e0a9f02e5207b3e
	fmt.Printf("originTxHashWit=%s\n",originTxHashWit.String())
	//originTxHashFull := originTx.TxHashFull()
	//fmt.Printf("originTxHashFull=%s\n",originTxHashFull.String())
	//buf,_ := originTx.Serialize(types.TxSerializeFull)
	buf := new(bytes.Buffer)
	originTx.Serialize(buf)
	fmt.Printf("originTxDump=%x\n",buf)
	//originTxDump=01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff020000ffffffff0100e1f505000000001976a9143dee47716e3cfa57df45113473a6312ebeaef31188ac00000000

	/*
	$ ./btc.sh decode 01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff020000ffffffff0100e1f505000000001976a9143dee47716e3cfa57df45113473a6312ebeaef31188ac00000000
{
 "txid": "349f60d2bbddfb156536de22c5cd8c4c5a14baa0fadc65dc4e0a9f02e5207b3e",
 "hash": "349f60d2bbddfb156536de22c5cd8c4c5a14baa0fadc65dc4e0a9f02e5207b3e",
 "version": 1,
 "size": 87,
 "vsize": 87,
 "locktime": 0,
 "vin": [
   {
     "coinbase": "0000",
     "sequence": 4294967295
   }
 ],
 "vout": [
   {
     "value": 1.00000000,
     "n": 0,
     "scriptPubKey": {
       "asm": "OP_DUP OP_HASH160 3dee47716e3cfa57df45113473a6312ebeaef311 OP_EQUALVERIFY OP_CHECKSIG",
       "hex": "76a9143dee47716e3cfa57df45113473a6312ebeaef31188ac",
       "reqSigs": 1,
       "type": "pubkeyhash",
       "addresses": [
         "16eTfd5Qsh3CRjW2bPKAF3iQqmYs1MJcZR"
       ]
     }
   }
 ]
}
	 */

	// Create the transaction to redeem the fake transaction.
	// redeemTx := types.NewTransaction()
	redeemTx := btctypes.NewBtcTx(btctypes.TxVersion)

	// Add the input(s) the redeeming transaction will spend.  There is no
	// signature script at this point since it hasn't been created or signed
	// yet, hence nil is provided for it.
	// newPrevOut := types.NewOutPoint(&originTxHash, 0)
	prevOut = btctypes.NewOutPoint(&originTxHash, 0)
	// newTxIn := types.NewTxInput(newPrevOut, 0,nil)
	txIn = btctypes.NewTxIn(prevOut, nil,nil)
	redeemTx.AddTxIn(txIn)

	// Ordinarily this would contain that actual destination of the funds,
	// but for this example don't bother.
	// newTxOut := types.NewTxOutput(0, nil)
	txOut = btctypes.NewTxOut(0, nil)
	redeemTx.AddTxOut(txOut)

	// Sign the redeeming transaction.

	lookupKey := func (types.Address) (ecc.PrivateKey, bool, error){
		// Ordinarily this function would involve looking up the private
		// key for the provided address, but since the only thing being
		// signed in this example uses the address associated with the
		// private key from above, simply return it with the compressed
		// flag set since the address is using the associated compressed
		// public key.
		//
		// NOTE: If you want to prove the code is actually signing the
		// transaction properly, uncomment the following line which
		// intentionally returns an invalid key to sign with, which in
		// turn will result in a failure during the script execution
		// when verifying the signature.
		//
		// privKey.D.SetInt64(12345)
		//
		return privKey, true, nil
	}

	// Notice that the script database parameter is nil here since it isn't
	// used.  It must be specified when pay-to-script-hash transactions are
	// being signed.
	sigScript, err := txscript.SignTxOut(
		redeemTx, 0, originTx.TxOut[0].PkScript, txscript.SigHashAll,
		txscript.KeyClosure(lookupKey), nil, nil, ecc.ECDSA_Secp256k1)
	if err != nil {
		fmt.Println(err)
		return
	}
	//redeemTx.TxIn[0].SignScript = sigScript
	redeemTx.TxIn[0].SignatureScript = sigScript

	sigScriptHash := btchash.HashH(sigScript)
	fmt.Printf("sigScriptHash=%s\n",sigScriptHash) //sigScriptHash=be48b666469c05f92a307998f4a1830b08906ddbd4590202ee340bac77ec46ec


	// Prove that the transaction has been validly signed by executing the
	// script pair.
	flags := txscript.ScriptBip16 | txscript.ScriptVerifyDERSignatures |
		//txscript.ScriptStrictMultiSig |
		txscript.ScriptDiscourageUpgradableNops
	vm, err := txscript.NewEngine2(originTx.TxOut[0].PkScript, redeemTx, 0,
		flags, 0, nil,)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := vm.Execute(); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Transaction successfully signed")
	// Output:
	// pk=02a673638cb9587cb68ea08dbef685c6f2d2a751a8b3c6f2a7e9a4999e6e4bfaf5
	// pkhash=3dee47716e3cfa57df45113473a6312ebeaef311
	// addr=16eTfd5Qsh3CRjW2bPKAF3iQqmYs1MJcZR
	// prevOut=0000000000000000000000000000000000000000000000000000000000000000:4294967295
	// pkScript=76a9143dee47716e3cfa57df45113473a6312ebeaef31188ac
	// txOut=&{5f5e100 76a9143dee47716e3cfa57df45113473a6312ebeaef31188ac}
	// originTxHash=349f60d2bbddfb156536de22c5cd8c4c5a14baa0fadc65dc4e0a9f02e5207b3e
	// originTxHashWit=349f60d2bbddfb156536de22c5cd8c4c5a14baa0fadc65dc4e0a9f02e5207b3e
    // originTxDump=01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff020000ffffffff0100e1f505000000001976a9143dee47716e3cfa57df45113473a6312ebeaef31188ac00000000
    // sigScriptHash=be48b666469c05f92a307998f4a1830b08906ddbd4590202ee340bac77ec46ec
    // Transaction successfully signed
}

