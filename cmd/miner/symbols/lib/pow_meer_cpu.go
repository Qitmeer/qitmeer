//+build !asic

/**
Qitmeer
james
*/
package lib

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer/cmd/miner/common"
	"github.com/Qitmeer/qitmeer/cmd/miner/core"
	"github.com/Qitmeer/qng-core/common/hash"
	"github.com/Qitmeer/qng-core/core/types"
	"github.com/Qitmeer/qng-core/core/types/pow"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"
)

type MeerCrypto struct {
	core.Device
	Work   *QitmeerWork
	header MinerBlockData
}

func (this *MeerCrypto) InitDevice() {
	this.Started = time.Now().Unix()
	common.MinerLoger.Debug(fmt.Sprintf("CPUMiner [%d] NonceStart %d==============Mining MeerCrypto==============", this.MinerId, this.NonceStart))
}

func (this *MeerCrypto) Update() {
	//update coinbase tx hash
	this.Device.Update()
	if this.Pool {
		this.Work.PoolWork.ExtraNonce2 = fmt.Sprintf("%08x", this.CurrentWorkID<<this.MinerId)[:8]
		this.header.Exnonce2 = this.Work.PoolWork.ExtraNonce2
		this.Work.PoolWork.WorkData = this.Work.PoolWork.PrepQitmeerWork()
		this.header.PackagePoolHeader(this.Work, pow.MEERXKECCAKV1)
	} else {
		this.header.PackageRpcHeader(this.Work)
	}
}

func (this *MeerCrypto) Mine(wg *sync.WaitGroup) {
	defer wg.Done()
	defer this.Release()
	var w core.BaseWork
	this.Started = time.Now().Unix()
	this.AllDiffOneShares = 0
	this.IsRunning = true
	for {
		this.AllDiffOneShares = 0
		select {
		case w = <-this.NewWork:
			this.Work = w.(*QitmeerWork)
		case <-this.Quit.Done():
			common.MinerLoger.Debug("mining service exit")
			return
		}
		if !this.IsValid {
			return
		}
		if this.ForceStop {
			continue
		}
		if !this.HasNewWork || this.Work == nil {
			continue
		}

		if len(this.Work.PoolWork.WorkData) <= 0 && this.Pool {
			continue
		}
		this.HasNewWork = false
		this.CurrentWorkID = 0
		this.header = MinerBlockData{
			HeaderData: make([]byte, 0),
			TargetDiff: &big.Int{},
			JobID:      "",
		}
		nonce := this.NonceStart
		this.Started = time.Now().Unix()
		this.Update()
		for {
			select {
			case <-this.Quit.Done():
				common.MinerLoger.Debug("mining service exit")
				return
			default:
			}
			// if has new work ,current calc stop
			if this.HasNewWork || this.ForceStop {
				break
			}

			hData := make([]byte, 128)
			copy(hData[0:types.MaxBlockHeaderPayload-pow.PROOFDATA_LENGTH], this.header.BlockData())
			nonce++
			this.AllDiffOneShares++
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, nonce)
			copy(hData[NONCESTART:NONCEEND], b)
			h := hash.HashMeerXKeccakV1(hData[:117])
			if pow.HashToBig(&h).Cmp(this.header.TargetDiff) <= 0 {
				common.MinerLoger.Debug(fmt.Sprintf("device #%d found hash : %s nonce:%d target:%064x", this.MinerId, h, nonce, this.header.TargetDiff))
				subm := hex.EncodeToString(hData[:117])
				if !this.Pool {
					subm += "-" + fmt.Sprintf("%d", this.Work.Block.GBTID)
				} else {
					subm += "-" + this.header.JobID + "-" + this.header.Exnonce2
				}
				this.SubmitData <- subm
			}
		}
	}
}

func (this *MeerCrypto) GetDiff() float64 {
	s := fmt.Sprintf("%064x", this.header.TargetDiff)
	diff := float64(1)
	for i := 0; i < 64; i++ {
		if strings.ToLower(s[i:i+1]) == "f" {
			break
		}
		a, _ := strconv.ParseInt(s[i:i+1], 16, 64)
		diff *= 16 / float64(a+1)
		if strings.ToLower(s[i:i+1]) != "0" {
			break
		}
	}
	common.MinerLoger.Debug("[current target]", "value", s, "diff", diff/1e9)
	return diff
}
func (this *MeerCrypto) Status(wg *sync.WaitGroup) {
	common.MinerLoger.Info("start listen hashrate")
	t := time.NewTicker(time.Second * 20)
	defer t.Stop()
	defer wg.Done()
	for {
		select {
		case <-this.Quit.Done():
			common.MinerLoger.Debug(fmt.Sprintf("# %d device stats service exit", this.MinerId))
			return
		case <-t.C:
			if !this.IsValid {
				return
			}
			secondsElapsed := time.Now().Unix() - this.Started
			if this.AllDiffOneShares <= 0 || secondsElapsed <= 0 {
				continue
			}
			diff := this.GetDiff()
			hashrate := float64(this.AllDiffOneShares) / float64(secondsElapsed)
			mayBlockTime := diff / hashrate // sec
			hour := mayBlockTime / 3600     // hour
			// diff
			unit := "H/s"
			start := time.Unix(this.Started, 0)
			common.MinerLoger.Info(fmt.Sprintf("# %d Start time: %s  Diff: %s HashRate: %s may-block-out-per %.2f hour",
				this.MinerId,
				start.Format(time.RFC3339),
				common.FormatHashRate(diff, unit),
				common.FormatHashRate(hashrate, unit), hour))
		}
	}
}
