package bip32

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"errors"
	"fmt"
	"reflect"
)

const (
	// FirstHardenedChild is the index of the firxt "harded" child key as per the
	// bip32 spec
	FirstHardenedChild = uint32(0x80000000)

	// PublicKeyCompressedLength is the byte count of a compressed public key
	PublicKeyCompressedLength = 33
)

var (
	DefaultBip32Version = Bip32Version{
		// The default bip32 version flag for serialized private keys
		[]byte{0x04,0x88,0xad,0xe4},
		// The default bip32 version flag for serialized public keys
		[]byte{0x04,0x88,0xb2,0x1e},
	}
	// ErrSerializedKeyWrongSize is returned when trying to deserialize a key that
	// has an incorrect length
	ErrSerializedKeyWrongSize = errors.New("Serialized keys should by exactly 82 bytes")

	// ErrHardnedChildPublicKey is returned when trying to create a harded child
	// of the public key
	ErrHardnedChildPublicKey = errors.New("Can't create hardened child for public key")

	// ErrInvalidChecksum is returned when deserializing a key with an incorrect
	// checksum
	ErrInvalidChecksum = errors.New("Checksum doesn't match")

	// ErrInvalidPrivateKey is returned when a derived private key is invalid
	ErrInvalidPrivateKey = errors.New("Invalid private key")

	// ErrInvalidPublicKey is returned when a derived public key is invalid
	ErrInvalidPublicKey = errors.New("Invalid public key")
)

// Key represents a bip32 extended key
type Key struct {
	Key         []byte // 33 bytes
	Version     []byte // 4 bytes
	ChildNumber []byte // 4 bytes
	FingerPrint []byte // 4 bytes
	ChainCode   []byte // 32 bytes
	Depth       byte   // 1 bytes
	IsPrivate   bool   // unserialized
	Bip32Ver    Bip32Version //unserialized
}

type Bip32Version struct {
	PrivKeyVersion []byte
	PubKeyVersion  []byte
}
func (v Bip32Version) isEmpty() bool {
	return len(v.PubKeyVersion) + len(v.PrivKeyVersion) == 0
}
func (v Bip32Version) IsPrivkeyVersion(privKeyVer []byte) bool {
	return reflect.DeepEqual(v.PrivKeyVersion,privKeyVer)
}
func (v Bip32Version) IsPubkeyVersion(pukKeyVer []byte) bool {
	return reflect.DeepEqual(v.PubKeyVersion,pukKeyVer)
}

// NewMasterKey creates a new master extended key from a seed
func NewMasterKey(seed []byte) (*Key, error) {
	return NewMasterKey2(seed,DefaultBip32Version)
}

func NewMasterKey2(seed []byte, version Bip32Version) (*Key, error) {
	// Generate key and chaincode
	hmac := hmac.New(sha512.New, []byte("Bitcoin seed"))
	_, err := hmac.Write(seed)
	if err != nil {
		return nil, err
	}
	intermediary := hmac.Sum(nil)

	// Split it into our key and chain code
	keyBytes := intermediary[:32]
	chainCode := intermediary[32:]

	// Validate key
	err = validatePrivateKey(keyBytes)
	if err != nil {
		return nil, err
	}

	if version.isEmpty() {
		version = DefaultBip32Version
	}

	// Create the key struct
	key := &Key{
		Version:     version.PrivKeyVersion,
		ChainCode:   chainCode,
		Key:         keyBytes,
		Depth:       0x0,
		ChildNumber: []byte{0x00, 0x00, 0x00, 0x00},
		FingerPrint: []byte{0x00, 0x00, 0x00, 0x00},
		IsPrivate:   true,
		Bip32Ver:    version,
	}

	return key, nil
}

// NewChildKey derives a child key from a given parent as outlined by bip32
func (key *Key) NewChildKey(childIdx uint32) (*Key, error) {
	// Fail early if trying to create hardned child from public key
	if !key.IsPrivate && childIdx >= FirstHardenedChild {
		return nil, ErrHardnedChildPublicKey
	}

	intermediary, err := key.getIntermediary(childIdx)
	if err != nil {
		return nil, err
	}

	// Create child Key with data common to all both scenarios
	childKey := &Key{
		ChildNumber: uint32Bytes(childIdx),
		ChainCode:   intermediary[32:],
		Depth:       key.Depth + 1,
		IsPrivate:   key.IsPrivate,
		Bip32Ver:    key.Bip32Ver,
	}

	// Bip32 CKDpriv
	if key.IsPrivate {
		childKey.Version = key.Bip32Ver.PrivKeyVersion
		fingerprint, err := hash160(publicKeyForPrivateKey(key.Key))
		if err != nil {
			return nil, err
		}
		childKey.FingerPrint = fingerprint[:4]
		childKey.Key = addPrivateKeys(intermediary[:32], key.Key)

		// Validate key
		err = validatePrivateKey(childKey.Key)
		if err != nil {
			return nil, err
		}
		// Bip32 CKDpub
	} else {
		keyBytes := publicKeyForPrivateKey(intermediary[:32])

		// Validate key
		err := validateChildPublicKey(keyBytes)
		if err != nil {
			return nil, err
		}
		childKey.Version = key.Bip32Ver.PubKeyVersion
		fingerprint, err := hash160(key.Key)
		if err != nil {
			return nil, err
		}
		childKey.FingerPrint = fingerprint[:4]
		childKey.Key = addPublicKeys(keyBytes, key.Key)
	}

	return childKey, nil
}

