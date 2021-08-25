// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2014-2016 The btcsuite developers
// Copyright (c) 2015-2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package params

import (
	"github.com/Qitmeer/qitmeer/common"
	"github.com/Qitmeer/qitmeer/core/protocol"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"math/big"
	"time"
)

// mainPowLimit is the highest proof of work value a block can
// have for the main network. It is the value 2^224 - 1.
var mainPowLimit = new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 224), common.Big1)

// target time per block unit second(s)
const mainTargetTimePerBlock = 60 * 5

// MainNetParams defines the network parameters for the main network.
var MainNetParams = Params{
	Name:           "mainnet",
	Net:            protocol.MainNet,
	DefaultPort:    "8130",
	DefaultUDPPort: 8140,
	Bootstrap:      []string{},

	// Chain parameters
	GenesisBlock: &genesisBlock,
	GenesisHash:  &genesisHash,
	PowConfig: &pow.PowConfig{
		Blake2bdPowLimit:             mainPowLimit,
		Blake2bdPowLimitBits:         0x1d00ffff,
		X16rv3PowLimit:               mainPowLimit,
		X16rv3PowLimitBits:           0x1d00ffff,
		X8r16PowLimit:                mainPowLimit,
		X8r16PowLimitBits:            0x1d00ffff,
		QitmeerKeccak256PowLimit:     mainPowLimit,
		QitmeerKeccak256PowLimitBits: 0x1d00ffff,
		//hash ffffffffffffffff000000000000000000000000000000000000000000000000 corresponding difficulty is 48 for edge bits 24
		// Uniform field type uint64 value is 48 . bigToCompact the uint32 value
		// 24 edge_bits only need hash 1*4 times use for privnet if GPS is 2. need 50 /2 * 4 find once
		CuckarooMinDifficulty:  0x1300000 * 4,
		CuckaroomMinDifficulty: 0x1300000 * 4,
		CuckatooMinDifficulty:  0x1300000 * 4,

		Percent: map[pow.MainHeight]pow.PercentItem{
			pow.MainHeight(0): {
				pow.QITMEERKECCAK256: 100,
			},
		},
		// after this height the big graph will be the main pow graph
		AdjustmentStartMainHeight: 45 * 1440 * 60 / mainTargetTimePerBlock,
	},
	CoinbaseConfig:           CoinbaseConfigs{},
	ReduceMinDifficulty:      false,
	MinDiffReductionTime:     0, // Does not apply since ReduceMinDifficulty false
	GenerateSupported:        false,
	WorkDiffAlpha:            1,
	WorkDiffWindowSize:       144,
	WorkDiffWindows:          20,
	MaximumBlockSizes:        []int{393216},
	MaxTxSize:                393216,
	TargetTimePerBlock:       time.Second * mainTargetTimePerBlock,
	TargetTimespan:           time.Second * mainTargetTimePerBlock * 144, // TimePerBlock * WindowSize
	RetargetAdjustmentFactor: 4,

	// Subsidy parameters.
	BaseSubsidy:              3119582664, // 21m
	MulSubsidy:               100,
	DivSubsidy:               101,
	SubsidyReductionInterval: 6144,
	WorkRewardProportion:     9,
	StakeRewardProportion:    0,
	BlockTaxProportion:       1,

	// Checkpoints ordered from oldest to newest.
	Checkpoints: []Checkpoint{},

	// Address encoding magics
	NetworkAddressPrefix: "M",
	PubKeyAddrID:         [2]byte{0x1f, 0xc5}, // starts with Mk
	PubKeyHashAddrID:     [2]byte{0x0b, 0xb1}, // starts with Mm
	PKHEdwardsAddrID:     [2]byte{0x0b, 0x9f}, // starts with Me
	PKHSchnorrAddrID:     [2]byte{0x0b, 0xbd}, // starts with Mr
	ScriptHashAddrID:     [2]byte{0x0b, 0x81}, // starts with MS
	PrivateKeyID:         [2]byte{0x22, 0xdc}, // starts with Pm

	// BIP32 hierarchical deterministic extended key magics
	HDPrivateKeyID: [4]byte{0x03, 0xb8, 0xc4, 0x22}, // starts with nprv
	HDPublicKeyID:  [4]byte{0x03, 0xb8, 0xc8, 0x58}, // starts with npub

	// BIP44 coin type used in the hierarchical deterministic path for
	// address generation.
	// TODO : register coin type
	// https://github.com/satoshilabs/slips/blob/master/slip-0044.md
	HDCoinType: 223,

	CoinbaseMaturity: 512,

	OrganizationPkScript: hexMustDecode("76a914c0f0b73c320e1fe38eb1166a57b953e509c8f93e88ac"),
}
