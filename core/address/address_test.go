// Copyright (c) 2013, 2014 The btcsuite developers
// Copyright (c) 2015-2019 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package address

import (
	// "fmt"
	"reflect"
	"bytes"
	"qitmeer/common/encode/base58"
	"qitmeer/core/types"
	"qitmeer/params"
	"qitmeer/crypto/ecc"
	// "fmt"
	// "qitmeer/params"
	"encoding/hex"
	// "qitmeer/common/encode/base58"
	"golang.org/x/crypto/ripemd160"
	// "qitmeer/common/hash"
	"testing"
)

func TestAddress(t *testing.T){
	mainNetParams := &params.MainNetParams
	testNetParams := &params.TestNetParams
	privNetParams := &params.PrivNetParams
	tests:=[]struct{
		name string
		addr  string
		pubkeystr string
		encoded string
		// verify that result and function results are consistent
		valid bool
		result types.Address
		f   func() (types.Address,error)
		net *params.Params
	}{
		{
			// prikey dbee90783adebfa5250af47e4dd2a0ccdd9b01b3d3d537a1fd916a39dd2d2fed
			name:"privNet p2pkh NewPubKeyHashAddress",
			addr :"RmCYoUMqKZopUkai2YhUFHR9UeqjeyjTAgW",
			pubkeystr:"023e4cc6f9ca030345a9e4e6f3859eae50169d6336001989f11ca71c67fd499541",
			encoded:"RmCYoUMqKZopUkai2YhUFHR9UeqjeyjTAgW",
			valid :true,
			result:&PubKeyHashAddress{
				net:privNetParams,
				netID:privNetParams.PubKeyHashAddrID,
				hash:[ripemd160.Size]byte{
					0x44,0xd9,0x59,0xaf,0xb6,0xdb,0x4a,0xd7,0x30,0xa6,
					0xe2,0xc0,0xda,0xf4,0x6c,0xee,0xb9,0x8c,0x53,0xa0},
			},
			f:func()(types.Address,error){
				pushData :=[]byte{
					0x44,0xd9,0x59,0xaf,0xb6,0xdb,0x4a,0xd7,0x30,0xa6,
					0xe2,0xc0,0xda,0xf4,0x6c,0xee,0xb9,0x8c,0x53,0xa0}
				return NewPubKeyHashAddress(pushData,privNetParams,ecc.ECDSA_Secp256k1)
			},
			net:privNetParams,
		},
		{
			// prikey dbee90783adebfa5250af47e4dd2a0dbdd9b01b3d3d537a1fd916a39dd2d2fed
			name:"mainNet p2pkh NewPubKeyHashAddress",
			addr :"Nmd9qorkwJuKUw3a3yGmD9h35Hk97bnC1Ea",
			pubkeystr:"038fa26a090ae75759fd8f23a6cd9110d229b962a9cc80d8cebb30b863f12c94df",
			encoded:"Nmd9qorkwJuKUw3a3yGmD9h35Hk97bnC1Ea",
			valid :true,
			result:&PubKeyHashAddress{
				net:mainNetParams,
				netID:mainNetParams.PubKeyHashAddrID,
				hash:[ripemd160.Size]byte{
					0xe5,0x27,0x77,0xd3,0xcc,0x8c,0x18,0xe5,0x0a,0x02,
					0xe8,0x84,0x3a,0xd2,0xe6,0xb9,0xd8,0x61,0xa6,0x3f},
			},
			f:func()(types.Address,error){
				pushData :=[]byte{
					0xe5,0x27,0x77,0xd3,0xcc,0x8c,0x18,0xe5,0x0a,0x02,
					0xe8,0x84,0x3a,0xd2,0xe6,0xb9,0xd8,0x61,0xa6,0x3f}
				return NewPubKeyHashAddress(pushData,mainNetParams,ecc.ECDSA_Secp256k1)
			},
			net:mainNetParams,
		},
		{
			// prikey dbee90783adebfa5250af37e4dd2a0dbdd9b01b3d3d537a1fd916a39dd2d2fed
			name:"testNet p2pkh NewPubKeyHashAddress",
			addr :"TmghL12FzxT8bXMpnNjjwmNQ4hTa9LWK5sH",
			pubkeystr:"035f25055b418fc2ec3be35ff34e0e4bd5aff7e49bfa5b6d99642e4a11ed866cc2",
			encoded:"TmghL12FzxT8bXMpnNjjwmNQ4hTa9LWK5sH",
			valid :true,
			result:&PubKeyHashAddress{
				net:testNetParams,
				netID:testNetParams.PubKeyHashAddrID,
				hash:[ripemd160.Size]byte{
					0xc2,0xa7,0xf5,0x10,0x36,0x75,0x28,0x14,0xdd,0xd0,
					0xb0,0x44,0x9b,0x50,0xe3,0xbf,0xa1,0x09,0x89,0x1f},
			},
			f:func()(types.Address,error){
				pushData :=[]byte{
					0xc2,0xa7,0xf5,0x10,0x36,0x75,0x28,0x14,0xdd,0xd0,
					0xb0,0x44,0x9b,0x50,0xe3,0xbf,0xa1,0x09,0x89,0x1f}
				return NewPubKeyHashAddress(pushData,testNetParams,ecc.ECDSA_Secp256k1)
			},
			net:testNetParams,
		},
		{
			// p2pkh wrong hash length
			// prikey dbee90783adebfa5250af37e4dd2a0dbdd9b01b3d3d537a1fd916a39dd2d2fed
			name:"testNet p2pkh wrong hash length NewPubKeyHashAddress",
			pubkeystr:"035f25055b418fc2ec3be35ff34e0e4bd5aff7e49bfa5b6d99642e4a11ed866cc2",
			encoded:"TmghL12FzxT8bXMpnNjjwmNQ4hTa9LWK5sH",
			valid :false,
			f:func()(types.Address,error){
				pushData :=[]byte{
					0xc2,0xa7,0xf5,0x10,0x36,0x75,0x28,0x14,0xdd,0xd0,
					0xb0,0x44,0x9b,0x50,0xe3,0xbf,0xa1,0x09,0x89,0x1f,0x12}
				return NewPubKeyHashAddress(pushData,testNetParams,ecc.ECDSA_Secp256k1)
			},
			net:testNetParams,
		},
		//  p2sh address
		{
			name:"privNet p2sh NewAddressScriptHashFromHash",
			addr :"RSNJopyNcHBnzGt8WHHBfnz8Mxob85fVAGG",
			encoded:"RSNJopyNcHBnzGt8WHHBfnz8Mxob85fVAGG",
			valid :true,
			result:&ScriptHashAddress{
				net:privNetParams,
				netID:privNetParams.ScriptHashAddrID,
				hash:[ripemd160.Size]byte{
					0xcd,0x91,0xf0,0xec,0xd6,0xa3,0x61,0xb1,0x0c,0x26,
					0x6b,0x91,0xbb,0xcc,0x2a,0x96,0xb2,0x08,0x7a,0xa8},
			},
			f:func()(types.Address,error){
				pushData :=[]byte{
					0xcd,0x91,0xf0,0xec,0xd6,0xa3,0x61,0xb1,0x0c,0x26,
					0x6b,0x91,0xbb,0xcc,0x2a,0x96,0xb2,0x08,0x7a,0xa8}
				return NewAddressScriptHashFromHash(pushData,privNetParams)
			},
			net:privNetParams,
		},
		{
			name:"mainNet p2sh NewAddressScriptHashFromHash",
			addr :"NSYJEU4c9ZFcUydu5i3HzJpYW4tHczukQPz",
			encoded:"NSYJEU4c9ZFcUydu5i3HzJpYW4tHczukQPz",
			valid :true,
			result:&ScriptHashAddress{
				net:mainNetParams,
				netID:mainNetParams.ScriptHashAddrID,
				hash:[ripemd160.Size]byte{
					0xcd,0x91,0xf0,0xec,0xd6,0xa3,0x61,0xb1,0x0c,0x26,
					0x6b,0x91,0xbb,0xcc,0x2a,0x96,0xb2,0x08,0x7a,0xa8},
			},
			f:func()(types.Address,error){
				pushData :=[]byte{
					0xcd,0x91,0xf0,0xec,0xd6,0xa3,0x61,0xb1,0x0c,0x26,
					0x6b,0x91,0xbb,0xcc,0x2a,0x96,0xb2,0x08,0x7a,0xa8}
				return NewAddressScriptHashFromHash(pushData,mainNetParams)
			},
			net:mainNetParams,
		},
		{
			name:"testNet p2sh NewAddressScriptHashFromHash",
			addr :"TSFeXQFDam9FKoiHTL77TT6WwE5nnviGr8k",
			encoded:"TSFeXQFDam9FKoiHTL77TT6WwE5nnviGr8k",
			valid :true,
			result:&ScriptHashAddress{
				net:testNetParams,
				netID:testNetParams.ScriptHashAddrID,
				hash:[ripemd160.Size]byte{
					0xcd,0x91,0xf0,0xec,0xd6,0xa3,0x61,0xb1,0x0c,0x26,
					0x6b,0x91,0xbb,0xcc,0x2a,0x96,0xb2,0x08,0x7a,0xa8},
			},
			f:func()(types.Address,error){
				pushData :=[]byte{
					0xcd,0x91,0xf0,0xec,0xd6,0xa3,0x61,0xb1,0x0c,0x26,
					0x6b,0x91,0xbb,0xcc,0x2a,0x96,0xb2,0x08,0x7a,0xa8}
				return NewAddressScriptHashFromHash(pushData,testNetParams)
			},
			net:testNetParams,
		},
		
		
	}
	for _,test:= range tests{
		decoded,err:=DecodeAddress(test.addr)
		if(err==nil) !=test.valid{
			t.Errorf("%v:decoding test failed: %v",test.name,err)
			return
		}
		if err== nil{
			// Ensure the stringer returns the same address as the
			// original.
			decodeStr := decoded.String()
			if test.addr !=decodeStr {
				t.Errorf("%v: String on decoded value does not match expected value: %v != %v",
					test.name, test.addr, decodeStr)
				return
			}
			encode:=decoded.Encode()
			if test.encoded !=encode{
				t.Errorf("%v: encode value does not match expected value: %v != %v",
						test.name, test.encoded, decoded.Encode())
					return
			}
			// base:=base58.Decode(encode)
			var saddr []byte
			switch d:=decoded.(type){
				case *PubKeyHashAddress:
					decoded :=base58.Decode(encode)
					saddr=decoded[2:2+ripemd160.Size]
				case *ScriptHashAddress:
					decoded :=base58.Decode(encode)
					saddr =decoded[2:2+ripemd160.Size]
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
			// Check script address, as well as the Hash160 method for P2PKH and
			// P2SH addresses.
			if !bytes.Equal(saddr, decoded.ScriptAddress()) {
				t.Errorf("%v: script addresses do not match:\n%x != \n%x",
					test.name, saddr, decoded.ScriptAddress())
				return
			}
			switch a := decoded.(type) {
			case *PubKeyHashAddress:
				if h := a.Hash160()[:]; !bytes.Equal(saddr, h) {
					t.Errorf("%v: hashes do not match:\n%x != \n%x",
						test.name, saddr, h)
					return
				}
				if !bytes.Equal(a.netID[:],test.net.PubKeyHashAddrID[:]) {
					t.Errorf("%v: calculated network does not match expected",test.name)
					return 
				}

			case *ScriptHashAddress:
				if h := a.Hash160()[:]; !bytes.Equal(saddr, h) {
					t.Errorf("%v: hashes do not match:\n%x != \n%x",
						test.name, saddr, h)
					return
				}
				if !bytes.Equal(a.netID[:],test.net.ScriptHashAddrID[:]) {
					t.Errorf("%v: calculated network does not match expected",test.name)
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
						test.name,ad.String())
				}else{
					t.Errorf("%v:creating new address failed with error: %v",
				test.name, err)
				}
			}
			continue
		}
		addr,err:=test.f()
		if err != nil {
			t.Errorf("%v: address is valid but creating new address failed with error %v",
				test.name, err)
			return
		}
		if !reflect.DeepEqual(addr.ScriptAddress(), test.result.ScriptAddress()) {
			t.Errorf("%v: created address does not match expected result \n "+
				"	got %x, expected %x",
				test.name, addr.ScriptAddress(), test.result.ScriptAddress())
			return
		}
	}
}