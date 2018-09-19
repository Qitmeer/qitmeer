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
)

const (
	NX_VERSION = "0.0.1"
)

func init() {

}

func usage() {
	fmt.Fprintf(os.Stderr,"Usage: nx [--version] [--help] <command> [<args>]\n")
	fmt.Fprintf(os.Stderr,`
Nox commmands :
    base58check-encode    Encode base58check hex string from stdin
`)
	os.Exit(1)
}

func version() {
	fmt.Fprintf(os.Stderr,"Nx Version : %q\n",NX_VERSION)
	os.Exit(1)
}



func main() {
	encodeCommand := flag.NewFlagSet("base58check-encode", flag.ExitOnError)
	var base58version string
	encodeCommand.StringVar(&base58version, "v","0df1","base58check version")

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
				fmt.Fprintf(os.Stderr, "Usage: nx base58check-encode [hexstring]\n")
				encodeCommand.PrintDefaults()
			default:
				noxBase58CheckEncode(base58version,os.Args[len(os.Args)-1])
			}
		}else {  //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				panic(err)
			}
			noxBase58CheckEncode(base58version,string(src))
		}
	}
}

func noxBase58CheckEncode(version string, input string){
	data, err := hex.DecodeString(input)
	if err!=nil {
		panic(err)
	}
	ver, err := hex.DecodeString(version)
	if err !=nil {
		panic(err)
	}
	if len(ver) != 2 {
		panic("invaid version byte")
	}
	var versionByte [2]byte
	versionByte[0] = ver[0]
	versionByte[1] = ver[1]
	encoded := base58.CheckEncode(data, versionByte)
	// Show the encoded data.
	//fmt.Printf("Encoded Data ver[%v] : %s\n",ver, encoded)
	fmt.Printf("%s\n",encoded)
}