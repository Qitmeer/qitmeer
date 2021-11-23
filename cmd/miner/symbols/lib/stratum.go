// Copyright (c) 2019 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package lib

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer/cmd/miner/common"
	"github.com/Qitmeer/qitmeer/cmd/miner/core"
	qitmeer "github.com/Qitmeer/qng-core/common/hash"
	"github.com/Qitmeer/qng-core/core/types/pow"
	"github.com/Qitmeer/qng-core/params"
	"math/big"
	"strconv"
	"strings"
	"time"
)

// ErrStratumStaleWork indicates that the work to send to the pool was stale.
var ErrStratumStaleWork = fmt.Errorf("Stale work, throwing away")

// NotifyRes models the json from a mining.notify message.
type NotifyRes struct {
	JobID          string
	Hash           string
	GenTX1         string
	GenTX2         string
	MerkleBranches []string
	BlockVersion   string
	Nbits          string
	Ntime          string
	CleanJobs      bool
	StateRoot      string
	Height         int64
	CB3            string
	CB4            string
}

// Submit models a submission message.
type Submit struct {
	Params []string    `json:"params"`
	ID     interface{} `json:"id"`
	Method string      `json:"method"`
}

// SubscribeReply models the server response to a subscribe message.
type SubscribeReply struct {
	SubscribeID       string
	ExtraNonce1       string
	ExtraNonce2Length float64
}

// Basic reply is a reply type for any of the simple messages.
type BasicReply struct {
	ID     interface{} `json:"id"`
	Error  interface{} `json:"error,omitempty"`
	Result bool        `json:"result"`
}

// StratumRsp is the basic response type from stratum.
type StratumRsp struct {
	Method string `json:"method"`
	// Need to make generic.
	ID     interface{}      `json:"id"`
	Error  StratErr         `json:"error,omitempty"`
	Result *json.RawMessage `json:"result,omitempty"`
}

// StratErr is the basic error type (a number and a string) sent by
// the stratum server.
type StratErr struct {
	ErrNum uint64
	ErrStr string
	Result *json.RawMessage `json:"result,omitempty"`
}

// StratumMsg is the basic message object from stratum.
type StratumMsg struct {
	Method string `json:"method"`
	// Need to make generic.
	Params []string    `json:"params"`
	ID     interface{} `json:"id"`
}

// NotifyWork holds all the info recieved from a mining.notify message along
// with the Work data generate from it.
type NotifyWork struct {
	Clean             bool
	Target            *big.Int
	ExtraNonce1       string
	ExtraNonce2       string
	ExtraNonce2Length float64
	Nonce2            uint32
	CB1               string
	CB2               string
	CB3               string
	CB4               string
	Height            int64
	NtimeDelta        int64
	JobID             string
	Hash              string
	Nbits             string
	Ntime             string
	Version           string
	NewWork           bool
	StateRoot         string
	MerkleBranches    []string
	WorkData          []byte
	LatestJobTime     uint64
	PowType           pow.PowType
	CuckooProof       [169]byte
}
type QitmeerStratum struct {
	core.Stratum
	Target   *big.Int
	Diff     float64
	PoolWork NotifyWork
}

func (s *QitmeerStratum) CalcBasePowLimit() *big.Int {
	switch s.PowType {
	case pow.MEERXKECCAKV1:
		return new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), s.Cfg.OptionConfig.BaseDiff), big.NewInt(1))
	}
	return params.MainNetParams.PowConfig.Blake2bdPowLimit
}

func (this *QitmeerStratum) HandleReply() {
	var resp interface{}
	var err error
	this.Stratum.Listen(func(data string) {
		resp, err = this.Unmarshal([]byte(data))
		if err != nil {
			common.MinerLoger.Error(data + err.Error())
			return
		}
		switch resp.(type) {
		case StratumMsg:
			this.handleStratumMsg(resp)
		case NotifyRes:
			common.MinerLoger.Debug(fmt.Sprintf("[pool notify message]:%s", data))
			this.handleNotifyRes(resp)
		case *SubscribeReply:
			this.handleSubscribeReply(resp)
		case *BasicReply:
			this.HandleSubmitReply(resp)
		default:
			this.HandleSubmitReply(resp)
			common.MinerLoger.Debug("[Unhandled message]: ", "data", data)
		}
	})
}

func (s *QitmeerStratum) handleSubscribeReply(resp interface{}) {
	nResp := resp.(*SubscribeReply)
	s.PoolWork.ExtraNonce1 = nResp.ExtraNonce1
	s.PoolWork.ExtraNonce2Length = nResp.ExtraNonce2Length
}

