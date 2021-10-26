// Copyright (c) 2019 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package lib

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer/cmd/miner/common"
	"github.com/Qitmeer/qitmeer/cmd/miner/core"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"github.com/Qitmeer/qitmeer/rpc/client"
	"strconv"
	"strings"
	"sync"
	"time"
)

var ErrSameWork = fmt.Errorf("Same work, Had Submitted!")

type QitmeerWork struct {
	core.Work
	Block       *BlockHeader
	PoolWork    NotifyWork
	stra        *QitmeerStratum
	StartWork   bool
	ForceUpdate bool
	Ing         bool
	WorkLock    sync.Mutex
	WsClient    *client.Client
}

func (this *QitmeerWork) GetPowType() pow.PowType {
	switch this.Cfg.NecessaryConfig.Pow {
	case POW_MEER_CRYPTO:
		return pow.MEERXKECCAKV1
	default:
		return pow.BLAKE2BD
	}
}

// GetBlockTemplate
func (this *QitmeerWork) Get() bool {
	if this.Ing {
		return false
	}
	defer func() {
		this.Ing = false
	}()
	this.Ing = true
	for {
		if this.WsClient == nil || this.WsClient.Disconnected() {
			return false
		}
		this.ForceUpdate = false
		this.Rpc.GbtID++
		header, err := this.WsClient.GetRemoteGBT(this.GetPowType())
		if err != nil {
			time.Sleep(time.Duration(this.Cfg.OptionConfig.TaskInterval) * time.Millisecond)
			common.MinerLoger.Error("GetRemoteGBT Error", "err", err.Error())
			continue
		}
		if this.Block != nil && this.Block.ParentRoot == header.ParentRoot &&
			(time.Now().Unix()-this.GetWorkTime) < int64(this.Cfg.OptionConfig.Timeout)*10 {
			//not has new work
			return false
		}
		this.Block = &BlockHeader{}
		this.Block.ParentRoot = header.ParentRoot
		this.Block.WorkData = header.BlockData()
		this.Block.Target = fmt.Sprintf("%064x", pow.CompactToBig(header.Difficulty))
		common.MinerLoger.Info(fmt.Sprintf("getRemoteBlockTemplate , target :%s", this.Block.Target))
		return true
	}
}

// submit
func (this *QitmeerWork) Submit(header *types.BlockHeader, gbtID string) (string, int, error) {
	this.Lock()
	defer this.Unlock()
	this.Rpc.SubmitID++
	id := fmt.Sprintf("miner_submit_gbtID:%s_id:%d", gbtID, this.Rpc.SubmitID)
	fmt.Println("header", header.ParentRoot.String())
	var res string
	var err error
	common.Timeout(func() {
		res, err = this.WsClient.SubmitBlockHeader(header)
	}, 15, func() {
		err = errors.New("submit timeout")
	})
	if err != nil {
		common.MinerLoger.Error("[submit error] " + id + " " + err.Error())
		if strings.Contains(err.Error(), "The tips of block is expired") {
			return "", 0, ErrSameWork
		}
		if strings.Contains(err.Error(), "worthless") {
			return "", 0, ErrSameWork
		}
		return "", 0, errors.New("[submit data failed]" + err.Error())
	}
	arr := strings.Split(res, "coinbaseTx:")
	arr = strings.Split(arr[1], " ")
	txID := arr[0]
	arr = strings.Split(res, "height:")
	arr = strings.Split(arr[1], " ")
	height, _ := strconv.Atoi(arr[0])
	return txID, height, err
}

// pool get work
func (this *QitmeerWork) PoolGet() bool {
	if !this.stra.PoolWork.NewWork {
		return false
	}
	err := this.stra.PoolWork.PrepWork()
	if err != nil {
		common.MinerLoger.Error(err.Error())
		return false
	}

	if (this.stra.PoolWork.JobID != "" && this.stra.PoolWork.Clean) || this.PoolWork.JobID != this.stra.PoolWork.JobID {
		this.stra.PoolWork.Clean = false
		this.Cfg.OptionConfig.Target = fmt.Sprintf("%064x", common.BlockBitsToTarget(this.stra.PoolWork.Nbits, 2))
		this.PoolWork = this.stra.PoolWork
		common.CurrentHeight = uint64(this.stra.PoolWork.Height)
		common.JobID = this.stra.PoolWork.JobID
		return true
	}

	return false
}

//pool submit work
func (this *QitmeerWork) PoolSubmit(subm string) error {
	if this.LastSub == subm {
		return ErrSameWork
	}
	this.LastSub = subm
	arr := strings.Split(subm, "-")
	data, err := hex.DecodeString(arr[0])
	if err != nil {
		return err
	}
	sub, err := this.stra.PrepSubmit(data, arr[1], arr[2])
	if err != nil {
		return err
	}
	m, err := json.Marshal(sub)
	if err != nil {
		return err
	}
	_, err = this.stra.Conn.Write(m)
	if err != nil {
		common.MinerLoger.Debug("[submit error][pool connect error]", "error", err)
		return err
	}
	_, err = this.stra.Conn.Write([]byte("\n"))
	if err != nil {
		common.MinerLoger.Debug(err.Error())
		return err
	}

	return nil
}
