// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2014 The btcsuite developers
// Copyright (c) 2015 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package base58_test

import (
	"testing"

	"github.com/Qitmeer/qitmeer/common/encode/base58"
)

var checkEncodingStringTests = []struct {
	version0 byte
	version1 byte
	in       string
	out      string
}{
	{0x42, 20, "", "3MNQE1X"},
	{0x42, 20, " ", "B2Kr6dBE"},
	{0x42, 20, "-", "B3jv1Aft"},
	{0x42, 20, "0", "B482yuaX"},
	{0x42, 20, "1", "B4CmeGAC"},
	{0x42, 20, "-1", "mM7eUf6kB"},
	{0x42, 20, "11", "mP7BMTDVH"},
	{0x42, 20, "abc", "4QiVtDjUdeq"},
	{0x42, 20, "1234598760", "ZmNb8uQn5zvnUohNCEPP"},
	{0x42, 20, "abcdefghijklmnopqrstuvwxyz", "K2RYDcKfupxwXdWhSAxQPCeiULntKm63UXyx5MvEH2"},
	{0x42, 20, "00000000000000000000000000000000000000000000000000000000000000", "bi1EWXwJay2udZVxLJozuTb8Meg4W9c6xnmJaRDjg6pri5MBAxb9XwrpQXbtnqEoRV5U2pixnFfwyXC8tRAVC8XxnjK"},

	{0x44, 20, "", "auJfeHwC"}, // test 11
	{0x44, 20, " ", "3adhVZxihM"},
	{0x44, 20, "-", "3adj1TRkKZ"},
	{0x44, 20, "0", "3adjJWGj89"},
	{0x44, 20, "1", "3adjR6HCdi"},
	{0x44, 20, "-1", "CPT64YbXmzR"},
	{0x44, 20, "11", "CPT84AuHhM1"},
	{0x44, 20, "abc", "sG884ALZzrd4"},
	{0x44, 20, "1234598760", "9Qb3AD8psZVrjGEfEVMKkM"},
	{0x44, 20, "abcdefghijklmnopqrstuvwxyz", "5akkPDZM6XaNjyht71YXqR4kFNtkoVFQo2SvDDAikLYE"},
	{0x44, 20, "00000000000000000000000000000000000000000000000000000000000000", "9uayvqtptCkt9inFUhW4t68L1YBZMsZZjqLmmfszj3kx4uLDEJPZ65wGAbxmnY9hZiK74ggerwV7ga9HCUHkhMUchLPf4"},

	{0x4e, 20, "", "ft9ZcnPM"}, // test 22
	{0x4e, 20, " ", "3xcWA3JEKm"},
	{0x4e, 20, "-", "3xcXdqa1XB"},
	{0x4e, 20, "0", "3xcY2te4UB"},
	{0x4e, 20, "1", "3xcY4hggyU"},
	{0x4e, 20, "-1", "E4TpCcMT8Yy"},
	{0x4e, 20, "11", "E4TrCCzXwtL"},
	{0x4e, 20, "abc", "zeKL3ntPqDcf"},
	{0x4e, 20, "1234598760", "AeD3mvFdyCt8akCaar1Muf"},
	{0x4e, 20, "abcdefghijklmnopqrstuvwxyz", "6FnprApC2tXWptudC4CzngX2qUYTrzGUdhXoFAK53LFq"},
	{0x4e, 20, "00000000000000000000000000000000000000000000000000000000000000", "BDU3ycLHFvYhCRqEXvg2qH8WqUs1scJiQVz86VNAmeq5JuiarpT7TQAvgWikTdeMfuW7EofPmHmMru7JYzSJZxAedRZE8"},
}

func TestBase58Check(t *testing.T) {
	for x, test := range checkEncodingStringTests {
		var ver [2]byte
		ver[0] = test.version0
		ver[1] = test.version1
		// test encoding
		var eRes []byte
		switch ver[0] {
		case 0x42:
			eRes,_ = base58.BtcCheckEncode([]byte(test.in), ver[1])
		case 0x44:
			eRes,_ = base58.DcrCheckEncode([]byte(test.in), ver)
		default:
			eRes,_ = base58.QitmeerCheckEncode([]byte(test.in), ver[:])
		}
		if string(eRes) != test.out {
			t.Errorf("CheckEncode test #%d failed: got %s, want: %s", x, eRes, test.out)
		}
		var res []byte
		var version [2]byte
		var err error
		// test decoding
		switch ver[0] {
		case 0x42:
			version[0] = 0x42
			res, version[1], err = base58.BtcCheckDecode(test.out)
		case 0x44:
			res, version, err = base58.DcrCheckDecode(test.out)
		default:
			res, version, err = base58.QitmeerCheckDecode(test.out)

		}
		if err != nil {
			t.Errorf("CheckDecode test #%d failed with err: %v", x, err)
		} else if version[0] != ver[0] {
			t.Errorf("CheckDecode test #%d failed: got version: %d want: %d", x, version, ver)
		} else if version[1] != ver[1] {
			t.Errorf("CheckDecode test #%d failed: got version: %d want: %d", x, version, ver)
		} else if string(res) != test.in {
			t.Errorf("CheckDecode test #%d failed: got: %s want: %s", x, res, test.in)
		}

	}

	// test the two decoding failure cases
	// case 1: checksum error
	_, _, err := base58.QitmeerCheckDecode("Axk2WA6M")
	if err != base58.ErrChecksum {
		t.Error("Checkdecode test failed, expected ErrChecksum")
	}
	// case 2: invalid formats (string lengths below 5 mean the version byte and/or the checksum
	// bytes are missing).
	testString := ""
	for len := 0; len < 4; len++ {
		// make a string of length `len`
		_, _, err = base58.QitmeerCheckDecode(testString)
		if err != base58.ErrInvalidFormat {
			t.Error("Checkdecode test failed, expected ErrInvalidFormat")
		}
	}

}
