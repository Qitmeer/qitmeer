// Copyright (c) 2017-2018 The qitmeer developers

package hash

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/Qitmeer/qitmeer/common/util"
)

// Ensure same result with the normal way
func TestHashWithPoolGotSameResult(t *testing.T){
	data := []byte("Test data")
	h := HashB(data)
	h2 :=HashB_pool(data)
	assert.Equal(t, h,h2)

	dh := DoubleHashB(data)
	dh2 := DoubleHashB_pool(data)
	assert.Equal(t,dh,dh2)
}

// TODO revisit the cached hasher pool
// TODO add test for GC
//
// Note : Using sync.Pool for hash is a little worse than normal way
//
// go test -bench=.
//	goos: darwin
//	goarch: amd64
//	pkg: qitmeer/common/hash
//	BenchmarkHashWithPool-8      	  500000	      2544 ns/op
//	BenchmarkHashWithoutPool-8   	  500000	      2390 ns/op
//	PASS
//	ok  	qitmeer/common/hash	2.541s
//

func BenchmarkHashWithPool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		data := util.ReadSizedRand(nil,32)
		DoubleHashB_pool(data)
	}
}

func BenchmarkHashWithoutPool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		data := util.ReadSizedRand(nil,32)
		DoubleHashB(data)
	}
}

