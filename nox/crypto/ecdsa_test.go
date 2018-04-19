// Copyright 2017-2018 The nox developers

package crypto

import (
	"fmt"
	"testing"
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"math/big"
)

func TestCreatePrivateKey(t *testing.T) {
	_, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
}

func TestRecoveryPrivateKey(t *testing.T) {

	var testData = struct{
		X,Y,D,Hex string
	}{
		X:"87950662733769106189209133184670574697837350395849211932318447659544876731804",
		Y:"105695193418456567358484820636781991762237673468026920925214955808655271236414",
		D:"10033073139661513675084157146470084444134961864185077454113967581855315496919",
		Hex:"162e84a5ffd1a131ac7d9fa127b66d1bf203b161f16b32ab74643bc6f1158fd7",   /* -> binary for storage */
	}
	keyBytes,err := hex.DecodeString(testData.Hex)
	if (err!=nil){
		t.Fatal(err)
	}
	privKey, err :=ToECDSA(keyBytes,true)

	// test for D
	d := new(big.Int)
	d.SetString(testData.D,10)
	assert.Equal(t,d,privKey.D);

	d2 := []byte(testData.D) // notice, wrong convert
	d.SetBytes(d2)
	assert.NotEqual(t,d,privKey.D); //! not equal here !

	d3 := new(big.Int)
	fmt.Sscan(testData.D, d3)
	assert.Equal(t,d3,privKey.D)

	// test for X
	x := new(big.Int)
	x.SetString(testData.X,10)
	assert.Equal(t,x,privKey.X)

	// test for Y
	y := new(big.Int)
	y.SetString(testData.Y,10)
	assert.Equal(t,y,privKey.Y)

	// test []byte key is identical
	newKeyBytes := FromECDSA(privKey)
	assert.Equal(t,keyBytes, newKeyBytes)
	assert.Equal(t,testData.Hex,hex.EncodeToString(newKeyBytes))
}
