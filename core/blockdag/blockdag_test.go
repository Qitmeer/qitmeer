package blockdag

import (
	"encoding/json"
	"fmt"
	"github.com/Qitmeer/qng-core/common/hash"
	"github.com/Qitmeer/qng-core/common/roughtime"
	"github.com/Qitmeer/qng-core/config"
	"github.com/Qitmeer/qng-core/database"
	l "github.com/Qitmeer/qng-core/log"
	"github.com/Qitmeer/qng-core/params"
	"io"
	"math/rand"
	"os"
	"path/filepath"
)

// Structure of blocks data
type TestBlocksData struct {
	Tag     string   `json:"tag"`
	Parents []string `json:"parents"`
}

// Test input and output structure
type TestInOutData struct {
	Input  string   `json:"in"`
	Output []string `json:"out"`
}

// Test input and output structure2
type TestInOutData2 struct {
	Input  string `json:"in"`
	Output int    `json:"out"`
}

type TestInOutData3 struct {
	Input  []string `json:"in"`
	Output bool     `json:"out"`
}

// Structure of test data
type TestData struct {
	PH_Fig2Blocks      []TestBlocksData `json:"PH_fig2-blocks"`
	PH_Fig4Blocks      []TestBlocksData `json:"PH_fig4-blocks"`
	PH_GetFutureSet    TestInOutData
	PH_GetAnticone     TestInOutData
	PH_BlueSetFig2     TestInOutData
	PH_BlueSetFig4     TestInOutData
	PH_OrderFig2       TestInOutData
	PH_OrderFig4       TestInOutData
	PH_IsOnMainChain   TestInOutData
	PH_GetLayer        TestInOutData
	CO_Blocks          []TestBlocksData
	CO_GetMainChain    TestInOutData
	CO_GetOrder        TestInOutData
	SP_Blocks          []TestBlocksData
	PH_LocateBlocks    TestInOutData
	PH_LocateMaxBlocks TestInOutData
	CP_Blocks          []TestBlocksData
	PH_MPConcurrency   TestInOutData2
	PH_BConcurrency    TestInOutData2
	PH_MainChainTip    []TestInOutData3
}

// Load some data that phantom test need,it can use to build the dag ;This is the
// raw input data.
func loadTestData(fileName string, testData *TestData) error {
	if len(fileName) == 0 {
		return fmt.Errorf("file name error")
	}

	var f *os.File
	var err error

	f, err = os.Open(fileName)
	if err != nil {
		return err
	}

	defer func() {
		cErr := f.Close()
		if err == nil {
			err = cErr
		}
	}()
	//
	err = json.NewDecoder(f).Decode(testData)
	return err
}

// DAG block data
type TestBlock struct {
	hash      hash.Hash
	parents   []*hash.Hash
	timeStamp int64
}

// Return the hash
func (tb *TestBlock) GetHash() *hash.Hash {
	return &tb.hash
}

// Get all parents set,the dag block has more than one parent
func (tb *TestBlock) GetParents() []*hash.Hash {
	return tb.parents
}

func (tb *TestBlock) GetTimestamp() int64 {
	return tb.timeStamp
}

// Acquire the weight of block
func (tb *TestBlock) GetWeight() uint64 {
	return 1
}

func (tb *TestBlock) GetPriority() int {
	return MaxPriority
}

// This is the interface for Block DAG,can use to call public function.
var bd BlockDAG

// Used to simulate block hash,It's just a test program,beacause
// we only care about the block DAG.
var tempHash int = 0

var randTool *rand.Rand = rand.New(rand.NewSource(roughtime.Now().UnixNano()))

// It contains all of test data. Convenient for you to use different input data
// and output data.
var testData *TestData

// This is the test data file name
var testDataFilePath string = "./testData.json"

var tbMap map[string]IBlock

