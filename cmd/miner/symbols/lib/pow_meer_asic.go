//+build asic

/**
Qitmeer
james
*/
package lib

/*
#include "../../asic/meer/main.h"
#include "../../asic/meer/main.c"
#include "../../asic/meer/algo_meer.c"
#include "../../asic/meer/meer.h"
#include "../../asic/meer/meer_drv.c"
#include "../../asic/meer/meer_drv.h"
#include "../../asic/meer/uart.h"
#include "../../asic/meer/uart.c"
#cgo CFLAGS: -Wno-unused-result
#cgo CFLAGS: -Wno-int-conversion
*/
import "C"
import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer/cmd/miner/common"
	"github.com/Qitmeer/qitmeer/cmd/miner/core"
	"github.com/Qitmeer/qng-core/common/hash"
	"github.com/Qitmeer/qng-core/core/types/pow"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
)

type MeerCrypto struct {
	core.Device
	Work   *QitmeerWork
	header MinerBlockData
}

func (this *MeerCrypto) InitDevice() {
	common.MinerLoger.Debug("==============Mining MeerCrypto ==============",
		"chips num", this.Cfg.OptionConfig.NumOfChips, "UART", this.UartPath, "NonceStart", this.NonceStart)
}

func (this *MeerCrypto) UpdateWork() {
	// update coinbase tx hash
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

type MiningResultItem struct {
	Nonce  uint64
	JobId  byte
	ChipId byte
}

type Work struct {
	ChipId    byte
	Height    uint64
	Header    []byte
	Target    *big.Int
	SubmitStr string
	GBTID     int64
	Pool      bool
}

type MiningResult map[uint64]MiningResultItem

func (this *MeerCrypto) Mine(wg *sync.WaitGroup) {
	start := false
	fd := 0
	arr := strings.Split(this.UartPath, ":")
	uartPath := C.CString(arr[0])
	gpio := C.CString(arr[1])
	defer func() {
		// recover from panic caused by writing to a closed channel
		if fd > 0 {
			common.MinerLoger.Info(fmt.Sprintf("[%s][meer_drv_deinit] miner chips exit", this.UartPath))
			C.meer_drv_deinit((C.int)(fd), gpio)
			C.free(unsafe.Pointer(uartPath))
			C.free(unsafe.Pointer(gpio))
		}

		wg.Done()
		this.Release()
	}()
	nonceBytes := make([]byte, 8) // nonce bytes
	var w core.BaseWork
	for {
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
		reject := 0
		this.HasNewWork = false
		this.CurrentWorkID = 0
		this.header = MinerBlockData{
			HeaderData: make([]byte, 0),
			TargetDiff: &big.Int{},
			JobID:      "",
		}
		t1 := time.Now().Nanosecond()
		common.MinerLoger.Debug("receive new task")
	gotoWork:
		for !this.HasNewWork && !this.ForceStop {
			// if has new work ,current calc stop
			select {
			case <-this.Quit.Done():
				common.MinerLoger.Debug("mining service exit")
				return
			default:
				if !start && fd == 0 {
					// init chips
					fd = int(C.init_drv((C.int)(this.Cfg.OptionConfig.NumOfChips), uartPath, gpio))
					if fd <= 0 {
						this.SetIsValid(false)
						return
					}
					start = true
					// set freqs
					freqsArr := strings.Split(this.Cfg.OptionConfig.Freqs, "|")
					for k := 0; k < len(freqsArr); k++ {
						arr := strings.Split(freqsArr[k], ",")
						if len(arr) != 2 {
							continue
						}
						microTime, err := strconv.Atoi(arr[0])
						if err != nil {
							common.MinerLoger.Error("freqs setting error", "value", freqsArr[k])
							return
						}
						freqVal, err := strconv.Atoi(arr[1])
						if err != nil {
							common.MinerLoger.Error("freqs setting error", "value", freqsArr[k])
							return
						}
						C.meer_drv_set_freq((C.int)(fd), (C.uint)(freqVal))
						time.Sleep(time.Duration(microTime) * time.Millisecond)
					}
					this.Started = time.Now().Unix()
					this.IsRunning = true
				}
				nonces := MiningResult{}
				works := map[byte]Work{}

				for j := 1; j <= this.Cfg.OptionConfig.NumOfChips; j++ {
					this.UpdateWork()
					nonceStart := this.NonceStart + uint64(j-1)*this.NonceStep
					nonceA := make([]byte, 8)
					nonceB := make([]byte, 8)
					nonceC := make([]byte, 8)
					binary.LittleEndian.PutUint64(nonceA, nonceStart/3*0)
					binary.LittleEndian.PutUint64(nonceB, nonceStart/3*1)
					binary.LittleEndian.PutUint64(nonceC, nonceStart/3*2)
					if this.IsRunning && common.CurrentHeight > 0 && this.header.Height > 0 && this.header.Height != common.CurrentHeight {
						common.MinerLoger.Warn("current work is stale", "height",
							this.header.Height, "cheight", common.CurrentHeight)
						break gotoWork
					}
					jid := int64(0)
					if !this.Pool {
						jid = this.Work.Block.GBTID
					}
					works[byte(j)] = Work{
						ChipId:    byte(j),
						Height:    this.header.Height,
						Header:    make([]byte, 117),
						Target:    this.header.TargetDiff,
						SubmitStr: this.GetSubmitStr(),
						GBTID:     jid,
						Pool:      this.Pool,
					}
					copy(works[byte(j)].Header[0:117], this.header.BlockData())
					C.set_work(
						(C.int)(fd),
						(*C.uchar)(unsafe.Pointer(&works[byte(j)].Header[0])),
						(C.int)(len(works[byte(j)].Header)),
						(*C.uchar)(unsafe.Pointer(&this.header.Target2[0])),
						(C.int)(j),
						(*C.uchar)(unsafe.Pointer(&nonceA[0])),
						(*C.uchar)(unsafe.Pointer(&nonceB[0])),
						(*C.uchar)(unsafe.Pointer(&nonceC[0])),
					)
				}
				t2 := time.Now().Nanosecond()
				common.MinerLoger.Debug("Notify New Task To Chips",
					"spent seconds(s)", float64(t2-t1)/1000000000.00)
				// set work
				t1 := time.Now().Unix()

				// 10 mill second next task
				for time.Now().Unix()-t1 < int64(this.Cfg.OptionConfig.Timeout) && !this.HasNewWork && !this.ForceStop {
					select {
					case <-this.Quit.Done():
						common.MinerLoger.Debug("mining service exit")
						return
					default:
					}
					chipId := make([]byte, 1)
					jobId := make([]byte, 1)
					nonceBytes = make([]byte, 8)
					if fd != 0 && C.get_nonce((C.int)(fd),
						(*C.uchar)(unsafe.Pointer(&nonceBytes[0])),
						(*C.uchar)(unsafe.Pointer(&chipId[0])),
						(*C.uchar)(unsafe.Pointer(&jobId[0])),
					) {
						if chipId[0] < 1 || chipId[0] > byte(this.Cfg.OptionConfig.NumOfChips) {
							time.Sleep(10 * time.Millisecond)
							continue
						}
						cwork := works[chipId[0]]
						lastNonce := binary.LittleEndian.Uint64(nonceBytes)
						if _, ok := nonces[lastNonce]; !ok {
							nonces[lastNonce] = MiningResultItem{
								Nonce:  lastNonce,
								JobId:  jobId[0],
								ChipId: chipId[0],
							}
							copy(cwork.Header[NONCESTART:NONCEEND], nonceBytes)
							h := hash.HashMeerXKeccakV1(cwork.Header[:117])
							if pow.HashToBig(&h).Cmp(cwork.Target) <= 0 {
								common.MinerLoger.Debug(fmt.Sprintf("[%s]ChipId #%d JobId #%d Height #%d Found hash : %s nonce:%s target:%064x",
									this.UartPath,
									chipId[0],
									jobId[0],
									cwork.Height,
									h,
									hex.EncodeToString(nonceBytes), cwork.Target))
								this.AllDiffOneShares++
								this.SubmitData <- cwork.ReplaceNonce(nonceBytes)
							} else {
								// 1T
								if this.GetDiff() > 1e12 {
									reject++
									if reject >= this.Cfg.OptionConfig.NumOfChips*1000 {
										common.MinerLoger.Warn("Chips return exception,wait init again", "ChipID", this.UartPath)
										C.meer_drv_deinit((C.int)(fd), gpio)
										start = false
										fd = 0
										reject = 0
										break gotoWork
									}
								}
							}
						} else {
							common.MinerLoger.Debug(fmt.Sprintf("[%s][DUP Shares]ChipId #%d JobId #%d nonce:%d  Last ChipId: %d Last JobId :%d ",
								this.UartPath,
								chipId[0], jobId[0],
								lastNonce,
								nonces[lastNonce].ChipId, nonces[lastNonce].JobId))
						}
						time.Sleep(10 * time.Millisecond)
					}
				}
			}
		}
	}
}

func (this *MeerCrypto) GetSubmitStr() string {
	headerData := this.header.BlockData()
	subm := hex.EncodeToString(headerData)
	if this.Pool {
		subm += "-" + this.header.JobID + "-" + this.header.Exnonce2
	}
	return subm
}

func (this *MeerCrypto) Status(wg *sync.WaitGroup) {
	common.MinerLoger.Info("start listen hashrate")
	t := time.NewTicker(time.Second * 10)
	defer t.Stop()
	defer wg.Done()
	for {
		select {
		case <-this.Quit.Done():
			common.MinerLoger.Debug("device stats service exit")
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
			hashrate := float64(this.AllDiffOneShares) / float64(secondsElapsed) * diff
			// diff
			unit := "H/s"
			start := time.Unix(this.Started, 0)
			common.MinerLoger.Info(fmt.Sprintf("[%s]Start time: %s  Diff: %s All Shares: %d HashRate: %s",
				this.UartPath,
				start.Format("2006-01-02 15:04:05"),
				common.FormatHashRate(diff, unit),
				this.AllDiffOneShares,
				common.FormatHashRate(hashrate, unit)))
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

func (this *Work) ReplaceNonce(nonce []byte) string {
	arr := strings.Split(this.SubmitStr, "-")
	b, err := hex.DecodeString(arr[0])
	if err != nil {
		return this.SubmitStr
	}
	copy(b[0:117], this.Header)
	copy(b[NONCESTART:NONCEEND], nonce)
	arr[0] = hex.EncodeToString(b)
	subm := strings.Join(arr, "-")
	if !this.Pool {
		subm += "-" +
			fmt.Sprintf("%d", this.GBTID)
	}
	return subm
}
