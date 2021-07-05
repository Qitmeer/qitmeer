package main

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/database"
	_ "github.com/Qitmeer/qitmeer/database/ffldb"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"github.com/Qitmeer/qitmeer/ledger"
	"github.com/Qitmeer/qitmeer/log"
	"github.com/Qitmeer/qitmeer/params"
	_ "github.com/Qitmeer/qitmeer/services/common"
	"sort"
)

const (
	defaultSuffixFilename = "payouts.go"
	defaultPayoutDirPath  = "./../../ledger"
)

func main() {
	//log.PrintOrigins(true)
	// Load configuration and parse command line.  This function also
	// initializes logging and configures it accordingly.
	cfg, _, err := LoadConfig()
	if err != nil {
		log.Error(err.Error())
		return
	}
	fmt.Println(cfg.DebugAddress)
	if len(cfg.DebugAddress) > 0 {
		node := &DebugAddressNode{}
		err = node.init(cfg)
		if err != nil {
			log.Error(err.Error())
			return
		}

		node.exit()
		return
	}
	if blockInfo(cfg) {
		return
	}
	srcnode := &SrcNode{}
	err = srcnode.init(cfg)
	defer func() {
		srcnode.exit()
	}()
	if err != nil {
		log.Error(err.Error())
		return
	}

	if cfg.Last {
		// Just show last result
		buildLedger(srcnode, cfg)
		return
	}

	if cfg.ShowEndPoints > 0 {
		showEndBlocks(srcnode)
		return
	}

	if len(cfg.CheckEndPoint) > 0 {
		checkEndBlocks(srcnode)
		return
	}

	if len(cfg.EndPoint) > 0 {
		useWhole := false
		var ib blockdag.IBlock
		var blockHash *hash.Hash
		if cfg.EndPoint == "*" {
			// Use the whole UTXO database directly
			useWhole = true
		} else {
			blockH, err := hash.NewHashFromStr(cfg.EndPoint)
			if err != nil {
				log.Error(fmt.Sprintf("Error load endPoint hash: %s", err.Error()))
				return
			}
			blockHash = blockH
			ib = srcnode.bc.GetBlock(blockHash)
			if ib == nil {
				log.Error(fmt.Sprintf("Can't find block:%s", blockHash.String()))
				return
			}
			mainIB := srcnode.bc.BlockDAG().GetMainChainTip()
			if mainIB.GetHash().IsEqual(blockHash) {
				useWhole = true
			}
		}
		if useWhole {
			buildLedger(srcnode, cfg)
			// Must save data
			if Exists(cfg.DataDir) {
				RemovePath(cfg.DataDir)
			}
			if !CopyPath(cfg.SrcDataDir, cfg.DataDir) {
				log.Error(fmt.Sprintf("Can't copy %s to %s.", cfg.SrcDataDir, cfg.DataDir))
			}
		} else if ib != nil {
			if !srcnode.bc.BlockDAG().IsHourglass(ib.GetID()) {
				log.Error(fmt.Sprintf("%s is not good\n", ib.GetHash()))
				return
			}
			node := &Node{}
			err = node.init(cfg, srcnode, ib)
			if err != nil {
				log.Error(err.Error())
				return
			}
			buildLedger(node, cfg)
			node.exit()
		} else {
			log.Error(fmt.Sprintf("%s is not good\n", blockHash))
		}
	}
	return
}

func showEndBlocks(node *SrcNode) {
	fmt.Println("\nShow some recommended blocks for building ledger:")
	skipNum := node.cfg.EndPointSkips
	total := 0
	for cur := node.bc.BlockDAG().GetMainChainTip(); cur != nil; cur = node.bc.BlockDAG().GetBlockById(cur.GetMainParent()) {
		if skipNum > 0 {
			skipNum--
			continue
		}

		if node.bc.BlockDAG().IsHourglass(cur.GetID()) {
			fmt.Println(fmt.Sprintf("Great! order:%d  hash:%s  main_height:%d", cur.GetOrder(), cur.GetHash().String(), cur.GetHeight()))
			total++
		} else if node.bc.BlockDAG().IsOnMainChain(cur.GetID()) {
			fmt.Println(fmt.Sprintf("So-so! order:%d  hash:%s  main_height:%d", cur.GetOrder(), cur.GetHash().String(), cur.GetHeight()))
			total++
		} else {
			continue
		}

		if total >= node.cfg.ShowEndPoints {
			break
		}
	}
	fmt.Printf("Total:%d\n\n", total)
}

func checkEndBlocks(node *SrcNode) {
	blockHash, err := hash.NewHashFromStr(node.cfg.CheckEndPoint)
	if err != nil {
		log.Error(err.Error())
		return
	}
	ib := node.bc.GetBlock(blockHash)
	if ib == nil {
		log.Error(fmt.Sprintf("Can't find block:%s", blockHash.String()))
		return
	}
	if node.bc.BlockDAG().IsHourglass(ib.GetID()) {
		fmt.Printf("%s is OK\n", node.cfg.CheckEndPoint)
	} else {
		fmt.Printf("%s is not good\n", node.cfg.CheckEndPoint)
	}
}

