package core

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"github.com/Qitmeer/qitmeer/cmd/miner/common"
	"github.com/Qitmeer/qitmeer/cmd/miner/common/socks"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"net"
	"strings"
	"sync"
	"time"
)

// ErrJsonType is an error for json that we do not expect.
var ErrJsonType = errors.New("Unexpected type in json.")

type StratumMsg struct {
	Method string `json:"method"`
	// Need to make generic.
	Params []string    `json:"params"`
	ID     interface{} `json:"id"`
}

type Stratum struct {
	sync.Mutex
	Cfg           *common.GlobalConfig
	Conn          net.Conn
	Reader        *bufio.Reader
	ID            uint64
	Started       uint32
	Timeout       uint32
	ValidShares   uint64
	InvalidShares uint64
	StaleShares   uint64
	SubmitIDs     []uint64
	SubID         uint64
	AuthID        uint64
	PowType       pow.PowType
	Quit          context.Context
}

func GetPowType(powName string) pow.PowType {
	switch powName {
	case "blake2bd":
		return pow.BLAKE2BD
	case "x8r16":
		return pow.X8R16
	case "x16rv3":
		return pow.X16RV3
	case "qitmeer_keccak256":
		return pow.QITMEERKECCAK256
	case "cuckaroo":
		return pow.CUCKAROO
	case "cuckaroom":
		return pow.CUCKAROOM
	case "cuckatoo":
		return pow.CUCKATOO
	case "meer_crypto":
		return pow.MEERXKECCAKV1
	}
	return pow.BLAKE2BD
}

// StratumConn starts the initial connection to a stratum pool and sets defaults
// in the pool object.
func (this *Stratum) StratumConn(cfg *common.GlobalConfig, ctx context.Context) error {
	this.Cfg = cfg
	pool := cfg.PoolConfig.Pool
	common.MinerLoger.Debug("[Connect pool]", "address", pool)
	proto := "stratum+tcp://"
	if strings.HasPrefix(this.Cfg.PoolConfig.Pool, proto) {
		pool = strings.Replace(pool, proto, "", 1)
	} else {
		err := errors.New("[error] Only stratum pools supported.stratum+tcp://")
		return err
	}
	this.Cfg.PoolConfig.Pool = pool
	this.ID = 1
	this.Quit = ctx
	this.PowType = GetPowType(cfg.NecessaryConfig.Pow)
	this.ConnectRetry()
	return nil
}

func (this *Stratum) ConnectRetry() {
	var err error
	for {
		select {
		case <-this.Quit.Done():
			common.MinerLoger.Info("pool service exit")
			return
		default:
		}
		common.Usleep(2)
		err = this.Reconnect()
		if err != nil {
			common.MinerLoger.Debug("[Connect error , It will reconnect after 2s].", "error", err.Error())
			continue
		}
		break
	}
}

func (this *Stratum) Listen(handle func(data string)) {
	common.MinerLoger.Debug("Starting Stratum Listener")
	var data string
	var err error
	// start := time.Now().Unix()
	for {
		select {
		case <-this.Quit.Done():
			common.MinerLoger.Info("pool service exit")
			return
		default:
		}
		if this.Reader != nil {
			data, err = this.Reader.ReadString('\n')
			if err != nil {
				common.MinerLoger.Error("TCP Read Error:", "detail", err.Error())
			}
		} else {
			err = errors.New("network wrong!")
		}
		if err != nil {
			this.ConnectRetry()
		}
		handle(data)
		this.Timeout = uint32(time.Now().Unix())
		// if time.Now().Unix() - start > 10{
		// 	_ = this.Conn.Close()
		// }
	}
}

// Reconnect reconnects to a stratum server if the connection has been lost.
func (s *Stratum) Reconnect() error {
	var conn net.Conn
	var err error
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}
	if s.Cfg.OptionConfig.Proxy != "" {
		proxy := &socks.Proxy{
			Addr:     s.Cfg.OptionConfig.Proxy,
			Username: s.Cfg.OptionConfig.ProxyUser,
			Password: s.Cfg.OptionConfig.ProxyPass,
		}
		conn, err = proxy.Dial("tcp", s.Cfg.PoolConfig.Pool)
	} else {
		conn, err = tls.Dial("tcp", s.Cfg.PoolConfig.Pool, conf)
	}
	if err != nil {
		common.MinerLoger.Debug("[init reconnect error]", "error", err)
		return err
	}
	s.Conn = conn
	s.Reader = bufio.NewReader(s.Conn)
	err = s.Subscribe()
	if err != nil {
		common.MinerLoger.Debug("[subscribe reconnect]", "error", err)
		return nil
	}
	// XXX Do I really need to re-auth here?
	err = s.Auth()
	if err != nil {
		common.MinerLoger.Debug("[auth reconnect]", "error", err)
		return nil
	}
	// If we were able to reconnect, restart counter
	s.Started = uint32(time.Now().Unix())
	s.Timeout = uint32(time.Now().Unix())
	return nil
}

// Auth sends a message to the pool to authorize a worker.
func (s *Stratum) Auth() error {
	/*
		defaultUser := "XmWHCtdUtPyuPCNZVzHj4rhDNN7ioCG5zA8"*/
	msg := StratumMsg{
		Method: "mining.authorize",
		ID:     s.ID,
		Params: []string{s.Cfg.PoolConfig.PoolUser, s.Cfg.PoolConfig.PoolPassword},
	}
	// Auth reply has no method so need a way to identify it.
	// Ugly, but not much choice.
	id, ok := msg.ID.(uint64)
	if !ok {
		return ErrJsonType
	}
	s.ID++
	s.AuthID = id
	m, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = s.Conn.Write(m)
	if err != nil {
		common.MinerLoger.Debug("[auth connect]", "error", err)
		return err
	}
	_, err = s.Conn.Write([]byte("\n"))
	if err != nil {
		return err
	}
	return nil
}

// Subscribe sends the subscribe message to get mining info for a worker.
func (s *Stratum) Subscribe() error {
	msg := StratumMsg{
		Method: "mining.subscribe",
		ID:     s.ID,
		Params: []string{"github.com/Qitmeer/qitmeer/cmd/miner/v0.0.1"},
	}
	s.SubID = msg.ID.(uint64)
	s.ID++
	m, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = s.Conn.Write(m)
	if err != nil {
		common.MinerLoger.Debug("[subscribe connect]", "error", err)
		return err
	}
	_, err = s.Conn.Write([]byte("\n"))
	if err != nil {
		return err
	}
	return nil
}
