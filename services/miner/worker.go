/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package miner

const (
	CPUWorkerType = "CPU_Worker"
	GBTWorkerType = "GBT_Worker"
	RemoteType    = "Remote_Worker"
	PoolType      = "Pool_Worker"
)

type IWorker interface {
	GetType() string
	Start() error
	Stop()
	IsRunning() bool
	Update()
}
