// Copyright 2017-2018 The nox developers

package crypto

import (
	"crypto/ecdsa"
	"crypto/rand"
	"golang.org/x/crypto/ed25519"
	"fmt"
)


func GenerateKeyEd25519() (*ecdsa.PrivateKey, error) {
	if _, privkey, err := ed25519.GenerateKey(rand.Reader); err!=nil {
		return nil,err
	}else {
		return parsePrivKeyFromBytes(privkey)
	}
}

func parsePrivKeyFromBytes(pkBytes []byte) (*ecdsa.PrivateKey, error) {

	return nil,fmt.Errorf("error parsePrivKeyFromBytes for ed25519")
}
