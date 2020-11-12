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

// testMixNetPowLimit is the highest proof of work value a block can
// have for the test network. It is the value 2^250 - 1.
// target 0x03ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff
var testMixNetPowLimit = new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 250), common.Big1)

// target time per block unit second(s)
const mixTargetTimePerBlock = 15

// Difficulty check interval is about 60*15 = 15 mins
const mixWorkDiffWindowSize = 60

// testPowNetParams defines the network parameters for the test network.
var MixNetParams = Params{
	Name:           "mixnet",
	Net:            protocol.MixNet,
	DefaultPort:    "28130",
	DefaultUDPPort: 28140,
	Bootstrap: []string{
		"/dns4/ns.qitmeer.top/tcp/28230/p2p/16Uiu2HAmRtp5CjNv3WvPYuh7kNXXZQDYegwFFeDH9vWY3JY4JS1W",
	},

	// Chain parameters
	GenesisBlock:         &testPowNetGenesisBlock,
	GenesisHash:          &testPowNetGenesisHash,
	ReduceMinDifficulty:  false,
	MinDiffReductionTime: 0, // Does not apply since ReduceMinDifficulty false
	GenerateSupported:    true,
	PowConfig: &pow.PowConfig{
		Blake2bdPowLimit:             testMixNetPowLimit,
		Blake2bdPowLimitBits:         0x2003ffff,
		X16rv3PowLimit:               testMixNetPowLimit,
		X16rv3PowLimitBits:           0x2003ffff,
		X8r16PowLimit:                testMixNetPowLimit,
		X8r16PowLimitBits:            0x2003ffff,
		QitmeerKeccak256PowLimit:     testMixNetPowLimit,
		QitmeerKeccak256PowLimitBits: 0x2003ffff,
		CryptoNightPowLimit:          testMixNetPowLimit,
		CryptoNightPowLimitBits:      0x2003ffff,
		//hash ffffffffffffffff000000000000000000000000000000000000000000000000 corresponding difficulty is 48 for edge bits 24
		// Uniform field type uint64 value is 48 . bigToCompact the uint32 value
		// 24 edge_bits only need hash 1*4 times use for privnet if GPS is 2. need 50 /2 * 2 ≈ 1min find once
		CuckarooMinDifficulty:  0x1600000, // 96
		CuckatooMinDifficulty:  0x1600000, // 96
		CuckaroomMinDifficulty: 0x1600000, // 96

		Percent: map[pow.MainHeight]pow.PercentItem{
			pow.MainHeight(0): {
				pow.BLAKE2BD:         20,
				pow.CUCKAROO:         30,
				pow.QITMEERKECCAK256: 30,
				pow.CUCKAROOM:        20,
			},
			pow.MainHeight(35000): {
				pow.BLAKE2BD:         20,
				pow.CUCKAROO:         30,
				pow.QITMEERKECCAK256: 10,
				pow.CRYPTONIGHT:      30,
				pow.CUCKAROOM:        10,
			},
		},
		// after this height the big graph will be the main pow graph
		AdjustmentStartMainHeight: 1440 * 15 / mixTargetTimePerBlock,
	},

	WorkDiffAlpha:            1,
	WorkDiffWindowSize:       mixWorkDiffWindowSize,
	WorkDiffWindows:          20,
	MaximumBlockSizes:        []int{1310720},
	MaxTxSize:                1000000,
	TargetTimePerBlock:       time.Second * mixTargetTimePerBlock,
	TargetTimespan:           time.Second * mixTargetTimePerBlock * mixWorkDiffWindowSize, // TimePerBlock * WindowSize
	RetargetAdjustmentFactor: 2,

	// Subsidy parameters.
	BaseSubsidy:              12000000000, // 120 Coin, stay same with testnet
	MulSubsidy:               100,
	DivSubsidy:               10000000000000, // Coin-base reward reduce to zero at 1540677 blocks created
	SubsidyReductionInterval: 1669066,        // 120 * 1669066 (blocks) *= 200287911 (200M) -> 579 ~ 289 days
	WorkRewardProportion:     10,
	StakeRewardProportion:    0,
	BlockTaxProportion:       0,

	// Checkpoints ordered from oldest to newest.
	Checkpoints: []Checkpoint{},

	// Consensus rule change deployments.
	//
	Deployments: map[uint32][]ConsensusDeployment{},

	// Address encoding magics
	NetworkAddressPrefix: "X",
	PubKeyAddrID:         [2]byte{0x11, 0x6e}, // starts with Xx
	PubKeyHashAddrID:     [2]byte{0x11, 0x53}, // starts with Xm
	PKHEdwardsAddrID:     [2]byte{0x11, 0x3c}, // starts with Xc
	PKHSchnorrAddrID:     [2]byte{0x11, 0x5f}, // starts with Xr
	ScriptHashAddrID:     [2]byte{0x11, 0x3e}, // starts with Xd
	PrivateKeyID:         [2]byte{0x11, 0x64}, // starts with Xt

	// BIP32 hierarchical deterministic extended key magics
	HDPrivateKeyID: [4]byte{0x01, 0x9d, 0x0b, 0xe1}, // starts with LsFC
	HDPublicKeyID:  [4]byte{0x01, 0x9d, 0x0d, 0x62}, // starts with LsG9

	// BIP44 coin type used in the hierarchical deterministic path for
	// address generation.
	HDCoinType: 223,

	CoinbaseMaturity: 720,
	//OrganizationPkScript:  hexMustDecode("76a914868b9b6bc7e4a9c804ad3d3d7a2a6be27476941e88ac"),
}
