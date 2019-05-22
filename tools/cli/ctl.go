package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:               "cli",
	Long:              `cli is a RPC tool for noxd`,
	PersistentPreRunE: rootCmdPreRun,
}

func init() {
	rootCmd.AddCommand(GenerateCmd)

	rootCmd.AddCommand(GetBlockCountCmd)
	rootCmd.AddCommand(GetBlockHashCmd)
	rootCmd.AddCommand(GetBlockCmd)

	rootCmd.AddCommand(GetMempoolCmd)

	rootCmd.AddCommand(GetRawTransactionCmd)

	rootCmd.AddCommand(CreateRawTransactionCmd)
	rootCmd.AddCommand(DecodeRawTransactionCmd)
	rootCmd.AddCommand(SendRawTransactionCmd)
	rootCmd.AddCommand(TxSignCmd)
	rootCmd.AddCommand(GetUtxoCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return
}

//GetBlockCountCmd get block count
var GetBlockCountCmd = &cobra.Command{
	Use:   "getblockcount",
	Short: "get block count",
	Example: `
		getblockcount 
	`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {

		params := []interface{}{}
		blockCount, err := getResString("getBlockCount", params)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(blockCount)
		}
	},
}

//GetBlockHashCmd get block hash by number
var GetBlockHashCmd = &cobra.Command{
	Use:   "getblockhash {number}",
	Short: "get block hash by number",
	Example: `
		getblockhash 100 
	`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		blockNUmber, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Println("block number is not int")
			return
		}

		params := []interface{}{blockNUmber}

		blockHash, err := getResString("getBlockhash", params)
		if err != nil {
			fmt.Println(err)
		} else {
			// blockHash= " \"xxxx\" "
			blockHash = strings.Trim(blockHash, "\"")
			fmt.Println(blockHash)
		}
	},
}

//GetBlockCmd get block by number or hash
var GetBlockCmd = &cobra.Command{
	Use:   "getblock {number|hash} {bool,show detail,defalut true}",
	Short: "get block by number or hash",
	Example: `
		getblock 100 false
		getblock 100
		getblock 000000e4c6b7f5b89827711d412957bfff5c51730df05c2eedd1352468313eca
		getblock 000000e4c6b7f5b89827711d412957bfff5c51730df05c2eedd1352468313eca true
	`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		var err error
		var blockHash string

		if len(args[0]) != 64 {
			//block number
			var blockNUmber int64
			blockNUmber, err = strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				fmt.Println("block number is not int or hash wrong")
				return
			}

			blockHash, err = getResString("getBlockhash", []interface{}{blockNUmber})
			if err != nil {
				fmt.Println(err)
				return
			}
			// blockHash= " \"xxxx\" "
			blockHash = strings.Trim(blockHash, "\"")
		} else {
			blockHash = args[0]
		}

		var isDetail bool = true
		if len(args) > 1 {
			isDetail, err = strconv.ParseBool(args[1])
			if err != nil {
				fmt.Println("isDetail bool true or false", err)
				return
			}
		}

		getBlockParam := []interface{}{}
		getBlockParam = append(getBlockParam, blockHash)
		getBlockParam = append(getBlockParam, isDetail)

		var blockInfo string
		blockInfo, err = getResString("getBlock", getBlockParam)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(blockInfo)
		}
	},
}

//GetMempoolCmd get mempool
var GetMempoolCmd = &cobra.Command{
	Use:   "getmempool {type string, defalut regular} {verbose bool,defalut false}",
	Short: "get mempool",
	Example: `
		getmempool
		getmempool regular false
		getmempool false
	`,
	Run: func(cmd *cobra.Command, args []string) {

		var err error
		var gtype string = "regular"
		var verbose bool = false

		if len(args) == 1 {
			if args[0] == "true" || args[0] == "false" {
				verbose, _ = strconv.ParseBool(args[0])
			} else {
				gtype = args[0]
			}
		} else if len(args) > 1 {
			if verbose, err = strconv.ParseBool(args[1]); err != nil {
				fmt.Println("verbose true or false", err)
				return
			}
		}

		getBlockParam := []interface{}{}
		getBlockParam = append(getBlockParam, gtype)
		getBlockParam = append(getBlockParam, verbose)

		var blockInfo string
		blockInfo, err = getResString("getMempool", getBlockParam)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(blockInfo)
		}
	},
}

