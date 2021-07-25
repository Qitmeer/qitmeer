// Copyright (c) 2013, 2014 The btcsuite developers
// Copyright (c) 2015-2019 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package address

import (
	"bytes"
	"github.com/Qitmeer/qitmeer/common/encode/base58"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/crypto/ecc"
	"github.com/Qitmeer/qitmeer/params"
	// "fmt"
	"reflect"
	// "fmt"
	// "qitmeer/params"
	"encoding/hex"
	// "qitmeer/common/encode/base58"
	"golang.org/x/crypto/ripemd160"
	// "qitmeer/common/hash"
	"testing"
)

func TestAddress(t *testing.T) {
	mainNetParams := &params.MainNetParams
	testNetParams := &params.TestNetParams
	privNetParams := &params.PrivNetParams
	mixNetParams := &params.MixNetParams
	tests := []struct {
		name      string
		addr      string
		pubkeystr string
		encoded   string
		// verify that result and function results are consistent
		valid  bool
		result types.Address
		f      func() (types.Address, error)
		net    *params.Params
	}{
		{
			// prikey 7e445aa5ffd834cb2d3b2db50f8997dd21af29bec3d296aaa066d902b93f484b
			name:      "privNet p2pkh NewPubKeyHashAddress",
			addr:      "XmNZ52HdwsBWUK4TAhyXfYhZ1NcJ9HC88oh",
			pubkeystr: "0354455a60d86273d322eebb913d87f428988ce97922a366f0a0867a426df78bc9",
			encoded:   "XmNZ52HdwsBWUK4TAhyXfYhZ1NcJ9HC88oh",
			valid:     true,
			result: &PubKeyHashAddress{
				net:   mixNetParams,
				netID: mixNetParams.PubKeyHashAddrID,
				hash: [ripemd160.Size]byte{
					0x8d, 0xc2, 0x68, 0xa8, 0xe2, 0x87, 0x6b, 0x94, 0x1b, 0x95,
					0xf3, 0xbd, 0x3b, 0x71, 0x05, 0x0f, 0x00, 0x3c, 0x53, 0x83},
			},
			f: func() (types.Address, error) {
				pushData := []byte{
					0x8d, 0xc2, 0x68, 0xa8, 0xe2, 0x87, 0x6b, 0x94, 0x1b, 0x95,
					0xf3, 0xbd, 0x3b, 0x71, 0x05, 0x0f, 0x00, 0x3c, 0x53, 0x83}
				return NewPubKeyHashAddress(pushData, mixNetParams, ecc.ECDSA_Secp256k1)
			},
			net: mixNetParams,
		},
		{
			// prikey 7e445aa5ffd834cb2d3b2db50f8997dd21af29bec3d296aaa066d902b93f484b
			name:      "privNet p2pkh NewPubKeyHashAddress",
			addr:      "RmKCKKMoji8gau8rEY5Qz6F7VxGSCL4bSWV",
			pubkeystr: "0354455a60d86273d322eebb913d87f428988ce97922a366f0a0867a426df78bc9",
			encoded:   "RmKCKKMoji8gau8rEY5Qz6F7VxGSCL4bSWV",
			valid:     true,
			result: &PubKeyHashAddress{
				net:   privNetParams,
				netID: privNetParams.PubKeyHashAddrID,
				hash: [ripemd160.Size]byte{
					0x8d, 0xc2, 0x68, 0xa8, 0xe2, 0x87, 0x6b, 0x94, 0x1b, 0x95,
					0xf3, 0xbd, 0x3b, 0x71, 0x05, 0x0f, 0x00, 0x3c, 0x53, 0x83},
			},
			f: func() (types.Address, error) {
				pushData := []byte{
					0x8d, 0xc2, 0x68, 0xa8, 0xe2, 0x87, 0x6b, 0x94, 0x1b, 0x95,
					0xf3, 0xbd, 0x3b, 0x71, 0x05, 0x0f, 0x00, 0x3c, 0x53, 0x83}
				return NewPubKeyHashAddress(pushData, privNetParams, ecc.ECDSA_Secp256k1)
			},
			net: privNetParams,
		},
		{
			// prikey 7e445aa5ffd834cb2d3b2db50f8997dd21af29bec3d296aaa066d902b93f484b
			name:      "mainNet p2pkh NewPubKeyHashAddress",
			addr:      "MmYWtAp7nkDmuqUYLSRZQn2LMRi2rsAdJLr",
			pubkeystr: "0354455a60d86273d322eebb913d87f428988ce97922a366f0a0867a426df78bc9",
			encoded:   "MmYWtAp7nkDmuqUYLSRZQn2LMRi2rsAdJLr",
			valid:     true,
			result: &PubKeyHashAddress{
				net:   mainNetParams,
				netID: mainNetParams.PubKeyHashAddrID,
				hash: [ripemd160.Size]byte{
					0x8d, 0xc2, 0x68, 0xa8, 0xe2, 0x87, 0x6b, 0x94, 0x1b, 0x95,
					0xf3, 0xbd, 0x3b, 0x71, 0x05, 0x0f, 0x00, 0x3c, 0x53, 0x83},
			},
			f: func() (types.Address, error) {
				pushData := []byte{
					0x8d, 0xc2, 0x68, 0xa8, 0xe2, 0x87, 0x6b, 0x94, 0x1b, 0x95,
					0xf3, 0xbd, 0x3b, 0x71, 0x05, 0x0f, 0x00, 0x3c, 0x53, 0x83}
				return NewPubKeyHashAddress(pushData, mainNetParams, ecc.ECDSA_Secp256k1)
			},
			net: mainNetParams,
		},
		{
			// prikey 7e445aa5ffd834cb2d3b2db50f8997dd21af29bec3d296aaa066d902b93f484b
			name:      "testNet p2pkh NewPubKeyHashAddress",
			addr:      "TnQYqqxYaKdXYtGRSfALjCjKS74QhEXaQZp",
			pubkeystr: "0354455a60d86273d322eebb913d87f428988ce97922a366f0a0867a426df78bc9",
			encoded:   "TnQYqqxYaKdXYtGRSfALjCjKS74QhEXaQZp",
			valid:     true,
			result: &PubKeyHashAddress{
				net:   testNetParams,
				netID: testNetParams.PubKeyHashAddrID,
				hash: [ripemd160.Size]byte{
					0x8d, 0xc2, 0x68, 0xa8, 0xe2, 0x87, 0x6b, 0x94, 0x1b, 0x95,
					0xf3, 0xbd, 0x3b, 0x71, 0x05, 0x0f, 0x00, 0x3c, 0x53, 0x83},
			},
			f: func() (types.Address, error) {
				pushData := []byte{
					0x8d, 0xc2, 0x68, 0xa8, 0xe2, 0x87, 0x6b, 0x94, 0x1b, 0x95,
					0xf3, 0xbd, 0x3b, 0x71, 0x05, 0x0f, 0x00, 0x3c, 0x53, 0x83}
				return NewPubKeyHashAddress(pushData, testNetParams, ecc.ECDSA_Secp256k1)
			},
			net: testNetParams,
		},
		{
			// p2pkh wrong hash length
			// prikey 7e445aa5ffd834cb2d3b2db50f8997dd21af29bec3d296aaa066d902b93f484b
			name:      "testNet p2pkh wrong hash length NewPubKeyHashAddress",
			pubkeystr: "035f25055b418fc2ec3be35ff34e0e4bd5aff7e49bfa5b6d99642e4a11ed866cc2",
			encoded:   "TmghL12FzxT8bXMpnNjjwmNQ4hTa9LWK5sH",
			valid:     false,
			f: func() (types.Address, error) {
				pushData := []byte{
					0x8d, 0xc2, 0x68, 0xa8, 0xe2, 0x87, 0x6b, 0x94, 0x1b, 0x95,
					0xf3, 0xbd, 0x3b, 0x71, 0x05, 0x0f, 0x00, 0x3c, 0x53, 0x83, 0x12}
				return NewPubKeyHashAddress(pushData, testNetParams, ecc.ECDSA_Secp256k1)
			},
			net: testNetParams,
		},
		//  p2sh address
		{
			name:    "privNet p2sh NewAddressScriptHashFromHash",
			addr:    "RSNJopyNcHBnzGt8WHHBfnz8Mxob85fVAGG",
			encoded: "RSNJopyNcHBnzGt8WHHBfnz8Mxob85fVAGG",
			valid:   true,
			result: &ScriptHashAddress{
				net:   privNetParams,
				netID: privNetParams.ScriptHashAddrID,
				hash: [ripemd160.Size]byte{
					0x77, 0xca, 0x77, 0xb8, 0x27, 0x72, 0xbb, 0xf6, 0x86, 0x27,
					0xe5, 0x00, 0x44, 0xa0, 0x82, 0x3e, 0xa9, 0xaf, 0x4f, 0x30},
			},
			f: func() (types.Address, error) {
				pushData := []byte{
					0x77, 0xca, 0x77, 0xb8, 0x27, 0x72, 0xbb, 0xf6, 0x86, 0x27,
					0xe5, 0x00, 0x44, 0xa0, 0x82, 0x3e, 0xa9, 0xaf, 0x4f, 0x30}
				return NewScriptHashAddressFromHash(pushData, privNetParams)
			},
			net: privNetParams,
		},
		{
			name:    "mainNet p2sh NewAddressScriptHashFromHash",
			addr:    "MSCHmhKPNc6RSPngXACzmze5RojvqusFGeN",
			encoded: "MSCHmhKPNc6RSPngXACzmze5RojvqusFGeN",
			valid:   true,
			result: &ScriptHashAddress{
				net:   mainNetParams,
				netID: mainNetParams.ScriptHashAddrID,
				hash: [ripemd160.Size]byte{
					0x77, 0xca, 0x77, 0xb8, 0x27, 0x72, 0xbb, 0xf6, 0x86, 0x27,
					0xe5, 0x00, 0x44, 0xa0, 0x82, 0x3e, 0xa9, 0xaf, 0x4f, 0x30},
			},
			f: func() (types.Address, error) {
				pushData := []byte{
					0x77, 0xca, 0x77, 0xb8, 0x27, 0x72, 0xbb, 0xf6, 0x86, 0x27,
					0xe5, 0x00, 0x44, 0xa0, 0x82, 0x3e, 0xa9, 0xaf, 0x4f, 0x30}
				return NewScriptHashAddressFromHash(pushData, mainNetParams)
			},
			net: mainNetParams,
		},
		{
			name:    "testNet p2sh NewAddressScriptHashFromHash",
			addr:    "TSFeXQFDam9FKoiHTL77TT6WwE5nnviGr8k",
			encoded: "TSFeXQFDam9FKoiHTL77TT6WwE5nnviGr8k",
			valid:   true,
			result: &ScriptHashAddress{
				net:   testNetParams,
				netID: testNetParams.ScriptHashAddrID,
				hash: [ripemd160.Size]byte{
					0x77, 0xca, 0x77, 0xb8, 0x27, 0x72, 0xbb, 0xf6, 0x86, 0x27,
					0xe5, 0x00, 0x44, 0xa0, 0x82, 0x3e, 0xa9, 0xaf, 0x4f, 0x30},
			},
			f: func() (types.Address, error) {
				pushData := []byte{
					0x77, 0xca, 0x77, 0xb8, 0x27, 0x72, 0xbb, 0xf6, 0x86, 0x27,
					0xe5, 0x00, 0x44, 0xa0, 0x82, 0x3e, 0xa9, 0xaf, 0x4f, 0x30}
				return NewScriptHashAddressFromHash(pushData, testNetParams)
			},
			net: testNetParams,
		},
		{
			name:    "mixNet p2sh NewAddressScriptHashFromHash",
			addr:    "XSRfZXuCpSEcsgojSTBJMFSZsP9T58yB1Z5",
			encoded: "XSRfZXuCpSEcsgojSTBJMFSZsP9T58yB1Z5",
			valid:   true,
			result: &ScriptHashAddress{
				net:   mixNetParams,
				netID: mixNetParams.ScriptHashAddrID,
				hash: [ripemd160.Size]byte{
					0x77, 0xca, 0x77, 0xb8, 0x27, 0x72, 0xbb, 0xf6, 0x86, 0x27,
					0xe5, 0x00, 0x44, 0xa0, 0x82, 0x3e, 0xa9, 0xaf, 0x4f, 0x30},
			},
			f: func() (types.Address, error) {
				pushData := []byte{
					0x77, 0xca, 0x77, 0xb8, 0x27, 0x72, 0xbb, 0xf6, 0x86, 0x27,
					0xe5, 0x00, 0x44, 0xa0, 0x82, 0x3e, 0xa9, 0xaf, 0x4f, 0x30}
				return NewScriptHashAddressFromHash(pushData, mixNetParams)
			},
			net: mixNetParams,
		},
	}
	for _, test := range tests {
		decoded, err := DecodeAddress(test.addr)
		if (err == nil) != test.valid {
			t.Errorf("%v:decoding test failed: %v", test.name, err)
			return
		}
		if err == nil {
			// Ensure the stringer returns the same address as the
			// original.
			decodeStr := decoded.String()
			if test.addr != decodeStr {
				t.Errorf("%v: String on decoded value does not match expected value: %v != %v",
					test.name, test.addr, decodeStr)
				return
			}
			encode := decoded.Encode()
			if test.encoded != encode {
				t.Errorf("%v: encode value does not match expected value: %v != %v",
					test.name, test.encoded, decoded.Encode())
				return
			}
			// base:=base58.Decode(encode)
			var saddr []byte
			switch d := decoded.(type) {
			case *PubKeyHashAddress:
				decoded := base58.Decode([]byte(encode))
				saddr = decoded[2 : 2+ripemd160.Size]
			case *ScriptHashAddress:
				decoded := base58.Decode([]byte(encode))
				saddr = decoded[2 : 2+ripemd160.Size]
			case *SecpPubKeyAddress:
				// Ignore the error here since the script
				// address is checked below.
				saddr, err = hex.DecodeString(d.String())
				if err != nil {
					saddr, _ = hex.DecodeString(test.pubkeystr)
				}
			case *SecSchnorrPubKeyAddress:
				// Ignore the error here since the script
				// address is checked below.
				saddr, _ = hex.DecodeString(d.String())
			case *EdwardsPubKeyAddress:
				// Ignore the error here since the script
				// address is checked below.
				saddr, _ = hex.DecodeString(d.String())
			}
			// Check address script, as well as the Hash160 method for P2PKH and
			// P2SH addresses.
			if !bytes.Equal(saddr, decoded.Script()) {
				t.Errorf("%v: addresses script do not match:\n%x != \n%x",
					test.name, saddr, decoded.Script())
				return
			}
			switch a := decoded.(type) {
			case *PubKeyHashAddress:
				if h := a.Hash160()[:]; !bytes.Equal(saddr, h) {
					t.Errorf("%v: hashes do not match:\n%x != \n%x",
						test.name, saddr, h)
					return
				}
				if !bytes.Equal(a.netID[:], test.net.PubKeyHashAddrID[:]) {
					t.Errorf("%v: calculated network does not match expected", test.name)
					return
				}

			case *ScriptHashAddress:
				if h := a.Hash160()[:]; !bytes.Equal(saddr, h) {
					t.Errorf("%v: hashes do not match:\n%x != \n%x",
						test.name, saddr, h)
					return
				}
				if !bytes.Equal(a.netID[:], test.net.ScriptHashAddrID[:]) {
					t.Errorf("%v: calculated network does not match expected", test.name)
					return
				}
			}
		}

		if !test.valid {
			// If address is invalid, but a creation function exists,
			// verify that it returns a nil addr and non-nil error.
			if test.f != nil {
				ad, err := test.f()
				if err == nil {
					t.Errorf("%v: address is invalid but creating new address succeeded new address:%v",
						test.name, ad.String())
				} else {
					t.Logf("%v:creating new address failed with error: %v",
						test.name, err)
				}
			}
			continue
		}
		addr, err := test.f()
		if err != nil {
			t.Errorf("%v: address is valid but creating new address failed with error %v",
				test.name, err)
			return
		}
		if !reflect.DeepEqual(addr.Script(), test.result.Script()) {
			t.Errorf("%v: created address does not match expected result \n "+
				"	got %x, expected %x",
				test.name, addr.Script(), test.result.Script())
			return
		}
		if addr.String() != test.addr {
			t.Errorf("%v: created address does not match expected result \n "+
				"	got %s, expected %s",
				test.name, addr.String(), test.addr)
			return
		}
	}
}
