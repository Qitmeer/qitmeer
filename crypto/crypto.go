package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/crypto/ecc"
	"github.com/Qitmeer/qitmeer/crypto/ecc/secp256k1"
	"math/big"
)

var errInvalidPubkey = errors.New("invalid secp256k1 public key")

// Keccak256 calculates and returns the Keccak256 hash of the input data.
func Keccak256(data ...[]byte) []byte {
	d := hash.GetHasher(hash.Keccak_256)
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

// Keccak256Hash calculates and returns the Keccak256 hash of the input data,
// converting it to an internal Hash data structure.
func Keccak256Hash(data ...[]byte) (h hash.Hash) {
	d := hash.GetHasher(hash.Keccak_256)
	for _, b := range data {
		d.Write(b)
	}
	d.Sum(h[:0])
	return h
}

// <head><R><S>
func Sign(digestHash []byte, prv *ecdsa.PrivateKey) (sig []byte, err error) {
	prvKey := secp256k1.NewPrivateKey(prv.D)
	sig, err = secp256k1.SignCompact(prvKey, digestHash, false)
	if err != nil {
		return nil, fmt.Errorf("can't sign: %v", err)
	}
	return sig, nil
}

// VerifySignature checks that the given public key created signature over digest.
func VerifySignature(pubkey *ecdsa.PublicKey, digestHash, signature []byte) bool {
	sig, err := ParseSignature(signature)
	if err != nil {
		return false
	}
	publicKey := secp256k1.NewPublicKey(pubkey.X, pubkey.Y)
	return ecc.Secp256k1.Verify(publicKey, digestHash, sig.GetR(), sig.GetS())
}

// <R><S>
func ParseSignature(signature []byte) (*secp256k1.Signature, error) {
	bitlen := (secp256k1.S256().BitSize + 7) / 8
	if len(signature) != bitlen*2 {
		return nil, errors.New("invalid compact signature size")
	}
	// format is <header byte><bitlen R><bitlen S>
	sig := &secp256k1.Signature{
		R: new(big.Int).SetBytes(signature[:bitlen]),
		S: new(big.Int).SetBytes(signature[bitlen:]),
	}
	return sig, nil
}

// Ecrecover returns the uncompressed public key that created the given signature.
func Ecrecover(digestHash, sig []byte) (ecc.PublicKey, error) {
	pubKey, _, err := ecc.Secp256k1.RecoverCompact(sig, digestHash)
	if err != nil {
		return nil, err
	}
	return pubKey, nil
}

// HexToECDSA parses a secp256k1 private key.
func HexToECDSA(hexkey string) (*ecdsa.PrivateKey, error) {
	b, err := hex.DecodeString(hexkey)
	if byteErr, ok := err.(hex.InvalidByteError); ok {
		return nil, fmt.Errorf("invalid hex character %q in private key", byte(byteErr))
	} else if err != nil {
		return nil, errors.New("invalid hex data for private key")
	}
	return ToECDSA(b)
}

// ToECDSA creates a private key with the given D value.
func ToECDSA(d []byte) (*ecdsa.PrivateKey, error) {
	privateKey, _ := ecc.Secp256k1.PrivKeyFromBytes(d)
	return privateKey.(*secp256k1.PrivateKey).ToECDSA(), nil
}

// CompressPubkey encodes a public key to the 33-byte compressed format.
func CompressPubkey(pubkey *ecdsa.PublicKey) []byte {
	publicKey := secp256k1.NewPublicKey(pubkey.X, pubkey.Y)
	return publicKey.SerializeCompressed()
}

// DecompressPubkey parses a public key in the 33-byte compressed format.
func DecompressPubkey(pubkey []byte) (*ecdsa.PublicKey, error) {
	publicKey, err := secp256k1.ParsePubKey(pubkey)
	if err != nil {
		return nil, fmt.Errorf("invalid public key:%v", err)
	}
	return publicKey.ToECDSA(), nil
}

// UnmarshalPubkey converts bytes to a secp256k1 public key.
func UnmarshalPubkey(pub []byte) (*ecdsa.PublicKey, error) {
	x, y := elliptic.Unmarshal(secp256k1.S256(), pub)
	if x == nil {
		return nil, errInvalidPubkey
	}
	return &ecdsa.PublicKey{Curve: secp256k1.S256(), X: x, Y: y}, nil
}

func FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(secp256k1.S256(), pub.X, pub.Y)
}
