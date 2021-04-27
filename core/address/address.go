// Copyright 2017-2018 The qitmeer developers

package address

import (
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/encode/base58"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/crypto/ecc"
	"github.com/Qitmeer/qitmeer/params"
	"golang.org/x/crypto/ripemd160"
)

// encodeAddress returns a human-readable payment address given a ripemd160 hash
// and netID which encodes the network and address type.  It is used in both
// pay-to-pubkey-hash (P2PKH) and pay-to-script-hash (P2SH) address encoding.
func encodeAddress(hash160 []byte, netID [2]byte) string {
	// Format is 2 bytes for a network and address class (i.e. P2PKH vs
	// P2SH), 20 bytes for a RIPEMD160 hash, and 4 bytes of checksum.
	res, _ := base58.QitmeerCheckEncode(hash160[:ripemd160.Size], netID[:])
	return string(res)
}

// encodePKAddress returns a human-readable payment address to a public key
// given a serialized public key, a netID, and a signature suite.
func encodePKAddress(serializedPK []byte, netID [2]byte, algo ecc.EcType) string {
	pubKeyBytes := []byte{0x00}

	switch algo {
	case ecc.ECDSA_Secp256k1:
		pubKeyBytes[0] = byte(ecc.ECDSA_Secp256k1)
	case ecc.EdDSA_Ed25519:
		pubKeyBytes[0] = byte(ecc.EdDSA_Ed25519)
	case ecc.ECDSA_SecpSchnorr:
		pubKeyBytes[0] = byte(ecc.ECDSA_SecpSchnorr)
	}

	// Pubkeys are encoded as [0] = type/ybit, [1:33] = serialized pubkey
	compressed := serializedPK
	if algo == ecc.ECDSA_Secp256k1 || algo == ecc.ECDSA_SecpSchnorr {
		pub, err := ecc.Secp256k1.ParsePubKey(serializedPK)
		if err != nil {
			return ""
		}
		pubSerComp := pub.SerializeCompressed()

		// Set the y-bit if needed.
		if pubSerComp[0] == 0x03 {
			pubKeyBytes[0] |= (1 << 7)
		}

		compressed = pubSerComp[1:]
	}

	pubKeyBytes = append(pubKeyBytes, compressed...)
	res, _ := base58.QitmeerCheckEncode(pubKeyBytes, netID[:])
	return string(res)
}

// PubKeyHashAddress is an Address for a pay-to-pubkey-hash (P2PKH)
// transaction.
type PubKeyHashAddress struct {
	net   *params.Params
	netID [2]byte
	hash  [ripemd160.Size]byte
}

// NewAddressPubKeyHash returns a new AddressPubKeyHash.  pkHash must
// be 20 bytes.
func NewPubKeyHashAddress(pkHash []byte, net *params.Params, algo ecc.EcType) (*PubKeyHashAddress, error) {
	var addrID [2]byte
	switch algo {
	case ecc.ECDSA_Secp256k1:
		addrID = net.PubKeyHashAddrID
	case ecc.EdDSA_Ed25519:
		addrID = net.PKHEdwardsAddrID
	case ecc.ECDSA_SecpSchnorr:
		addrID = net.PKHSchnorrAddrID
	default:
		return nil, errors.New("unknown ECDSA algorithm")
	}
	apkh, err := newPubKeyHashAddress(pkHash, addrID)
	if err != nil {
		return nil, err
	}
	apkh.net = net
	return apkh, nil
}

// NewPubKeyHashAddressByNetId returns a new PubKeyHashAddress from net id directly instead from params
func NewPubKeyHashAddressByNetId(pkHash []byte, netID [2]byte) (*PubKeyHashAddress,
	error) {
	apkh, err := newPubKeyHashAddress(pkHash, netID)
	if err != nil {
		return nil, err
	}
	return apkh, nil
}

// newPubKeyHashAddress is the internal API to create a pubkey hash address
// with a known leading identifier byte for a network, rather than looking
// it up through its parameters.  This is useful when creating a new address
// structure from a string encoding where the identifer byte is already
// known.
func newPubKeyHashAddress(pkHash []byte, netID [2]byte) (*PubKeyHashAddress,
	error) {
	// Check for a valid pubkey hash length.
	if len(pkHash) != ripemd160.Size {
		return nil, errors.New("pkHash must be 20 bytes")
	}
	addr := &PubKeyHashAddress{netID: netID}
	copy(addr.hash[:], pkHash)
	return addr, nil
}

