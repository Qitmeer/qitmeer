package main

import (
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qng-core/crypto/bip32"
	"github.com/Qitmeer/qng-core/crypto/bip39"
	"os"
)

func main() {

	args := os.Args[1:]
	mnemonic := args[0]
	// Generate a Bip32 HD wallet for the mnemonic and a user supplied password
	seed := bip39.NewSeed(mnemonic, "")

	masterKey, _ := bip32.NewMasterKey(seed)
	publicKey := masterKey.PublicKey()

	// Display mnemonic and keys
	fmt.Println("Mnemonic: ", mnemonic)
	fmt.Println("Seed: ", hex.EncodeToString(seed))
	fmt.Println("Master private key: ", masterKey)
	fmt.Println("Master public key: ", publicKey)
}
