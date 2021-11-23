package rpc

import (
	"github.com/Qitmeer/qng-core/core/address"
	"github.com/Qitmeer/qng-core/core/types"
	"github.com/Qitmeer/qng-core/crypto/ecc/secp256k1"
	"golang.org/x/crypto/ripemd160"
	"sync"
)

// wsClientFilter tracks relevant addresses for each websocket client for
// the `rescanblocks` extension. It is modified by the `loadtxfilter` command.
//
// NOTE: This extension was ported from github.com/decred/dcrd
type wsClientFilter struct {
	mu sync.Mutex

	// Implemented fast paths for address lookup.
	pubKeyHashes        map[[ripemd160.Size]byte]struct{}
	scriptHashes        map[[ripemd160.Size]byte]struct{}
	compressedPubKeys   map[[33]byte]struct{}
	uncompressedPubKeys map[[65]byte]struct{}

	// A fallback address lookup map in case a fast path doesn't exist.
	// Only exists for completeness.  If using this shows up in a profile,
	// there's a good chance a fast path should be added.
	otherAddresses map[string]struct{}

	// Outpoints of unspent outputs.
	unspent map[types.TxOutPoint]struct{}
}

type rescanKeys struct {
	addrs   map[string]struct{}
	unspent map[types.TxOutPoint]struct{}
}

// unspentSlice returns a slice of currently-unspent outpoints for the rescan
// lookup keys.  This is primarily intended to be used to register outpoints
// for continuous notifications after a rescan has completed.
func (r *rescanKeys) unspentSlice() []*types.TxOutPoint {
	ops := make([]*types.TxOutPoint, 0, len(r.unspent))
	for op := range r.unspent {
		opCopy := op
		ops = append(ops, &opCopy)
	}
	return ops
}

// newWSClientFilter creates a new, empty wsClientFilter struct to be used
// for a websocket client.
//
// NOTE: This extension was ported from github.com/decred/dcrd
func newWSClientFilter(addresses []string, unspentOutPoints []types.TxOutPoint) *wsClientFilter {
	filter := &wsClientFilter{
		pubKeyHashes:        map[[ripemd160.Size]byte]struct{}{},
		scriptHashes:        map[[ripemd160.Size]byte]struct{}{},
		compressedPubKeys:   map[[33]byte]struct{}{},
		uncompressedPubKeys: map[[65]byte]struct{}{},
		otherAddresses:      map[string]struct{}{},
		unspent:             make(map[types.TxOutPoint]struct{}, len(unspentOutPoints)),
	}

	for _, s := range addresses {
		filter.addAddressStr(s)
	}
	for i := range unspentOutPoints {
		filter.addUnspentOutPoint(&unspentOutPoints[i])
	}

	return filter
}

// addAddress adds an address to a wsClientFilter, treating it correctly based
// on the type of address passed as an argument.
//
// NOTE: This extension was ported from github.com/decred/dcrd
func (f *wsClientFilter) addAddress(a types.Address) {
	switch a := a.(type) {
	case *address.PubKeyHashAddress:
		f.pubKeyHashes[*a.Hash160()] = struct{}{}
		return
	case *address.ScriptHashAddress:
		f.scriptHashes[*a.Hash160()] = struct{}{}
		return
	case *address.SecpPubKeyAddress:
		serializedPubKey := a.Script()
		switch len(serializedPubKey) {
		case secp256k1.PubKeyBytesLenCompressed: // compressed
			var compressedPubKey [secp256k1.PubKeyBytesLenCompressed]byte
			copy(compressedPubKey[:], serializedPubKey)
			f.compressedPubKeys[compressedPubKey] = struct{}{}
			return
		case secp256k1.PubKeyBytesLenUncompressed: // uncompressed
			var uncompressedPubKey [secp256k1.PubKeyBytesLenUncompressed]byte
			copy(uncompressedPubKey[:], serializedPubKey)
			f.uncompressedPubKeys[uncompressedPubKey] = struct{}{}
			return
		}
	}

	f.otherAddresses[a.Encode()] = struct{}{}
}

// addAddressStr parses an address from a string and then adds it to the
// wsClientFilter using addAddress.
//
// NOTE: This extension was ported from github.com/decred/dcrd
func (f *wsClientFilter) addAddressStr(s string) {
	// If address can't be decoded, no point in saving it since it should also
	// impossible to create the address from an inspected transaction output
	// script.
	a, err := address.DecodeAddress(s)
	if err != nil {
		return
	}
	f.addAddress(a)
}