// EcType returns the digital signature algorithm for the associated public key
// hash.
func (a *PubKeyHashAddress) EcType() ecc.EcType {
	switch a.netID {
	case a.net.PubKeyHashAddrID:
		return ecc.ECDSA_Secp256k1
	case a.net.PKHEdwardsAddrID:
		return ecc.EdDSA_Ed25519
	case a.net.PKHSchnorrAddrID:
		return ecc.ECDSA_SecpSchnorr
	}
	return -1
}
func (a *PubKeyHashAddress) Encode() string {
	//TODO error handling
	return encodeAddress(a.hash[:], a.netID)
}

// String returns a human-readable string for the pay-to-pubkey-hash address.
// This is equivalent to calling EncodeAddress, but is provided so the type can
// be used as a fmt.Stringer.
func (a *PubKeyHashAddress) String() string {
	return a.Encode()
}

func (a *PubKeyHashAddress) Hash160() *[ripemd160.Size]byte {
	return &a.hash
}

type PubKeyAddress struct {
	pk       []byte
	addrType types.AddressType
}

// NewAddressPubKey returns a new Address. decoded must
// be 33 bytes.
func NewPubKeyAddress(decoded []byte, net *params.Params) (types.Address, error) {
	if len(decoded) == 33 {
		// First byte is the signature suite and ybit.
		suite := decoded[0]
		suite &= ^uint8(1 << 7)
		ybit := !(decoded[0]&(1<<7) == 0)
		toAppend := uint8(0x02)
		if ybit {
			toAppend = 0x03
		}
		switch ecc.EcType(suite) {
		case ecc.ECDSA_Secp256k1:
			return NewSecpPubKeyAddress(
				append([]byte{toAppend}, decoded[1:]...),
				net)
		case ecc.EdDSA_Ed25519:
			return NewEdwardsPubKeyAddress(decoded, net)
		case ecc.ECDSA_SecpSchnorr:
			return NewSecSchnorrPubKeyAddress(
				append([]byte{toAppend}, decoded[1:]...),
				net)
		}
		return nil, ErrUnknownAddressType
	}
	return nil, ErrUnknownAddressType
}

// Script returns the bytes to be included in a txout script to pay
// to a pubkey hash.  Part of the Address interface.
func (a *PubKeyHashAddress) Script() []byte {
	return a.hash[:]
}

// ScriptHashAddress is an Address for a pay-to-script-hash (P2SH)
// transaction.
type ScriptHashAddress struct {
	net   *params.Params
	hash  [ripemd160.Size]byte
	netID [2]byte
}

// NewAddressScriptHash returns a new AddressScriptHash.
func NewScriptHashAddress(serializedScript []byte, net *params.Params) (*ScriptHashAddress, error) {
	scriptHash := hash.Hash160(serializedScript)
	return newScriptHashAddressFromHash(scriptHash, net.ScriptHashAddrID)
}

// NewAddressScriptHashFromHash returns a new AddressScriptHash.  scriptHash
// must be 20 bytes.
func NewScriptHashAddressFromHash(scriptHash []byte, net *params.Params) (*ScriptHashAddress, error) {
	ash, err := newScriptHashAddressFromHash(scriptHash, net.ScriptHashAddrID)
	if err != nil {
		return nil, err
	}
	ash.net = net

	return ash, nil
}

// newAddressScriptHashFromHash is the internal API to create a script hash
// address with a known leading identifier byte for a network, rather than
// looking it up through its parameters.  This is useful when creating a new
// address structure from a string encoding where the identifer byte is already
// known.
func newScriptHashAddressFromHash(scriptHash []byte, netID [2]byte) (*ScriptHashAddress, error) {
	// Check for a valid script hash length.
	if len(scriptHash) != ripemd160.Size {
		return nil, errors.New("scriptHash must be 20 bytes")
	}

	addr := &ScriptHashAddress{netID: netID}
	copy(addr.hash[:], scriptHash)
	return addr, nil
}

// Hash160 returns the underlying array of the script hash.  This can be useful
// when an array is more appropriate than a slice (for example, when used as map
// keys).
func (a *ScriptHashAddress) Hash160() *[ripemd160.Size]byte {
	return &a.hash
}