func (s *QitmeerStratum) HandleSubmitReply(resp interface{}) {
	aResp := resp.(*BasicReply)
	if int(aResp.ID.(float64)) == int(s.AuthID) {
		if aResp.Result {
			common.MinerLoger.Info("[pool reply]Logged in")
		} else {
			common.MinerLoger.Error("[pool reply]Auth failure.")
		}
	} else {
		if aResp.Result {
			s.ValidShares++
			common.MinerLoger.Info("[pool reply]Share accepted")
		} else {
			s.InvalidShares++
			common.MinerLoger.Error("[pool reply]Share rejected:", "reason", aResp.Error)
		}
	}
}

func (s *QitmeerStratum) handleStratumMsg(resp interface{}) {
	nResp := resp.(StratumMsg)
	// Too much is still handled in unmarshaler.  Need to
	// move stuff other than unmarshalling here.
	switch nResp.Method {
	case "client.show_message":
		common.MinerLoger.Debug("client.show_message", "params", nResp.Params)
	case "client.reconnect":
		common.MinerLoger.Debug("Reconnect requested")
		wait, err := strconv.Atoi(nResp.Params[2])
		if err != nil {
			common.MinerLoger.Debug(err.Error())
			return
		}
		common.Usleep(wait)
		pool := nResp.Params[0] + ":" + nResp.Params[1]
		s.Cfg.PoolConfig.Pool = pool
		err = s.Reconnect()
		if err != nil {
			common.MinerLoger.Debug(err.Error())
			// XXX should just die at this point
			// but we don't really have access to
			// the channel to end everything.
			return
		}

	case "client.get_version":
		common.MinerLoger.Debug("get_version request received.")
		msg := StratumMsg{
			Method: nResp.Method,
			ID:     nResp.ID,
			Params: []string{"github.com/Qitmeer/qitmeer/cmd/miner/v0.0.1"},
		}
		m, err := json.Marshal(msg)
		if err != nil {
			common.MinerLoger.Debug(err.Error())
			return
		}
		_, err = s.Conn.Write(m)
		if err != nil {
			common.MinerLoger.Debug(err.Error())
			return
		}
		_, err = s.Conn.Write([]byte("\n"))
		if err != nil {
			common.MinerLoger.Debug(err.Error())
			return
		}
	}
}

func (s *QitmeerStratum) handleNotifyRes(resp interface{}) {
	s.Lock()
	defer s.Unlock()
	nResp := resp.(NotifyRes)
	s.PoolWork.JobID = nResp.JobID
	s.PoolWork.CB1 = nResp.GenTX1
	s.PoolWork.Hash = nResp.Hash
	s.PoolWork.MerkleBranches = nResp.MerkleBranches
	s.PoolWork.CB2 = nResp.GenTX2
	s.PoolWork.CB3 = nResp.CB3
	s.PoolWork.CB4 = nResp.CB4
	s.PoolWork.Nbits = nResp.Nbits
	s.PoolWork.Version = nResp.BlockVersion
	s.PoolWork.CuckooProof = [169]byte{}
	s.PoolWork.PowType = s.PowType
	s.PoolWork.StateRoot = nResp.StateRoot
	s.PoolWork.NewWork = true
	hei, _ := strconv.Atoi(nResp.JobID)
	s.PoolWork.Height = int64(hei)
	parsedNtime, err := strconv.ParseInt(nResp.Ntime, 16, 64)
	if err != nil {
		common.MinerLoger.Error(err.Error())
	}
	//sync the pool base difficulty
	s.Target, _ = common.DiffToTarget(s.Diff, s.CalcBasePowLimit(), s.PowType)
	s.PoolWork.Ntime = nResp.Ntime
	s.PoolWork.NtimeDelta = parsedNtime - time.Now().Unix()
	s.PoolWork.Clean = nResp.CleanJobs
}