func (key *Key) getIntermediary(childIdx uint32) ([]byte, error) {
	// Get intermediary to create key and chaincode from
	// Hardened children are based on the private key
	// NonHardened children are based on the public key
	childIndexBytes := uint32Bytes(childIdx)

	var data []byte
	if childIdx >= FirstHardenedChild {
		data = append([]byte{0x0}, key.Key...)
	} else {
		if key.IsPrivate {
			data = publicKeyForPrivateKey(key.Key)
		} else {
			data = key.Key
		}
	}
	data = append(data, childIndexBytes...)

	hmac := hmac.New(sha512.New, key.ChainCode)
	_, err := hmac.Write(data)
	if err != nil {
		return nil, err
	}
	return hmac.Sum(nil), nil
}

// PublicKey returns the public version of key or return a copy
// The 'Neuter' function from the bip32 spec
func (key *Key) PublicKey() *Key {
	keyBytes := key.Key

	if key.IsPrivate {
		keyBytes = publicKeyForPrivateKey(keyBytes)
	}

	return &Key{
		Version:     key.Bip32Ver.PubKeyVersion,
		Key:         keyBytes,
		Depth:       key.Depth,
		ChildNumber: key.ChildNumber,
		FingerPrint: key.FingerPrint,
		ChainCode:   key.ChainCode,
		IsPrivate:   false,
		Bip32Ver:    key.Bip32Ver,
	}
}

// Serialize a Key to a 78 byte byte slice
func (key *Key) Serialize() ([]byte, error) {
	// Private keys should be prepended with a single null byte
	keyBytes := key.Key
	if key.IsPrivate {
		keyBytes = append([]byte{0x0}, keyBytes...)
	}

	// Write fields to buffer in order
	buffer := new(bytes.Buffer)
	buffer.Write(key.Version)
	buffer.WriteByte(key.Depth)
	buffer.Write(key.FingerPrint)
	buffer.Write(key.ChildNumber)
	buffer.Write(key.ChainCode)
	buffer.Write(keyBytes)

	// Append the standard doublesha256 checksum
	serializedKey, err := addChecksumToBytes(buffer.Bytes())
	if err != nil {
		return nil, err
	}

	return serializedKey, nil
}

// B58Serialize encodes the Key in the standard Bitcoin base58 encoding
func (key *Key) B58Serialize() string {
	serializedKey, err := key.Serialize()
	if err != nil {
		return ""
	}

	return base58Encode(serializedKey)
}

// String encodes the Key in the standard Bitcoin base58 encoding
func (key *Key) String() string {
	return key.B58Serialize()
}

func Deserialize(data []byte) (*Key, error) {
	return Deserialize2(data, DefaultBip32Version)
}
// Deserialize a byte slice into a Key
func Deserialize2(data []byte, version Bip32Version) (*Key, error) {
	if len(data) != 82 {
		return nil, ErrSerializedKeyWrongSize
	}
	var key = &Key{}
	// check version correct
	if version.IsPrivkeyVersion(data[0:4]) || version.IsPubkeyVersion(data[0:4]) {
		key.Version = data[0:4]
		key.Bip32Ver = version
	}else {
		return nil, fmt.Errorf("fail to derserilize, unmatched version data %x with {%x,%x}",data[0:4],version.PrivKeyVersion,version.PubKeyVersion)
	}
	key.Depth = data[4]
	key.FingerPrint = data[5:9]
	key.ChildNumber = data[9:13]
	key.ChainCode = data[13:45]

	if data[45] == byte(0) {
		key.IsPrivate = true
		key.Key = data[46:78]
	} else {
		key.IsPrivate = false
		key.Key = data[45:78]
	}

	// validate checksum
	cs1, err := checksum(data[0 : len(data)-4])
	if err != nil {
		return nil, err
	}

	cs2 := data[len(data)-4:]
	for i := range cs1 {
		if cs1[i] != cs2[i] {
			return nil, ErrInvalidChecksum
		}
	}
	return key, nil
}

// B58Deserialize deserializes a Key encoded in base58 encoding
func B58Deserialize(data string, version Bip32Version) (*Key, error) {
	b, err := base58Decode(data)
	if err != nil {
		return nil, err
	}
	return Deserialize2(b,version)
}

// NewSeed returns a cryptographically secure seed
func NewSeed() ([]byte, error) {
	// Well that easy, just make go read 256 random bytes into a slice
	s := make([]byte, 256)
	_, err := rand.Read(s)
	return s, err
}