// EncodeAddress returns the string encoding of a pay-to-script-hash
// address.  Part of the Address interface.
func (a *ScriptHashAddress) Encode() string {
	return encodeAddress(a.hash[:], a.netID)
}

func (a *ScriptHashAddress) String() string {
	return a.Encode()
}

func (a *ScriptHashAddress) EcType() ecc.EcType {
	return ecc.ECDSA_Secp256k1
}

// Script returns the bytes to be included in a txout script to pay
// to a script hash.  Part of the Address interface.
func (a *ScriptHashAddress) Script() []byte {
	return a.hash[:]
}

// PubKeyFormat describes what format to use for a pay-to-pubkey address.
type PubKeyFormat int

const (
	// PKFUncompressed indicates the pay-to-pubkey address format is an
	// uncompressed public key.
	PKFUncompressed PubKeyFormat = iota

	// PKFCompressed indicates the pay-to-pubkey address format is a
	// compressed public key.
	PKFCompressed
)

// SecpPubKeyAddress is an Address for a secp256k1 pay-to-pubkey transaction.
type SecpPubKeyAddress struct {
	net          *params.Params
	pubKeyFormat PubKeyFormat
	pubKey       ecc.PublicKey
	pubKeyHashID [2]byte
}

// NewSecpPubKeyCompressedAddress creates a new address using a compressed public key
func NewSecpPubKeyCompressedAddress(pubkey ecc.PublicKey, params *params.Params) (*SecpPubKeyAddress, error) {
	return NewSecpPubKeyAddress(pubkey.SerializeCompressed(), params)
}

// ErrInvalidPubKeyFormat indicates that a serialized pubkey is unusable as it
// is neither in the uncompressed or compressed format.
var ErrInvalidPubKeyFormat = errors.New("invalid pubkey format")

// NewAddressSecpPubKey returns a new AddressSecpPubKey which represents a
// pay-to-pubkey address, using a secp256k1 pubkey.  The serializedPubKey
// parameter must be a valid pubkey and must be uncompressed or compressed.
func NewSecpPubKeyAddress(serializedPubKey []byte,
	net *params.Params) (*SecpPubKeyAddress, error) {
	pubKey, err := ecc.Secp256k1.ParsePubKey(serializedPubKey)

	if err != nil {
		return nil, err
	}

	// Set the format of the pubkey.  This probably should be returned
	// from the crypto layer, but do it here to avoid API churn.  We already know the
	// pubkey is valid since it parsed above, so it's safe to simply examine
	// the leading byte to get the format.
	var pkFormat PubKeyFormat
	switch serializedPubKey[0] {
	case 0x02, 0x03:
		pkFormat = PKFCompressed
	case 0x04:
		pkFormat = PKFUncompressed
	default:
		return nil, ErrInvalidPubKeyFormat
	}

	return &SecpPubKeyAddress{
		net:          net,
		pubKeyFormat: pkFormat,
		pubKey:       pubKey,
		pubKeyHashID: net.PubKeyHashAddrID,
	}, nil
}

// EncodeAddress returns the string encoding of the public key as a
// pay-to-pubkey-hash.  Note that the public key format (uncompressed,
// compressed, etc) will change the resulting address.  This is expected since
// pay-to-pubkey-hash is a hash of the serialized public key which obviously
// differs with the format.
// Part of the Address interface.
func (a *SecpPubKeyAddress) Encode() string {
	return encodeAddress(hash.Hash160(a.serialize()), a.pubKeyHashID)
}

// String returns the hex-encoded human-readable string for the pay-to-pubkey
// address.  This is not the same as calling EncodeAddress.
func (a *SecpPubKeyAddress) String() string {
	return encodePKAddress(a.serialize(), a.net.PubKeyAddrID,
		ecc.ECDSA_Secp256k1)
}

// serialize returns the serialization of the public key according to the
// format associated with the address.
func (a *SecpPubKeyAddress) serialize() []byte {
	switch a.pubKeyFormat {
	default:
		fallthrough
	case PKFUncompressed:
		return a.pubKey.SerializeUncompressed()

	case PKFCompressed:
		return a.pubKey.SerializeCompressed()
	}
}

func (a *SecpPubKeyAddress) EcType() ecc.EcType {
	return ecc.ECDSA_Secp256k1
}

