package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/noxproject/nox/common/hash"
	"github.com/noxproject/nox/common/hash/btc"
	"github.com/noxproject/nox/core/serialization"
	"github.com/noxproject/nox/crypto/ecc"
	"github.com/noxproject/nox/crypto/ecc/secp256k1"
)

const (
	BTCMsgSignaturePrefixMagic = "Bitcoin Signed Message:\n"
	NoxMsgSignaturePrefixMagic = "Nox Signed Message:\n"
)

func decodeSignature(signStr string) {
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


func msgSign(mode string, showSignDetail bool, wif string, msg string){
	decoded,compressed,err := decodeWIF(wif)
	if err!= nil {
		errExit(err)
	}
	privateKey,_ := ecc.Secp256k1.PrivKeyFromBytes(decoded)

	var msgHash []byte
	switch (mode) {
	case "btc" :
		var buf bytes.Buffer
		serialization.WriteVarString(&buf, 0, BTCMsgSignaturePrefixMagic)
		serialization.WriteVarString(&buf, 0, msg)
		msgHash = btc.DoubleHashB(buf.Bytes())

	default :
		var buf bytes.Buffer
		serialization.WriteVarString(&buf, 0, BTCMsgSignaturePrefixMagic)
		serialization.WriteVarString(&buf, 0, msg)
		msgHash = hash.HashB(buf.Bytes())
	}

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
	/*
	verified := ecc.Secp256k1.Verify(publicKey,msgHash,r,s)

	fmt.Printf("Call verfiy, verified= %v \n", verified )

	signature,err := ecc.Secp256k1.ParseSignature(sigHex)
	if err != nil {
		errExit(err)
	}
	fmt.Printf("R=%x,S=%x\n",signature.GetR(),signature.GetS())

	signatureDer, err :=ecc.Secp256k1.ParseDERSignature(sigHex)
	if err != nil {
		errExit(err)
	}
	fmt.Printf("R=%x,S=%x\n",signatureDer.GetR(),signatureDer.GetS())

	sign_uc, err := secp256k1.SignCompact(secp256k1.NewPrivateKey(privateKey.GetD()),msgHash,false)
	if err!=nil {
		errExit(err)
	}
	fmt.Printf("SignCompact(uck)   =%x\n",sign_uc[:])
	fmt.Printf("SignCompact(base64)=%s\n",base64.StdEncoding.EncodeToString(sign_uc[:]))


    // the recovery mode of btc
    // That number between 0 and 3 we call the recovery id, or recid.
    // Therefore, we return an extra byte, which also functions as a header byte,
    // by using 27+recid (for uncompressed recovered pubkeys) or 31+recid (for compressed recovered pubkeys).
    // https://bitcoin.stackexchange.com/a/38909
    // 27 + recId (uncompressed pubkey)
    // 31 + recId (compressed pubkey)
    pubKey, uncompressed,err := ecc.Secp256k1.RecoverCompact(sign_c,msgHash)
	if err!=nil {
		errExit(err)
	}
	fmt.Printf("Recover compact :%v\n",uncompressed)
	fmt.Printf("uncompressed pubkey=%x\n", pubKey.SerializeUncompressed())
	fmt.Printf("  compressed pubkey=%x\n", pubKey.SerializeCompressed())
	*/
}
