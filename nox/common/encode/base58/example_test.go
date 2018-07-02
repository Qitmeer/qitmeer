// Copyright (c) 2017-2018 The nox developers
// Copyright (c) 2014 The btcsuite developers
// Copyright (c) 2015 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package base58_test

import (
	"fmt"

	"github.com/noxproject/nox/common/encode/base58"
)



// This example demonstrates how to decode modified base58 encoded data.
func ExampleDecode() {
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
func ExampleEncode() {
	// Encode example data with the modified base58 encoding scheme.
	data := []byte("Test data")
	encoded := base58.Encode(data)

	// Show the encoded data.
	fmt.Println("Encoded Data:", encoded)

	// Output:
	// Encoded Data: 25JnwSn7XKfNQ
}

// This example demonstrates how to decode Base58Check encoded data.
func ExampleCheckDecodeBtc() {
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

// This example demonstrates how to encode data using the Base58Check encoding
// scheme.
func ExampleCheckEncodeBtc() {
	// Encode example data with the Base58Check encoding scheme.
	data := []byte("Test data")
	encoded := base58.CheckEncode(data, [2]byte{0x42,0})

	// Show the encoded data.
	fmt.Println("Encoded Data:", encoded)

	// Output:
	// Encoded Data: 182iP79GRURMp7oMHDU
}

func ExampleCheckEncodeDcr() {
	// Encode example data with the Base58Check encoding scheme.
	data := []byte("Test data")
	var ver [2]byte
	ver[0] = 0x44
	ver[1] = 0

	encoded := base58.CheckEncode(data, ver)

	// Show the encoded data.
	fmt.Println("Encoded Data:", encoded)

	// Output:
	// Encoded Data: 2uLtqkeVgFqTUBnjicK8o
}

func ExampleCheckDecodeDcr() {
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

func ExampleCheckEncode() {
	// Encode example data with the Base58Check encoding scheme.
	data := []byte("Test data")
	var ver [2]byte
	ver[0] = 0
	ver[1] = 0

	encoded := base58.CheckEncode(data, ver)

	// Show the encoded data.
	fmt.Println("Encoded Data:", encoded)

	// Output:
	// Encoded Data: 1182iP79GRURMp6Rsz9X
}

func ExampleCheckDecode() {
	encoded := "1182iP79GRURMp6Rsz9X"
	decoded, version, err := base58.CheckDecode(encoded)
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
