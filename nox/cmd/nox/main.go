// Copyright 2017-2018 The nox developers

package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/noxproject/nox/core/types"
)

func init() {

}

func main() {
	initGenesis()
}

// initialize the genesis block from the genesis json
func initGenesis() {
	// open genesis JSON
	genesisFile := "/work/temp/nox-private/genesis_test.json"
	file, err := os.Open(genesisFile)
	if err != nil {
		log.Fatal(err)
	}
	genesis := new(types.Genesis)
	if err := json.NewDecoder(file).Decode(genesis); err != nil {
		log.Fatal(err)
	}
	log.Printf("Genesis=%v\n", genesis)

	// load genesis object by parse json
	// Open database & save genesis
}
