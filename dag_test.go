package main

import (
	"fmt"
	"github.com/noxproject/nox/node"
	"github.com/noxproject/nox/database"
	"github.com/noxproject/nox/log"
	"testing"
	"os"
	"runtime"
	"runtime/debug"
	"github.com/noxproject/nox/common/hash"
)


func TestFig2(t *testing.T) {
	Gen:=initDag()

	B:=buildBlock("B",[]*hash.Hash{Gen.Hash})
	C:=buildBlock("C",[]*hash.Hash{Gen.Hash})
	D:= buildBlock("D",[]*hash.Hash{Gen.Hash})
	E:= buildBlock("E",[]*hash.Hash{Gen.Hash})

	F := buildBlock("F", []*hash.Hash{B.GetHash(), C.GetHash()})

	G := buildBlock("G", []*hash.Hash{C.GetHash(), D.GetHash()})

	H := buildBlock("H", []*hash.Hash{E.GetHash()})

	//I :=
	buildBlock("I", []*hash.Hash{F.GetHash(), D.GetHash()})

	//J :=
	buildBlock("J", []*hash.Hash{B.GetHash(), G.GetHash(), E.GetHash()})

	//K :=
	buildBlock("K", []*hash.Hash{D.GetHash(), H.GetHash()})
	//

	//
	fmt.Println()
	for _,v:=range barray{
		fmt.Printf("%s = %s\n",v.Name,v.Hash.String())
	}
	//
	fmt.Println()
	lastNode:=ser.GetNoxFull().GetBlockManager().GetChain().DAG().GetLastBlock()
	fmt.Printf("The Fig.2 Order: %d\n",lastNode.GetHeight())

	for i:=0;i<=int(lastNode.GetHeight());i++ {
		hash:=ser.GetNoxFull().GetBlockManager().GetChain().DAG().GetBlockByOrder(i)
		hb:=bmap[*hash]
		if hb==nil {
			continue
		}
		if hb.Hash.IsEqual(Gen.Hash) {
			fmt.Printf("%s",hb.Name)
		}else{
			fmt.Printf(" -> %s",hb.Name)
		}

	}
	fmt.Println()
	fmt.Println()
	//

	//time.Sleep(5*time.Second)
	End()
}
//some base function
var ser *node.Node=nil
var db database.DB
/////////////////////
var bmap map[hash.Hash]*HelpBlock
var barray []*HelpBlock

type HelpBlock struct {
	Name string
	Hash *hash.Hash
}
func (hb *HelpBlock) GetHash() *hash.Hash{
	return hb.Hash
}
func initDag() *HelpBlock{
	err:=Run()
	if err!=nil {
		return nil
	}
	//
	bmap=make(map[hash.Hash]*HelpBlock)

	Gen:=&HelpBlock{"Gen",ser.Params.GenesisHash}
	bmap[*Gen.Hash]=Gen
	barray=append(barray,Gen)

	return Gen
}
func buildBlock(name string,parents []*hash.Hash) *HelpBlock{
	hb:=&HelpBlock{Name:name}

	hash,err:=ser.GetNoxFull().GetCpuMiner().GenerateBlockByParents(parents)
	if err!=nil {
		fmt.Printf("ERROR:%s [%s]\n",hb.Name,err.Error())
		return nil
	}
	if hash==nil {
		fmt.Printf("%s = nil\n",hb.Name)
		return nil
	}
	hb.Hash=hash

	//fmt.Printf("%s = %s\n",hb.Name,hb.Hash.String())

	bmap[*hash]=hb
	barray=append(barray,hb)

	return hb
}
////////////////////
func Run() error {
	os.Args=[]string{}
	os.Args=append(os.Args,"dag_test")
	os.Args=append(os.Args,"-A=./bin")

	//
	runtime.GOMAXPROCS(runtime.NumCPU())
	debug.SetGCPercent(20)

	tcfg, _, err := loadConfig()
	if err != nil {
		return err
	}
	glogger.Verbosity(log.LvlCrit)
	cfg = tcfg

	interrupt := interruptListener()
	defer log.Info("Shutdown complete")

	log.Info("System info", "Nox Version", version(), "Go version",runtime.Version())
	log.Info("System info", "Home dir", cfg.HomeDir)
	if cfg.NoFileLogging {
		log.Info("File logging disabled")
	}
	dbPath := blockDbPath(cfg.DbType)

	fi, err := os.Stat(dbPath)
	if err == nil {
		log.Info(fmt.Sprintf("Removing test database from '%s'", dbPath))
		if fi.IsDir() {
			err := os.RemoveAll(dbPath)
			if err != nil {
				return err
			}
		} else {
			err := os.Remove(dbPath)
			if err != nil {
				return err
			}
		}
	}
	// Load the block database.
	db, err = loadBlockDB()
	if err != nil {
		log.Error("load block database","error", err)
		return err
	}

	// Return now if an interrupt signal was triggered.
	if interruptRequested(interrupt) {
		return nil
	}

	// Create node and start it.
	ser, err = node.NewNode(cfg,db,activeNetParams.Params)
	if err != nil {
		log.Error("Unable to start server","listeners",cfg.Listeners,"error", err)
		return err
	}
	err = ser.RegisterService()
	if err != nil {
		return err
	}
	err = ser.Start()
	if err != nil {
		log.Error("Uable to start server", "error",err)
		return err
	}

	return nil
}
func End(){
	db.Close()

	ser.Stop()
	ser.WaitForShutdown()
}
