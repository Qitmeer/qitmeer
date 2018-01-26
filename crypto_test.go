// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package crypto

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/dindinw/dagproject/common"
)

var testAddrHex = "970e8128ab834e8eac17ab8e3812f010678cf791"
var testPrivHex = "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032"

// These tests are sanity checks.
// They should ensure that we don't e.g. use Sha3-224 instead of Sha3-256
// and that the sha3 library uses keccak-f permutation.
func TestKeccak256Hash(t *testing.T) {
	msg := []byte("abc")
	exp, _ := hex.DecodeString("4e03657aea45a94fc7d47ba826c8d667c0d1e6e33a64a036ec44f58fa12d6c45")
	checkhash(t, "Sha3-256-array", func(in []byte) []byte { h := Keccak256Hash(in); return h[:] }, msg, exp)
}

func TestEthKeccak256Hash(t *testing.T) {
	msg := []byte("testing")
	exp, _ := hex.DecodeString("5f16f4c7f149ac4f9510d9cf8cf384038ad348b3bcdc01915f95de12df9d1b02")
	checkhash(t, "Sha3-256-keccak", func(in []byte) []byte { h := Keccak256Hash(in); return h[:] }, msg, exp)
}

func TestSha3256Hash(t *testing.T) { //SHA3 NIST Standard
	msg := []byte("testing")
	exp, _ := hex.DecodeString("7f5979fb78f082e8b1c676635db8795c4ac6faba03525fb708cb5fd68fd40c5e")
	checkhash(t, "Sha3-256-NIST", func(in []byte) []byte { h := Sha3Nist256Hash(in); return h[:] }, msg, exp)
}

/*
    key, err := crypto.HexToECDSA("000000...")
	returns nil key and an error instead of a non-nil key with nil X and Y inside.
 */

func TestToECDSAErrors(t *testing.T) {
	if _, err := HexToECDSA("0000000000000000000000000000000000000000000000000000000000000000"); err == nil {
		t.Fatal("HexToECDSA should've returned error")
	}
	if _, err := HexToECDSA("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"); err == nil {
		t.Fatal("HexToECDSA should've returned error")
	}
}

/*
	Nearly every 256-bit number is a valid ECDSA private key.
	Specifically, any 256-bit number from 0x1 to 0xFFFF FFFF FFFF FFFF FFFF FFFF FFFF FFFE BAAE DCE6 AF48 A03B BFD2 5E8C D036 4140
	is a valid private key. The range of valid private keys is governed by the secp256k1 ECDSA standard.
	see also : http://www.secg.org/sec2-v2.pdf
	2.4.1 Recommended Parameters secp256k1
 */

func TestToECDSAInputRange(t *testing.T){
	badPrivateKeys := []string{
		"0000000000000000000000000000000000000000000000000000000000000000", //not included

		"FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141", //same with 0x0, not included
		"FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364142", //same with 0x1, not included
		"FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364143", //same with 0x2, not included
		"fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe", //same with 0x0000...bebd, not included
		"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", //same with 0x0000...bebe, not included
	}
	specialPrivateKeys := []string{
		"0000000000000000000000000000000000000000000000000000000000000001", //same with 0xffff...ebaa...4142
		"0000000000000000000000000000000000000000000000000000000000000002",
		"0000000000000000000000000000000000000000000000000000000000000003",
		//000000000000000000000000000000014551231950b75fc4402da1732fc9bebe
		// 	= FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF
		// 		- FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364140 - 1
		//000000000000000000000000000000014551231950b75fc4402da1732fc9bebe
		//  = FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF
		//      mod (FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364140+1)
		"000000000000000000000000000000014551231950b75fc4402da1732fc9bebd", //same with 0xffff...e
		"000000000000000000000000000000014551231950b75fc4402da1732fc9bebe", //same with 0xffff...f
		"000000000000000000000000000000014551231950b75fc4402da1732fc9bebf",
		// n=2^256-0x14551231950b75fc4402da1732fc9bebf   [1,n-1]
		"FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364140", // 1x ~ 0xffff....EBAAEDCE6AF48A03BBFD25E8CD0364140 is valid
	}
	for _, hex := range badPrivateKeys {
		_,err := HexToECDSA(hex)
		if  err == nil  {
			t.Errorf("%s should return error",hex)
		}
	}
	for _, hex := range specialPrivateKeys{
		_,err := HexToECDSA(hex)
		if err != nil {
			t.Errorf("%s should not return error ",hex)
		}
	}
}