// existsAddress returns true if the passed address has been added to the
// wsClientFilter.
//
// NOTE: This extension was ported from github.com/decred/dcrd
func (f *wsClientFilter) existsAddress(a types.Address) bool {
	switch a := a.(type) {
	case *address.PubKeyHashAddress:
		_, ok := f.pubKeyHashes[*a.Hash160()]
		return ok
	case *address.ScriptHashAddress:
		_, ok := f.scriptHashes[*a.Hash160()]
		return ok
	case *address.SecSchnorrPubKeyAddress:
	case *address.EdwardsPubKeyAddress:
	case *address.SecpPubKeyAddress:
		serializedPubKey := a.Script()
		switch len(serializedPubKey) {
		case secp256k1.PubKeyBytesLenCompressed: // compressed
			var compressedPubKey [secp256k1.PubKeyBytesLenCompressed]byte
			copy(compressedPubKey[:], serializedPubKey)
			_, ok := f.compressedPubKeys[compressedPubKey]
			if !ok {
				_, ok = f.pubKeyHashes[*a.PKHAddress().Hash160()]
			}
			return ok
		case secp256k1.PubKeyBytesLenUncompressed: // uncompressed
			var uncompressedPubKey [secp256k1.PubKeyBytesLenUncompressed]byte
			copy(uncompressedPubKey[:], serializedPubKey)
			_, ok := f.uncompressedPubKeys[uncompressedPubKey]
			if !ok {
				_, ok = f.pubKeyHashes[*a.PKHAddress().Hash160()]
			}
			return ok
		}
	}

	_, ok := f.otherAddresses[a.Encode()]
	return ok
}

// removeAddress removes the passed address, if it exists, from the
// wsClientFilter.
//
// NOTE: This extension was ported from github.com/decred/dcrd
func (f *wsClientFilter) removeAddress(a types.Address) {
	switch a := a.(type) {
	case *address.PubKeyHashAddress:
		delete(f.pubKeyHashes, *a.Hash160())
		return
	case *address.ScriptHashAddress:
		delete(f.scriptHashes, *a.Hash160())
		return
	case *address.SecSchnorrPubKeyAddress:
	case *address.EdwardsPubKeyAddress:
	case *address.SecpPubKeyAddress:
		serializedPubKey := a.Script()
		switch len(serializedPubKey) {
		case secp256k1.PubKeyBytesLenCompressed: // compressed
			var compressedPubKey [secp256k1.PubKeyBytesLenCompressed]byte
			copy(compressedPubKey[:], serializedPubKey)
			delete(f.compressedPubKeys, compressedPubKey)
			return
		case secp256k1.PubKeyBytesLenUncompressed: // uncompressed
			var uncompressedPubKey [secp256k1.PubKeyBytesLenUncompressed]byte
			copy(uncompressedPubKey[:], serializedPubKey)
			delete(f.uncompressedPubKeys, uncompressedPubKey)
			return
		}
	}

	delete(f.otherAddresses, a.Encode())
}

// removeAddressStr parses an address from a string and then removes it from the
// wsClientFilter using removeAddress.
//
// NOTE: This extension was ported from github.com/decred/dcrd
func (f *wsClientFilter) removeAddressStr(s string) {
	a, err := address.DecodeAddress(s)
	if err == nil {
		f.removeAddress(a)
	} else {
		delete(f.otherAddresses, s)
	}
}

// addUnspentOutPoint adds an outpoint to the wsClientFilter.
//
// NOTE: This extension was ported from github.com/decred/dcrd
func (f *wsClientFilter) addUnspentOutPoint(op *types.TxOutPoint) {
	f.unspent[*op] = struct{}{}
}

// existsUnspentOutPoint returns true if the passed outpoint has been added to
// the wsClientFilter.
//
// NOTE: This extension was ported from github.com/decred/dcrd
func (f *wsClientFilter) existsUnspentOutPoint(op *types.TxOutPoint) bool {
	_, ok := f.unspent[*op]
	return ok
}

// removeUnspentOutPoint removes the passed outpoint, if it exists, from the
// wsClientFilter.
//
// NOTE: This extension was ported from github.com/decred/dcrd
func (f *wsClientFilter) removeUnspentOutPoint(op *types.TxOutPoint) {
	delete(f.unspent, *op)
}
