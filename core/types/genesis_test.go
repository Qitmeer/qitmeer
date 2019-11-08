// Copyright 2017-2018 The qitmeer developers

package types

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"math/big"
	"os"
	"strings"
	"testing"
)

func TestGenesis_MarshalJSON(t *testing.T) {
	genesis := Genesis{
		Config: &Config{Id: big.NewInt(2)},
		Nonce:  1000,
	}
	genesisFile, err := ioutil.TempFile("", "test_genesis_json")
	if err != nil {
		t.Fatal(err)
	}
	defer genesisFile.Close()
	data, err := genesis.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Fprintf(os.Stdout, "%s\n", genesisFile.Name())
	fmt.Fprintf(os.Stdout, "%s\n", data)
	genesisFile.Write(data)
	r, err := os.Open(genesisFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	g := new(Genesis)
	if err := json.NewDecoder(r).Decode(g); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, big.NewInt(2), g.Config.Id)
	assert.Equal(t, uint64(1000), g.Nonce)
}

func TestGenesis_UnmarshalJSON(t *testing.T) {
	for _, input := range []string{
		`{"config": {"Id":10}, "nonce":1000}  `,
		`{"config": {"Id":"0xa"}, "nonce":"0x3e8"} `,
	} {
		r := strings.NewReader(input)
		genesis := new(Genesis)
		if err := json.NewDecoder(r).Decode(genesis); err != nil {
			t.Fatal(err)
		}
		g, err := genesis.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}
		fmt.Fprintf(os.Stdout, "%s\n", g)

		assert.Equal(t, uint64(1000), genesis.Nonce)
		assert.Equal(t, big.NewInt(10), genesis.Config.Id)
	}
}

func TestNounceShouldLargeThanZero(t *testing.T) {
	input := `{"config": {"Id":10}, "nonce":0}  `
	r := strings.NewReader(input)
	g := new(Genesis)
	err := json.NewDecoder(r).Decode(&g)
	assert.NotNil(t, err)
	assert.EqualError(t, err, "error field 'nonce' for Genesis, minimal is 1")
}
