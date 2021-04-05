package types

const (
	// BloomUpdateNone indicates the filter is not adjusted when a match is
	// found.
	BloomUpdateNone BloomUpdateType = 0

	// BloomUpdateAll indicates if the filter matches any data element in a
	// public key script, the outpoint is serialized and inserted into the
	// filter.
	BloomUpdateAll BloomUpdateType = 1

	// BloomUpdateP2PubkeyOnly indicates if the filter matches a data
	// element in a public key script and the script is of the standard
	// pay-to-pubkey or multisig, the outpoint is serialized and inserted
	// into the filter.
	BloomUpdateP2PubkeyOnly BloomUpdateType = 2

	// MaxFilterLoadHashFuncs is the maximum number of hash functions to
	// load into the Bloom filter.
	MaxFilterLoadHashFuncs = 50

	// MaxFilterLoadFilterSize is the maximum size in bytes a filter may be.
	MaxFilterLoadFilterSize = 36000
)

// MsgFilterAdd implements the Message interface and represents a qitmeer
// filteradd message.  It is used to add a data element to an existing Bloom
// filter.
//
// This message was not added until protocol version BIP0037Version.
type MsgFilterAdd struct {
	Data []byte
}

// MsgFilterClear implements the Message interface and represents a qitmeer
// filterclear message which is used to reset a Bloom filter.
//
// This message was not added until protocol version BIP0037Version and has
// no payload.
type MsgFilterClear struct{}

// BloomUpdateType specifies how the filter is updated when a match is found
type BloomUpdateType uint8

// MsgFilterLoad implements the Message interface and represents a bitcoin
// filterload message which is used to reset a Bloom filter.
//
// This message was not added until protocol version BIP0037Version.
type MsgFilterLoad struct {
	Filter    []byte
	HashFuncs uint32
	Tweak     uint32
	Flags     BloomUpdateType
}

// NewMsgFilterLoad returns a new bitcoin filterload message that conforms to
// the Message interface.  See MsgFilterLoad for details.
func NewMsgFilterLoad(filter []byte, hashFuncs uint32, tweak uint32, flags BloomUpdateType) *MsgFilterLoad {
	return &MsgFilterLoad{
		Filter:    filter,
		HashFuncs: hashFuncs,
		Tweak:     tweak,
		Flags:     flags,
	}
}
