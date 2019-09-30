// Copyright 2017-2018 The qitmeer developers

package rlp

import (
	"fmt"
)

func ExampleEncodeToBytes() {
	inputs := []interface{}{
		[]byte{0},
		[]byte{1},
		[]byte{127},
		[]byte{128}, // 128 -> 0x80,
		[]byte{129},
		[]byte{192}, // 192 -> 0xc0
		[]byte{255},
		[]byte{'\xaa'}, // 170
		[]byte{'\xbb'}, // 187
		[]byte{0, 0},
		[]byte{0, 128},
		[]byte{1, 128},
		[]byte{128, 0},
		[]byte{0, 0, 0},
		[]byte{'A', 'B', 'C'},      // [65, 66, 67]
		[]byte{'a', 'b', 'c', 'd'}, // [97,98,99,100]
		[]uint{0},
		[]uint{1},
		[]uint{127},
		[]uint{128},
		[]uint{129},
		[]uint{255},
		[]uint{0, 0},
		[]uint{0, 0, 0},
		[]string{"0"}, //0x30
		[]uint{'\x30'},
		[]string{"0", "0"},
		[]string{"0", "0", "0"},
		[]string{"a"}, //same with uint 97
		[]uint{97},
		[]string{"abc"},
		[]string{"abc", "ABC"},
		[]string{"00"},
		[]uint{'\x30', '\x30', '\x00'},
		[]uint{'\x30', '\x30'},
		[]byte{'\x30', '\x30'},
		"00",
		[]interface{}{[]byte{'\x30', '\x30'}}, //same with []string{"00"}
	}

	for _, v := range inputs {
		bytes, _ := EncodeToBytes(v)
		fmt.Printf("%#v -> %X\n", v, bytes) //%X	base 16, with upper-case letters for A-F
	}
	// Output: []byte{0x0} -> 00
	// []byte{0x1} -> 01
	// []byte{0x7f} -> 7F
	// []byte{0x80} -> 8180
	// []byte{0x81} -> 8181
	// []byte{0xc0} -> 81C0
	// []byte{0xff} -> 81FF
	// []byte{0xaa} -> 81AA
	// []byte{0xbb} -> 81BB
	// []byte{0x0, 0x0} -> 820000
	// []byte{0x0, 0x80} -> 820080
	// []byte{0x1, 0x80} -> 820180
	// []byte{0x80, 0x0} -> 828000
	// []byte{0x0, 0x0, 0x0} -> 83000000
	// []byte{0x41, 0x42, 0x43} -> 83414243
	// []byte{0x61, 0x62, 0x63, 0x64} -> 8461626364
	// []uint{0x0} -> C180
	// []uint{0x1} -> C101
	// []uint{0x7f} -> C17F
	// []uint{0x80} -> C28180
	// []uint{0x81} -> C28181
	// []uint{0xff} -> C281FF
	// []uint{0x0, 0x0} -> C28080
	// []uint{0x0, 0x0, 0x0} -> C3808080
	// []string{"0"} -> C130
	// []uint{0x30} -> C130
	// []string{"0", "0"} -> C23030
	// []string{"0", "0", "0"} -> C3303030
	// []string{"a"} -> C161
	// []uint{0x61} -> C161
	// []string{"abc"} -> C483616263
	// []string{"abc", "ABC"} -> C88361626383414243
	// []string{"00"} -> C3823030
	// []uint{0x30, 0x30, 0x0} -> C3303080
	// []uint{0x30, 0x30} -> C23030
	// []byte{0x30, 0x30} -> 823030
	// "00" -> 823030
	// []interface {}{[]uint8{0x30, 0x30}} -> C3823030
}

/*
  https://github.com/ethereum/wiki/wiki/%5BEnglish%5D-RLP
    * The string "dog" = [ 0x83, 'd', 'o', 'g' ]
    * The list [ "cat", "dog" ] = [ 0xc8, 0x83, 'c', 'a', 't', 0x83, 'd', 'o', 'g' ]
    * The empty string ('null') = [ 0x80 ]
    * The empty list = [ 0xc0 ]
    * The encoded integer 15 ('\x0f') = [ 0x0f ]
    * The encoded integer 1024 ('\x04\x00') = [ 0x82, 0x04, 0x00 ]
    * The set theoretical representation of two, [ [], [[]], [ [], [[]] ] ] = [ 0xc7, 0xc0, 0xc1, 0xc0, 0xc3, 0xc0, 0xc1, 0xc0 ]
    * The string "Lorem ipsum dolor sit amet, consectetur adipisicing elit" = [ 0xb8, 0x38, 'L', 'o', 'r', 'e', 'm', ' ', ... , 'e', 'l', 'i', 't' ]
*/
func ExampleEncodeToBytes_inSpec() {
	inputs := []interface{}{
		"dog",
		[]string{"cat", "dag"},
		"",
		[]string{},
		[]byte{'\x0f'},         //15
		[]byte{'\x04', '\x00'}, //1024
		[]interface{}{
			[]string{}, []interface{}{[]string{}}, []interface{}{[]string{}, []interface{}{[]string{}}},
		}, // [ [], [[]], [ [], [[]] ] ]
		"Lorem ipsum dolor sit amet, consectetur adipisicing elit",
	}
	for _, v := range inputs {
		bytes, _ := EncodeToBytes(v)
		fmt.Printf("%#v -> %X\n", v, bytes) //%X	base 16, with upper-case letters for A-F
	}
	//Output: "dog" -> 83646F67
	//[]string{"cat", "dag"} -> C88363617483646167
	//"" -> 80
	//[]string{} -> C0
	//[]byte{0xf} -> 0F
	//[]byte{0x4, 0x0} -> 820400
	//[]interface {}{[]string{}, []interface {}{[]string{}}, []interface {}{[]string{}, []interface {}{[]string{}}}} -> C7C0C1C0C3C0C1C0
	//"Lorem ipsum dolor sit amet, consectetur adipisicing elit" -> B8384C6F72656D20697073756D20646F6C6F722073697420616D65742C20636F6E7365637465747572206164697069736963696E6720656C6974

}
