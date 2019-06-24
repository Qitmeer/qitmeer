// Copyright (c) 2017-2018 The nox developers

package ecc

import (
	"math/big"
	"crypto/ecdsa"
)

// The Ec Type
type EcType int
const (
	// the secp256k1 curve and ECDSA system used in Bitcoin and Ethereum
	ECDSA_Secp256k1 EcType = iota      // 0

	// the Ed25519 ECDSA signature system.
	EdDSA_Ed25519                   // 1

	// the Schnorr signature scheme
	// TODO
	// 1.) the secp256k1 curve implemented in libsecp256k1
	// 2.) the Schnorr signatures over Curve25519
	ECDSA_SecpSchnorr               // 2

	// the Sm2 ecdsa, SM2-P-256
	// TODO, try github.com/tjfoc/gmsm/sm2
	ECDSA_SM2
)

// Key represents a ec key
type Key interface {

	// returns a serialized representation of this key
	Serialize() []byte

	// GetType returns the ECDSA type of this key.
	GetType() int
}

// PublicKey is an interface representing a public key and its associated
// functions.
type PublicKey interface {

	Key

	// SerializeUncompressed serializes to the uncompressed format (if
	// available).
	SerializeUncompressed() []byte

	// SerializeCompressed serializes to the compressed format (if
	// available).
	SerializeCompressed() []byte

	// ToECDSA converts the public key to an ECDSA public key.
	ToECDSA() *ecdsa.PublicKey

	// GetCurve returns the current curve as an interface.
	GetCurve() interface{}

	// GetX returns the point's X value.
	GetX() *big.Int

	// GetY returns the point's Y value.
	GetY() *big.Int


}

// PrivateKey is an interface representing a private key and its associated
// functions.
type PrivateKey interface {
	Key

	// SerializeSecret serializes the secret to the default serialization
	// format. Used for Ed25519.
	SerializeSecret() []byte

	// Public returns the (X,Y) coordinates of the point produced
	// by scalar multiplication of the scalar by the base point,
	// AKA the public key.
	Public() (*big.Int, *big.Int)

	// GetD returns the value of the private scalar.
	GetD() *big.Int
}

// Signature is an interface representing a signature and its associated
// functions.
type Signature interface {
	Key

	// GetR gets the R value of the signature.
	GetR() *big.Int

	// GetS gets the S value of the signature.
	GetS() *big.Int

}
