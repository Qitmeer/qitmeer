// This file is ignored during the regular build due to the following build tag.
// It is called by go generate and used to automatically generate pre-computed
// tables used to accelerate operations.
// +build ignore

package main

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/address"
	"github.com/Qitmeer/qitmeer/core/protocol"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"github.com/Qitmeer/qitmeer/params"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	GENE_PAYOUT_TYPE_STANDARD = iota
	GENE_PAYOUT_TYPE_AUTO_LOCK_WITH_CONFIG
	GENE_PAYOUT_TYPE_LOCK_WITH_HEIGHT
	ONE_DAY_SECONDS = 86400
	YEAR_DAYS       = 365
)

var (
	defaultPayoutDirPath  = "./"
	defaultSuffixFilename = "ledgerpayout_gen"
	oneDayHeight          = int64(0)
	yearMiningAmount      = 0.00
	yearReleaseAmount     = 0.00
	originAmount          = int64(0)
	genesisAmount         = int64(0)
)

type GenesisInitPayout struct {
	CoinID            types.CoinID
	Address           string
	Amount            int64
	GenesisPayoutType int
	LockHeight        int64 // amount lock with height
}

func GeneratePayoutFile(param *params.Params, geneData []GenesisInitPayout, geneDataImport []string) {
	importData, err := FormatDataFromImport(geneDataImport, param)
	if err != nil {
		fmt.Println(err)
		return
	}
	geneData = append(geneData, importData...)
	seedHash, err := GenerateUniqueSeedHash(geneData)
	if err != nil {
		fmt.Println(err)
		return
	}
	payList, payKeys := ReSortAndSliceGeneDataWithSeed(geneData, seedHash, param)
	savePayoutsFileBySliceShuffle(param, payList, payKeys)
}

func ReSortAndSliceGeneDataWithSeed(data []GenesisInitPayout, seedHash []byte, p *params.Params) ([]GenesisInitPayout, []int) {
	payList := make([]GenesisInitPayout, 0)
	i := 0
	payKeys := make([]int, 0)
	for _, v := range data {
		for {
			if v.GenesisPayoutType == GENE_PAYOUT_TYPE_STANDARD || v.GenesisPayoutType == GENE_PAYOUT_TYPE_LOCK_WITH_HEIGHT {
				payList = append(payList, v)
				payKeys = append(payKeys, i)
				i++
				break
			}
			if int64(v.Amount) > p.LedgerParams.GenesisAmountUnit {
				payList = append(payList, GenesisInitPayout{
					v.CoinID, v.Address, p.LedgerParams.GenesisAmountUnit, v.GenesisPayoutType, v.LockHeight,
				})
				payKeys = append(payKeys, i)
				i++
				v.Amount -= p.LedgerParams.GenesisAmountUnit
			} else {
				payList = append(payList, v)
				payKeys = append(payKeys, i)
				i++
				break
			}
		}
	}
	payKeys = GenesisShuffle(payKeys, seedHash)
	return payList, payKeys
}

// generate unique map with certainly input data
func GenerateUniqueSeedHash(data []GenesisInitPayout) ([]byte, error) {
	uniqueString := make([]string, 0)
	for _, v := range data {
		uniqueString = append(uniqueString, fmt.Sprintf("%d:%s:%d:%d:%f", v.CoinID, v.Address, v.GenesisPayoutType, v.LockHeight, v.Amount))
	}
	sort.Strings(uniqueString)
	b, err := json.Marshal(uniqueString)
	if err != nil {
		return nil, err
	}
	seedHash := hash.HashB(b)
	return seedHash, nil
}

func ConvertStringToInt64(s string) int64 {
	a, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return 0
	}
	return a.Int64()
}

func CalcAllNeedReleaseAmount(data []string) int64 {
	for _, v := range data {
		// CoinID,address,amount,locktype,height
		arr := strings.Split(v, ",")
		amount := ConvertStringToInt64(arr[2])
		if amount <= 0 {
			continue
		}
		originAmount += amount
	}
	return originAmount
}