// Hash160 returns the underlying array of the pubkey hash.  This can be useful
// when an array is more appropriate than a slice (for example, when used as map
// keys).
func (a *SecpPubKeyAddress) Hash160() *[ripemd160.Size]byte {
	h160 := hash.Hash160(a.pubKey.SerializeCompressed())
	array := new([ripemd160.Size]byte)
	copy(array[:], h160)

	return array
}

// PubKey returns the underlying public key for the address.
func (a *SecpPubKeyAddress) PubKey() ecc.PublicKey {
	return a.pubKey
}

// PKHAddress returns the pay-to-pubkey address converted to a
// pay-to-pubkey-hash address.  Note that the public key format (uncompressed,
// compressed, etc) will change the resulting address.  This is expected since
// pay-to-pubkey-hash is a hash of the serialized public key which obviously
// differs with the format.
func (a *SecpPubKeyAddress) PKHAddress() *PubKeyHashAddress {
	return toPKHAddress(a.net, a.pubKeyHashID, a.serialize())
}

func toPKHAddress(net *params.Params, netID [2]byte, b []byte) *PubKeyHashAddress {
	addr := &PubKeyHashAddress{net: net, netID: netID}
	copy(addr.hash[:], hash.Hash160(b))
	return addr
}

// Script returns the bytes to be included in a txout script to pay
// to a public key.  Setting the public key format will affect the output of
// this function accordingly.  Part of the Address interface.
func (a *SecpPubKeyAddress) Script() []byte {
	return a.serialize()
}

// EdwardsPubKeyAddress is an Address for an Ed25519 pay-to-pubkey transaction.
type EdwardsPubKeyAddress struct {
	net          *params.Params
	pubKey       ecc.PublicKey
	pubKeyHashID [2]byte
}

// NewAddressEdwardsPubKey returns a new AddressEdwardsPubKey which represents a
// pay-to-pubkey address, using an Ed25519 pubkey.  The serializedPubKey
// parameter must be a valid 32 byte serialized public key.
func NewEdwardsPubKeyAddress(serializedPubKey []byte,
	net *params.Params) (*EdwardsPubKeyAddress, error) {
	pubKey, err := ecc.Ed25519.ParsePubKey(serializedPubKey)
	if err != nil {
		return nil, err
	}

	return &EdwardsPubKeyAddress{
		net:          net,
		pubKey:       pubKey,
		pubKeyHashID: net.PKHEdwardsAddrID,
	}, nil
}

func (a *EdwardsPubKeyAddress) EcType() ecc.EcType {
	return ecc.EdDSA_Ed25519
}
func (a *EdwardsPubKeyAddress) Encode() string {
	return encodeAddress(hash.Hash160(a.serialize()), a.pubKeyHashID)
}

// String returns the hex-encoded human-readable string for the pay-to-pubkey
// address.  This is not the same as calling EncodeAddress.
func (a *EdwardsPubKeyAddress) String() string {
	return encodePKAddress(a.serialize(), a.net.PubKeyAddrID,
		ecc.EdDSA_Ed25519)
}

// serialize returns the serialization of the public key.
func (a *EdwardsPubKeyAddress) serialize() []byte {
	return a.pubKey.Serialize()
}

// Hash160 returns the underlying array of the pubkey hash.  This can be useful
// when an array is more appropriate than a slice (for example, when used as map
// keys).
func (a *EdwardsPubKeyAddress) Hash160() *[ripemd160.Size]byte {
	h160 := hash.Hash160(a.pubKey.Serialize())
	array := new([ripemd160.Size]byte)
	copy(array[:], h160)
	return array
}

func (a *EdwardsPubKeyAddress) PKHAddress() *PubKeyHashAddress {
	return toPKHAddress(a.net, a.pubKeyHashID, a.serialize())
}

// Script returns the bytes to be included in a txout script to pay
// to a public key.  Setting the public key format will affect the output of
// this function accordingly.  Part of the Address interface.
func (a *EdwardsPubKeyAddress) Script() []byte {
	return a.serialize()
}

// SecSchnorrPubKeyAddress is an Address for a secp256k1-schnorr pay-to-pubkey transaction.
type SecSchnorrPubKeyAddress struct {
	net          *params.Params
	pubKey       ecc.PublicKey
	pubKeyHashID [2]byte
}

