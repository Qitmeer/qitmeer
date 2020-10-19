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

// testNetPowLimit is the highest proof of work value a block can
// have for the test network. It is the value 2^208 - 1.
var testNetPowLimit = new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 208), common.Big1)

// target time per block unit second(s)
const testTargetTimePerBlock = 30

// Difficulty check interval is about 60*30 = 30 mins
const testWorkDiffWindowSize = 60

// TestNetParams defines the network parameters for the test network.
var TestNetParams = Params{
	Name:        "testnet",
	Net:         protocol.TestNet,
	DefaultPort: "18130",
	DNSSeeds: []DNSSeed{
		{"testnet-seed.hlcwallet.info", true},
		{"testnet-seed.qitmeer.xyz", true},
		{"seed.qitmir.info", true},
	},

	// Chain parameters
	GenesisBlock: &testNetGenesisBlock,
	GenesisHash:  &testNetGenesisHash,
	PowConfig: &pow.PowConfig{
		Blake2bdPowLimit:             testNetPowLimit,
		Blake2bdPowLimitBits:         0x1b7fffff, // compact from of testNetPowLimit (2^215-1)
		X16rv3PowLimit:               testNetPowLimit,
		X16rv3PowLimitBits:           0x1b7fffff, // compact from of testNetPowLimit (2^215-1)
		X8r16PowLimit:                testNetPowLimit,
		X8r16PowLimitBits:            0x1b7fffff, // compact from of testNetPowLimit (2^215-1)
		QitmeerKeccak256PowLimit:     testNetPowLimit,
		QitmeerKeccak256PowLimitBits: 0x1b00ffff, // compact from of testNetPowLimit (2^208-1) 453050367
		//hash ffffffffffffffff000000000000000000000000000000000000000000000000 corresponding difficulty is 48 for edge bits 24
		// Uniform field type uint64 value is 48 . bigToCompact the uint32 value
		// 24 edge_bits only need hash 1*4 times use for privnet if GPS is 2. need 50 /2 * 4 = 1min find once
		CuckarooMinDifficulty:  0x2018000, // 96 * 4 = 384
		CuckatooMinDifficulty:  0x2074000, // 1856
		CuckaroomMinDifficulty: 0x34ad1ec, // compact : 55235052 diff : 4903404

		Percent: map[pow.MainHeight]pow.PercentItem{
			pow.MainHeight(0): {
				pow.BLAKE2BD:         0,
				pow.X16RV3:           0,
				pow.QITMEERKECCAK256: 30,
				pow.CUCKAROOM:        70,
				pow.CUCKATOO:         0,
			},
			// | time	| timestamp	| mainHeight |
			// | ---| --- | --- |
			// | 2020-08-30 10:31:46 | 1598754706 | 192266
			// | 2020-09-15 12:00 | 1600142400 | 238522
			// The soft forking mainHeight was calculated according to the average time of 30s
			// In other words, pmeer will be produced by the pow of QitmeerKeccak256 only after mainHeight arrived 238522
			pow.MainHeight(238522): {
				pow.BLAKE2BD:         0,
				pow.X16RV3:           0,
				pow.QITMEERKECCAK256: 100,
				pow.CUCKAROOM:        0,
				pow.CUCKATOO:         0,
			},
		},
		// after this height the big graph will be the main pow graph
		AdjustmentStartMainHeight: 365 * 1440 * 60 / testTargetTimePerBlock,
	},
	ReduceMinDifficulty:      false,
	MinDiffReductionTime:     0, // Does not apply since ReduceMinDifficulty false
	GenerateSupported:        true,
	WorkDiffAlpha:            1,
	WorkDiffWindowSize:       testWorkDiffWindowSize,
	WorkDiffWindows:          20,
	MaximumBlockSizes:        []int{1310720},
	MaxTxSize:                1000000,
	TargetTimePerBlock:       time.Second * testTargetTimePerBlock,
	TargetTimespan:           time.Second * testTargetTimePerBlock * testWorkDiffWindowSize, // TimePerBlock * WindowSize
	RetargetAdjustmentFactor: 2,                                                             // equal to 2 hour vs. 4

	// Subsidy parameters.
	BaseSubsidy:              12000000000, // 120 Coin , daily supply is 120*2*60*24 = 345600 ~ 345600 * 2 (DAG factor)
	MulSubsidy:               100,
	DivSubsidy:               10000000000000,   // Coin-base reward reduce to zero at 1540677 blocks created
	SubsidyReductionInterval: 1669066 - 541194, // 120 * 1669066 (blocks) *= 200287911 (200M) -> 579 ~ 289 days
	// && subsidy has to reduce the 0.8.5 mining_rewarded blocks (541194)
	WorkRewardProportion:  10,
	StakeRewardProportion: 0,
	BlockTaxProportion:    0,

	// Maturity
	CoinbaseMaturity: 720, // coinbase required 720 * 30 = 6 hours before repent

	// Checkpoints ordered from oldest to newest.
	Checkpoints: []Checkpoint{},

	// Consensus rule change deployments.
	//
	Deployments: map[uint32][]ConsensusDeployment{},

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
	HDCoinType: 223,

	//OrganizationPkScript:  hexMustDecode("76a914868b9b6bc7e4a9c804ad3d3d7a2a6be27476941e88ac"),
}
