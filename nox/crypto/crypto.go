// Copyright 2017-2018 The nox developers

package crypto

import (
	"math/big"
	"crypto"
)

type PublicKey  crypto.PublicKey
type PrivateKey crypto.PrivateKey

 // Secp256k1 is the secp256k1 curve and ECDSA system used in Bitcoin.
var Secp256k1 = newSecp256k1DSA()

func newSecp256k1DSA() interface{} {
	
}

// Edwards is the Ed25519 ECDSA signature system.
var Edwards = newEdwardsDSA()

// SecSchnorr is a Schnorr signature scheme about the secp256k1 curve
// implemented in libsecp256k1.
var SecSchnorr = newSecSchnorrDSA()

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




