// Copyright (c) 2017-2018 The nox developers
// Copyright (c) 2014-2016 The btcsuite developers
// Copyright (c) 2015-2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package params

import (
	"time"
	"math/big"
	"github.com/noxproject/nox/common"
	"github.com/noxproject/nox/core/protocol"
)

// mainPowLimit is the highest proof of work value a block can
// have for the main network. It is the value 2^224 - 1.
var mainPowLimit    = new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 224), common.Big1)

// MainNetParams defines the network parameters for the main network.
var MainNetParams = Params{
	Name:        "mainnet",
	Net:         protocol.MainNet,
	DefaultPort: "8130",
	DNSSeeds: []DNSSeed{
		{"mainnet-seed.alice.nox.io", true},
		{"mainnet-seed.bob.nox.io", true},
		{"mainnet-seed.charis.nox.io", true},
	},

	// Chain parameters
	GenesisBlock:             &genesisBlock,
	GenesisHash:              &genesisHash,
	PowLimit:                 mainPowLimit,
	PowLimitBits:             0x1d00ffff,
	ReduceMinDifficulty:      false,
	MinDiffReductionTime:     0, // Does not apply since ReduceMinDifficulty false
	GenerateSupported:        false,
	MaximumBlockSizes:        []int{393216},
	MaxTxSize:                393216,
	TargetTimePerBlock:       time.Minute * 5,
	TargetTimespan:           time.Minute * 5 * 144, // TimePerBlock * WindowSize
	RetargetAdjustmentFactor: 4,

	// Subsidy parameters.
	SubsidyReductionInterval: 210000,  //bitcoin mainnet

	// Checkpoints ordered from oldest to newest.
	Checkpoints: []Checkpoint{
	},

	Deployments: map[uint32][]ConsensusDeployment{
	},

	// Address encoding magics
	NetworkAddressPrefix: "D",
	PubKeyHashAddrID:     [2]byte{0x07, 0x3f}, // starts with Ds
	ScriptHashAddrID:     [2]byte{0x07, 0x1a}, // starts with Dc
	PrivateKeyID:         [2]byte{0x22, 0xde}, // starts with Pm

	// BIP32 hierarchical deterministic extended key magics
	HDPrivateKeyID: [4]byte{0x02, 0xfd, 0xa4, 0xe8}, // starts with dprv
	HDPublicKeyID:  [4]byte{0x02, 0xfd, 0xa9, 0x26}, // starts with dpub

	// BIP44 coin type used in the hierarchical deterministic path for
	// address generation.
	HDCoinType: 20,

	CoinbaseMaturity:        256,
}
