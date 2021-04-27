package main

import (
	"github.com/Qitmeer/qitmeer/params"
)

func InitNet(net string, config *Config) {
	genesisMapData := map[string]map[uint16]float64{}
	switch net {
	case "privnet":
		params.ActiveNetParams = &params.PrivNetParam
		genesisMapData = PrivGenesisMapData
	case "mixnet":
		params.ActiveNetParams = &params.MixNetParam
		genesisMapData = MixGenesisMapData
	case "testnet":
		params.ActiveNetParams = &params.TestNetParam
		genesisMapData = TestGenesisMapData
	case "mainnet":
		params.ActiveNetParams = &params.MainNetParam
		genesisMapData = MainGenesisMapData
	default:
		return
	}
	p := params.ActiveNetParams.Params
	config.GenesisAmountUnit = p.GenesisAmountUnit
	config.UnlocksPerHeightStep = p.UnlocksPerHeightStep
	config.UnlocksPerHeight = p.UnlocksPerHeight
	buildLedgerByTxData(config, genesisMapData)
}

var PrivGenesisMapData = map[string]map[uint16]float64{
	"RmBKxMWg4C4EMzYowisDEGSBwmnR6tPgjLs": {
		0: 50000, // meer
		1: 5000,  //qit
	},
	"RmHFARk5xmoMNUVJ6UCHFiWQML1vxwUhw1b": {
		0: 5000.343546,
		1: 1000.5,
	},
}

var MixGenesisMapData = map[string]map[uint16]float64{}

var TestGenesisMapData = map[string]map[uint16]float64{}

var MainGenesisMapData = map[string]map[uint16]float64{}
