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

// mainPowLimit is the highest proof of work value a block can
// have for the main network. It is the value 2^212 - 1.
// target 00000000000fffffffffffffffffffffffffffffffffffffffffffffffffffff
// Min Diff 17 T
// compact 0x1b0fffff
var mainPowLimit = new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 212), common.Big1)

// target time per block unit second(s)
const mainTargetTimePerBlock = 30

// Difficulty check interval is about 60*30 = 30 mins
const mainWorkDiffWindowSize = 60

// MainNetParams defines the network parameters for the main network.
var MainNetParams = Params{
	Name:           "mainnet",
	Net:            protocol.MainNet,
	DefaultPort:    "8130",
	DefaultUDPPort: 8140,
	Bootstrap: []string{
		"/dns4/node.meerscan.io/tcp/28130/p2p/16Uiu2HAmTdcrQ2S4MD6UxeR81Su8DQdt2eB7vLzJA7LrawNf93T2",
		"/dns4/ns-cn.qitmeer.xyz/tcp/18150/p2p/16Uiu2HAm45YEQXf5sYgpebp1NvPS96ypvvpz5uPx7iPHmau94vVk",
		"/dns4/ns.qitmeer.top/tcp/28230/p2p/16Uiu2HAmRtp5CjNv3WvPYuh7kNXXZQDYegwFFeDH9vWY3JY4JS1W",
		"/dns4/boot.qitmir.info/tcp/2001/p2p/16Uiu2HAmJ8qBBgoNoHH84ntLuXB9sqDngh82zZgaEejdFUYGR59Y",
	},
	LedgerParams: ledger.LedgerParams{
		GenesisAmountUnit: 1000 * 1e8,                              // 1000 MEER every utxo
		MaxLockHeight:     86400 / mixTargetTimePerBlock * 365 * 5, // max lock height
	},
	// Chain parameters
	GenesisBlock: &genesisBlock,
	GenesisHash:  &genesisHash,
	PowConfig: &pow.PowConfig{
		Blake2bdPowLimit:             mainPowLimit,
		Blake2bdPowLimitBits:         0x1b0fffff,
		X16rv3PowLimit:               mainPowLimit,
		X16rv3PowLimitBits:           0x1b0fffff,
		X8r16PowLimit:                mainPowLimit,
		X8r16PowLimitBits:            0x1b0fffff,
		QitmeerKeccak256PowLimit:     mainPowLimit,
		QitmeerKeccak256PowLimitBits: 0x1b0fffff,
		//hash ffffffffffffffff000000000000000000000000000000000000000000000000 corresponding difficulty is 48 for edge bits 24
		// Uniform field type uint64 value is 48 . bigToCompact the uint32 value
		// 24 edge_bits only need hash 1*4 times use for privnet if GPS is 2. need 50 /2 * 4 find once
		CuckarooMinDifficulty:     0x1300000 * 4,
		CuckaroomMinDifficulty:    0x1300000 * 4,
		CuckatooMinDifficulty:     0x1300000 * 4,
		MeerXKeccakV1PowLimit:     mainPowLimit,
		MeerXKeccakV1PowLimitBits: 0x1b0fffff,
		Percent: map[pow.MainHeight]pow.PercentItem{
			pow.MainHeight(0): {
				pow.MEERXKECCAKV1: 100,
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
	WorkDiffWindowSize:       mainWorkDiffWindowSize,
	WorkDiffWindows:          20, //
	MaximumBlockSizes:        []int{1310720},
	MaxTxSize:                1000000,
	TargetTimePerBlock:       time.Second * mainTargetTimePerBlock,
	TargetTimespan:           time.Second * mainTargetTimePerBlock * mainWorkDiffWindowSize, // TimePerBlock * WindowSize
	RetargetAdjustmentFactor: 2,

	// Subsidy parameters.
	BaseSubsidy:              10 * 1e8, // POW daily supply is almost 24*60*(60/30)*10 = 28880, ignore the DAG concurrent increment.
	MulSubsidy:               0,        // subsidy reduce to zero after 7986093 block created
	DivSubsidy:               101,
	SubsidyReductionInterval: 7986093, // (210240000-50518130)/2/BaseSubsidy = 7986093
	// total POW + locked supply almost = 79860930*2*10 = 159721860, ignore DAG increment.
	WorkRewardProportion:  10,
	StakeRewardProportion: 0,
	BlockTaxProportion:    0,

	// Checkpoints ordered from oldest to newest.
	Checkpoints: []Checkpoint{},

	// Consensus rule change deployments.
	//
	// The miner confirmation window is defined as:
	//   target proof of work timespan / target proof of work spacing
	RuleChangeActivationThreshold: 57,
	MinerConfirmationWindow:       mainWorkDiffWindowSize,
	Deployments: []ConsensusDeployment{
		DeploymentTestDummy: {
			BitNumber: 28,
		},
		DeploymentToken: {
			BitNumber:  0,
			StartTime:  0,
			ExpireTime: mainWorkDiffWindowSize * 2,
		},
	},

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

	CoinbaseMaturity: 720,

	OrganizationPkScript: hexMustDecode("76a914e99ebf409dda2a10ea9970651021d8e552f286de88ac"),
	TokenAdminPkScript: hexMustDecode("00000000c96d6d76a914b8834294977b26a44094fe2216f8a7d59af1130888ac"),

	// MmQitmeerMainNetGuardAddressXd7b76q
	GuardAddrPkScript: hexMustDecode("76a9143846e53e5e952b5cd6023e3ad3cfc75cb93fce0388ac"),
	// MmQitmeerMainNetHonorAddressXY9JH2y
	HonorAddrPkScript: hexMustDecode("76a9143846e53e5e952b5cd60240ad9c4cf6164dd5090988ac"),
}