//GetRawTransactionCmd getrawtransaction
var GetRawTransactionCmd = &cobra.Command{
	Use:     "getrawtransaction {tx_hash} {verbose bool,show detail,defalut true}",
	Aliases: []string{"tx", "getrawtx", "getRawTransaction"},
	Short:   "getrawtransaction",
	Example: `
		getrawtransaction 000000e4c6b7f5b89827711d412957bfff5c51730df05c2eedd1352468313eca
		getrawtransaction 000000e4c6b7f5b89827711d412957bfff5c51730df05c2eedd1352468313eca true
	`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		var err error

		var txHash string
		var verbose bool = true

		if len(args) > 1 {
			verbose, err = strconv.ParseBool(args[1])
			if err != nil {
				fmt.Println("verbose bool true or false", err)
				return
			}
		}

		var txInfo string
		txInfo, err = getResString("getRawTransaction", []interface{}{txHash, verbose})
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(txInfo)
		}
	},
}

//CreateRawTransactionCmd CreateRawTransactionCmd
var CreateRawTransactionCmd = &cobra.Command{
	Use:     "createrawtransaction {tx}",
	Aliases: []string{"createrawtx", "createRawTransaction"},
	Short:   "createRawTransaction",
	Example: `
		createRawTransaction xx
	`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		var err error

		var tx string

		var rawTx string
		rawTx, err = getResString("createRawTransaction", []interface{}{tx})
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(rawTx)
		}
	},
}

//DecodeRawTransactionCmd DecodeRawTransactionCmd
var DecodeRawTransactionCmd = &cobra.Command{
	Use:     "decoderawtransaction {raw_tx}",
	Aliases: []string{"decoderawtx", "decodeRawTransaction"},
	Short:   "decodeRawTransaction",
	Example: `
		decoderawtx xx
	`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		var err error
		var tx string
		tx, err = getResString("decodeRawTransaction", []interface{}{args[0]})
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(tx)
		}
	},
}

//SendRawTransactionCmd SendRawTransactionCmdl
var SendRawTransactionCmd = &cobra.Command{
	Use:     "sendrawtransaction {raw_tx} {allow_high_fee bool,defalut false}",
	Aliases: []string{"sendRawTx", "sendrawtx", "sendRawTransaction"},
	Short:   "sendRawTransaction",
	Example: `
		sendRawTransaction raw_tx
		sendRawTransaction raw_tx true
	`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		var allowHighFee bool = false
		if len(args) > 1 {
			allowHighFee, err = strconv.ParseBool(args[1])
			if err != nil {
				fmt.Println("allowHighFee bool true or false", err)
				return
			}
		}

		var rs string
		rs, err = getResString("sendRawTransaction", []interface{}{args[0], allowHighFee})
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(rs)
		}
	},
}

//TxSignCmd TxSignCmd
var TxSignCmd = &cobra.Command{
	Use:   "txSign {private_key} {raw_tx}",
	Short: "txSign private_key raw_tx",
	Example: `
	txSign private_key raw_tx
	`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var rs string
		rs, err = getResString("txSign", []interface{}{args[0], args[1]})
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(rs)
		}
	},
}

//GetUtxoCmd GetUtxoCmd
var GetUtxoCmd = &cobra.Command{
	Use:     "getUtxo {tx_hash} {vout index} {include_mempool,bool,defalut true}",
	Short:   "getUtxo tx_hash vout include_mempool,",
	Aliases: []string{"getutxo"},
	Example: `
		getUtxo xx
	`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		params := []interface{}{}
		params = append(params, args[0])

		var vout int64
		vout, err = strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			fmt.Println("vout not number", err)
			return
		}
		params = append(params, vout)

		var includeMempool bool = true
		if len(args) > 2 {
			includeMempool, err = strconv.ParseBool(args[2])
			if err != nil {
				fmt.Println("include_mempool true or false", err)
				return
			}
		}
		params = append(params, includeMempool)

		var tx string

		tx, err = getResString("getUtxo", params)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(tx)
		}
	},
}

//GenerateCmd cpu mine block
var GenerateCmd = &cobra.Command{
	Use:   "generate {number,default latest}",
	Short: "generate {n}, cpu mine n blocks",
	Example: `
		generate
		generate 1
	`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		params := []interface{}{}
		if len(args) == 0 {
			params = append(params, "latest")
		} else {
			number, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				fmt.Println("number error:", err)
				return
			}
			params = append(params, number)
		}

		var rs string
		rs, err = getResString("generate", params)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(rs)
		}
	},
}