// Unmarshal provides a json unmarshaler for the commands.
// I'm sure a lot of this can be generalized but the json we deal with
// is pretty yucky.
func (s *QitmeerStratum) Unmarshal(blob []byte) (interface{}, error) {
	s.Lock()
	defer s.Unlock()
	var (
		objmap map[string]json.RawMessage
		method string
		id     uint64
	)

	err := json.Unmarshal(blob, &objmap)
	if err != nil {
		return nil, err
	}
	// decode command
	// Not everyone has a method.
	if _, ok := objmap["method"]; ok {
		err = json.Unmarshal(objmap["method"], &method)
		if err != nil {
			method = ""
		}
	}
	if _, ok := objmap["id"]; ok {
		err = json.Unmarshal(objmap["id"], &id)
		if err != nil {
			return nil, err
		}
		if id == s.SubID {
			var resi []interface{}
			err := json.Unmarshal(objmap["result"], &resi)
			if err != nil {
				return nil, err
			}
			resp := &SubscribeReply{}

			var objmap2 map[string]json.RawMessage
			err = json.Unmarshal(blob, &objmap2)
			if err != nil {
				return nil, err
			}

			var resJS []json.RawMessage
			err = json.Unmarshal(objmap["result"], &resJS)
			if err != nil {
				return nil, err
			}

			if len(resJS) == 0 {
				return nil, errors.New("json wrong")
			}
			var msgPeak []interface{}
			err = json.Unmarshal(resJS[0], &msgPeak)
			if err != nil {
				return nil, err
			}
			// The pools do not all agree on what this message looks like
			// so we need to actually look at it before unmarshalling for
			// real so we can use the right form.  Yuck.
			if msgPeak[0] == "mining.notify" {
				var innerMsg []string
				err = json.Unmarshal(resJS[0], &innerMsg)
				if err != nil {
					return nil, err
				}
				resp.SubscribeID = innerMsg[1]
			} else {
				var innerMsg [][]string
				err = json.Unmarshal(resJS[0], &innerMsg)
				if err != nil {
					return nil, err
				}

				for i := 0; i < len(innerMsg); i++ {
					if innerMsg[i][0] == "mining.notify" {
						resp.SubscribeID = innerMsg[i][1]
					}
					if innerMsg[i][0] == "mining.set_difficulty" {
						// Not all pools correctly put something
						// in here so we will ignore it (we
						// already have the default value of 1
						// anyway and pool can send a new one.
						// dcr.coinmine.pl puts something that
						// is not a difficulty here which is why
						// we ignore.
					}
				}
			}

			resp.ExtraNonce1 = resi[1].(string)
			resp.ExtraNonce2Length = resi[2].(float64)
			return resp, nil
		}
	}
	switch method {
	case "mining.notify":
		var resi []interface{}
		err := json.Unmarshal(objmap["params"], &resi)
		if err != nil {
			return nil, err
		}
		var nres = NotifyRes{}
		if len(resi) < 9 {
			common.MinerLoger.Debug("[error pool notify data]", "error", resi)
			return nil, errors.New("data error")
		}
		jobID, ok := resi[0].(string)
		if !ok {
			return nil, core.ErrJsonType
		}
		nres.JobID = jobID
		hash, ok := resi[1].(string)
		if !ok {
			return nil, core.ErrJsonType
		}
		nres.Hash = hash
		genTX1, ok := resi[2].(string)
		if !ok {
			return nil, core.ErrJsonType
		}
		nres.GenTX1 = genTX1
		genTX2, ok := resi[3].(string)
		if !ok {
			return nil, core.ErrJsonType
		}
		nres.GenTX2 = genTX2
		cb3, ok := resi[4].(string)
		if !ok {
			return nil, core.ErrJsonType
		}
		nres.CB3 = cb3
		cb4, ok := resi[5].(string)
		if !ok {
			return nil, core.ErrJsonType
		}
		nres.CB4 = cb4
		//ccminer code also confirms this
		transactions := resi[6].([]interface{})
		for _, v := range transactions {
			nres.MerkleBranches = append(nres.MerkleBranches, v.(string))
		}
		blockVersion, ok := resi[7].(string)
		if !ok {
			return nil, core.ErrJsonType
		}
		nres.BlockVersion = blockVersion
		nbits, ok := resi[8].(string)
		if !ok {
			return nil, core.ErrJsonType
		}
		nres.Nbits = nbits
		ntime, ok := resi[9].(string)
		if !ok {
			return nil, core.ErrJsonType
		}
		nres.Ntime = ntime

		if len(resi) <= 11 {
			cleanJobs, ok := resi[10].(bool)
			if !ok {
				return nil, core.ErrJsonType
			}
			nres.CleanJobs = cleanJobs
		} else { //add stateroot
			stateRoot, ok := resi[10].(string)
			if !ok {
				return nil, core.ErrJsonType
			}
			nres.StateRoot = stateRoot
			cleanJobs, ok := resi[11].(bool)
			if !ok {
				return nil, core.ErrJsonType
			}
			nres.CleanJobs = cleanJobs
		}

		return nres, nil

	case "mining.set_difficulty":
		var resi []interface{}
		err := json.Unmarshal(objmap["params"], &resi)
		if err != nil {
			return nil, err
		}

		difficulty, ok := resi[0].(float64)
		if !ok {
			return nil, core.ErrJsonType
		}
		powLimit := s.CalcBasePowLimit()
		s.Target, err = common.DiffToTarget(difficulty, powLimit, s.PowType)
		if err != nil {
			return nil, err
		}
		s.Diff = difficulty
		var nres = StratumMsg{}
		nres.Method = method
		diffStr := strconv.FormatFloat(difficulty, 'E', -1, 32)
		var param []string
		param = append(param, diffStr)
		nres.Params = param
		s.PoolWork.Clean = true // clean task
		common.MinerLoger.Debug("[pool reply]Stratum difficulty set to ", "value", difficulty)
		return nres, nil
	default:
		resp := &BasicReply{}
		err := json.Unmarshal(blob, &resp)
		if err != nil {
			common.MinerLoger.Debug(string(blob))
			return nil, err
		}
		return resp, nil
	}
}

