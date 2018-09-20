// Copyright 2017-2018 The nox developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

const (
	NX_VERSION = "0.0.1"
)

func usage() {
	fmt.Fprintf(os.Stderr,"Usage: nx [--version] [--help] <command> [<args>]\n")
	fmt.Fprintf(os.Stderr,`
Nox commmands :
    base58check-encode    Encode base58check hex string from stdin
    base58check-decode    Decode base58check hex string from stdin
`)
	os.Exit(1)
}

func version() {
	fmt.Fprintf(os.Stderr,"Nx Version : %q\n",NX_VERSION)
	os.Exit(1)
}

func errExit(err error){
	fmt.Fprintf(os.Stderr, "Nx Error : %q\n",err)
	os.Exit(1)
}

var base58CheckVer string
var showDecodeDetails bool
var decodeMode string

func main() {

	base58CheckEncodeCommand := flag.NewFlagSet("base58check-encode", flag.ExitOnError)
	base58CheckEncodeCommand.StringVar(&base58CheckVer, "v","0df1","base58check version")

	base58CheckDecodeCommand := flag.NewFlagSet("base58check-decode", flag.ExitOnError)
	base58CheckDecodeCommand.BoolVar(&showDecodeDetails,"d",false, "show decode datails")
	base58CheckDecodeCommand.StringVar(&decodeMode,"m","nox", "base58 decode mode : [nox|btc]")

	base58EncodeCmd := flag.NewFlagSet("base58encode",flag.ExitOnError)
	base58DecodeCmd := flag.NewFlagSet("base58decode",flag.ExitOnError)

	if len(os.Args) == 1 {
		usage()
	}
	switch os.Args[1]{
	case "help","--help" :
		usage()
	case "version","--version":
		version()
	case "base58check-encode" :
		base58CheckEncodeCommand.Parse(os.Args[2:])
	case "base58check-decode" :
		base58CheckDecodeCommand.Parse(os.Args[2:])
	default:
		invalid := os.Args[1]
		if invalid[0] == '-' {
			fmt.Fprintf(os.Stderr, "unknown option: %q \n", invalid)
		}else {
			fmt.Fprintf(os.Stderr, "%q is not valid command\n", invalid)
		}
		os.Exit(1)
	}
	// Handle base58check-encode
	if base58CheckEncodeCommand.Parsed(){
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			switch os.Args[2] {
			case "help","--help":
				fmt.Fprintf(os.Stderr, "Usage: nx base58check-encode [-v <ver>] [hexstring]\n")
				base58CheckEncodeCommand.PrintDefaults()
			default:
				base58CheckEncode(base58CheckVer,os.Args[len(os.Args)-1])
			}
		}else {  //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			base58CheckEncode(base58CheckVer,str)
		}
	}

	// Handle base58check-decode
	if base58CheckDecodeCommand.Parsed(){
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			switch os.Args[2] {
			case "help","--help":
				fmt.Fprintf(os.Stderr, "Usage: nx base58check-decode [-d] [hexstring]\n")
				base58CheckDecodeCommand.PrintDefaults()
			default:
				base58CheckDecode(decodeMode,os.Args[len(os.Args)-1])
			}
		}else {  //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			base58CheckDecode(decodeMode,str)
		}
	}

	// Handle base58-encode
	if base58EncodeCmd.Parsed(){

	}
	// Handle base58-decode
	if base58DecodeCmd.Parsed(){

	}
}

