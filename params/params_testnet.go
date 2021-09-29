// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2014-2016 The btcsuite developers
// Copyright (c) 2015-2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package params

import (
	"github.com/Qitmeer/qitmeer/common"
	"github.com/Qitmeer/qitmeer/common/math"
	"github.com/Qitmeer/qitmeer/core/protocol"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"github.com/Qitmeer/qitmeer/ledger"
	"math/big"
	"time"
)

// testNetPowLimit is the highest proof of work value a block can
// have for the test network. It is the value 2^242 - 1.
var testNetPowLimit = new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 242), common.Big1)
var maxNetPowLimit = new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 0), common.Big1)

// target time per block unit second(s)
const testTargetTimePerBlock = 30

// Difficulty check interval is about 60*30 = 30 mins
const testWorkDiffWindowSize = 60

// TestNetParams defines the network parameters for the test network.
var TestNetParams = Params{
	Name:           "testnet",
	Net:            protocol.TestNet,
	DefaultPort:    "18150",
	DefaultUDPPort: 18160,
	Bootstrap: []string{
		"/dns4/node.meerscan.io/tcp/28130/p2p/16Uiu2HAmTdcrQ2S4MD6UxeR81Su8DQdt2eB7vLzJA7LrawNf93T2",
		"/dns4/ns-cn.qitmeer.xyz/tcp/18150/p2p/16Uiu2HAm45YEQXf5sYgpebp1NvPS96ypvvpz5uPx7iPHmau94vVk",
		"/dns4/ns.qitmeer.top/tcp/28230/p2p/16Uiu2HAmRtp5CjNv3WvPYuh7kNXXZQDYegwFFeDH9vWY3JY4JS1W",
		"/dns4/boot.qitmir.info/tcp/2001/p2p/16Uiu2HAmJ8qBBgoNoHH84ntLuXB9sqDngh82zZgaEejdFUYGR59Y",
	},
	LedgerParams: ledger.LedgerParams{ // lock tx release rule in genesis
		GenesisAmountUnit: 1000 * 1e8, // 1000 MEER every utxo
		MaxLockHeight:     2880 * 365 * 5,
	},
	// Chain parameters
	GenesisBlock: &testNetGenesisBlock,
	GenesisHash:  &testNetGenesisHash,
	PowConfig: &pow.PowConfig{
		Blake2bdPowLimit:             maxNetPowLimit,
		Blake2bdPowLimitBits:         0x0, // compact from of testNetPowLimit 0
		X16rv3PowLimit:               maxNetPowLimit,
		X16rv3PowLimitBits:           0x0, // compact from of testNetPowLimit 0
		X8r16PowLimit:                maxNetPowLimit,
		X8r16PowLimitBits:            0x0, // compact from of testNetPowLimit 0
		QitmeerKeccak256PowLimit:     maxNetPowLimit,
		QitmeerKeccak256PowLimitBits: 0x0, // compact from of testNetPowLimit 0
		MeerXKeccakV1PowLimit:        testNetPowLimit,
		MeerXKeccakV1PowLimitBits:    0x1f0198f2, // compact from of testNetPowLimit (2^242-1)
		//hash ffffffffffffffff000000000000000000000000000000000000000000000000 corresponding difficulty is 48 for edge bits 24
		// Uniform field type uint64 value is 48 . bigToCompact the uint32 value
		// 24 edge_bits only need hash 1*4 times use for privnet if GPS is 2. need 50 /2 * 4 = 1min find once
		CuckarooMinDifficulty:  0x87fffff, // diff: max int64
		CuckatooMinDifficulty:  0x87fffff, // diff: max int64
		CuckaroomMinDifficulty: 0x87fffff, // diff: max int64

		Percent: map[pow.MainHeight]pow.PercentItem{
			pow.MainHeight(0): {
				pow.MEERXKECCAKV1: 100,
			},
		},
		// after this height the big graph will be the main pow graph
		AdjustmentStartMainHeight: 365 * 1440 * 60 / testTargetTimePerBlock,
	},
	CoinbaseConfig: CoinbaseConfigs{
		{
			Height:                    61279,
			Version:                   "0.10.4",
			ExtraDataIncludedVer:      true,
			ExtraDataIncludedNodeInfo: true,
		},
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
	DivSubsidy:               10000000000000, //
	SubsidyReductionInterval: math.MaxInt64,
	WorkRewardProportion:     10,
	StakeRewardProportion:    0,
	BlockTaxProportion:       0,

	// Maturity
	CoinbaseMaturity: 720, // coinbase required 720 * 30 = 6 hours before repent

	// Checkpoints ordered from oldest to newest.
	Checkpoints: []Checkpoint{},

	// Address encoding magics
	NetworkAddressPrefix: "T",
	PubKeyAddrID:         [2]byte{0x28, 0xf5}, // starts with Tk
	PubKeyHashAddrID:     [2]byte{0x0f, 0x14}, // starts with Tn (to distinguish 0.9.x testnet)
	PKHEdwardsAddrID:     [2]byte{0x0f, 0x01}, // starts with Te
	PKHSchnorrAddrID:     [2]byte{0x0f, 0x1e}, // starts with Tr
	ScriptHashAddrID:     [2]byte{0x0e, 0xe2}, // starts with TS
	PrivateKeyID:         [2]byte{0x23, 0x0b}, // starts with Pt

	// BIP32 hierarchical deterministic extended key magics
	HDPrivateKeyID: [4]byte{0x04, 0x35, 0x83, 0x97}, // starts with tprv
	HDPublicKeyID:  [4]byte{0x04, 0x35, 0x87, 0xd1}, // starts with tpub

	// BIP44 coin type used in the hierarchical deterministic path for
	// address generation.
	HDCoinType:         223,
	OrganizationPkScript: hexMustDecode("76a91429209320e66d96839785dd07e643a7f1592edc5a88ac"),
	TokenAdminPkScript: hexMustDecode("00000000c96d6d76a914b8834294977b26a44094fe2216f8a7d59af1130888ac"),
}
