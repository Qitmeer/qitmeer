package main

import (
	"github.com/ethereum/go-ethereum/crypto"
	"fmt"
	"os"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common/math"
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
	return elliptic.Marshal(crypto.S256(), pub.X, pub.Y)
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
	return BytesToAddress(crypto.Keccak256(pubBytes[1:])[12:])
}

func main() {

	args := os.Args[1:]
	inputPrviKey := args[0]

	privkey, err := crypto.HexToECDSA(inputPrviKey)
	if err != nil {
		panic(err)
	}
	pubKey := privkey.PublicKey


	fmt.Printf("priv_big %v \n", privkey)
	fmt.Printf("priv %s\n",hex.EncodeToString(FromECDSAPriv(privkey)))
	fmt.Printf("pub %s \n", hex.EncodeToString(FromECDSAPub(&pubKey)))

	pubkeyHash0 := crypto.Keccak256(FromECDSAPub(&pubKey))
	pubkeyHash1 := crypto.Keccak256(FromECDSAPub(&pubKey)[1:])
	pubkeyHash40 := crypto.Keccak256(FromECDSAPub(&pubKey)[1:])[12:]

	fmt.Printf("pkhash %s \n", hex.EncodeToString(pubkeyHash0))
	fmt.Printf("pkhash %s \n", hex.EncodeToString(pubkeyHash1))
	fmt.Printf("pkhash %s \n", hex.EncodeToString(pubkeyHash40))

	addr := PubkeyToAddress(pubKey)
	fmt.Printf("add %v \n", addr)
	fmt.Printf("add %s \n", hex.EncodeToString(addr[:]))
}
