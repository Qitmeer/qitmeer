package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/Qitmeer/qng-core/common/hash"
	"github.com/Qitmeer/qng-core/common/math"
	"github.com/Qitmeer/qng-core/crypto/ecc/secp256k1"
	"math/big"
	"os"
)

// Lengths of hashes and addresses in bytes.
const (
	HashLength    = 32
	AddressLength = 20
)

// Address represents the 20 byte address of an Ethereum account.
type Address [AddressLength]byte

func BytesToAddress(b []byte) Address {
	var a Address
	a.SetBytes(b)
	return a
}

// SetBytes sets the address to the value of b.
// If b is larger than len(a) it will panic.
func (a *Address) SetBytes(b []byte) {
	if len(b) > len(a) {
		b = b[len(b)-AddressLength:]
	}
	copy(a[AddressLength-len(b):], b)
}

// FromECDSAPub exports a pub key into a binary dump.
func FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(secp256k1.S256(), pub.X, pub.Y)
}

// FromECDSAPrv exports a private key into a binary dump.
func FromECDSAPriv(priv *ecdsa.PrivateKey) []byte {
	if priv == nil {
		return nil
	}
	return math.PaddedBigBytes(priv.D, priv.Params().BitSize/8)
}

func PubkeyToAddress(p ecdsa.PublicKey) Address {
	pubBytes := FromECDSAPub(&p)
	return BytesToAddress(keccak256(pubBytes[1:])[12:])
}

func main() {

	args := os.Args[1:]
	inputPrviKey := args[0]

	privkey, err := hexToECDSA(inputPrviKey)
	if err != nil {
		panic(err)
	}
	pubKey := privkey.PublicKey

	fmt.Printf("priv_big %v \n", privkey)
	fmt.Printf("priv %s\n", hex.EncodeToString(FromECDSAPriv(privkey)))
	fmt.Printf("pub %s \n", hex.EncodeToString(FromECDSAPub(&pubKey)))

	pubkeyHash0 := keccak256(FromECDSAPub(&pubKey))
	pubkeyHash1 := keccak256(FromECDSAPub(&pubKey)[1:])
	pubkeyHash40 := keccak256(FromECDSAPub(&pubKey)[1:])[12:]

	fmt.Printf("pkhash %s \n", hex.EncodeToString(pubkeyHash0))
	fmt.Printf("pkhash %s \n", hex.EncodeToString(pubkeyHash1))
	fmt.Printf("pkhash %s \n", hex.EncodeToString(pubkeyHash40))

	addr := PubkeyToAddress(pubKey)
	fmt.Printf("add %v \n", addr)
	fmt.Printf("add %s \n", hex.EncodeToString(addr[:]))
}

func keccak256(data []byte) []byte {
	return hash.CalcHash(data, hash.GetHasher(hash.Keccak_256))
}

func hexToECDSA(hexkey string) (*ecdsa.PrivateKey, error) {
	b, err := hex.DecodeString(hexkey)
	if err != nil {
		return nil, errors.New("invalid hex string")
	}
	return toECDSA(b, true)
}

// toECDSA creates a private key with the given D value. The strict parameter
// controls whether the key's length should be enforced at the curve size or
// it can also accept legacy encodings (0 prefixes).
func toECDSA(d []byte, strict bool) (*ecdsa.PrivateKey, error) {
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = secp256k1.S256()
	if strict && 8*len(d) != priv.Params().BitSize {
		return nil, fmt.Errorf("invalid length, need %d bits", priv.Params().BitSize)
	}
	priv.D = new(big.Int).SetBytes(d)

	// The priv.D must < N
	if priv.D.Cmp(secp256k1.S256().N) >= 0 {
		return nil, fmt.Errorf("invalid private key, >=N")
	}
	// The priv.D must not be zero or negative.
	if priv.D.Sign() <= 0 {
		return nil, fmt.Errorf("invalid private key, zero or negative")
	}

	priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(d)
	if priv.PublicKey.X == nil {
		return nil, errors.New("invalid private key")
	}
	return priv, nil
}
