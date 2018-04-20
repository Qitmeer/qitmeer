// Copyright 2017-2018 The nox developers

package crypto

import (
	"math/big"
	"crypto"
	"fmt"
	"crypto/ecdsa"
	"crypto/elliptic"
)


type PublicKey  crypto.PublicKey
type PrivateKey interface {
	Public() crypto.PublicKey
}
// EccPublicKey represents a public key using an elliptic eurves algorithm. (included edwards curves)
type EccPublicKey interface {

	// GetCurve returns the current elliptic curve
	GetCurve() elliptic.Curve

	// GetX returns the point's X value.
	GetX() *big.Int

	// GetY returns the point's Y value.
	GetY() *big.Int

	// GetEcType returns the current elliptic curve type, ex. secp256k1
	GetEcType() EcType
}

// EccPrivateKey represents a private key using an elliptic curves algorithm. (included Edwards curves)
type EccPrivateKey interface {

	getPubKey() (*big.Int, *big.Int)

	Public() crypto.PublicKey
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

type DSAType byte

const (

	// the secp256k1 curve and ECDSA system used in Bitcoin and Ethereum
	Secp256k1DSA DSAType = iota

	// the Ed25519 ECDSA signature system.
	Ed25519DSA

	// the Schnorr signature scheme
	// TODO
	// 1.) the secp256k1 curve implemented in libsecp256k1
	// 2.) the Schnorr signatures over Curve25519
	SchnorrDSA

	// the Sm2 ecdsa, SM2-P-256
	// TODO, try github.com/tjfoc/gmsm/sm2
	SM2
)

type EcType byte

const (
	Secp256k1 EcType = iota
	Edwards
)

func GenKeyPair() (PrivateKey,PublicKey, error){
	return GenerateKeyPair(Secp256k1DSA)
}

func GenerateKeyPair(dsa DSAType) (EccPrivateKey,EccPublicKey, error) {
	switch dsa {
	case Secp256k1DSA:
		if privKey,err := GenerateKeySecp256k1(); err !=nil {
			return nil, nil, err
		}else {
			return ecdsaPrivateKey{privKey},
				ecdsaPubKey{privKey.PublicKey,Secp256k1}, nil
		}
	case Ed25519DSA:
		if privKey,err := GenerateKeyEd25519(); err !=nil {
			return nil, nil, err
		}else {
			return ecdsaPrivateKey{privKey},
				ecdsaPubKey{privKey.PublicKey,Edwards}, nil
		}
	}
	return nil,nil,fmt.Errorf("unsupport DSA type %v",dsa)
}

type ecdsaPubKey struct {
	ecdsa.PublicKey
	ectype EcType
}

func (pk ecdsaPubKey) GetX() *big.Int {
	return pk.X
}
func (pk ecdsaPubKey) GetY() *big.Int {
	return pk.Y
}
func (pk ecdsaPubKey) GetCurve() elliptic.Curve {
	return pk.Curve
}
func (pk ecdsaPubKey) GetEcType() EcType {
	return pk.ectype
}

type ecdsaPrivateKey struct{
	*ecdsa.PrivateKey
}

func (k ecdsaPrivateKey) getPubKey() (*big.Int, *big.Int) {
	return k.PublicKey.X, k.PublicKey.Y
}

func (k ecdsaPrivateKey) Public() crypto.PublicKey {
	return k.PublicKey
}