func InitBlockDAG(dagType string, graph string) IBlockDAG {
	output := io.Writer(os.Stdout)
	glogger := l.NewGlogHandler(l.StreamHandler(output, l.TerminalFormat(false)))
	glogger.Verbosity(l.LvlError)
	l.Root().SetHandler(glogger)
	blockdaglogger := l.New(l.Ctx{"module": "blockdag"})
	UseLogger(blockdaglogger)

	testData = &TestData{}
	err := loadTestData(testDataFilePath, testData)
	if err != nil {
		return nil
	}
	var tbd []TestBlocksData
	if graph == "PH_fig2-blocks" {
		tbd = testData.PH_Fig2Blocks
	} else if graph == "PH_fig4-blocks" {
		tbd = testData.PH_Fig4Blocks
	} else if graph == "CO_Blocks" {
		tbd = testData.CO_Blocks
	} else if graph == "SP_Blocks" {
		tbd = testData.SP_Blocks
	} else if graph == "CP_Blocks" {
		tbd = testData.CP_Blocks
	} else {
		return nil
	}
	blen := len(tbd)
	if blen < 2 {
		return nil
	}

	cfg := &config.Config{DbType: "ffldb", DataDir: "."}
	db, err := loadBlockDB(cfg)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	bd = BlockDAG{}
	instance := bd.Init(dagType, CalcBlockWeight, -1, db, nil)
	tbMap = map[string]IBlock{}
	for i := 0; i < blen; i++ {
		parents := []*hash.Hash{}
		for _, parent := range tbd[i].Parents {
			parents = append(parents, tbMap[parent].GetHash())
		}
		block := buildBlock(parents)
		l, _, ib, _ := bd.AddBlock(block)
		if l != nil && l.Len() > 0 {
			tbMap[tbd[i].Tag] = ib
			err = bd.Commit()
			if err != nil {
				return nil
			}
		} else {
			fmt.Printf("Error:%d  %s\n", tempHash, tbd[i].Tag)
			return nil
		}

	}

	return instance
}

func buildBlock(parents []*hash.Hash) *TestBlock {
	tempHash++
	hashStr := fmt.Sprintf("%d", tempHash)
	h := hash.MustHexToDecodedHash(hashStr)
	tBlock := &TestBlock{hash: h}
	tBlock.parents = parents
	tBlock.timeStamp = roughtime.Now().Unix()

	//
	return tBlock
}

func getBlockTag(id uint) string {
	for k, v := range tbMap {
		if v.GetID() == id {
			return k
		}
	}
	return ""
}

func changeToIDList(list []string) []uint {
	length := len(list)
	if length == 0 {
		return nil
	}
	result := []uint{}
	for i := 0; i < length; i++ {
		result = append(result, tbMap[list[i]].GetID())
	}
	return result
}

func processResult(calRet interface{}, theory []uint) bool {

	var ret bool = true
	switch calRet.(type) {
	case []uint:
		result := calRet.([]uint)
		rLen := len(result)

		if rLen != len(theory) {
			ret = false
		}
		for i := 0; i < rLen; i++ {
			if result[i] != theory[i] {
				ret = false
				break
			}
		}
	case *IdSet:
		result := calRet.(*IdSet)
		okResult := NewIdSet()
		okResult.AddList(theory)
		if !result.IsEqual(okResult) {
			ret = false
		}
	}

	if ret {
		fmt.Println("Congratulations，The result of the function is completely correct！！！")
	} else {
		fmt.Println("Failed，The result of the operation of a function is incompatible with the expectation！！！")
	}
	return ret
}

func printBlockChainTag(list []uint) {
	var result string
	for i := 0; i < len(list); i++ {
		name := getBlockTag(list[i])
		if i == 0 {
			result += name
		} else {
			result += fmt.Sprintf("-->%s", name)
		}
	}
	fmt.Println(result)
}

func printBlockSetTag(set *IdSet) {
	var result string = "["
	isFirst := true
	for k := range set.GetMap() {
		name := getBlockTag(k)
		if isFirst {
			result += name
			isFirst = false
		} else {
			result += fmt.Sprintf(",%s", name)
		}

	}
	result += "]"
	fmt.Println(result)
}

func reverseBlockList(s []uint) []uint {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func CalcBlockWeight(ib IBlock, bi *BlueInfo) int64 {
	if ib.(*PhantomBlock).blueNum == 0 {
		return 0
	} else if ib.(*PhantomBlock).blueNum < 3 {
		return 2
	}
	return 1
}

func loadBlockDB(cfg *config.Config) (database.DB, error) {
	dbName := "blocks_" + cfg.DbType
	dbPath := filepath.Join(cfg.DataDir, dbName)
	err := removeBlockDB(dbPath)
	if err != nil {
		return nil, err
	}
	db, err := database.Create(cfg.DbType, dbPath, params.ActiveNetParams.Net)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func removeBlockDB(dbPath string) error {
	fi, err := os.Stat(dbPath)
	if err == nil {
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

	return nil
}

func getBlocksByTag(tags []string) []*hash.Hash {
	result := []*hash.Hash{}
	for _, v := range tags {
		ib, ok := tbMap[v]
		if !ok {
			continue
		}
		result = append(result, ib.GetHash())
	}
	return result
}

func exit() {
	removeBlockDB("./blocks_ffldb")
}
