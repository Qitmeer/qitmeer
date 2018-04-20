// Copyright 2017-2018 The nox developers

package types

import (
	"hash"
	"crypto"
	_ "crypto/sha256"
    _ "golang.org/x/crypto/sha3"
    _ "golang.org/x/crypto/ripemd160"
    _ "golang.org/x/crypto/blake2b"
)

type Hash interface{

}

type Hash160 [20]byte

type Hash256 [32]byte

type Hash512 [64]byte

type Hasher interface{
	hash.Hash
}

type HashType byte

const (
	sha256 HashType = iota
	keccak_256
	keccak_512
	ripemd160
	blake2b_256
	blake2b_512
)

func GetHasher(ht HashType) Hasher{
	switch ht {
	case sha256:
		return crypto.SHA256.New()
	case keccak_256:
		return crypto.SHA3_256.New()
	case keccak_512:
		return crypto.SHA3_512.New()
	case ripemd160:
		return crypto.RIPEMD160.New()
	case blake2b_256:
		return crypto.BLAKE2b_256.New()
	case blake2b_512:
		return crypto.BLAKE2b_512.New()
	}
	return nil
}