// NewAddressSecSchnorrPubKey returns a new AddressSecpPubKey which represents a
// pay-to-pubkey address, using a secp256k1 pubkey.  The serializedPubKey
// parameter must be a valid pubkey and must be compressed.
func NewSecSchnorrPubKeyAddress(serializedPubKey []byte,
	net *params.Params) (*SecSchnorrPubKeyAddress, error) {
	pubKey, err := ecc.SecSchnorr.ParsePubKey(serializedPubKey)
	if err != nil {
		return nil, err
	}
	return &SecSchnorrPubKeyAddress{
		net:          net,
		pubKey:       pubKey,
		pubKeyHashID: net.PKHSchnorrAddrID,
	}, nil
}

func (a *SecSchnorrPubKeyAddress) EcType() ecc.EcType {
	return ecc.ECDSA_SecpSchnorr
}

func (a *SecSchnorrPubKeyAddress) Encode() string {
	return encodeAddress(hash.Hash160(a.serialize()), a.pubKeyHashID)
}

// String returns the hex-encoded human-readable string for the pay-to-pubkey
// address.  This is not the same as calling EncodeAddress.
func (a *SecSchnorrPubKeyAddress) String() string {
	return encodePKAddress(a.serialize(), a.net.PubKeyAddrID,
		ecc.ECDSA_SecpSchnorr)
}

func (a *SecSchnorrPubKeyAddress) serialize() []byte {
	return a.pubKey.Serialize()
}

// Hash160 returns the underlying array of the pubkey hash.  This can be useful
// when an array is more appropriate than a slice (for example, when used as map
// keys).
func (a *SecSchnorrPubKeyAddress) Hash160() *[ripemd160.Size]byte {
	h160 := hash.Hash160(a.pubKey.Serialize())
	array := new([ripemd160.Size]byte)
	copy(array[:], h160)
	return array
}

func (a *SecSchnorrPubKeyAddress) PKHAddress() *PubKeyHashAddress {
	return toPKHAddress(a.net, a.pubKeyHashID, a.serialize())
}

// Script returns the bytes to be included in a txout script to pay
// to a public key.  Setting the public key format will affect the output of
// this function accordingly.  Part of the Address interface.
func (a *SecSchnorrPubKeyAddress) Script() []byte {
	return a.serialize()
}

type ContractAddress struct {
	pk       []byte
	addrType types.AddressType
}

// DecodeAddress decodes the string encoding of an address and returns
// the Address if addr is a valid encoding for a known address type
func DecodeAddress(addr string) (types.Address, error) {
	// Switch on decoded length to determine the type.
	decoded, netID, err := base58.QitmeerCheckDecode(addr)
	if err != nil {
		if err == base58.ErrChecksum {
			return nil, ErrChecksumMismatch
		}
		return nil, fmt.Errorf("decoded address is of unknown format: %v",
			err.Error())
	}

	// TODO, refactor the params design for address
	net, err := detectNetworkForAddress(addr)
	if err != nil {
		return nil, ErrUnknownAddressType
	}

	// TODO, refactor the params design for address
	switch netID {
	case net.PubKeyAddrID:
		return NewPubKeyAddress(decoded, net)

	case net.PubKeyHashAddrID:
		return NewPubKeyHashAddress(decoded, net, ecc.ECDSA_Secp256k1)

	case net.PKHEdwardsAddrID:
		return NewPubKeyHashAddress(decoded, net, ecc.EdDSA_Ed25519)

	case net.PKHSchnorrAddrID:
		return NewPubKeyHashAddress(decoded, net, ecc.ECDSA_SecpSchnorr)

	case net.ScriptHashAddrID:
		return NewScriptHashAddressFromHash(decoded, net)

	default:
		return nil, ErrUnknownAddressType
	}
}

// TODO, refactor the params design for address
// detectNetworkForAddress pops the first character from a string encoded
// address and detects what network type it is for.
func detectNetworkForAddress(addr string) (*params.Params, error) {
	if len(addr) < 1 {
		return nil, fmt.Errorf("empty string given for network detection")
	}

	networkChar := addr[0:1]
	switch networkChar {
	case params.MainNetParams.NetworkAddressPrefix:
		return &params.MainNetParams, nil
	case params.TestNetParams.NetworkAddressPrefix:
		return &params.TestNetParams, nil
	case params.PrivNetParams.NetworkAddressPrefix:
		return &params.PrivNetParams, nil
	case params.MixNetParam.NetworkAddressPrefix:
		return &params.MixNetParams, nil
	}

	return nil, fmt.Errorf("unknown network type in string encoded address")
}
