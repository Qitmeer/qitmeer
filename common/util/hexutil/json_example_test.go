/*
 * Copyright (c) 2020.
 * Project:qitmeer
 * File:json_example_test.go
 * Date:7/4/20 9:00 PM
 * Author:Jin
 * Email:lochjin@gmail.com
 */

package hexutil

import (
	"encoding/json"
	"fmt"
)

type MyType [5]byte

func (v *MyType) UnmarshalText(input []byte) error {
	return UnmarshalFixedText("MyType", input, v[:])
}

func (v MyType) String() string {
	return Bytes(v[:]).String()
}

func ExampleUnmarshalFixedText() {
	var v1, v2 MyType
	fmt.Println("v1 error:", json.Unmarshal([]byte(`"0x01"`), &v1))
	fmt.Println("v2 error:", json.Unmarshal([]byte(`"0x0101010101"`), &v2))
	fmt.Println("v2:", v2)
	// Output:
	// v1 error: hex string has length 2, want 10 for MyType
	// v2 error: <nil>
	// v2: 0x0101010101
}
