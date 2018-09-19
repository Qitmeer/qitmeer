// Copyright 2017-2018 The nox developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import (
	"encoding/hex"
	"fmt"
	"github.com/noxproject/nox/common/encode/base58"
	"github.com/noxproject/nox/crypto/bip32"
	"github.com/noxproject/nox/crypto/bip39"
	"github.com/stretchr/testify/assert"
	"testing"
)

var testAddresses = []struct{
	ver [2]byte
	addr string
}{
	{[2]byte{0x0c,0x48},"NpEpq4L8vHqADsUKMtWpxu7SU2z69UEhydX"},
	{[2]byte{0x0c,0x40},"Nm281BTkccPTDL1CfhAAR27GAzx2bnKjZdM"},
	{[2]byte{0x0c,0x4f},"Ns4C3x6DwG6QMbWHy4TACHzLySWtkXEgP9J"},
	{[2]byte{0x0c,0xf3},"Q13cyRu7BhgQ3ibAB5FnTpYyzfF9T1SjfZm"},
	{[2]byte{0x0c,0x11},"NS7EeuMk5Eug2tdebYSWSEDUCREvGRW9UCF"},
	{[2]byte{0x0c,0xd1},"PmN8TxCyPZYeFurREEzTdLHjFG6PNtQsmvU"},
	{[2]byte{0x0d,0xee},"Rk3TNa9vVuxhxzP9vDZPTY2JTdrpDYUi8Ba"},
	{[2]byte{0x0d,0xf1},"RmFUBXUpN3W6bSgaBHpPQzQ7pXNb3fG59Kn"},
	{[2]byte{0x0d,0xdf},"Re1PKoXTBGFkpit4crGPgG9DfCHx4oo4jsm"},
	{[2]byte{0x0d,0xfd},"Rr5XRLnPpZff7FtFDarPEouPG5RgNE1TN9R"},
	{[2]byte{0x0e,0x01},"RsgsqHDayQPWcXcoZgXiWkQUQbSi91XLzvT"},
	{[2]byte{0x0c,0xdd},"PrCBhmWYr5iCmj46GY2TT9nzgp9UhUoRLes"},
	{[2]byte{0x0f,0x0f},"TkL8h8Y4m76dBLeSxHoeZgFxpXeGpxsFKq6"},
	{[2]byte{0x0f,0x11},"Tm8ou6kfLXTYvyWj8LeKCeWWPneniPb1kcP"},
	{[2]byte{0x0f,0x01},"TehQFM1tjAa8utaVjwvz6tW9oiafd3FkG4k"},
	{[2]byte{0x0f,0x1e},"TrNCjuAY5koaKc9YFf6eLx93cyD8ydhqP9L"},
	{[2]byte{0x0f,0x20},"TsAswsP8fBAW5F1pRhwJyvPbCEDes4TnaAa"},
	{[2]byte{0x0c,0xe2},"PtCsih43HdcX9pDnhf883aRMcxfmQxxB8M9"},
}

var versions = []struct{
	version [2]byte
}{
	{[2]byte{0x0c, 0x48}}, // starts with Nk
	{[2]byte{0x0c, 0x40}}, // starts with Nm
	{[2]byte{0x0c, 0x4f}}, // starts with Ns
	{[2]byte{0x0c, 0xf3}}, // starts with NE
	{[2]byte{0x0c, 0x11}}, // starts with NS
	{[2]byte{0x0c, 0xd1}}, // starts with Pm

	{[2]byte{0x0d, 0xee}}, // starts with Rk
	{[2]byte{0x0d, 0xf1}}, // starts with Rm
	{[2]byte{0x0d, 0xdf}}, // starts with Re
	{[2]byte{0x0d, 0xfd}}, // starts with Rr
	{[2]byte{0x0e, 0x01}}, // starts with Rs
	{[2]byte{0x0c, 0xdd}}, // starts with Pr

	{[2]byte{0x0f, 0x0f}}, // starts with Tk
	{[2]byte{0x0f, 0x11}}, // starts with Tm
	{[2]byte{0x0f, 0x01}}, // starts with Te
	{[2]byte{0x0f, 0x1e}}, // starts with Tr
	{[2]byte{0x0f, 0x20}}, // starts with Ts
	{[2]byte{0x0c, 0xe2}}, // starts with Pt
}

func TestNoxBase58CheckEncode(t *testing.T) {
	// Encode example data with the Base58Check encoding scheme.
	data := []byte{ 0x64, 0xe2, 0x0e, 0xb6, 0x07, 0x55, 0x61, 0xd3, 0x0c, 0x23, 0xa5, 0x17,
		0xc5, 0xb7, 0x3b, 0xad, 0xbc, 0x12, 0x0f, 0x05}
	/*
	for _,ver := range versions {
		encoded := base58.CheckEncode(data, ver.version)
		// Show the encoded data.
		fmt.Printf("{[2]byte{0x%.2x,0x%.2x},%q},\n", ver.version[0], ver.version[1], encoded)
	}
	*/
	for _, addrtest := range testAddresses {
		encoded := base58.CheckEncode(data, addrtest.ver)
		assert.Equal(t,fmt.Sprintf("%s",encoded),addrtest.addr)
	}
}

func TestNoxHd(t *testing.T) {
	// Use bx to verify the result
	// $ echo "skin join dog sponsor camera puppy ritual diagram arrow poverty boy elbow" | bx mnemonic-to-seed
	//   e8b9a1929f72ccd322b5241dd9a62c08d02129c386de9dc505bb4decc8eb33ffd89575a037be4fbb9a4cf258bff88bbf89745f2f523298e98810e5099502d1b6
	//
	// And double check https://iancoleman.io/bip39/
	var mnemonic = "skin join dog sponsor camera puppy ritual diagram arrow poverty boy elbow"
	// Generate a Bip32 HD wallet for the mnemonic and a user supplied password
	seed := bip39.NewSeed(mnemonic, "")

	masterKey, _ := bip32.NewMasterKey(seed)
	publicKey := masterKey.PublicKey()

	assert.Equal(t, hex.EncodeToString(seed),
		"e8b9a1929f72ccd322b5241dd9a62c08d02129c386de9dc505bb4decc8eb33ffd89575a037be4fbb9a4cf258bff88bbf89745f2f523298e98810e5099502d1b6")
	assert.Equal(t, fmt.Sprintf("%s",masterKey),
		"xprv9s21ZrQH143K4JrrZbKUXWM2MLy4zm6MHSSCJ3X4X1dUHnuYMeMwpWetk7ovL2uyzbvyoEpA6DTrtqFGExCGifFCJjHoDxYSaeerYN4CgrZ")
	assert.Equal(t, fmt.Sprintf("%s",publicKey),
		"xpub661MyMwAqRbcGnwKfcrUteHkuNoZQDpCefMo6Rvg5MATAbEguBgCNJyNbPcHrg9vDcYmas8e2fEUm7mqrWX4xoMrdCQj849PgaU2ubvBJTt")
}
