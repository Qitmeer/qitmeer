// Copyright 2017-2018 The nox developers

package types

import (
	"testing"
	"strings"
	"encoding/json"
	"fmt"
	"os"
	"github.com/stretchr/testify/assert"
	"math/big"
)

func TestConfig_UnmarshalJSON(t *testing.T) {
	for _,input := range []string{
		`{"Id":10}  `,
		`{"Id":"0xa"} `,
	} {
		r := strings.NewReader(input)
		conf := Config{}
		if err := json.NewDecoder(r).Decode(&conf); err != nil {
			t.Fatal(err)
		}
		raw, err := conf.MarshalJSON();
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
	assert.NotNil(t,err)
	assert.Equal(t,ErrConfigIdRange,err)
}
