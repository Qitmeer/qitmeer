// Copyright 2017-2018 The nox developers

package crypto

import (
	"math/big"
	"crypto"
	"fmt"
)

type PublicKey  crypto.PublicKey
type PrivateKey crypto.PrivateKey

type DSAType byte

const (

	// the secp256k1 curve and ECDSA system used in Bitcoin and Ethereum
	Secp256k1DSA DSAType = iota

	// the Ed25519 ECDSA signature system.
    Ed25519DSA

	// the Schnorr signature scheme about the secp256k1 curve
	// implemented in libsecp256k1.
    SchnorrDSA

    // the Sm2 ecdsa, SM2-P-256
    // TODO, try github.com/tjfoc/gmsm/sm2
    SM2
)

func GenKeyPair() (PublicKey,PrivateKey, error){
	return GenerateKeyPair(Secp256k1DSA)
}

func GenerateKeyPair(dsa DSAType) (PublicKey,PrivateKey, error) {
	switch dsa {
	case Secp256k1DSA:
		GenerateKeySecp256k1()
	}
	return nil,nil,fmt.Errorf("unsupport DSA type %v",dsa)
}


// PublicKey represents a public key using an unspecified algorithm.
type EcdsaPubKey interface {

	// GetCurve returns the current curve as an interface.
	GetCurve() interface{}

	// GetX returns the point's X value.
	GetX() *big.Int

	// GetY returns the point's Y value.
	GetY() *big.Int
}

// PrivateKey represents a private key using an unspecified algorithm.
type EcdsaPrivKey interface {

	getPubKey() (*big.Int, *big.Int)
}

// DSA is an encapsulating interface for all the functions of a digital
// signature algorithm.
type DSA interface {
		// ----------------------------------------------------------------------------
	// Constants
	//
	// GetP gets the prime modulus of the curve.
	GetP() *big.Int

	// GetN gets the prime order of the curve.
	GetN() *big.Int

}




