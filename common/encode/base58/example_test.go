// Copyright (c) 2017-2018 The nox developers
// Copyright (c) 2014 The btcsuite developers
// Copyright (c) 2015 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package base58_test

import (
	"fmt"
	"testing"

	"qitmeer/common/encode/base58"
	"encoding/hex"
)



// This example demonstrates how to decode modified base58 encoded data.
func Test_ExampleDecode(t *testing.T) {
	// Decode example modified base58 encoded data.
	encoded := "25JnwSn7XKfNQ"
	decoded := base58.Decode(encoded)

	// Show the decoded data.
	fmt.Println("Decoded Data:", string(decoded))

	// Output:
	// Decoded Data: Test data
}

// This example demonstrates how to encode data using the modified base58
// encoding scheme.
func Test_ExampleEncode(t *testing.T) {
	// Encode example data with the modified base58 encoding scheme.
	data := []byte("Test data")
	encoded := base58.Encode(data)

	// Show the encoded data.
	fmt.Println("Encoded Data:", encoded)

	// Output:
	// Encoded Data: 25JnwSn7XKfNQ
}

// This example demonstrates how to decode Base58Check encoded data.
func Test_ExampleCheckDecodeBtc(t *testing.T) {
	// Decode an example Base58Check encoded data.
	encoded := "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"
	decoded, version, err := base58.BtcCheckDecode(encoded)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Show the decoded data.
	fmt.Printf("Decoded data: %x\n", decoded)
	fmt.Println("Version Byte:", version)

	// Output:
	// Decoded data: 62e907b15cbf27d5425399ebf6f0fb50ebb88f18
	// Version Byte: 0
}

func Test_ExampleCheckEncodeBtc1(t *testing.T) {
	// Encode example data with the Base58Check encoding scheme.
	data,_ := hex.DecodeString("62e907b15cbf27d5425399ebf6f0fb50ebb88f18a")
	encoded := base58.BtcCheckEncode(data, 0x0)

	// Show the encoded data.
	fmt.Println("Encoded Data:", encoded)

	// Output:
	// Encoded Data: 1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa
}

// This example demonstrates how to encode data using the Base58Check encoding
// scheme.
func Test_ExampleCheckEncodeBtc(t *testing.T) {
	// Encode example data with the Base58Check encoding scheme.
	data := []byte("Test data")
	encoded := base58.BtcCheckEncode(data, 0x0)

	// Show the encoded data.
	fmt.Println("Encoded Data:", encoded)

	// Output:
	// Encoded Data: 182iP79GRURMp7oMHDU
}

func Test_ExampleCheckEncodeDcr(t *testing.T) {
	// Encode example data with the Base58Check encoding scheme.
	data := []byte("Test data")
	ver := [2]byte{0x44, 0x0}

	encoded := base58.DcrCheckEncode(data, ver)

	// Show the encoded data.
	fmt.Println("Encoded Data:", encoded)

	// Output:
	// Encoded Data: 2uLtqkeVgFqTUBnjicK8o
}



func Test_ExampleCheckDecodeDcr(t *testing.T) {
	encoded := "2uLtqkeVgFqTUBnjicK8o"
	decoded, version, err := base58.DcrCheckDecode(encoded)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Show the decoded data.
	fmt.Printf("Decoded data: %x\n", decoded)
	fmt.Println("Version Byte:", version)
	// Output:
	// Decoded data: 546573742064617461
	// Version Byte: [68 0]
}

func Test_ExampleCheckDecode_ds_addr(t *testing.T) {
	encoded := "DsaAKsMvZ6HrqhmbhLjV9qVbPkkzF7FnNFY"
	decoded, version, err := base58.NoxCheckDecode(encoded)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Show the decoded data.
	fmt.Printf("Decoded data: %x\n", decoded)
	fmt.Println("Version Byte:", version)
	// Output:
	// Decoded data: 64e20eb6075561d30c23a517c5b73badbc120f05
	// Version Byte: [7 63]
}

func Test_ExampleCheckEncode_addr(t *testing.T) {

	data,_ := hex.DecodeString("64e20eb6075561d30c23a517c5b73badbc120f05")
	ver  := [2]byte{0x0c, 0x40}  //Nox main
	encoded := base58.NoxCheckEncode(data, ver[:])
	fmt.Println("Address (sha256) : Nm281BTkccPTDL1CfhAAR27GAzx2bqFLQx5")
	fmt.Println("Address (b2b)    :",encoded)
	encoded = base58.DcrCheckEncode(data, ver)
	fmt.Println("Address (b256)   :",encoded)
	// Output:
	// Address (sha256) : Nm281BTkccPTDL1CfhAAR27GAzx2bqFLQx5
	// Address (b2b)    : Nm281BTkccPTDL1CfhAAR27GAzx2bnKjZdM
	// Address (b256)   : Nm281BTkccPTDL1CfhAAR27GAzx2br4Aebi
}


func Test_ExampleCheckEncode(t *testing.T) {
	// Encode example data with the Base58Check encoding scheme.
	data := []byte("Test data")
	var ver [2]byte
	ver[0] = 0
	ver[1] = 0

	encoded := base58.NoxCheckEncode(data, ver[:])

	// Show the encoded data.
	fmt.Println("Encoded Data:", encoded)

	// Output:
	// Encoded Data: 1182iP79GRURMp6Rsz9X
}

func Test_ExampleCheckDecode(t *testing.T) {
	encoded := "1182iP79GRURMp6Rsz9X"
	decoded, version, err := base58.NoxCheckDecode(encoded)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Show the decoded data.
	fmt.Printf("Decoded data: %x\n", decoded)
	fmt.Println("Version Byte:", version)
	// Output:
	// Decoded data: 546573742064617461
	// Version Byte: [0 0]
}
