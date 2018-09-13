package main

import (
	"os"
	"github.com/ethereum/go-ethereum/crypto"
	"encoding/hex"
	"fmt"
)

func main() {
	args := os.Args[1:]
	input := args[0]
	inputBytes, err := hex.DecodeString(input)
	if err != nil {
		panic(err)
	}
	out := crypto.Keccak256(inputBytes[:])
	fmt.Println(hex.EncodeToString(out))
}
