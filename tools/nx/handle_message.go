package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"qitmeer/common/encode/base58"
	"qitmeer/common/hash"
	"qitmeer/common/hash/btc"
	"qitmeer/core/serialization"
	"qitmeer/crypto/ecc"
	"qitmeer/crypto/ecc/secp256k1"
	"reflect"
)

const (
	BTCMsgSignaturePrefixMagic = "Bitcoin Signed Message:\n"
	NoxMsgSignaturePrefixMagic = "Nox Signed Message:\n"
)

func decodeSignature(signatureType string, signStr string) {
	signHex,err := base64.StdEncoding.DecodeString(signStr)
	if err!=nil {
		errExit(err)
	}
	signature,err := ecc.Secp256k1.ParseSignature(signHex)
	if err!=nil {
		errExit(err)
	}
	fmt.Printf("R=%x,S=%x\n",signature.GetR(),signature.GetS())
}

func decodeAddr (mode string, addrStr string) ([]byte, error) {
	switch(mode) {
		case "btc" :
			addrHash160, _, err := base58.BtcCheckDecode(addrStr)
			return addrHash160,err
		default :
			addrHash160,_,err := base58.NoxCheckDecode(addrStr)
			return addrHash160,err
	}
}
func verifyMsgSignature(mode string, addrStr string, signStr string, msgStr string){

	msgHash := buildMsgHash(mode,msgStr)

	addrHash160, err := decodeAddr(mode,addrStr)
	if err!=nil {
		errExit(err)
	}

	sign_c, err :=base64.StdEncoding.DecodeString(signStr)
	if err!=nil {
		errExit(err)
	}

    // the recovery mode of btc
    // That number between 0 and 3 we call the recovery id, or recid.
    // Therefore, we return an extra byte, which also functions as a header byte,
    // by using 27+recid (for uncompressed recovered pubkeys) or 31+recid (for compressed recovered pubkeys).
    // https://bitcoin.stackexchange.com/a/38909
    // 27 + recId (uncompressed pubkey)
    // 31 + recId (compressed pubkey)
    pubKey, compressed,err := ecc.Secp256k1.RecoverCompact(sign_c,msgHash)
	if err!=nil {
		errExit(err)
	}

    var data []byte
   	if compressed {
		data = pubKey.SerializeCompressed()
	}else {
		data = pubKey.SerializeUncompressed()
	}

    fmt.Printf("%v\n",reflect.DeepEqual(calcHash160(mode,data),addrHash160))
}

func calcHash160(mode string,data []byte) []byte {
	switch(mode) {
	case "btc":
		return btc.Hash160(data)
	default:
		return hash.Hash160(data)
	}
}


func buildMsgHash(mode string, msg string) []byte {
	var msgHash []byte
	switch (mode) {
	case "btc" :
		var buf bytes.Buffer
		serialization.WriteVarString(&buf, 0, BTCMsgSignaturePrefixMagic)
		serialization.WriteVarString(&buf, 0, msg)
		msgHash = btc.DoubleHashB(buf.Bytes())

	default :
		var buf bytes.Buffer
		serialization.WriteVarString(&buf, 0, NoxMsgSignaturePrefixMagic)
		serialization.WriteVarString(&buf, 0, msg)
		msgHash = hash.HashB(buf.Bytes())
	}
	return msgHash
}

func msgSign(mode string, showSignDetail bool, wif string, msg string){
	decoded,compressed,err := decodeWIF(wif)
	if err!= nil {
		errExit(err)
	}
	privateKey,_ := ecc.Secp256k1.PrivKeyFromBytes(decoded)

	msgHash := buildMsgHash(mode,msg)
	//
	r,s, err :=ecc.Secp256k1.Sign(privateKey,msgHash)
	if err!=nil {
		errExit(err)
	}
	// got the der signature
	sigHex := ecc.Secp256k1.NewSignature(r,s).Serialize()

	sign_c, err := secp256k1.SignCompact(secp256k1.NewPrivateKey(privateKey.GetD()),msgHash,compressed)
	if err!=nil {
		errExit(err)
	}

	if showDetails {
		fmt.Printf("        mode: %s\n",mode)
		fmt.Printf("        hash: %x\n",msgHash)
		fmt.Printf("   signature: %x\n",sigHex)
		fmt.Printf("    (base64): %s\n",base64.StdEncoding.EncodeToString(sigHex))
		fmt.Printf("           R: %x\n",r)
		fmt.Printf("           S: %x\n",s)
		fmt.Printf(" compactsign: %x\n",sign_c[:])
		fmt.Printf("    (base64): %s\n",base64.StdEncoding.EncodeToString(sign_c[:]))
		fmt.Printf("  compressed: %v\n",compressed)

	}else{
		fmt.Printf("%s\n",base64.StdEncoding.EncodeToString(sign_c[:]))
	}

}
