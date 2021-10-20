// Copyright (c) 2019 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import (
	"context"
	"github.com/Qitmeer/qitmeer/cmd/miner/common"
	"github.com/Qitmeer/qitmeer/cmd/miner/core"
	qitmeer "github.com/Qitmeer/qitmeer/cmd/miner/symbols/lib"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"
)

var robotminer core.Robot

//init the config file
func init() {
	cfg, _, err := common.LoadConfig()
	if err != nil {
		log.Fatal("[error] config error,please check it.[", err, "]")
		return
	}
	//init miner robot
	robotminer = GetRobot(cfg)
}

func main() {
	common.UpdateSysTime()
	// Use all processor cores.
	runtime.GOMAXPROCS(runtime.NumCPU())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-c
		cancel()
		common.MinerLoger.Info("Got Control+C, exiting... wait 20 second")
		time.Sleep(20 * time.Second)
		os.Exit(0)
	}()
	if robotminer == nil {
		common.MinerLoger.Error("[error] Please check the coin in the README.md! if this coin is supported, use -S to set")
		return
	}
	robotminer.Run(ctx)
	common.MinerLoger.Info("All services exited")
}

//get current coin miner
func GetRobot(cfg *common.GlobalConfig) core.Robot {
	switch strings.ToUpper(cfg.NecessaryConfig.Symbol) {
	case core.SYMBOL_PMEER:
		r := &qitmeer.QitmeerRobot{}
		r.Cfg = cfg
		r.NeedGBT = make(chan struct{}, 1)
		r.Started = uint32(time.Now().Unix())
		r.Rpc = &common.RpcClient{Cfg: cfg}
		r.SubmitStr = make(chan string)
		r.PendingBlocks = map[string]qitmeer.PendingBlock{}
		return r
	default:

	}
	return nil
}
