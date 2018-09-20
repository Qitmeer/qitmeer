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
encode and decode :
    base58-encode         encode a base16 string to a base58 string
    base58-decode         decode a base58 string to a base16 string
    base58check-encode    encode a base58check string
    base58check-decode    decode a base58check string

hash :
    blake2b256            calculate Blake2b 256 hash of a base16 data.
    sha256                calculate SHA256 hash of a base16 data. 
	ripemd160             calculate ripemd160 hash of a base16 data.
	bitcion160            calculate ripemd160(sha256(data))   
	hash160               calculate ripemd160(blake2b256(data))

seed & mnemoic & hd

addr & pbkey

sign 
`)
	os.Exit(1)
}

func cmdUsage (cmd *flag.FlagSet, usage string){
	fmt.Fprintf(os.Stderr, usage)
	cmd.PrintDefaults()
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
	base58CheckEncodeCommand.Usage = func() {
		cmdUsage(base58CheckEncodeCommand,"Usage: nx base58check-encode [-v <ver>] [hexstring]\n")
	}

	base58CheckDecodeCommand := flag.NewFlagSet("base58check-decode", flag.ExitOnError)
	base58CheckDecodeCommand.BoolVar(&showDecodeDetails,"d",false, "show decode datails")
	base58CheckDecodeCommand.StringVar(&decodeMode,"m","nox", "base58 decode mode : [nox|btc]")
	base58CheckDecodeCommand.Usage = func() {
		cmdUsage(base58CheckDecodeCommand,"Usage: nx base58check-encode [-v <ver>] [hexstring]\n")
	}

	base58EncodeCmd := flag.NewFlagSet("base58encode",flag.ExitOnError)
	base58EncodeCmd.Usage = func() {
		cmdUsage(base58EncodeCmd ,"Usage: nx base58-encode [hexstring]\n")
	}
	base58DecodeCmd := flag.NewFlagSet("base58decode",flag.ExitOnError)
	base58DecodeCmd.Usage = func() {
		cmdUsage(base58DecodeCmd, "Usage: nx base58-decode [hexstring]\n")
	}

	sha256cmd := flag.NewFlagSet("sha256",flag.ExitOnError)
	sha256cmd.Usage = func() {
		cmdUsage(sha256cmd, "Usage: nx sha256 [hexstring]\n")
	}

	blake2b256cmd := flag.NewFlagSet("blake2b256",flag.ExitOnError)
	blake2b256cmd.Usage = func() {
		cmdUsage(blake2b256cmd, "Usage: nx blak2b256 [hexstring]\n")
	}

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
	case "base58-encode" :
		base58EncodeCmd.Parse(os.Args[2:])
	case "base58-decode" :
		base58DecodeCmd.Parse(os.Args[2:])
	case "sha256":
		sha256cmd.Parse(os.Args[2:])
	case "blake2b256":
		blake2b256cmd.Parse(os.Args[2:])
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
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				base58CheckEncodeCommand.Usage()
			}else{
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
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				base58CheckDecodeCommand.Usage()
			}else{
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
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				base58EncodeCmd.Usage()
		 	}else{
				base58Encode(os.Args[len(os.Args)-1])
			}
		}else {  //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			base58Encode(str)
		}

	}
	// Handle base58-decode
	if base58DecodeCmd.Parsed(){
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				base58DecodeCmd.Usage()
			}else{
				base58Decode(os.Args[len(os.Args)-1])
			}
		}else {  //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			base58Decode(str)
		}
	}

	if sha256cmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				sha256cmd.Usage()
			}else{
				sha256(os.Args[len(os.Args)-1])
			}
		}else {  //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			sha256(str)
		}
	}

	if blake2b256cmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				blake2b256cmd.Usage()
			}else{
				blake2b256(os.Args[len(os.Args)-1])
			}
		}else {  //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			blake2b256(str)
		}
	}
}