func (s *NotifyWork) PrepQitmeerWork() []byte {
	cH1 := s.CB2 + s.ExtraNonce1 + s.ExtraNonce2 + s.CB3
	coinbaseD1, _ := hex.DecodeString(cH1)
	coinbaseH1 := qitmeer.DoubleHashH(coinbaseD1)
	coinbase := s.CB1 + hex.EncodeToString(coinbaseH1[:]) + s.CB4
	coinbaseD, _ := hex.DecodeString(coinbase)
	coinbaseH := qitmeer.DoubleHashH(coinbaseD)
	coinbase_hash_bin := coinbaseH[:]
	merkle_root := string(coinbase_hash_bin)
	for _, h := range s.MerkleBranches {
		d, _ := hex.DecodeString(h)
		bs := merkle_root + string(d)
		merkle_root = string(qitmeer.DoubleHashB([]byte(bs)))
	}
	merkleRootStr := hex.EncodeToString([]byte(merkle_root))
	ddd, _ := hex.DecodeString(merkleRootStr)

	ddd = common.Reverse(ddd)
	merkleRootStr2 := hex.EncodeToString(ddd)
	nonceStr := fmt.Sprintf("%016x", 0)
	//pool tx hash has converse every 4 bit
	// prevHash :=s.Hash
	tmpBytes, _ := hex.DecodeString(s.Hash)
	normalBytes := common.ReverseByWidth(tmpBytes, 1)
	prevHash := hex.EncodeToString(normalBytes)
	stateBytes, _ := hex.DecodeString(s.StateRoot)
	stateBytes = common.ReverseByWidth(stateBytes, 1)
	stateRoot := hex.EncodeToString(stateBytes)
	ntime, _ := hex.DecodeString(s.Ntime)
	blockheader := s.Version + prevHash + merkleRootStr2 + stateRoot + s.Nbits + hex.EncodeToString(ntime) +
		hex.EncodeToString([]byte{uint8(s.PowType)}) + nonceStr + hex.EncodeToString(s.CuckooProof[:])
	workData, _ := hex.DecodeString(blockheader)

	return workData
}

// PrepWork converts the stratum notify to getwork style data for mining.
func (s *NotifyWork) PrepWork() error {
	var givenTs uint32
	s.ExtraNonce2 = "00000000"
	s.WorkData = s.PrepQitmeerWork()
	if s.WorkData == nil {
		return errors.New("Not Have New Work")
	}
	givenTs = binary.LittleEndian.Uint32(s.WorkData[TIMESTART:TIMEEND])
	s.LatestJobTime = uint64(givenTs)
	return nil
}

func (s *QitmeerStratum) PrepSubmit(data []byte, jobID string, ExtraNonce2 string) (Submit, error) {
	sub := Submit{}
	sub.Method = "mining.submit"
	// Format data to send off.
	s.ID++
	sub.ID = s.ID
	s.SubmitIDs = append(s.SubmitIDs, s.ID)
	var timestampStr, nonceStr string
	timeD := data[TIMESTART:TIMEEND]
	timestampStr = hex.EncodeToString(common.Reverse(timeD[:])[0:4])
	nonceStr = hex.EncodeToString(common.Reverse(data[NONCESTART:NONCEEND]))
	if jobID != s.PoolWork.JobID && s.PoolWork.Clean {
		return sub, ErrStratumStaleWork
	}
	workArr := strings.Split(s.Cfg.PoolConfig.PoolUser, ".")
	workId := workArr[0]
	if len(workArr) > 1 {
		workId = workArr[1]
	}
	// every 100 shares 1 author  1% fee
	sub.Params = []string{workId, jobID, ExtraNonce2, timestampStr, nonceStr}
	common.MinerLoger.Debug("[submit]:", "{PoolUser, jobID, ExtraNonce2, timestampStr,nonceStr}",
		[]string{workId, jobID, ExtraNonce2, timestampStr, nonceStr})
	return sub, nil
}
