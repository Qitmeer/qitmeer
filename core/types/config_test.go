// Copyright 2017-2018 The qitmeer developers

package types

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/big"
	"os"
	"strings"
	"testing"
)

func TestConfig_UnmarshalJSON(t *testing.T) {
	for _, input := range []string{
		`{"Id":10}  `,
		`{"Id":"0xa"} `,
	} {
		r := strings.NewReader(input)
		conf := Config{}
		if err := json.NewDecoder(r).Decode(&conf); err != nil {
			t.Fatal(err)
		}
		raw, err := conf.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}
		fmt.Fprintf(os.Stdout, "%s\n", raw)

		assert.Equal(t, big.NewInt(10), conf.Id)
	}
}

func TestConfigIdShouldNotLessThan(t *testing.T) {
	input := `{"Id":-10}`
	r := strings.NewReader(input)
	conf := new(Config)
	err := json.NewDecoder(r).Decode(&conf)
	assert.NotNil(t, err)
	assert.EqualError(t, err, "error field 'Id' for Config, minimal is 0")
}