func FormatDataFromImport(data []string, params *params.Params) ([]GenesisInitPayout, error) {
	allNeedRelease := CalcAllNeedReleaseAmount(data)
	newData := make([]GenesisInitPayout, 0)
	oneDayHeight = int64(float64(ONE_DAY_SECONDS) / params.TargetTimePerBlock.Seconds())
	yearMiningAmount = float64(oneDayHeight) * float64(params.BaseSubsidy) * YEAR_DAYS
	for _, v := range data {
		// CoinID,address,amount,locktype,height
		arr := strings.Split(v, ",")
		if len(arr) < 5 {
			return nil, errors.New("data format error")
		}
		CoinID, err := strconv.Atoi(arr[0])
		if err != nil {
			return nil, errors.New("CoinID data error" + arr[0])
		}
		amount := ConvertStringToInt64(arr[2])
		if amount <= 0 {
			continue
		}
		payouttype, err := strconv.Atoi(arr[3])
		if err != nil {
			return nil, errors.New("payouttype data error" + arr[3])
		}
		lockheight, err := strconv.Atoi(arr[4])
		if err != nil {
			return nil, errors.New("height data error" + arr[4])
		}
		if payouttype != GENE_PAYOUT_TYPE_AUTO_LOCK_WITH_CONFIG {
			newData = append(newData, GenesisInitPayout{
				types.CoinID(CoinID),
				arr[1],
				amount,
				payouttype,
				int64(lockheight),
			})
			continue
		}
		userYearReleaseAmount := float64(amount) * yearMiningAmount / float64(allNeedRelease)
		if userYearReleaseAmount > float64(amount) {
			userYearReleaseAmount = float64(amount)
		}
		yearReleaseAmount += userYearReleaseAmount
		if userYearReleaseAmount > 0 {
			newData = append(newData, GenesisInitPayout{
				types.CoinID(CoinID),
				arr[1],
				int64(userYearReleaseAmount),
				payouttype,
				int64(lockheight),
			})
		}

		leftAmount := float64(amount) - float64(int64(userYearReleaseAmount))
		if leftAmount > 0 {
			if leftAmount > float64(amount) {
				leftAmount = float64(amount)
			}
			newData = append(newData, GenesisInitPayout{
				types.CoinID(CoinID),
				arr[1],
				int64(leftAmount),
				GENE_PAYOUT_TYPE_LOCK_WITH_HEIGHT,
				int64(params.LedgerParams.MaxLockHeight),
			})
		}

	}
	return newData, nil
}

func GenesisShuffle(array []int, seed []byte) []int {
	for i := len(array) - 1; i > 0; i-- {
		p := RandShuffle(int64(i), seed)
		a := array[i]
		array[i] = array[p]
		array[p] = a
	}
	return array
}

func RandShuffle(max int64, seed []byte) int64 {
	originL := max
	if max > 24 {
		max = max % 24
	}
	if max <= 0 {
		max = 1
	}
	seedNum := binary.LittleEndian.Uint64(seed[max : max+8])
	return int64(seedNum % uint64(originL))
}

func savePayoutsFileBySliceShuffle(params *params.Params, genesisLedger []GenesisInitPayout, sortKeys []int) error {
	netName := ""
	switch params.Net {
	case protocol.MainNet:
		netName = "main"
	case protocol.TestNet:
		netName = "test"
	case protocol.PrivNet:
		netName = "priv"
	case protocol.MixNet:
		netName = "mix"
	}
	if len(genesisLedger) == 0 {
		fmt.Println(netName + " network No payouts need to deal with.")
		return nil
	}
	fileName := filepath.Join(defaultPayoutDirPath, defaultSuffixFilename+netName+".go")

	f, err := os.Create(fileName)

	if err != nil {
		fmt.Println(fmt.Sprintf("Save error:%s  %s", fileName, err))
		return err
	}
	defer func() {
		err = f.Close()
	}()

	funName := fmt.Sprintf("%s%s", strings.ToUpper(string(netName[0])), netName[1:])
	fileContent := fmt.Sprintf("// It is called by go generate and used to automatically generate pre-computed \n// Copyright 2017-2018 The qitmeer developers \n// This file is auto generate \npackage ledger\n\nimport (\n\t. \"github.com/Qitmeer/qitmeer/core/types\"\n)\n\nfunc init%s() {\n", funName)
	oneDayRelease := yearReleaseAmount / YEAR_DAYS
	fileContent += processLockingGenesisPayouts(genesisLedger, sortKeys, int64(oneDayRelease), oneDayHeight)

	fileContent += "}"

	f.WriteString(fileContent)

	return nil
}

