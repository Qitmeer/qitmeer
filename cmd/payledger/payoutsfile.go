package main

import (
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qng-core/core/address"
	"github.com/Qitmeer/qng-core/core/protocol"
	"github.com/Qitmeer/qng-core/core/types"
	"github.com/Qitmeer/qng-core/engine/txscript"
	"github.com/Qitmeer/qng-core/ledger"
	"github.com/Qitmeer/qng-core/log"
	"github.com/Qitmeer/qng-core/params"
	"os"
	"path/filepath"
	"strings"
)

func savePayoutsFile(params *params.Params, genesisLedger ledger.PayoutList2, config *Config) error {
	if len(genesisLedger) == 0 {
		log.Info("No payouts need to deal with.")
		return nil
	}
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

	fileName := filepath.Join(defaultPayoutDirPath, netName+defaultSuffixFilename)

	f, err := os.Create(fileName)

	if err != nil {
		log.Error(fmt.Sprintf("Save error:%s  %s", fileName, err))
		return err
	}
	defer func() {
		err = f.Close()
	}()

	funName := fmt.Sprintf("%s%s", strings.ToUpper(string(netName[0])), netName[1:])
	fileContent := fmt.Sprintf("package ledger\nfunc init%s() {\n", funName)

	if config.UnlocksPerHeight > 0 {
		fileContent += processLockingPayouts(genesisLedger, int64(config.UnlocksPerHeight))
	} else {
		fileContent += processNormalPayouts(genesisLedger)
	}

	fileContent += "}"

	f.WriteString(fileContent)

	log.Info(fmt.Sprintf("Finish save %s", fileName))

	return nil
}

func processNormalPayouts(genesisLedger ledger.PayoutList2) string {
	fileContent := ""
	for _, v := range genesisLedger {
		if v.Payout.Amount.Id != types.MEERID {
			continue
		}
		fileContent += fmt.Sprintf("	addPayout(\"%s\",%d,\"%s\")\n", v.Payout.Address, v.Payout.Amount.Value, hex.EncodeToString(v.Payout.PkScript))
	}
	return fileContent
}

func processLockingPayouts(genesisLedger ledger.PayoutList2, lockNum int64) string {
	fileContent := ""

	curMHeight := int64(0)
	curLockedNum := int64(0)
	for _, v := range genesisLedger {
		if v.Payout.Amount.Id != types.MEERID {
			continue
		}

		for v.Payout.Amount.Value > 0 {
			needLockNum := lockNum - curLockedNum

			amount := int64(0)
			if v.Payout.Amount.Value >= needLockNum {
				v.Payout.Amount.Value -= needLockNum
				amount = needLockNum
				curMHeight++
				curLockedNum = 0
			} else {
				amount = v.Payout.Amount.Value
				curLockedNum += amount
				v.Payout.Amount.Value = 0
			}
			script, err := PayToCltvAddrScriptWithMainHeight(v.Payout.Address, curMHeight)
			if err != nil {
				return err.Error()
			}
			fileContent += fmt.Sprintf("	addPayout(\"%s\",%d,\"%s\")\n", v.Payout.Address, amount, hex.EncodeToString(script))
		}

	}
	return fileContent
}

func PayToCltvAddrScriptWithMainHeight(addrStr string, mainHeight int64) ([]byte, error) {
	addr, err := address.DecodeAddress(addrStr)
	if err != nil {
		return nil, err
	}
	return txscript.PayToCLTVPubKeyHashScript(addr.Script(), mainHeight)
}
