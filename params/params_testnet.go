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

// testNetPowLimit is the highest proof of work value a block can
// have for the test network. It is the value 2^224 - 1.
var	testNetPowLimit = new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 224), common.Big1)

// TestNetParams defines the network parameters for the test network.
var TestNetParams = Params{
	Name:        "testnet",
	Net:         protocol.TestNet,
	DefaultPort: "18130",
	DNSSeeds: []DNSSeed{
		{"testnet-seed.alice.noxproject.io", true},
		{"testnet-seed.bob.noxproject.io", true},
		{"testnet-seed.charis.noxproject.io", true},
	},

	// Chain parameters
	GenesisBlock:             &testNetGenesisBlock,
	GenesisHash:              &testNetGenesisHash,
	PowLimit:                 testNetPowLimit,
	PowLimitBits:             0x1e00ffff,
	ReduceMinDifficulty:      false,
	MinDiffReductionTime:     0, // Does not apply since ReduceMinDifficulty false
	GenerateSupported:        true,
	MaximumBlockSizes:        []int{1310720},
	MaxTxSize:                1000000,
	TargetTimePerBlock:       time.Minute * 2,
	TargetTimespan:           time.Minute * 2 * 144, // TimePerBlock * WindowSize
	RetargetAdjustmentFactor: 4,

	// Subsidy parameters.
	BaseSubsidy:              2500000000, // 25 Coin
	MulSubsidy:               100,
	DivSubsidy:               101,
	SubsidyReductionInterval: 2048,
	WorkRewardProportion:     6,
	StakeRewardProportion:    3,
	BlockTaxProportion:       1,

	// Checkpoints ordered from oldest to newest.
	Checkpoints: []Checkpoint{
	},

	// Consensus rule change deployments.
	//
	Deployments: map[uint32][]ConsensusDeployment{
	},

	// Address encoding magics
	NetworkAddressPrefix: "T",
	PubKeyAddrID:         [2]byte{0x0f, 0x0f}, // starts with Tk
	PubKeyHashAddrID:     [2]byte{0x0f, 0x12}, // starts with Tm
	PKHEdwardsAddrID:     [2]byte{0x0f, 0x01}, // starts with Te
	PKHSchnorrAddrID:     [2]byte{0x0f, 0x1e}, // starts with Tr
	ScriptHashAddrID:     [2]byte{0x0e, 0xe2}, // starts with TS
	PrivateKeyID:         [2]byte{0x0c, 0xe2}, // starts with Pt

	// BIP32 hierarchical deterministic extended key magics
	HDPrivateKeyID: [4]byte{0x04, 0x35, 0x83, 0x97}, // starts with tprv
	HDPublicKeyID:  [4]byte{0x04, 0x35, 0x87, 0xd1}, // starts with tpub

	// BIP44 coin type used in the hierarchical deterministic path for
	// address generation.
	HDCoinType: 11,

	CoinbaseMaturity:        16,

	//DAG
	AnticoneSize:3,

}
