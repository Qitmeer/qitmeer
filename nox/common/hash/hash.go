// Copyright 2017-2018 The nox developers

package hash

import (
	"hash"
	"crypto"
	_ "crypto/sha256"
    _ "golang.org/x/crypto/sha3"
    _ "golang.org/x/crypto/ripemd160"
    _ "golang.org/x/crypto/blake2b"
)

const HashSize = 32

type Hash [HashSize]byte

// IsEqual returns true if target is the same as hash.
func (hash *Hash) IsEqual(target *Hash) bool {
	if hash == nil && target == nil {
		return true
	}
	if hash == nil || target == nil {
		return false
	}
	return *hash == *target
}

type Hash256 [32]byte

type Hash512 [64]byte

type Hasher interface{
	hash.Hash
}

type HashType byte

// TODO refactoring hasher
// consider to integrated https://github.com/multiformats/go-multihash
const (
	SHA256 HashType = iota
	keccak_256
	keccak_512
	ripemd160
	blake2b_256
	blake2b_512
)

func GetHasher(ht HashType) Hasher{
	switch ht {
	case SHA256:
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