func buildLedger(node INode, config *Config) error {
	params := params.ActiveNetParams.Params
	genesisLedger := map[string]*ledger.TokenPayoutReGen{}
	blueMap := map[uint]bool{}
	var totalAmount uint64
	var genAmount uint64
	mainChainTip := node.BlockChain().BlockDAG().GetMainChainTip()
	log.Info(fmt.Sprintf("Cur main tip:%s", mainChainTip.GetHash().String()))
	serializedUtxos := [][]byte{}
	cursorKeys := [][]byte{}
	err := node.DB().View(func(dbTx database.Tx) error {
		meta := dbTx.Metadata()
		utxoBucket := meta.Bucket(dbnamespace.UtxoSetBucketName)
		cursor := utxoBucket.Cursor()
		for ok := cursor.First(); ok; ok = cursor.Next() {
			serializedUtxo := utxoBucket.Get(cursor.Key())
			serializedUtxos = append(serializedUtxos, serializedUtxo)
			cursorKeys = append(cursorKeys, cursor.Key()[hash.HashSize:])
		}
		return nil
	})
	if err != nil {
		return err
	}

	for i := 0; i < len(serializedUtxos); i++ {
		serializedUtxo := serializedUtxos[i]
		txOutIdex, size := deserializeVLQ(cursorKeys[i])
		if size <= 0 {
			return fmt.Errorf("deserializeVLQ:%s", cursorKeys[i])
		}
		// Deserialize the utxo entry and return it.
		entry, err := blockchain.DeserializeUtxoEntry(serializedUtxo)
		if err != nil {
			return err
		}
		if entry.IsSpent() {
			continue
		}
		ib := node.BlockChain().GetBlock(entry.BlockHash())
		if ib.GetOrder() == blockdag.MaxBlockOrder {
			continue
		}
		if ib.GetStatus().KnownInvalid() {
			continue
		}
		if entry.IsCoinBase() {
			isblue, ok := blueMap[ib.GetID()]
			if !ok {
				isblue = node.BlockChain().BlockDAG().IsBlue(ib.GetID())
				blueMap[ib.GetID()] = isblue
			}
			if !isblue {
				continue
			}
		}
		_, addr, _, err := txscript.ExtractPkScriptAddrs(entry.PkScript(), params)
		if err != nil {
			return err
		}
		var addrStr string
		if len(addr) > 0 {
			for i := 0; i < len(addr); i++ {
				if i > 0 {
					addrStr += "-"
				}
				addrStr += addr[i].String()
			}
		}
		if _, ok := genesisLedger[addrStr]; !ok {
			tp := ledger.TokenPayout{Address: addrStr, PkScript: entry.PkScript(), Amount: types.Amount{Value: 0, Id: types.MEERID}}
			reTp := ledger.TokenPayoutReGen{tp, types.Amount{Value: 0, Id: types.MEERID}}
			genesisLedger[addrStr] = &reTp
		}

		if params.GenesisHash.IsEqual(entry.BlockHash()) {
			if genesisLedger[addrStr].GenAmount.Id == entry.Amount().Id {
				genesisLedger[addrStr].GenAmount.Value += entry.Amount().Value
				genAmount += uint64(entry.Amount().Value)
			}
		} else {
			eAmount := entry.Amount()
			if entry.IsCoinBase() && txOutIdex == 0 {
				eAmount.Value += node.BlockChain().GetFeeByCoinID(ib.GetHash(), eAmount.Id)
			}
			genesisLedger[addrStr].Payout.Amount = eAmount
			totalAmount += uint64(eAmount.Value)
		}
		log.Trace(fmt.Sprintf("Process Address:%s Amount:%d Block Hash:%s", addrStr, entry.Amount(), entry.BlockHash().String()))
	}

	if len(genesisLedger) == 0 {
		log.Info("No payouts need to deal with.")
		return nil
	}
	fmt.Println(fmt.Sprintf("Show Ledger:[Genesis------->%s]", mainChainTip.GetHash().String()))
	payList := make(ledger.PayoutList2, len(genesisLedger))
	i := 0
	for _, v := range genesisLedger {
		payList[i] = *v
		i++
	}
	sort.Sort(sort.Reverse(payList))
	for _, v := range payList {
		fmt.Printf("Address:%s  GenAmount:%15d  Amount:%15d  Total:%15d\n", v.Payout.Address, v.GenAmount, v.Payout.Amount, v.GenAmount.Value+v.Payout.Amount.Value)
	}
	fmt.Printf("-----------------\n")
	fmt.Printf("Total Ledger:%5d  GenAmount:%15d  Amount:%15d  Total:%15d\n", len(genesisLedger), genAmount, totalAmount, genAmount+totalAmount)

	if config.SavePayoutsFile {
		return savePayoutsFile(params, payList, config)
	}
	return nil
}

func blockInfo(cfg *Config) bool {
	if cfg.BlocksInfo {
		node := &BINode{}
		err := node.init(cfg)
		defer func() {
			node.exit()
		}()
		if err != nil {
			log.Error(err.Error())
		}
		return true
	}
	return false
}

func deserializeVLQ(serialized []byte) (uint64, int) {
	var n uint64
	var size int
	for _, val := range serialized {
		size++
		n = (n << 7) | uint64(val&0x7f)
		if val&0x80 != 0x80 {
			break
		}
		n++
	}

	return n, size
}
