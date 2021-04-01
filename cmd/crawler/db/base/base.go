package base

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type Base struct {
	db *leveldb.DB
}

func Open(path string) (*Base, error) {
	var err error
	opts := &opt.Options{
		OpenFilesCacheCapacity: 16,
		Strict:                 opt.DefaultStrict,
		Compression:            opt.NoCompression,
		BlockCacheCapacity:     8 * opt.MiB,
		WriteBuffer:            4 * opt.MiB,
	}
	b := &Base{}
	if b.db, err = leveldb.OpenFile(path, opts); err != nil {
		if b.db, err = leveldb.RecoverFile(path, nil); err != nil {
			return nil, errors.New(fmt.Sprintf(`err while recoverfile %s : %s`, path, err.Error()))
		}

	}
	return b, nil
}

func (b *Base) Close() error {
	return b.db.Close()
}

func (b *Base) Put(key []byte, value []byte) error {
	return b.db.Put(key, value, nil)
}

func (b *Base) Delete(key []byte) error {
	return b.db.Delete(key, nil)
}

func (b *Base) Get(key []byte) ([]byte, error) {
	return b.db.Get(key, nil)
}

func (b *Base) Has(key []byte) (bool, error) {
	_, err := b.db.Get(key, nil)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (b *Base) PutInBucket(bucket string, key, value []byte) error {
	return b.db.Put(Key(bucket, key), value, nil)
}

func (b *Base) GetFromBucket(bucket string, key []byte) ([]byte, error) {
	return b.db.Get(Key(bucket, key), nil)
}

func (b *Base) Clear(bucket string) {
	rs := b.Foreach(bucket)
	for key, _ := range rs {
		b.db.Delete([]byte(key), nil)
	}
}

func (b *Base) Foreach(bucket string) map[string][]byte {
	rs := make(map[string][]byte)
	iter := b.db.NewIterator(util.BytesPrefix(bytes.Join([][]byte{[]byte(bucket), []byte("-")}, []byte{})), nil)
	defer iter.Release()

	// Iter will affect RLP decoding and reallocate memory to value
	for iter.Next() {
		value := make([]byte, len(iter.Value()))
		copy(value, iter.Value())
		rs[string(LeafKeyToKey(bucket, iter.Key()))] = value
	}
	return rs
}

func Key(bucket string, key []byte) []byte {
	return bytes.Join([][]byte{
		[]byte(bucket + "-"), key}, []byte{})
}

func Prefix(bucket string) []byte {
	return []byte(bucket + "-")
}

func LeafKeyToKey(bucket string, key []byte) []byte {
	return key[len(Prefix(bucket)):]
}
