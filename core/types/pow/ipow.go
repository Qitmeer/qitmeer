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

// the pow length is 178
const POW_LENGTH = 178

// proof data length is 169
const PROOFDATA_LENGTH = 169

type PowType byte
type PowBytes []byte

const (
	//pow type enum
	BLAKE2BD         PowType = 0
	CUCKAROO         PowType = 1
	CUCKATOO         PowType = 2
	CUCKAROOM        PowType = 3
	X16RV3           PowType = 4
	X8R16            PowType = 5
	QITMEERKECCAK256 PowType = 6
	CRYPTONIGHT      PowType = 7
	MEERXKECCAKV1    PowType = 8
)

var PowMapString = map[PowType]interface{}{
	BLAKE2BD:         "blake2bd",
	CUCKAROO:         "cuckaroo",
	CUCKATOO:         "cuckatoo",
	CUCKAROOM:        "cuckaroom",
	X16RV3:           "x16rv3",
	X8R16:            "x8r16",
	QITMEERKECCAK256: "qitmeer_keccak256",
	CRYPTONIGHT:      "cryptonight",
	MEERXKECCAKV1:    "meer_xkeccak_v1",
}

func GetPowName(powType PowType) string {
	val, ok := PowMapString[powType]
	if !ok {
		return ""
	}
	return val.(string)
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
	SetNonce(nonce uint64)
	//calc next diff
	GetNextDiffBig(weightedSumDiv *big.Int, oldDiffBig *big.Int, currentPowPercent *big.Int) *big.Int
	GetNonce() uint64
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
	SetMainHeight(height MainHeight)
	CheckAvailable() bool
	CompareDiff(newtarget *big.Int, target *big.Int) bool
	FindSolver(headerData []byte, blockHash hash.Hash, targetDiffBits uint32) bool
}

type Pow struct {
	PowType    PowType       //header pow type 1 bytes
	Nonce      uint64        //header nonce 4 bytes
	ProofData  ProofDataType // 1 edge_bits  168  bytes circle length total 169 bytes
	params     *PowConfig
	mainHeight MainHeight
}

//get pow instance
func GetInstance(powType PowType, nonce uint64, proofData []byte) IPow {
	var instance IPow
	switch powType {
	case BLAKE2BD:
		instance = &Blake2bd{}
	case X16RV3:
		instance = &X16rv3{}
	case X8R16:
		instance = &X8r16{}
	case QITMEERKECCAK256:
		instance = &QitmeerKeccak256{}
	case CUCKAROO:
		instance = &Cuckaroo{}
	case CUCKAROOM:
		instance = &Cuckaroom{}
	case CUCKATOO:
		instance = &Cuckatoo{}
	case CRYPTONIGHT:
		instance = &CryptoNight{}
	case MEERXKECCAKV1:
		instance = &MeerXKeccakV1{}
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

func (this *Pow) SetMainHeight(mainHeight MainHeight) {
	this.mainHeight = mainHeight
}

func (this *Pow) GetPowType() PowType {
	return this.PowType
}

func (this *Pow) GetNonce() uint64 {
	return this.Nonce
}

func (this *Pow) SetNonce(nonce uint64) {
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

func (this *Pow) PowPercent() *big.Int {
	targetPercent := big.NewInt(int64(this.params.GetPercentByHeightAndType(this.mainHeight, this.GetPowType())))
	targetPercent.Lsh(targetPercent, 32)
	return targetPercent
}

//check pow is available
func (this *Pow) CheckAvailable() bool {
	return this.params.GetPercentByHeightAndType(this.mainHeight, this.PowType) > 0
}
