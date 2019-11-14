// Copyright (c) 2017-2018 The qitmeer developers

package benchmark

// $ go test -run='^$' -bench=. -benchmem
// goos: darwin
// goarch: amd64
// pkg: qitmeer/database/benchmark
// BenchmarkGetBadger-8             2000000               796 ns/op          11.30 MB/s         456 B/op          9 allocs/op
// BenchmarkGetLevelDB-8            3000000               443 ns/op          20.29 MB/s         112 B/op          4 allocs/op
// BenchmarkGetBolt-8               2000000               647 ns/op          13.90 MB/s         440 B/op          7 allocs/op
// PASS
// ok      qitmeer/database/benchmark    8.194s
/*
import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/coreos/bbolt"
	"github.com/dgraph-io/badger"
	"github.com/syndtr/goleveldb/leveldb"
)

var (
	testKey       = []byte("testKey")
	testValue     = []byte("testValue")
	testValueSize = int64(len(testValue))
)

func BenchmarkGetBadger(b *testing.B) {
	tmp, err := ioutil.TempDir(os.TempDir(), "badger")
	if err != nil {
		b.Fatal(err)
	}

	opt := badger.DefaultOptions
	opt.Dir = tmp
	opt.ValueDir = tmp
	db, err := badger.Open(opt)
	if err != nil {
		b.Fatal(err)
	}
	txn := db.NewTransaction(true)
	if err := txn.Set(testKey, testValue); err != nil {
		b.Fatal(err)
	}
	if err := txn.Commit(nil); err != nil {
		b.Fatal(err)
	}
	b.SetBytes(testValueSize)
	b.ResetTimer()


	for i := 0; i < b.N; i++ {
		err := db.View(func(txn *badger.Txn) error {
			_, err := txn.Get(testKey)
			if err != nil {
			 	b.Fatal(err)
			}
			return nil
		})
		if err != nil {
			b.Fatal(err)
		}
	}

	b.StopTimer()
	db.Close()
	os.RemoveAll(tmp)
}

func BenchmarkGetLevelDB(b *testing.B) {
	tmp, err := ioutil.TempDir(os.TempDir(), "leveldb")
	if err != nil {
		b.Fatal(err)
	}

	db, err := leveldb.OpenFile(tmp, nil)
	if err != nil {
		b.Fatal(err)
	}

	if err := db.Put(testKey, testValue, nil); err != nil {
		b.Fatal(err)
	}

	b.SetBytes(testValueSize)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		db.Get(testKey, nil)
	}

	b.StopTimer()
	db.Close()
	os.RemoveAll(tmp)
}

func BenchmarkGetBolt(b *testing.B) {
	tmp, err := ioutil.TempDir(os.TempDir(), "bolt")
	if err != nil {
		b.Fatal(err)
	}

	db, err := bolt.Open(filepath.Join(tmp, "test.db"), 0600, nil)
	if err != nil {
		b.Fatal(err)
	}

	bucketName := []byte("testBucket")
	updateFn := func(tx *bolt.Tx) error {
		bc, err := tx.CreateBucket(bucketName)
		if err != nil {
			b.Fatal(err)
		}
		bc.Put(testKey, testValue)
		return nil
	}
	if err := db.Update(updateFn); err != nil {
		b.Fatal(err)
	}

	b.SetBytes(testValueSize)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		db.View(func(tx *bolt.Tx) error {
			bc := tx.Bucket(bucketName)
			bc.Get(testKey)
			return nil
		})
	}

	b.StopTimer()
	db.Close()
	os.RemoveAll(tmp)
}
*/
