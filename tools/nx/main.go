// Copyright 2017-2018 The nox developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/noxproject/nox/common/encode/base58"
	"io/ioutil"
	"os"
	"strconv"
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

var base58version string
var showDecodeDetails bool


func main() {
	encodeCommand := flag.NewFlagSet("base58check-encode", flag.ExitOnError)
	encodeCommand.StringVar(&base58version, "v","0df1","base58check version")

	decodeCommand := flag.NewFlagSet("base58check-decode", flag.ExitOnError)
	decodeCommand.BoolVar(&showDecodeDetails,"d",false, "show decode datails")

	if len(os.Args) == 1 {
		usage()
	}
	switch os.Args[1]{
	case "help","--help" :
		usage()
	case "version","--version":
		version()
	case "base58check-encode" :
		encodeCommand.Parse(os.Args[2:])
	case "base58check-decode" :
		decodeCommand.Parse(os.Args[2:])
	default:
		invalid := os.Args[1]
		if invalid[0] == '-' {
			fmt.Fprintf(os.Stderr, "unknown option: %q \n", invalid)
		}else {
			fmt.Fprintf(os.Stderr, "%q is not valid command\n", invalid)
		}
		os.Exit(1)
	}

	if encodeCommand.Parsed(){
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			switch os.Args[2] {
			case "help","--help":
				fmt.Fprintf(os.Stderr, "Usage: nx base58check-encode [-v <ver>] [hexstring]\n")
				encodeCommand.PrintDefaults()
			default:
				base58CheckEncode(base58version,os.Args[len(os.Args)-1])
			}
		}else {  //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			base58CheckEncode(base58version,str)
		}
	}

	if decodeCommand.Parsed(){
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			switch os.Args[2] {
			case "help","--help":
				fmt.Fprintf(os.Stderr, "Usage: nx base58check-decode [-d] [hexstring]\n")
				decodeCommand.PrintDefaults()
			default:
				base58CheckDecode(os.Args[len(os.Args)-1])
			}
		}else {  //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			base58CheckDecode(str)
		}
	}
}

func base58CheckEncode(version string, input string){
	data, err := hex.DecodeString(input)
	if err!=nil {
		errExit(err)
	}
	ver, err := hex.DecodeString(version)
	if err !=nil {
		errExit(err)
	}
	if len(ver) != 2 {
		errExit(fmt.Errorf("invaid version byte"))
	}
	var versionByte [2]byte
	versionByte[0] = ver[0]
	versionByte[1] = ver[1]
	encoded := base58.CheckEncode(data, versionByte)
	// Show the encoded data.
	//fmt.Printf("Encoded Data ver[%v] : %s\n",ver, encoded)
	fmt.Printf("%s\n",encoded)
}

func base58CheckDecode(input string) {
	cksum, err := base58.CheckInput(input)
	if err != nil {
		errExit(err)
	}
	data, ver, err := base58.CheckDecode(input)
	if err != nil {
		errExit(err)
	}
	if showDecodeDetails {
		fmt.Printf("payload : %x\n", data)
		dec,err := strconv.ParseUint(hex.EncodeToString(cksum), 16, 64)
		if err!=nil {
			errExit(err)
		}
		fmt.Printf("checksum: %d (%x)\n",dec, cksum)
		fmt.Printf("version : %x\n", ver)
	}else {
		fmt.Printf("%x\n", data)
	}
}