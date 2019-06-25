// Copyright 2017-2018 The nox developers

package types

import (
	"github.com/HalalChain/qitmeer-lib/crypto/ecc"
)

type Key struct {
	Address Address
	PrivateKey *ecc.PrivateKey
}

type keyJSON struct {
	Address    string `json:"address"`
	PrivateKey string `json:"privatekey"`
}

type keyStore interface {
	// Loads and decrypts the key from disk.
	GetKey(addr Address, filename string, auth string) (*Key, error)
	// Writes and encrypts the key.
	StoreKey(filename string, k *Key, auth string) error
	// Joins filename with the key directory unless it is already absolute.
	JoinPath(filename string) string
}
