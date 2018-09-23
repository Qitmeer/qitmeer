// Copyright (c) 2017-2018 The nox developers
// Copyright (c) 2014-2016 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package seed

import (
	"crypto/rand"
	"fmt"
)

const (

	// The Default length in bytes for a seed
	DefaultSeedBytes = 32 // 256 bits

	// MinSeedBytes is the minimum number of bytes allowed for a seed
	MinSeedBytes = 16 // 128 bits

	// MaxSeedBytes is the maximum number of bytes allowed for a seed
	MaxSeedBytes = 256 // 1024 bits
)

var (
	// ErrInvalidSeedLen describes an error in which the provided seed or
	// seed length is not in the allowed range.
	ErrInvalidSeedLen = fmt.Errorf("seed length must be between %d and %d "+
	"bits", MinSeedBytes*8, MaxSeedBytes*8)
)

// GenerateSeed returns a cryptographically secure random seed
// that can be used as the input for the further usage.
//
// The length is in bytes and it must be between 16 and 256 (128 to 1024 bits).
// The default length is 32 (256 bits) as defined by the DefaultSeedLength constant.
func GenerateSeed(length uint16) ([]byte, error) {
	// Per [BIP32], the seed must be in range [MinSeedBytes, MaxSeedBytes].
	if length < MinSeedBytes || length > MaxSeedBytes {
		return nil, ErrInvalidSeedLen
	}

	buf := make([]byte, length)
	_, err := rand.Read(buf)
	if err != nil {
		return nil, err
	}

	return buf, nil
}
