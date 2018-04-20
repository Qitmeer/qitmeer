// Copyright 2017-2018 The nox developers

package crypto

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"crypto"
	"crypto/ecdsa"
)

func TestGenKeyPair(t *testing.T) {
	privKey, pubKey, err := GenKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t,privKey)
	assert.NotNil(t,pubKey)
	assert.NotNil(t,privKey.Public())

	_,ok := privKey.Public().(crypto.PublicKey)
	assert.True(t, ok)

	ecdpk,ok := privKey.Public().(ecdsa.PublicKey)
	assert.True(t, ok)
	assert.NotNil(t,ecdpk.X)
	assert.NotNil(t,ecdpk.Y)

}

func TestGenrateKeyPair(t *testing.T) {

	privKey, pubKey, err := GenerateKeyPair(Secp256k1DSA)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t,privKey)
	assert.NotNil(t,pubKey)
	assert.NotNil(t,pubKey.GetX())
	assert.NotNil(t,pubKey.GetY())
	assert.Equal(t,Secp256k1,pubKey.GetEcType())
	/*
	fmt.Printf("public key\n")
	fmt.Printf("x=%s\n",pubKey.GetX())
	fmt.Printf("y=%s\n",pubKey.GetY())
	fmt.Printf("curve=%v\n",pubKey.GetCurve())
	*/
}