func BenchmarkSha3(b *testing.B) {
	a := []byte("hello world")
	for i := 0; i < b.N; i++ {
		Keccak256(a)
	}
}

// Blake2b
// input : abc
// BLAKE2 224-bit: 9bd237b02a29e43bdd6738afa5b53ff0eee178d6210b618e4511aec8
// BLAKE2 256-bit: bddd813c634239723171ef3fee98579b94964e3bb1cb3e427262c8c068d52319
// BLAKE2 384-bit: 6f56a82c8e7ef526dfe182eb5212f7db9df1317e57815dbda46083fc30f54ee6c66ba83be64b302d7cba6ce15bb556f4
// BLAKE2 512-bit: ba80a53f981c4d0d6a2797b69f12f6e94c212f14685ac4b74b12bb6fdbffa2d17d87c5392aab792dc252d5de4533cc9518d38aa8dbf1925ab92386edd4009923
func TestBlake2b(t *testing.T) {
	msg := []byte("abc")
	exp_256, _ := hex.DecodeString("bddd813c634239723171ef3fee98579b94964e3bb1cb3e427262c8c068d52319")
	exp_512, _ := hex.DecodeString("ba80a53f981c4d0d6a2797b69f12f6e94c212f14685ac4b74b12bb6fdbffa2d17d87c5392aab792dc252d5de4533cc9518d38aa8dbf1925ab92386edd4009923")
	checkhash(t, "blake2b-256-array1", func(in []byte) []byte { h := Blake2b256(in); return h[:] }, msg, exp_256)
	checkhash(t, "blake2b-256-array2", func(in []byte) []byte { h := Blake2b256Hash(in); return h[:] }, msg, exp_256)
	checkhash(t, "blake2b-512-array", func(in []byte) []byte { h := Blake2b512(in); return h[:] }, msg, exp_512)
}

func BenchmarkBlake2b(b *testing.B) {
	a := []byte("hello world")
	amount := 1000000
	start := time.Now()
	for i := 0; i < amount; i++ {
		Blake2b256(a)
	}

	fmt.Println(amount, ":", time.Since(start))
}

func TestSign(t *testing.T) {
	key, _ := HexToECDSA(testPrivHex)
	addr := common.HexToEthAddress(testAddrHex)

	msg := Keccak256([]byte("foo"))
	sig, err := Sign(msg, key)
	if err != nil {
		t.Errorf("Sign error: %s", err)
	}
	recoveredPub, err := Ecrecover(msg, sig)
	if err != nil {
		t.Errorf("ECRecover error: %s", err)
	}
	pubKey := ToECDSAPub(recoveredPub)
	recoveredAddr := PubkeyToEthAddress(*pubKey)
	if addr != recoveredAddr {
		t.Errorf("Address mismatch: want: %x have: %x", addr, recoveredAddr)
	}

	// should be equal to SigToPub
	recoveredPub2, err := SigToPub(msg, sig)
	if err != nil {
		t.Errorf("ECRecover error: %s", err)
	}
	recoveredAddr2 := PubkeyToEthAddress(*recoveredPub2)
	if addr != recoveredAddr2 {
		t.Errorf("Address mismatch: want: %x have: %x", addr, recoveredAddr2)
	}
}

func TestInvalidSign(t *testing.T) {
	if _, err := Sign(make([]byte, 1), nil); err == nil {
		t.Errorf("expected sign with hash 1 byte to error")
	}
	if _, err := Sign(make([]byte, 33), nil); err == nil {
		t.Errorf("expected sign with hash 33 byte to error")
	}
}

func TestNewContractAddress(t *testing.T) {
	key, _ := HexToECDSA(testPrivHex)
	addr := common.HexToEthAddress(testAddrHex)
	genAddr := PubkeyToEthAddress(key.PublicKey)
	// sanity check before using addr to create contract address
	checkAddr(t, genAddr, addr)

	caddr0 := CreateEthContractAddress(addr, 0)
	caddr1 := CreateEthContractAddress(addr, 1)
	caddr2 := CreateEthContractAddress(addr, 2)
	checkAddr(t, common.HexToEthAddress("333c3310824b7c685133f2bedb2ca4b8b4df633d"), caddr0)
	checkAddr(t, common.HexToEthAddress("8bda78331c916a08481428e4b07c96d3e916d165"), caddr1)
	checkAddr(t, common.HexToEthAddress("c9ddedf451bc62ce88bf9292afb13df35b670699"), caddr2)
}

