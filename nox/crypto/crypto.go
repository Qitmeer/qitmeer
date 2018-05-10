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

// TODO: refactor to use stand api to do sign encoding,
// using the asn1.Unmarshal(b, sig)/asn1.Marshal(*sig)
// see https://github.com/google/tink/blob/master/go/subtle/signature/ecdsa.go
// see the EncodeEcdsaSignature for a example

type SignatureScheme byte

const (

	// the secp256k1 curve and ECDSA system used in Bitcoin and Ethereum
	ECDSA_Secp256k1 SignatureScheme = iota

	// the Ed25519 ECDSA signature system.
	EdDSA_Ed25519

	// the Schnorr signature scheme
	// TODO
	// 1.) the secp256k1 curve implemented in libsecp256k1
	// 2.) the Schnorr signatures over Curve25519
	ECDSA_Schnorr

	// the Sm2 ecdsa, SM2-P-256
	// TODO, try github.com/tjfoc/gmsm/sm2
	ECDSA_SM2
)

type EcType byte

const (
	Secp256k1 EcType = iota
	Edwards
)

func GenKeyPair() (PrivateKey,PublicKey, error){
	return GenerateKeyPair(ECDSA_Secp256k1)
}

func GenerateKeyPair(scheme SignatureScheme) (EccPrivateKey,EccPublicKey, error) {
	switch scheme {
	case ECDSA_Secp256k1:
		if privKey,err := GenerateKeySecp256k1(); err !=nil {
			return nil, nil, err
		}else {
			return ecdsaPrivateKey{privKey},
				ecdsaPubKey{privKey.PublicKey,Secp256k1}, nil
		}
	case EdDSA_Ed25519:
		if privKey,err := GenerateKeyEd25519(); err !=nil {
			return nil, nil, err
		}else {
			return ecdsaPrivateKey{privKey},
				ecdsaPubKey{privKey.PublicKey,Edwards}, nil
		}
	}
	return nil,nil,fmt.Errorf("unsupport SignatureScheme %v",scheme)
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
