/*
 * Copyright (c) 2020.
 * Project:qitmeer
 * File:bn256_slow.go
 * Date:7/8/20 8:24 AM
 * Author:Jin
 * Email:lochjin@gmail.com
 */

// +build !amd64,!arm64

// Package bn256 implements the Optimal Ate pairing over a 256-bit Barreto-Naehrig curve.
package bn256

import bn256 "github.com/Qitmeer/qitmeer/p2p/crypto/bn256/google"

// G1 is an abstract cyclic group. The zero value is suitable for use as the
// output of an operation, but cannot be used as an input.
type G1 = bn256.G1

// G2 is an abstract cyclic group. The zero value is suitable for use as the
// output of an operation, but cannot be used as an input.
type G2 = bn256.G2

// PairingCheck calculates the Optimal Ate pairing for a set of points.
func PairingCheck(a []*G1, b []*G2) bool {
	return bn256.PairingCheck(a, b)
}
