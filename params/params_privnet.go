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
	"github.com/Qitmeer/qitmeer/ledger"
	"math/big"
	"time"
)

// privNetPowLimit is the highest proof of work value a block can
// have for the private test network. It is the value 2^255 - 1.
var privNetPowLimit = new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 255), common.Big1)

// target time per block unit second(s)
const privTargetTimePerBlock = 30

// PirvNetParams defines the network parameters for the private test network.
// This network is similar to the normal test network except it is
// intended for private use within a group of individuals doing simulation
// testing.  The functionality is intended to differ in that the only nodes
// which are specifically specified are used to create the network rather than
// following normal discovery rules.  This is important as otherwise it would
// just turn into another public testnet.
var PrivNetParams = Params{
	Name:           "privnet",
	Net:            protocol.PrivNet,
	DefaultPort:    "38130",
	DefaultUDPPort: 38140,
	Bootstrap:      []string{},

	// Chain parameters
	GenesisBlock: &privNetGenesisBlock,
	GenesisHash:  &privNetGenesisHash,
	LedgerParams: ledger.LedgerParams{
		GenesisAmountUnit: 1000 * 1e8,
		MaxLockHeight:     10 * 365 * 5,
	},
	PowConfig: &pow.PowConfig{
		Blake2bdPowLimit:             privNetPowLimit,
		Blake2bdPowLimitBits:         0x207fffff,
		X8r16PowLimit:                privNetPowLimit,
		X8r16PowLimitBits:            0x207fffff,
		X16rv3PowLimit:               privNetPowLimit,
		X16rv3PowLimitBits:           0x207fffff,
		QitmeerKeccak256PowLimit:     privNetPowLimit,
		QitmeerKeccak256PowLimitBits: 0x207fffff,
		MeerXKeccakV1PowLimit:        privNetPowLimit,
		MeerXKeccakV1PowLimitBits:    0x207fffff,
		//hash ffffffffffffffff000000000000000000000000000000000000000000000000 corresponding difficulty is 48 for edge bits 24
		// Uniform field type uint64 value is 48 . bigToCompact the uint32 value
		// 24 edge_bits only need hash 1 times use for privnet if GPS is 2. need 50 /2 = 25s find once
		CuckarooMinDifficulty:  0x1300000,
		CuckatooMinDifficulty:  0x1300000,
		CuckaroomMinDifficulty: 0x1300000,

		Percent: map[pow.MainHeight]pow.PercentItem{
			pow.MainHeight(0): {
				pow.BLAKE2BD:      10,
				pow.CUCKAROO:      10,
				pow.CUCKATOO:      20,
				pow.CUCKAROOM:     10,
				pow.X16RV3:        10,
				pow.X8R16:         20,
				pow.MEERXKECCAKV1: 20,
			},
			pow.MainHeight(50): {
				pow.BLAKE2BD:         0,
				pow.CUCKAROO:         30,
				pow.CUCKATOO:         0,
				pow.CUCKAROOM:        30,
				pow.X16RV3:           10,
				pow.X8R16:            0,
				pow.QITMEERKECCAK256: 0,
				pow.MEERXKECCAKV1:    30,
			},
			pow.MainHeight(100): {
				pow.BLAKE2BD:      0,
				pow.CUCKAROO:      0,
				pow.CUCKATOO:      0,
				pow.CUCKAROOM:     70,
				pow.X16RV3:        0,
				pow.X8R16:         0,
				pow.MEERXKECCAKV1: 30,
			},
		},
		// after this height the big graph will be the main pow graph
		AdjustmentStartMainHeight: 45 * 1440 * 60 / privTargetTimePerBlock,
	},
	CoinbaseConfig: CoinbaseConfigs{
		{
			Height:  0,
			Version: "0.10.1",
		},
		{
			Height:                    10,
			Version:                   "0.10.2",
			ExtraDataIncludedVer:      true,
			ExtraDataIncludedNodeInfo: true,
		},
		{
			Height:                    20,
			Version:                   "0.10.3",
			ExtraDataIncludedVer:      true,
			ExtraDataIncludedNodeInfo: true,
		},
	},
	ReduceMinDifficulty:      false,
	MinDiffReductionTime:     0, // Does not apply since ReduceMinDifficulty false
	GenerateSupported:        true,
	MaximumBlockSizes:        []int{1000000, 1310720},
	MaxTxSize:                1000000,
	WorkDiffAlpha:            1,
	WorkDiffWindowSize:       160,
	WorkDiffWindows:          20,
	TargetTimePerBlock:       time.Second * privTargetTimePerBlock,
	TargetTimespan:           time.Second * privTargetTimePerBlock * 16, // TimePerBlock * WindowSize
	RetargetAdjustmentFactor: 2,

	// Subsidy parameters.
	BaseSubsidy:              50000000000,
	MulSubsidy:               100,
	DivSubsidy:               101,
	SubsidyReductionInterval: 128,
	WorkRewardProportion:     10,
	StakeRewardProportion:    0,
	BlockTaxProportion:       0,

	// Checkpoints ordered from oldest to newest.
	Checkpoints: nil,

	// Address encoding magics
	NetworkAddressPrefix: "R",
	PubKeyAddrID:         [2]byte{0x25, 0xe5}, // starts with Rk
	PubKeyHashAddrID:     [2]byte{0x0d, 0xf1}, // starts with Rm
	PKHEdwardsAddrID:     [2]byte{0x0d, 0xe0}, // starts with Re
	PKHSchnorrAddrID:     [2]byte{0x0d, 0xfe}, // starts with Rr
	ScriptHashAddrID:     [2]byte{0x0d, 0xc2}, // starts with RS
	PrivateKeyID:         [2]byte{0x22, 0xfe}, // starts with Pr

	// BIP32 hierarchical deterministic extended key magics
	HDPrivateKeyID: [4]byte{0x04, 0x0b, 0xee, 0x6e}, // starts with rprv
	HDPublicKeyID:  [4]byte{0x04, 0x0b, 0xf2, 0xa7}, // starts with rpub

	// BIP44 coin type used in the hierarchical deterministic path for
	// address generation.
	// TODO coin type
	HDCoinType: 223, // ASCII for s

	OrganizationPkScript:  hexMustDecode("76a91429209320e66d96839785dd07e643a7f1592edc5a88ac"),

	// Because it's only for testing, it comes from testwallet.go
	TokenAdminPkScript: hexMustDecode("00000000c96d6d76a914785bfbf4ecad8b72f2582be83616c5d364a3244288ac"),

	CoinbaseMaturity: 16,
}
