/**
Qitmeer
james
*/
package core

import (
	"context"
	"fmt"
	"github.com/Qitmeer/qitmeer/cmd/miner/common"
	"math"
	"sync"
	"time"
)

type BaseDevice interface {
	Mine(wg *sync.WaitGroup)
	Update()
	InitDevice()
	Status(wg *sync.WaitGroup)
	GetIsValid() bool
	SetIsValid(valid bool)
	GetMinerId() int
	GetAverageHashRate() float64
	GetName() string
	GetStart() uint64
	SetPool(pool bool)
	SetNewWork(w BaseWork)
	SetForceUpdate(force bool)
	Release()
	GetMinerType() string
	SubmitShare(substr chan string)
	StopTask()
	GetIsRunning() bool
}
type Device struct {
	Cfg              *common.GlobalConfig //must init
	DeviceName       string
	HasNewWork       bool
	ForceStop        bool
	AllDiffOneShares uint64
	AverageHashRate  float64
	MinerId          uint32
	NonceStep        uint64
	NonceStart       uint64
	LocalItemSize    int
	NonceOut         []byte
	Started          int64
	GlobalItemSize   int
	CurrentWorkID    uint64
	Quit             context.Context //must init
	sync.Mutex
	Wg           sync.WaitGroup
	Pool         bool        //must init
	IsValid      bool        //is valid
	SubmitData   chan string //must
	NewWork      chan BaseWork
	Err          error
	MiningType   string
	UartPath     string
	StopTaskChan chan bool
	IsRunning    bool
}

func (this *Device) Init(i int, pool bool, ctx context.Context, cfg *common.GlobalConfig, allCount uint64) {
	this.MinerId = uint32(i)
	this.NewWork = make(chan BaseWork, 1)
	this.Cfg = cfg
	this.DeviceName = "CPU Miner"
	this.CurrentWorkID = 0
	this.IsValid = true
	this.Pool = pool
	this.SubmitData = make(chan string, 1)
	this.Quit = ctx
	this.AllDiffOneShares = 0
	this.NonceStep = math.MaxUint64 / allCount
	this.NonceStart = this.NonceStep * uint64(i)
	this.StopTaskChan = make(chan bool, 1)
}

func (this *Device) GetIsValid() bool {
	return this.IsValid
}

func (this *Device) GetIsRunning() bool {
	return this.IsRunning
}

func (this *Device) SetNewWork(work BaseWork) {
	if !this.GetIsValid() {
		return
	}
	this.HasNewWork = true
	this.NewWork <- work
}

func (this *Device) StopTask() {
	if !this.GetIsValid() {
		return
	}
	this.StopTaskChan <- true
}

func (this *Device) SetForceUpdate(force bool) {
	if !this.GetIsValid() {
		return
	}
	this.ForceStop = force
}

func (this *Device) GetMinerType() string {
	return this.MiningType
}

func (this *Device) Update() {
	this.CurrentWorkID = common.RandUint64()
}

func (this *Device) InitDevice() {
}

func (this *Device) SetPool(b bool) {
	this.Pool = b
}

func (this *Device) SetIsValid(valid bool) {
	this.IsValid = valid
}

func (this *Device) GetMinerId() int {
	return int(this.MinerId)
}

func (this *Device) GetIntensity() int {
	return int(math.Log2(float64(this.GlobalItemSize)))
}

func (this *Device) GetWorkSize() int {
	return this.LocalItemSize
}

func (this *Device) SetWorkSize(size int) {
	this.LocalItemSize = size
}

func (this *Device) GetName() string {
	return this.DeviceName
}

func (this *Device) GetStart() uint64 {
	return uint64(this.Started)
}

func (this *Device) GetAverageHashRate() float64 {
	return this.AverageHashRate
}

func (d *Device) Release() {
}

func (this *Device) Status(wg *sync.WaitGroup) {
	common.MinerLoger.Info("start listen hashrate")
	t := time.NewTicker(time.Second * 10)
	defer t.Stop()
	defer wg.Done()
	for {
		select {
		case <-this.Quit.Done():
			common.MinerLoger.Info("device stats service exit")
			return
		case <-t.C:
			if !this.IsValid {
				return
			}
			secondsElapsed := time.Now().Unix() - this.Started
			//diffOneShareHashesAvg := uint64(0x00000000FFFFFFFF)
			if this.AllDiffOneShares <= 0 || secondsElapsed <= 0 {
				continue
			}
			averageHashRate := float64(this.AllDiffOneShares) /
				float64(secondsElapsed)
			if averageHashRate <= 0 {
				continue
			}
			if this.AverageHashRate <= 0 {
				this.AverageHashRate = averageHashRate
			}
			//recent stats 95% percent
			this.AverageHashRate = (this.AverageHashRate*50 + averageHashRate*950) / 1000
			unit := "H/s"
			if this.GetMinerType() != "blake2bd" && this.GetMinerType() != "keccak256" && this.GetMinerType() != "meer_crypto" {
				unit = " GPS"
			}
			common.MinerLoger.Info(fmt.Sprintf("# %d : %s", this.MinerId, common.FormatHashRate(this.AverageHashRate, unit)))
		}
	}
}

func (this *Device) SubmitShare(substr chan string) {
	if !this.GetIsValid() {
		return
	}
	defer func() {
		close(this.SubmitData)
		// recover from panic caused by writing to a closed channel
		if r := recover(); r != nil {
			common.MinerLoger.Debug(fmt.Sprintf("# %d submit service exit", this.MinerId))
			return
		}
		common.MinerLoger.Debug(fmt.Sprintf("# %d submit service exit", this.MinerId))
	}()
	for {
		select {
		case <-this.Quit.Done():
			return
		case str := <-this.SubmitData:
			if this.HasNewWork {
				//the stale submit
				continue
			}
			substr <- str
		}
	}
}
