package main

import (
	"testing"
	"os"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"math/big"
	"strings"
	"fmt"
	"io/ioutil"
)

func TestGenesis_MarshalJSON(t *testing.T) {
	genesis := Genesis{
		Config:&ChainConfig{ChainId:big.NewInt(2)},
		Nonce:1000,
	}
	//genesisFile, err := os.Create("/tmp/new.json")
	genesisFile, err := ioutil.TempFile("","test_genesis_json")
	if err != nil {
		t.Fatal(err)
	}
	defer genesisFile.Close()
	data, err := genesis.MarshalJSON()
	if err !=nil {
		t.Fatal(err)
	}
	fmt.Fprintf(os.Stdout,"%s\n",genesisFile.Name())
	fmt.Fprintf(os.Stdout,"%s\n",data)
	genesisFile.Write(data)
	r, err := os.Open(genesisFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	g := new(Genesis)
    if err := json.NewDecoder(r).Decode(g); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, big.NewInt(2), g.Config.ChainId)
	assert.Equal(t, uint64(1000), g.Nonce)
}

func TestGenesis_UnmarshalJSON(t *testing.T) {
	r := strings.NewReader(`{"config": {"chainI":1}} extra `)
	genesis := new(Genesis)
	if err := json.NewDecoder(r).Decode(genesis); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, uint64(100),genesis.Nonce)
	assert.Equal(t, big.NewInt(1), genesis.Config.ChainId)

}
