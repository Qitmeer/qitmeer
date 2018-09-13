// Copyright (c) 2017-2018 The nox developers

package hash

import (
	"sync"
	"golang.org/x/crypto/blake2b"
	"hash"
)

// Warning !!!: Testing only, don't use it in the production
//
// attempt to use 'sync.Pool' to cache hasher
//
// References
//   https://github.com/ethereum/go-ethereum/blob/master/trie/hasher.go
//
// Basically, sync.Pool reuses allocations between garbage collection cycles,
// so that you don’t have to allocate another object.
// Each time a garbage collection cycle starts, it clears items out of the pool.
//
// TODO revisit cached hasher in sync.pool

var pool = &sync.Pool{New: func() interface{} {
	h,err := blake2b.New256(nil)
	if err!= nil {
		panic(err)
	}
	h.Reset()
	return h}}

func getFromPool() hash.Hash {
	return pool.Get().(hash.Hash)
}

func putToPool(h hash.Hash) {
	h.Reset()
	pool.Put(h)
}

// using pool
func HashB_pool(b []byte) []byte {
	cached := getFromPool()
	defer putToPool(cached)
	cached.Write(b)
	h := cached.Sum(nil)
	cached.Reset()
	return h[:]
}

func HashH_pool(b []byte) Hash{
	var h [32]byte
	copy(h[:],HashB(b))
	return Hash(h)
}

func DoubleHashH_pool(b []byte) Hash {
	var h [32]byte
	copy(h[:],DoubleHashB_pool(b))
	return Hash(h)
}

func DoubleHashB_pool(b []byte) []byte {
	cached := getFromPool()
	defer putToPool(cached)
	cached.Write(b)
	first := cached.Sum(nil)
	cached.Reset()
	cached.Write(first)
	second := cached.Sum(nil)
	return second[:]
}

