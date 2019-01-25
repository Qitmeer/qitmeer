// Copyright 2017-2018 The nox developers
// Copyright (c) 2013-2014 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.


package types

const (
	// AtomsPerCent is the number of atomic units in one coin cent.
	// TODO, relocate the coin related item to chain's params
	AtomsPerCent = 1e6

	// AtomsPerCoin is the number of atomic units in one coin.
	// TODO, relocate the coin related item to chain's params
	AtomsPerCoin = 1e8

	// MaxAmount is the maximum transaction amount allowed in atoms.
	// TODO, relocate the coin related item to chain's params
	MaxAmount = 21e6 * AtomsPerCoin
)
