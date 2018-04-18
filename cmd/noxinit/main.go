// Copyright 2017-2018 The nox developers

package main

import (
	"os"
	"log"
	"encoding/json"
	"github.com/dindinw/dagproject/common/math"
	"math/big"
	"errors"
)

type ChainConfig struct {
	ChainId *big.Int `json:"chainId"`
}

type Genesis struct {
	Config     *ChainConfig
	Nonce      uint64
}

/*
To check for missing fields, you will have to use pointers in order to
distinguish between missing/null and zero values
 */
type genesisJSON struct {
	Config *ChainConfig             `json:"config"`
	Nonce *math.HexOrDecimal64      `json:"nonce"`
}

// MarshalJSON marshals Genesis as JSON.
func (g Genesis) MarshalJSON() ([]byte, error) {
	var enc genesisJSON
	enc.Config = g.Config
	nonce := math.HexOrDecimal64(g.Nonce)
	enc.Nonce = &nonce
	return json.Marshal(&enc)
}

// UnmarshalJSON unmarshals Genesis from JSON.
func (g *Genesis) UnmarshalJSON(input []byte) error {
	var dec genesisJSON
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.Config != nil {
		g.Config = dec.Config
	}else{
		return errors.New("missing required field 'config' for Genesis")
	}
	if dec.Nonce != nil {
		g.Nonce = uint64(*dec.Nonce)
	}
	return nil
}

func init() {

}

func main(){
	initGenesis()
}

// initialize the genesis block from the genesis json
func initGenesis(){
    // open genesis JSON
	genesisFile := "/work/temp/nox-private/genesis_test.json"
	file, err := os.Open(genesisFile)
	if err != nil {
		log.Fatal(err)
	}
	genesis := new(Genesis)
	if err := json.NewDecoder(file).Decode(genesis); err != nil {
		log.Fatal(err)
	}
	log.Printf("Genesis=%v\n",genesis)



   // load genesis object by parse json
   // Open database & save genesis
}
