package qx

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/Qitmeer/qng-core/common/encode/base58"
	"github.com/Qitmeer/qng-core/common/hash"
	"github.com/Qitmeer/qng-core/common/hash/btc"
	"github.com/Qitmeer/qng-core/core/serialization"
	"github.com/Qitmeer/qng-core/crypto/ecc"
	"github.com/Qitmeer/qng-core/crypto/ecc/secp256k1"
	"reflect"
)

const (
	BTCMsgSignaturePrefixMagic     = "Bitcoin Signed Message:\n"
	QitmeerMsgSignaturePrefixMagic = "Qitmeer Signed Message:\n"
)

func DecodeSignature(signatureType string, signStr string) {
	signHex, err := base64.StdEncoding.DecodeString(signStr)
	if err != nil {
		ErrExit(err)
	}
	signature, err := ecc.Secp256k1.ParseSignature(signHex)
	if err != nil {
		ErrExit(err)
	}
	fmt.Printf("R=%x,S=%x\n", signature.GetR(), signature.GetS())
}

func DecodeAddr(mode string, addrStr string) ([]byte, error) {
	switch mode {
	case "btc":
		addrHash160, _, err := base58.BtcCheckDecode(addrStr)
		return addrHash160, err
	default:
		addrHash160, _, err := base58.QitmeerCheckDecode(addrStr)
		return addrHash160, err
	}
}
func VerifyMsgSignature(mode string, addrStr string, signStr string, msgStr string) {

	msgHash := BuildMsgHash(mode, msgStr)

	addrHash160, err := DecodeAddr(mode, addrStr)
	if err != nil {
		ErrExit(err)
	}

	sign_c, err := base64.StdEncoding.DecodeString(signStr)
	if err != nil {
		ErrExit(err)
	}

	// the recovery mode of btc
	// That number between 0 and 3 we call the recovery id, or recid.
	// Therefore, we return an extra byte, which also functions as a header byte,
	// by using 27+recid (for uncompressed recovered pubkeys) or 31+recid (for compressed recovered pubkeys).
	// https://bitcoin.stackexchange.com/a/38909
	// 27 + recId (uncompressed pubkey)
	// 31 + recId (compressed pubkey)
	pubKey, compressed, err := ecc.Secp256k1.RecoverCompact(sign_c, msgHash)
	if err != nil {
		ErrExit(err)
	}

	var data []byte
	if compressed {
		data = pubKey.SerializeCompressed()
	} else {
		data = pubKey.SerializeUncompressed()
	}

	fmt.Printf("%v\n", reflect.DeepEqual(CalcHash160(mode, data), addrHash160))
}

func CalcHash160(mode string, data []byte) []byte {
	switch mode {
	case "btc":
		return btc.Hash160(data)
	default:
		return hash.Hash160(data)
	}
}

func BuildMsgHash(mode string, msg string) []byte {
	var msgHash []byte
	switch mode {
	case "btc":
		var buf bytes.Buffer
		serialization.WriteVarString(&buf, 0, BTCMsgSignaturePrefixMagic)
		serialization.WriteVarString(&buf, 0, msg)
		msgHash = btc.DoubleHashB(buf.Bytes())

	default:
		var buf bytes.Buffer
		serialization.WriteVarString(&buf, 0, QitmeerMsgSignaturePrefixMagic)
		serialization.WriteVarString(&buf, 0, msg)
		msgHash = hash.HashB(buf.Bytes())
	}
	return msgHash
}

func MsgSign(mode string, showSignDetail bool, wif string, msg string, showDetails bool) {
	decoded, compressed, err := DecodeWIF(wif)
	if err != nil {
		ErrExit(err)
	}
	privateKey, _ := ecc.Secp256k1.PrivKeyFromBytes(decoded)

	msgHash := BuildMsgHash(mode, msg)
	//
	r, s, err := ecc.Secp256k1.Sign(privateKey, msgHash)
	if err != nil {
		ErrExit(err)
	}
	// got the der signature
	sigHex := ecc.Secp256k1.NewSignature(r, s).Serialize()

	sign_c, err := secp256k1.SignCompact(secp256k1.NewPrivateKey(privateKey.GetD()), msgHash, compressed)
	if err != nil {
		ErrExit(err)
	}

	if showDetails {
		fmt.Printf("        mode: %s\n", mode)
		fmt.Printf("        hash: %x\n", msgHash)
		fmt.Printf("   signature: %x\n", sigHex)
		fmt.Printf("    (base64): %s\n", base64.StdEncoding.EncodeToString(sigHex))
		fmt.Printf("           R: %x\n", r)
		fmt.Printf("           S: %x\n", s)
		fmt.Printf(" compactsign: %x\n", sign_c[:])
		fmt.Printf("    (base64): %s\n", base64.StdEncoding.EncodeToString(sign_c[:]))
		fmt.Printf("  compressed: %v\n", compressed)

	} else {
		fmt.Printf("%s\n", base64.StdEncoding.EncodeToString(sign_c[:]))
	}

}
