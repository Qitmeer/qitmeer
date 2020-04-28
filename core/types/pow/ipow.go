// Copyright (c) 2017-2020 The qitmeer developers
// license that can be found in the LICENSE file.
// Reference resources of rust bitVector
package pow

import (
	"encoding/hex"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/json"
	"math/big"
)

// proof data length 188
const POW_LENGTH = 174

//except pow type 4bytes and nonce 8 bytes 176 bytes
const PROOFDATA_LENGTH = 169

type PowType byte
type PowBytes []byte

const (
	//pow type enum
	BLAKE2BD  PowType = 0
	CUCKAROO  PowType = 1
	CUCKATOO  PowType = 2
	CUCKAROOM PowType = 3
)

var PowMapString = map[PowType]interface{}{
	BLAKE2BD:  "blake2bd",
	CUCKAROO:  "cuckaroo",
	CUCKATOO:  "cuckatoo",
	CUCKAROOM: "cuckaroom",
}

type ProofDataType [PROOFDATA_LENGTH]byte

func (this *ProofDataType) String() string {
	return hex.EncodeToString(this[:])
}

func (this *ProofDataType) Bytes() []byte {
	return this[:]
}

type IPow interface {
	// verify result difficulty
	Verify(headerData []byte, blockHash hash.Hash, targetDiff uint32) error
	//set header nonce
	SetNonce(nonce uint32)
	//calc next diff
	GetNextDiffBig(weightedSumDiv *big.Int, oldDiffBig *big.Int, currentPowPercent *big.Int) *big.Int
	GetNonce() uint32
	GetPowType() PowType
	//set pow type
	SetPowType(powType PowType)
	GetProofData() string
	//set proof data
	SetProofData([]byte)
	Bytes() PowBytes
	BlockData() PowBytes
	//if cur_reduce_diff > 0 compare cur_reduce_diff with powLimitBits or minDiff ，the cur_reduce_diff should less than powLimitBits , and should more than min diff
	//if cur_reduce_diff <=0 return powLimit or min diff
	GetSafeDiff(cur_reduce_diff uint64) *big.Int
	// pow percent
	PowPercent() *big.Int
	//pow result
	GetPowResult() json.PowResult
	//SetParams
	SetParams(params *PowConfig)
	//SetHeight
	SetMainHeight(height int64)
	CheckAvailable() bool
	CompareDiff(newtarget *big.Int, target *big.Int) bool
}

type Pow struct {
	PowType    PowType       //header pow type 1 bytes
	Nonce      uint32        //header nonce 4 bytes
	ProofData  ProofDataType // 1 edge_bits  168  bytes circle length total 169 bytes
	params     *PowConfig
	mainHeight int64
}

//get pow instance
func GetInstance(powType PowType, nonce uint32, proofData []byte) IPow {
	var instance IPow
	switch powType {
	case BLAKE2BD:
		instance = &Blake2bd{}
	case CUCKAROO:
		instance = &Cuckaroo{}
	case CUCKAROOM:
		instance = &Cuckaroom{}
	case CUCKATOO:
		instance = &Cuckatoo{}
	default:
		instance = &Blake2bd{}
	}
	instance.SetPowType(powType)
	instance.SetNonce(nonce)
	instance.SetProofData(proofData)
	return instance
}

func (this *Pow) SetPowType(powType PowType) {
	this.PowType = powType
}

func (this *Pow) SetParams(params *PowConfig) {
	this.params = GetPowConfig().Set(params)
}

func (this *Pow) SetMainHeight(mainHeight int64) {
	this.mainHeight = mainHeight
}

func (this *Pow) GetPowType() PowType {
	return this.PowType
}

func (this *Pow) GetNonce() uint32 {
	return this.Nonce
}

func (this *Pow) SetNonce(nonce uint32) {
	this.Nonce = nonce
}

func (this *Pow) GetProofData() string {
	return this.ProofData.String()
}

//set proof data except pow type
func (this *Pow) SetProofData(data []byte) {
	l := len(data)
	copy(this.ProofData[0:l], data[:])
}
