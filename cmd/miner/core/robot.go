// Copyright (c) 2019 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package core

import (
	"context"
	"github.com/Qitmeer/qitmeer/cmd/miner/common"
	"strings"
	"sync"
)

//var devicesTypesForMining = cl.DeviceTypeAll

type Robot interface {
	Run(ctx context.Context) // uses device to calulate the nonce
	ListenWork()             //listen the solo or pool work
	SubmitWork()             //submit the work
}

type MinerRobot struct {
	Cfg              *common.GlobalConfig //config
	ValidShares      uint64
	PendingShares    uint64
	StaleShares      uint64
	InvalidShares    uint64
	AllDiffOneShares uint64
	Wg               *sync.WaitGroup
	Started          uint32
	Quit             context.Context
	Work             *Work
	ClDevices        []string
	Rpc              *common.RpcClient
	Pool             bool
	SubmitStr        chan string
	UseDevices       []string
}

//init GPU device
func (this *MinerRobot) InitDevice() {
	this.UseDevices = []string{}
	if this.Cfg.OptionConfig.UseDevices != "" {
		this.UseDevices = strings.Split(this.Cfg.OptionConfig.UseDevices, ",")
	}
}
