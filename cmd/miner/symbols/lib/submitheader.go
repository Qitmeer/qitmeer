package lib

import (
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer/cmd/miner/common"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"math/big"
)

type MinerBlockData struct {
	HeaderData  []byte
	TargetDiff  *big.Int
	Target2     []byte
	Exnonce2    string
	JobID       string
	HeaderBlock *types.BlockHeader
	Height      uint64
}

// Header structure of assembly pool
func BlockComputePoolData(b []byte) []byte {
	//the qitmeer order
	nonce := hex.EncodeToString(b[NONCESTART:NONCEEND])
	powType := hex.EncodeToString(b[POWTYPE_START:POWTYPE_END])
	ntime := hex.EncodeToString(b[TIMESTART:TIMEEND])
	nbits := hex.EncodeToString(b[NBITSTART:NBITEND])
	state := hex.EncodeToString(b[STATESTART:STATEEND])
	merkle := hex.EncodeToString(b[MERKLESTART:MERKLEEND])
	prevhash := hex.EncodeToString(b[PRESTART:PREEND])
	version := hex.EncodeToString(b[VERSIONSTART:VERSIONEND])
	//the pool order
	header := nonce + powType + ntime + nbits + state + merkle + prevhash + version

	bb, _ := hex.DecodeString(header)
	bb = common.Reverse(bb)
	return bb
}

//the pool work submit structure
func (this *MinerBlockData) PackagePoolHeader(work *QitmeerWork, powType pow.PowType) {
	this.HeaderData = BlockComputePoolData(work.PoolWork.WorkData) // 128
	this.TargetDiff = work.stra.Target
	nbitesBy, _ := hex.DecodeString(fmt.Sprintf("%064x", this.TargetDiff))
	this.Target2 = common.Reverse(nbitesBy[0:32])
	copy(this.HeaderData[NONCESTART:NONCEEND], nbitesBy[:])
	instance := pow.GetInstance(powType, 0, []byte{})
	proofData, _ := hex.DecodeString(instance.GetProofData())
	this.HeaderData = append(this.HeaderData, proofData...)
	this.JobID = work.PoolWork.JobID
	this.HeaderBlock = &types.BlockHeader{}
	_ = ReadBlockHeader(this.HeaderData, this.HeaderBlock)
	this.Height = uint64(work.PoolWork.Height)
}

//the solo work submit structure
func (this *MinerBlockData) PackageRpcHeader(work *QitmeerWork) {
	bitesBy, _ := hex.DecodeString(work.Block.Target)
	this.Target2 = common.Reverse(bitesBy[0:32])
	bitesBy = common.Reverse(bitesBy[:8])
	b1, _ := hex.DecodeString(work.Block.Target)
	var r [32]byte
	copy(r[:], common.Reverse(b1)[:])
	r1 := hash.Hash(r)
	this.TargetDiff = pow.HashToBig(&r1)
	this.HeaderData = make([]byte, 117)
	copy(this.HeaderData, work.Block.WorkData)
}

func (this *MinerBlockData) BlockData() []byte {
	return this.HeaderData
}
