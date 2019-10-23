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
// have for the test network. It is the value 2^232 - 1.
var	testMixNetPowLimit = new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 232), common.Big1)

// testPowNetParams defines the network parameters for the test network.
var MixNetParams = Params{
    Name:        "mixnet",
    Net:         protocol.MixNet,
    DefaultPort: "28132",
    DNSSeeds: []DNSSeed{
        {"mixnet-seed.hlcwallet.info", true},
        {"mixnet-seed.qitmeer.xyz", true},
        {"mixnet-seed.qitmeer.top", true},
    },

    // Chain parameters
    GenesisBlock:             &testPowNetGenesisBlock,
    GenesisHash:              &testPowNetGenesisHash,
    ReduceMinDifficulty:      false,
    MinDiffReductionTime:     0, // Does not apply since ReduceMinDifficulty false
    GenerateSupported:        true,
    PowConfig :&pow.PowConfig{
        Blake2bdPowLimit:      testMixNetPowLimit,
        Blake2bdPowLimitBits:  0x1e00ffff,
        Blake2bDPercent:       34,
        CuckarooPercent:       33,
        CuckatooPercent:       33,
        //hash ffffffffffffffff000000000000000000000000000000000000000000000000 corresponding difficulty is 48 for edge bits 24
        // Uniform field type uint64 value is 48 . bigToCompact the uint32 value
        // 24 edge_bits only need hash 1*4 times use for privnet if GPS is 2. need 50 /2 * 2 ≈ 1min find once
        CuckarooMinDifficulty:     0x1600000,
        CuckatooMinDifficulty:     0x1600000,
    },

    WorkDiffAlpha:            1,
    WorkDiffWindowSize:       60,
    WorkDiffWindows:          20,
    MaximumBlockSizes:        []int{1310720},
    MaxTxSize:                1000000,
    TargetTimePerBlock:       time.Minute * 1,
    TargetTimespan:           time.Minute * 1 * 1, // TimePerBlock * WindowSize
    RetargetAdjustmentFactor: 3,

    // Subsidy parameters.
    BaseSubsidy:              2500000000, // 25 Coin
    MulSubsidy:               100,
    DivSubsidy:               101,
    SubsidyReductionInterval: 2048,
    WorkRewardProportion:     10,
    StakeRewardProportion:    0,
    BlockTaxProportion:       0,

    // Checkpoints ordered from oldest to newest.
    Checkpoints: []Checkpoint{
    },

    // Consensus rule change deployments.
    //
    Deployments: map[uint32][]ConsensusDeployment{
    },

    // Address encoding magics
    NetworkAddressPrefix: "X",
    PubKeyAddrID:         [2]byte{0x11, 0x6e}, // starts with Xx
    PubKeyHashAddrID:     [2]byte{0x11, 0x53}, // starts with Xm
    PKHEdwardsAddrID:     [2]byte{0x11, 0x42}, // starts with Xe
    PKHSchnorrAddrID:     [2]byte{0x11, 0x5f}, // starts with Xr
    ScriptHashAddrID:     [2]byte{0x11, 0x24}, // starts with XS
    PrivateKeyID:         [2]byte{0x11, 0x64}, // starts with Xt

    // BIP32 hierarchical deterministic extended key magics
    HDPrivateKeyID: [4]byte{0x01, 0x9d, 0x0b, 0xe1}, // starts with xprv
    HDPublicKeyID:  [4]byte{0x01, 0x9d, 0x0d, 0x62}, // starts with xpub

    // BIP44 coin type used in the hierarchical deterministic path for
    // address generation.
    HDCoinType: 11,

    CoinbaseMaturity:        512,
    //OrganizationPkScript:  hexMustDecode("76a914868b9b6bc7e4a9c804ad3d3d7a2a6be27476941e88ac"),
}