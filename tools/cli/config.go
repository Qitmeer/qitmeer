package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

var cfg = &Config{}

/*  */
func init() {
	rootCmd.PersistentFlags().StringVar(&cfg.ConfigFile, "conf", "config.toml", "RPC username")

	rootCmd.PersistentFlags().StringVarP(&cfg.RPCUser, "user", "u", "", "RPC username")
	rootCmd.PersistentFlags().StringVarP(&cfg.RPCPassword, "password", "P", "", "RPC password")
	rootCmd.PersistentFlags().StringVarP(&cfg.RPCServer, "server", "s", "127.0.0.1:18131", "RPC server to connect to")
	rootCmd.PersistentFlags().StringVar(&cfg.RPCCert, "c", "", "RPC server certificate file path")

	rootCmd.PersistentFlags().BoolVar(&cfg.NoTLS, "notls", true, "Do not verify tls certificates (not recommended!)")
	rootCmd.PersistentFlags().BoolVar(&cfg.TLSSkipVerify, "skipverify", true, "Do not verify tls certificates (not recommended!)")

	rootCmd.PersistentFlags().StringVar(&cfg.Proxy, "proxy", "", "Connect via SOCKS5 proxy (eg. 127.0.0.1:9050)")
	rootCmd.PersistentFlags().StringVar(&cfg.ProxyUser, "proxyuser", "", "Username for proxy server")
	rootCmd.PersistentFlags().StringVar(&cfg.ProxyPass, "proxypass", "", "Password for proxy server")

	rootCmd.PersistentFlags().BoolVar(&cfg.TestNet, "testnet", false, "Connect to testnet")
	rootCmd.PersistentFlags().BoolVar(&cfg.SimNet, "simnet", false, "Connect to the simulation test network")
}

type Config struct {
	ConfigFile string

	RPCUser       string
	RPCPassword   string
	RPCServer     string
	RPCCert       string
	NoTLS         bool
	TLSSkipVerify bool

	Proxy     string
	ProxyUser string
	ProxyPass string

	TestNet bool
	SimNet  bool
}

//
func rootCmdPreRun(cmd *cobra.Command, args []string) error {
	cfg2 := &Config{}
	_, decodeErr := toml.DecodeFile(cfg.ConfigFile, cfg2)

	if decodeErr != nil {
		fmt.Println("config file err:", decodeErr)
	} else {
		if !cmd.Flag("user").Changed {
			cfg.RPCUser = cfg2.RPCUser
		}
		if !cmd.Flag("password").Changed {
			cfg.RPCPassword = cfg2.RPCPassword
		}
		if !cmd.Flag("server").Changed {
			cfg.RPCServer = cfg2.RPCServer
		}

	}

	//params.MainNetParams.DefaultPort
	//	preCfg := cfg

	// Multiple networks can't be selected simultaneously.
	numNets := 0
	if cfg.TestNet {
		numNets++
	}
	if cfg.SimNet {
		numNets++
	}
	if numNets > 1 {
		return fmt.Errorf("loadConfig: %s", "one of the testnet and simnet")
	}

	//save
	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(cfg); err != nil {
		log.Fatal(err)
	}

	return ioutil.WriteFile(cfg.ConfigFile, buf.Bytes(), 0666)
}
