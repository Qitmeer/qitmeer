// Copyright (c) 2017-2018 The nox developers
// Copyright (c) 2013-2014 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

/*
Package secp256k1 implements support for the elliptic curves needed for nox.

Nox uses elliptic curve cryptography using koblitz curves
(specifically secp256k1) for cryptographic functions.  See
http://www.secg.org/collateral/sec2_final.pdf for details on the
standard.

This package provides the data structures and functions implementing the
crypto/elliptic Curve interface in order to permit using these curves
with the standard crypto/ecdsa package provided with go. Helper
functionality is provided to parse signatures and public keys from
standard formats.
*/
package secp256k1
