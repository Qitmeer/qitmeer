// Copyright 2017-2018 The nox developers

package main

import (
    "encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"log"
	"io"
	"time"
	"io/ioutil"
	"strings"
)

var (
	args  []string  //trick to test command args from test-case
	hexString = flag.String("hex", "", "input is hex string")
	debug = flag.String("D", "", "show debug log")
	logger = log.New(ioutil.Discard,"",log.LstdFlags)
)

func init() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "[HexString]")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "Convert the given hex string to the base64 str.")
	}
}

func main() {
	var input string
	if args != nil {
		flag.CommandLine.Parse(args)
		defer resetFlags()  // it's required because the args trick
	}else {
		flag.Parse()  //from os.Args[1:]
	}
	switch *debug { // show debug log
	case "std" :
		logger = log.New(io.Writer(os.Stderr), "", log.LstdFlags)
	case "log" :
		currentTime := time.Now().Format("2006-01-02_15.04.05")
		logfile, _ := os.OpenFile("./"+os.Args[0]+"_"+currentTime+".log", os.O_RDWR|os.O_CREATE, 0666)
		logger = log.New(io.Writer(logfile), "", log.LstdFlags)
	case "":
	default:
		fmt.Fprintf(os.Stderr, "Error : wrong debug option '%s'.\n",*debug)
		os.Exit(1)
	}
	switch{
	case *hexString != "" :    // from -hex
		input = *hexString
		logger.Printf("from -hex, flag number=%d , input=%s\n",flag.NArg(),input)
	case flag.NArg() == 0 :    // from stdin terminal/file/pipe
		logger.Printf("from stdin,flag number is 0")
		stat, _ := os.Stdin.Stat()
		fileMode:=stat.Mode()
		if  (fileMode & os.ModeCharDevice) != 0 {
			logger.Printf( "stdin terminal %v",fileMode)   //from terminal stdin
		} else if (stat.Mode() & os.ModeNamedPipe) != 0 {
		    logger.Printf( "stdin pipe %v",fileMode)  //from stdin pipe
			myinput, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprint(os.Stderr,err)
				os.Exit(1)
			}
			input = strings.TrimSpace(string(myinput))
		}else{
			logger.Printf( "stdin file %v",fileMode)  //from stdin file
			fileInput,err := ioutil.ReadAll(os.Stdin)
			if err!=nil {
				fmt.Fprint(os.Stderr,err)
				os.Exit(1)
			}
			input = strings.TrimSpace(string(fileInput))
		}
	case flag.NArg() == 1 :    // from input str
		input = flag.Arg(0)
		logger.Printf("from arg 1,falg number is 1, input=%s\n", input)
	default :                  // error, print usage
		fmt.Fprintln(os.Stderr, "Error : too many arguments.")
		flag.Usage()
		os.Exit(1)
	}
	logger.Printf("input  is '%s'\n",input)
	base64str, err := HexStr2base64Str(input);
	if err != nil {
		fmt.Fprint(os.Stderr, err)
	}
	logger.Printf("output is '%s'\n",base64str)
	fmt.Fprintln(os.Stdout,base64str)
}

func resetFlags(){
	flag.Set("hex","")
	flag.Set("D","")
}

func HexStr2base64Str (hexStr string) (string,error) {
	decoded, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", err
	}
	encodeToStr := base64.StdEncoding.EncodeToString(decoded)
	return encodeToStr,nil
}
