// Copyright 2017-2018 The qitmeer developers

package hash

import (
	"golang.org/x/crypto/sha3"
	"hash"
	"crypto"
	_ "crypto/sha256"
    _ "golang.org/x/crypto/sha3"
    _ "golang.org/x/crypto/ripemd160"
    _ "golang.org/x/crypto/blake2b"
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer/core/json"
)

const HashSize = 32
// MaxHashStringSize is the maximum length of a Hash hash string.
const MaxHashStringSize = HashSize * 2

// ErrHashStrSize describes an error that indicates the caller specified a hash
// string that has too many characters.
var ErrHashStrSize = fmt.Errorf("max hash string length is %v bytes", MaxHashStringSize)

type Hash [HashSize]byte

type Hash256 [32]byte

type Hash512 [64]byte

type Hasher interface{
	hash.Hash
}

var ZeroHash = Hash([32]byte{ // Make go vet happy.
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		})

type HashType byte

// TODO refactoring hasher
// consider to integrated https://github.com/multiformats/go-multihash
const (
	SHA256 HashType = iota
	Keccak_256
	SHA3_256
	SHA3_512
	Ripemd160
	Blake2b_256
	Blake2b_512
)

func GetHasher(ht HashType) Hasher{
	switch ht {
	case SHA256:
		return crypto.SHA256.New()
	case Keccak_256:
		return sha3.NewLegacyKeccak256()
	case SHA3_256:
		return crypto.SHA3_256.New()
	case SHA3_512:
		return crypto.SHA3_512.New()
	case Ripemd160:
		return crypto.RIPEMD160.New()
	case Blake2b_256:
		return crypto.BLAKE2b_256.New()
	case Blake2b_512:
		return crypto.BLAKE2b_512.New()
	}
	return nil
}

// String returns the Hash as the hexadecimal string of the byte-reversed
// hash.
func (hash Hash) String() string {
	for i := 0; i < HashSize/2; i++ {
		hash[i], hash[HashSize-1-i] = hash[HashSize-1-i], hash[i]
	}
	return hex.EncodeToString(hash[:])
}

func (h Hash) Bytes() []byte { return h[:] }


// CloneBytes returns a copy of the bytes which represent the hash as a byte
// slice.
//
// NOTE: It is generally cheaper to just slice the hash directly thereby reusing
// the same bytes rather than calling this method.
func (hash *Hash) CloneBytes() []byte {
	newHash := make([]byte, HashSize)
	copy(newHash, hash[:])

	return newHash
}

// SetBytes sets the bytes which represent the hash.  An error is returned if
// the number of bytes passed in is not HashSize.
func (hash *Hash) SetBytes(newHash []byte) error {
	nhlen := len(newHash)
	if nhlen != HashSize {
		return fmt.Errorf("invalid hash length of %v, want %v", nhlen,
			HashSize)
	}
	copy(hash[:], newHash)

	return nil
}

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

// NewHash returns a new Hash from a byte slice.  An error is returned if
// the number of bytes passed in is not HashSize.
func NewHash(newHash []byte) (*Hash, error) {
	var sh Hash
	err := sh.SetBytes(newHash)
	if err != nil {
		return nil, err
	}
	return &sh, err
}

// NewHashFromStr creates a Hash from a hash string.  The string should be
// the hexadecimal string of a byte-reversed hash, but any missing characters
// result in zero padding at the end of the Hash.
func NewHashFromStr(hash string) (*Hash, error) {
	ret := new(Hash)
	err := Decode(ret, hash)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// convert hex string to a hash. Must means it panics for invalid input.
func MustHexToHash(i string) Hash {
	data, err := hex.DecodeString(i)
	if err != nil {
		panic(err)
	}

	var h Hash
	if len(data) > len(h) {
		data = data[len(data)-HashSize:]
	}
	copy(h[HashSize-len(data):], data)

	var nh Hash
	err = nh.SetBytes(h[:])
	if err != nil {
		panic(err)
	}
	return nh
}
// convert hex string to a byte-reversed hash, Must means it panics for invalid input.
func MustHexToDecodedHash(i string) Hash {
	h, err := NewHashFromStr(i)
	if err!=nil {
		panic(err)
	}
	return *h
}

// convert []byte to a hash, Must means it panics for invalid input.
func MustBytesToHash(b []byte) Hash {
	var h Hash
	if len(b) > len(h) {
		b = b[len(b)-HashSize:]
	}
	copy(h[HashSize-len(b):], b)

	hh, err :=NewHash(h[:])
	if err != nil {
		panic(err)
	}
	return *hh
}

// convert []byte to a byte-reversed hash, Must means it panics for invalid input.
func MustBytesToDecodeHash(b []byte) Hash {
	s := hex.EncodeToString(b)
	return MustHexToDecodedHash(s)
}

// Decode decodes the byte-reversed hexadecimal string encoding of a Hash to a
// destination.
func Decode(dst *Hash, src string) error {
	// Return error if hash string is too long.
	if len(src) > MaxHashStringSize {
		return ErrHashStrSize
	}

	// Hex decoder expects the hash to be a multiple of two.  When not, pad
	// with a leading zero.
	var srcBytes []byte
	if len(src)%2 == 0 {
		srcBytes = []byte(src)
	} else {
		srcBytes = make([]byte, 1+len(src))
		srcBytes[0] = '0'
		copy(srcBytes[1:], src)
	}

	// Hex decode the source bytes to a temporary destination.
	var reversedHash Hash
	_, err := hex.Decode(reversedHash[HashSize-hex.DecodedLen(len(srcBytes)):], srcBytes)
	if err != nil {
		return err
	}

	// Reverse copy from the temporary hash to destination.  Because the
	// temporary was zeroed, the written result will be correctly padded.
	for i, b := range reversedHash[:HashSize/2] {
		dst[i], dst[HashSize-1-i] = reversedHash[HashSize-1-i], b
	}

	return nil
}


// UnmarshalText decodes the hash from hex. The 0x prefix is optional.
// TODO clean-up the byte-reverse hash
func (h *Hash) UnmarshalText(input []byte) error {
 	var inputStr [HashSize]byte
	err := json.UnmarshalFixedUnprefixedText("UnprefixedHash", input, inputStr[:])
	if err!=nil {
		return err
	}
	//TODO, remove the need to reverse byte
	err = Decode(h, hex.EncodeToString(inputStr[:]))
	if err!=nil{
		return err
	}
	return nil
}

// MarshalText encodes the hash as hex.
// TODO, impl after the clean-up of byte-reverse hash
/*
func (h Hash) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeToString(h[:])), nil
}
*/