func processLockingGenesisPayouts(genesisLedger []GenesisInitPayout, sortKeys []int, oneDayRelease int64, heightStep int64) string {
	fileContent := ""
	curMHeight := int64(0)
	oneDayUsedAmount := int64(0)
	for i := 0; i < len(sortKeys); i++ {
		v := genesisLedger[sortKeys[i]]
		if v.GenesisPayoutType == GENE_PAYOUT_TYPE_STANDARD {
			addr, err := address.DecodeAddress(v.Address)
			if err != nil {
				return err.Error()
			}
			script, err := txscript.PayToAddrScript(addr)
			if err != nil {
				return err.Error()
			}
			genesisAmount += int64(v.Amount)
			fileContent += fmt.Sprintf("	addPayout2(\"%s\",Amount{Value: %d, Id: CoinID(%d)},\"%s\")\n", v.Address, int64(v.Amount), v.CoinID, hex.EncodeToString(script))
			continue
		}
		if v.GenesisPayoutType == GENE_PAYOUT_TYPE_LOCK_WITH_HEIGHT {
			script, err := PayToCltvAddrScriptWithMainHeight(v.Address, v.LockHeight)
			if err != nil {
				return err.Error()
			}
			genesisAmount += int64(v.Amount)
			fileContent += fmt.Sprintf("	addPayout2(\"%s\",Amount{Value: %d, Id: CoinID(%d)},\"%s\")\n", v.Address, int64(v.Amount), v.CoinID, hex.EncodeToString(script))
			continue
		}
		if v.GenesisPayoutType == GENE_PAYOUT_TYPE_AUTO_LOCK_WITH_CONFIG {
			for v.Amount > 0 {
				oneDayLeftAmount := oneDayRelease - oneDayUsedAmount
				amount := int64(0)
				script, err := PayToCltvAddrScriptWithMainHeight(v.Address, curMHeight)
				if err != nil {
					return err.Error()
				}
				if v.Amount >= oneDayLeftAmount {
					v.Amount -= oneDayLeftAmount
					amount = oneDayLeftAmount
					curMHeight += heightStep
					oneDayUsedAmount = 0
				} else {
					amount = v.Amount
					oneDayUsedAmount += int64(amount)
					v.Amount = 0
				}
				genesisAmount += int64(amount)
				fileContent += fmt.Sprintf("	addPayout2(\"%s\",Amount{Value: %d, Id: CoinID(%d)},\"%s\")\n", v.Address, int64(amount), v.CoinID, hex.EncodeToString(script))
			}
		}
	}
	if originAmount != genesisAmount {
		fmt.Println("not equal", "originAmount", originAmount, "genesisAmount", genesisAmount)
		return ""
	}
	fmt.Println("generate success")
	return fileContent
}

func PayToCltvAddrScriptWithMainHeight(addrStr string, mainHeight int64) ([]byte, error) {
	addr, err := address.DecodeAddress(addrStr)
	if err != nil {
		return nil, err
	}
	if mainHeight <= 0 {
		return txscript.PayToAddrScript(addr)
	}
	return txscript.PayToCLTVPubKeyHashScript(addr.Script(), mainHeight)
}