func TestLoadECDSAFile(t *testing.T) {
	keyBytes := common.FromHex(testPrivHex)
	fileName0 := "test_key0"
	fileName1 := "test_key1"
	checkKey := func(k *ecdsa.PrivateKey) {
		checkAddr(t, PubkeyToEthAddress(k.PublicKey), common.HexToEthAddress(testAddrHex))
		loadedKeyBytes := FromECDSA(k)
		if !bytes.Equal(loadedKeyBytes, keyBytes) {
			t.Fatalf("private key mismatch: want: %x have: %x", keyBytes, loadedKeyBytes)
		}
	}

	ioutil.WriteFile(fileName0, []byte(testPrivHex), 0600)
	defer os.Remove(fileName0)

	key0, err := LoadECDSA(fileName0)
	if err != nil {
		t.Fatal(err)
	}
	checkKey(key0)

	// again, this time with SaveECDSA instead of manual save:
	err = SaveECDSA(fileName1, key0)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(fileName1)

	key1, err := LoadECDSA(fileName1)
	if err != nil {
		t.Fatal(err)
	}
	checkKey(key1)
}

func TestValidateSignatureValues(t *testing.T) {
	check := func(expected bool, v byte, r, s *big.Int) {
		if ValidateSignatureValues(v, r, s, false) != expected {
			t.Errorf("mismatch for v: %d r: %d s: %d want: %v", v, r, s, expected)
		}
	}
	minusOne := big.NewInt(-1)
	one := common.Big1
	zero := common.Big0
	secp256k1nMinus1 := new(big.Int).Sub(secp256k1_N, common.Big1)

	// correct v,r,s
	check(true, 0, one, one)
	check(true, 1, one, one)
	// incorrect v, correct r,s,
	check(false, 2, one, one)
	check(false, 3, one, one)

	// incorrect v, combinations of incorrect/correct r,s at lower limit
	check(false, 2, zero, zero)
	check(false, 2, zero, one)
	check(false, 2, one, zero)
	check(false, 2, one, one)

	// correct v for any combination of incorrect r,s
	check(false, 0, zero, zero)
	check(false, 0, zero, one)
	check(false, 0, one, zero)

	check(false, 1, zero, zero)
	check(false, 1, zero, one)
	check(false, 1, one, zero)

	// correct sig with max r,s
	check(true, 0, secp256k1nMinus1, secp256k1nMinus1)
	// correct v, combinations of incorrect r,s at upper limit
	check(false, 0, secp256k1_N, secp256k1nMinus1)
	check(false, 0, secp256k1nMinus1, secp256k1_N)
	check(false, 0, secp256k1_N, secp256k1_N)

	// current callers ensures r,s cannot be negative, but let's test for that too
	// as crypto package could be used stand-alone
	check(false, 0, minusOne, one)
	check(false, 0, one, minusOne)
}

func checkhash(t *testing.T, name string, f func([]byte) []byte, msg, exp []byte) {
	sum := f(msg)
	if !bytes.Equal(exp, sum) {
		t.Fatalf("hash %s mismatch: want: %x have: %x", name, exp, sum)
	}
}

func checkAddr(t *testing.T, addr0, addr1 common.EthAddress) {
	if addr0 != addr1 {
		t.Fatalf("address mismatch: want: %x have: %x", addr0, addr1)
	}
}

// test to help Python team with integration of libsecp256k1
// skip but keep it after they are done
func TestPythonIntegration(t *testing.T) {
	kh := "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032"
	k0, _ := HexToECDSA(kh)

	msg0 := Keccak256([]byte("foo"))
	sig0, _ := Sign(msg0, k0)

	msg1 := common.FromHex("00000000000000000000000000000000")
	sig1, _ := Sign(msg0, k0)

	t.Logf("msg: %x, privkey: %s sig: %x\n", msg0, kh, sig0)
	t.Logf("msg: %x, privkey: %s sig: %x\n", msg1, kh, sig1)
}
