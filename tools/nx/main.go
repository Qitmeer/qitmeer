// Copyright 2017-2018 The nox developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"qitmeer/crypto/seed"
	"qitmeer/wallet"
	"strings"
)

const (
	NX_VERSION = "0.0.1"
	TX_VERION  = 1 //default version is 1
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: nx [--version] [--help] <command> [<args>]\n")
	fmt.Fprintf(os.Stderr, `
encode and decode :
    base58-encode         encode a base16 string to a base58 string
    base58-decode         decode a base58 string to a base16 string
    base58check-encode    encode a base58check string
    base58check-decode    decode a base58check string
    base64-encode         encode a base16 string to a base64 string
    base64-encode         encode a base64 string to a base16 string
    rlp-encode            encode a string to a rlp encoded base16 string 
    rlp-decode            decode a rlp base16 string to a human-readble representation

hash :
    blake2b256            calculate Blake2b 256 hash of a base16 data.
    blake2b512            calculate Blake2b 512 hash of a base16 data.
    sha256                calculate SHA256 hash of a base16 data. 
    sha3-256              calculate SHA3 256 hash of a base16 data.
    keccak-256            calculate legacy keccak 256 hash of a bash16 data.
    blake256              calculate blake256 hash of a base16 data.
    ripemd160             calculate ripemd160 hash of a base16 data.
    bitcion160            calculate ripemd160(sha256(data))   
    hash160               calculate ripemd160(blake2b256(data))

entropy (seed) & mnemoic & hd & ec 
    entropy               generate a cryptographically secure pseudorandom entropy (seed)
    hd-new                create a new HD(BIP32) private key from an entropy (seed)
    hd-to-ec              convert the HD (BIP32) format private/public key to a EC private/public key
    hd-to-public          derive the HD (BIP32) public key from a HD private key
    hd-decode             decode a HD (BIP32) private/public key serialization format
    hd-derive             Derive a child HD (BIP32) key from another HD public or private key.
    mnemonic-new          create a mnemonic world-list (BIP39) from an entropy
    mnemonic-to-entropy   return back to the entropy (the random seed) from a mnemonic world list (BIP39)
    mnemonic-to-seed      convert a mnemonic world-list (BIP39) to its 512 bits seed 
    ec-new                create a new EC private key from an entropy (seed).
    ec-to-public          derive the EC public key from an EC private key (the compressed format by default )
    ec-to-wif             convert an EC private key to a WIF, associates with the compressed public key by default.
    wif-to-ec             convert a WIF private key to an EC private key.
    wif-to-public         derive the EC public key from a WIF private key. 

addr & tx & sign
    ec-to-addr            convert an EC public key to a paymant address. default is nox address
    tx-encode             encode a unsigned transaction.
    tx-decode             decode a transaction in base16 to json format.
    tx-sign               sign a transactions using a private key.
    msg-sign              create a message signature
    msg-verify            validate a message signature
    signature-decode      decode a ECDSA signature
	
`)
	os.Exit(1)
}

func cmdUsage(cmd *flag.FlagSet, usage string) {
	fmt.Fprintf(os.Stderr, usage)
	cmd.PrintDefaults()
}

func version() {
	fmt.Fprintf(os.Stderr, "Nx Version : %q\n", NX_VERSION)
	os.Exit(1)
}

func errExit(err error) {
	fmt.Fprintf(os.Stderr, "Nx Error : %q\n", err)
	os.Exit(1)
}

var base58checkVersion noxBase58checkVersionFlag
var base58checkVersionSize int
var base58checkMode string
var showDetails bool
var base58checkHasher string
var base58checkCksumSize int
var seedSize uint
var hdVer bip32VersionFlag
var hdHarden bool
var hdIndex uint
var derivePath derivePathFlag
var mnemoicSeedPassphrase string
var curve string
var uncompressedPKFormat bool
var network string
var txInputs txInputsFlag
var txOutputs txOutputsFlag
var txVersion txVersionFlag
var txLockTime txLockTimeFlag
var privateKey string
var msgSignatureMode string

func main() {

	// ----------------------------
	// cmd for encoding & decoding
	// ----------------------------

	base58CheckEncodeCommand := flag.NewFlagSet("base58check-encode", flag.ExitOnError)
	base58checkVersion = noxBase58checkVersionFlag{}
	base58checkVersion.Set("privnet")
	base58CheckEncodeCommand.Var(&base58checkVersion, "v", "base58check `version` [mainnet|testnet|privnet]")
	base58CheckEncodeCommand.StringVar(&base58checkHasher, "a", "", "base58check hasher")
	base58CheckEncodeCommand.IntVar(&base58checkCksumSize, "c", 4, "base58check checksum size")
	base58CheckEncodeCommand.Usage = func() {
		cmdUsage(base58CheckEncodeCommand, "Usage: nx base58check-encode [-v <ver>] [hexstring]\n")
	}

	base58CheckDecodeCommand := flag.NewFlagSet("base58check-decode", flag.ExitOnError)
	base58CheckDecodeCommand.BoolVar(&showDetails, "d", false, "show decode details")
	base58CheckDecodeCommand.StringVar(&base58checkMode, "m", "nox", "base58check decode `mode`: [nox|btc]")
	base58CheckDecodeCommand.StringVar(&base58checkHasher, "a", "", "base58check `hasher`")
	base58CheckDecodeCommand.IntVar(&base58checkVersionSize, "vs", 2, "base58check version `size`")
	base58CheckDecodeCommand.IntVar(&base58checkCksumSize, "cs", 4, "base58check checksum `size`")
	base58CheckDecodeCommand.Usage = func() {
		cmdUsage(base58CheckDecodeCommand, "Usage: nx base58check-decode [hexstring]\n")
	}

	base58EncodeCmd := flag.NewFlagSet("base58-encode", flag.ExitOnError)
	base58EncodeCmd.Usage = func() {
		cmdUsage(base58EncodeCmd, "Usage: nx base58-encode [hexstring]\n")
	}
	base58DecodeCmd := flag.NewFlagSet("base58-decode", flag.ExitOnError)
	base58DecodeCmd.Usage = func() {
		cmdUsage(base58DecodeCmd, "Usage: nx base58-decode [hexstring]\n")
	}

	base64EncodeCmd := flag.NewFlagSet("base64-encode", flag.ExitOnError)
	base64EncodeCmd.Usage = func() {
		cmdUsage(base64EncodeCmd, "Usage: nx base64-encode [hexstring]\n")
	}
	base64DecodeCmd := flag.NewFlagSet("base64-decode", flag.ExitOnError)
	base64DecodeCmd.Usage = func() {
		cmdUsage(base64DecodeCmd, "Usage: nx base64-decode [hexstring]\n")
	}

	rlpEncodeCmd := flag.NewFlagSet("rlp-encode", flag.ExitOnError)
	rlpEncodeCmd.Usage = func() {
		cmdUsage(rlpEncodeCmd, "Usage: nx rlp-encode [string]\n")
	}

	rlpDecodeCmd := flag.NewFlagSet("rlp-decode", flag.ExitOnError)
	rlpDecodeCmd.Usage = func() {
		cmdUsage(rlpDecodeCmd, "Usage: nx rlp-decode [hexstring]\n")
	}

	// ----------------------------
	// cmd for hashing
	// ----------------------------

	sha256cmd := flag.NewFlagSet("sha256", flag.ExitOnError)
	sha256cmd.Usage = func() {
		cmdUsage(sha256cmd, "Usage: nx sha256 [hexstring]\n")
	}

	blake2b256cmd := flag.NewFlagSet("blake2b256", flag.ExitOnError)
	blake2b256cmd.Usage = func() {
		cmdUsage(blake2b256cmd, "Usage: nx blak2b256 [hexstring]\n")
	}

	blake2b512cmd := flag.NewFlagSet("blake2b512", flag.ExitOnError)
	blake2b512cmd.Usage = func() {
		cmdUsage(blake2b512cmd, "Usage: nx blak2b512 [hexstring]\n")
	}

	blake256cmd := flag.NewFlagSet("blake256", flag.ExitOnError)
	blake256cmd.Usage = func() {
		cmdUsage(blake256cmd, "Usage: nx blake256 [hexstring]\n")
	}

	sha3_256cmd := flag.NewFlagSet("sha3-256", flag.ExitOnError)
	sha3_256cmd.Usage = func() {
		cmdUsage(sha3_256cmd, "Usage: nx sha3-256 [hexstring]\n")
	}

	keccak256cmd := flag.NewFlagSet("keccak-256", flag.ExitOnError)
	keccak256cmd.Usage = func() {
		cmdUsage(keccak256cmd, "Usage: nx keccak-256 [hexstring]\n")
	}

	ripemd160Cmd := flag.NewFlagSet("ripemd160", flag.ExitOnError)
	ripemd160Cmd.Usage = func() {
		cmdUsage(ripemd160Cmd, "Usage: nx ripemd160 [hexstring]\n")
	}

	bitcion160Cmd := flag.NewFlagSet("bitcoin160", flag.ExitOnError)
	bitcion160Cmd.Usage = func() {
		cmdUsage(bitcion160Cmd, "Usage: nx bitcoin160 [hexstring]\n")
	}

	hash160Cmd := flag.NewFlagSet("hash160", flag.ExitOnError)
	hash160Cmd.Usage = func() {
		cmdUsage(bitcion160Cmd, "Usage: nx hash160 [hexstring]\n")
	}

	// ----------------------------
	// cmd for crypto
	// ----------------------------

	// Entropy (Seed)
	entropyCmd := flag.NewFlagSet("entropy", flag.ExitOnError)
	entropyCmd.Usage = func() {
		cmdUsage(entropyCmd, "Usage: nx entropy [-s size] \n")
	}
	entropyCmd.UintVar(&seedSize, "s", seed.DefaultSeedBytes*8, "The length in bits for a seed (entropy)")

	// HD (BIP32)
	hdNewCmd := flag.NewFlagSet("hd-new", flag.ExitOnError)
	hdNewCmd.Usage = func() {
		cmdUsage(hdNewCmd, "Usage: nx hd-new [-v version] [entropy] \n")
	}
	hdVer.Set("privnet")
	hdNewCmd.Var(&hdVer, "v", "The HD(BIP32) `version` [mainnet|testnet|privnet|bip32]")

	hdToPubCmd := flag.NewFlagSet("hd-to-public", flag.ExitOnError)
	hdToPubCmd.Usage = func() {
		cmdUsage(hdToPubCmd, "Usage: nx hd-to-public [hd_private_key] \n")
	}
	hdToPubCmd.Var(&hdVer, "v", "The HD(BIP32) `version` [mainnet|testnet|privnet|bip32]")

	hdToEcCmd := flag.NewFlagSet("hd-to-ec", flag.ExitOnError)
	hdToEcCmd.Usage = func() {
		cmdUsage(hdToEcCmd, "Usage: nx hd-to-ec [hd_private_key or hd_public_key] \n")
	}
	hdToEcCmd.Var(&hdVer, "v", "The HD(BIP32) `version` [mainnet|testnet|privnet|bip32]")

	hdDecodeCmd := flag.NewFlagSet("hd-decode", flag.ExitOnError)
	hdDecodeCmd.Usage = func() {
		cmdUsage(hdDecodeCmd, "Usage: nx hd-decode [hd_private_key or hd_public_key] \n")
	}

	hdDeriveCmd := flag.NewFlagSet("hd-derive", flag.ExitOnError)
	hdDeriveCmd.Usage = func() {
		cmdUsage(hdDeriveCmd, "Usage: nx hd-derive [hd_private_key or hd_public_key] \n")
	}
	hdDeriveCmd.UintVar(&hdIndex, "i", 0, "The HD `index`")
	hdDeriveCmd.BoolVar(&hdHarden, "d", false, "create a hardened key")
	derivePath = derivePathFlag{wallet.DerivationPath{}}
	hdDeriveCmd.Var(&derivePath, "p", "hd derive `path`. ex: m/44'/0'/0'/0")
	hdDeriveCmd.Var(&hdVer, "v", "The HD(BIP32) `version` [mainnet|testnet|privnet|bip32]")

	// Mnemonic (BIP39)
	mnemonicNewCmd := flag.NewFlagSet("mnemonic-new", flag.ExitOnError)
	mnemonicNewCmd.Usage = func() {
		cmdUsage(mnemonicNewCmd, "Usage: nx mnemonic-new [entropy]  \n")
	}

	mnemonicToEntropyCmd := flag.NewFlagSet("mnemonic-to-entropy", flag.ExitOnError)
	mnemonicToEntropyCmd.Usage = func() {
		cmdUsage(mnemonicToEntropyCmd, "Usage: nx mnemonic-to-entropy [mnemonic]  \n")
	}

	mnemonicToSeedCmd := flag.NewFlagSet("mnemonic-to-seed", flag.ExitOnError)
	mnemonicToSeedCmd.Usage = func() {
		cmdUsage(mnemonicToSeedCmd, "Usage: nx mnemonic-to-seed [mnemonic]  \n")
	}
	mnemonicToSeedCmd.StringVar(&mnemoicSeedPassphrase, "p", "", "An optional passphrase for converting the mnemonic to a seed")

	// EC
	ecNewCmd := flag.NewFlagSet("ec-new", flag.ExitOnError)
	ecNewCmd.Usage = func() {
		cmdUsage(ecNewCmd, "Usage: nx ec-new [entropy]  \n")
	}
	ecNewCmd.StringVar(&curve, "c", "secp256k1", "the elliptic curve is using")

	ecToPubCmd := flag.NewFlagSet("ec-to-public", flag.ExitOnError)
	ecToPubCmd.Usage = func() {
		cmdUsage(ecToPubCmd, "Usage: nx ec-to-public [ec_private_key] \n")
	}
	ecToPubCmd.BoolVar(&uncompressedPKFormat, "u", false, "using the uncompressed public key format")

	// Wif
	ecToWifCmd := flag.NewFlagSet("ec-to-wif", flag.ExitOnError)
	ecToWifCmd.Usage = func() {
		cmdUsage(ecToWifCmd, "Usage: nx ec-to-wif [ec_private_key] \n")
	}
	ecToWifCmd.BoolVar(&uncompressedPKFormat, "u", false, "using the uncompressed public key format")

	wifToEcCmd := flag.NewFlagSet("wif-to-ec", flag.ExitOnError)
	wifToEcCmd.Usage = func() {
		cmdUsage(wifToEcCmd, "Usage: nx wif-to-ec [WIF] \n")
	}

	wifToPubCmd := flag.NewFlagSet("wif-to-public", flag.ExitOnError)
	wifToPubCmd.Usage = func() {
		cmdUsage(wifToPubCmd, "Usage: nx wif-to-public [WIF] \n")
	}
	wifToPubCmd.BoolVar(&uncompressedPKFormat, "u", false, "using the uncompressed public key format")

	// Address
	ecToAddrCmd := flag.NewFlagSet("ec-to-addr", flag.ExitOnError)
	ecToAddrCmd.Usage = func() {
		cmdUsage(ecToAddrCmd, "Usage: nx ec-to-addr [ec_public_key] \n")
	}
	ecToAddrCmd.Var(&base58checkVersion, "v", "base58check `version` [mainnet|testnet|privnet]")

	// Transaction
	txDecodeCmd := flag.NewFlagSet("tx-decode", flag.ExitOnError)
	txDecodeCmd.Usage = func() {
		cmdUsage(txDecodeCmd, "Usage: nx tx-decode [base16_string] \n")
	}
	txDecodeCmd.StringVar(&network, "n", "privnet", "decode rawtx for the target network. (mainnet, testnet, privnet)")

	txEncodeCmd := flag.NewFlagSet("tx-encode", flag.ExitOnError)
	txEncodeCmd.Usage = func() {
		cmdUsage(txEncodeCmd, "Usage: nx tx-encode [-i tx-input] [-l tx-lock-time] [-o tx-output] [-v tx-version] \n")
	}
	txVersion = txVersionFlag(TX_VERION) //set default tx version
	txEncodeCmd.Var(&txVersion, "v", "the transaction version")
	txEncodeCmd.Var(&txLockTime, "l", "the transaction lock time")
	txEncodeCmd.Var(&txInputs, "i", `The set of transaction input points encoded as TXHASH:INDEX:SEQUENCE. 
TXHASH is a Base16 transaction hash. INDEX is the 32 bit input index
in the context of the transaction. SEQUENCE is the optional 32 bit 
input sequence and defaults to the maximum value.`)
	txEncodeCmd.Var(&txOutputs, "o", `The set of transaction output data encoded as TARGET:NOX. 
TARGET is an address (pay-to-pubkey-hash or pay-to-script-hash).
NOX is the 64 bit spend amount in nox.`)

	txSignCmd := flag.NewFlagSet("tx-sign", flag.ExitOnError)
	txSignCmd.Usage = func() {
		cmdUsage(txSignCmd, "Usage: nx tx-sign [raw_tx_base16_string] \n")
	}
	txSignCmd.StringVar(&privateKey, "k", "", "the ec private key to sign the raw transaction")

	msgSignCmd := flag.NewFlagSet("msg-sign", flag.ExitOnError)
	msgSignCmd.Usage = func() {
		cmdUsage(msgSignCmd, "Usage: msg-sign [wif] [message] \n")
	}
	msgSignCmd.StringVar(&msgSignatureMode, "m", "nox", "the msg signature mode")
	msgSignCmd.BoolVar(&showDetails, "d", false, "show signature details")

	msgVerifyCmd := flag.NewFlagSet("msg-verify", flag.ExitOnError)
	msgVerifyCmd.Usage = func() {
		cmdUsage(msgVerifyCmd, "Usage: msg-verify [addr] [signature] [message] \n")
	}
	msgVerifyCmd.StringVar(&msgSignatureMode, "m", "nox", "the msg signature mode")

	flagSet := []*flag.FlagSet{
		base58CheckEncodeCommand,
		base58CheckDecodeCommand,
		base58EncodeCmd,
		base58DecodeCmd,
		base64EncodeCmd,
		base64DecodeCmd,
		rlpEncodeCmd,
		rlpDecodeCmd,
		sha256cmd,
		blake2b256cmd,
		blake2b512cmd,
		blake256cmd,
		sha3_256cmd,
		keccak256cmd,
		ripemd160Cmd,
		bitcion160Cmd,
		hash160Cmd,
		entropyCmd,
		hdNewCmd,
		hdToPubCmd,
		hdToEcCmd,
		hdDecodeCmd,
		hdDeriveCmd,
		mnemonicNewCmd,
		mnemonicToEntropyCmd,
		mnemonicToSeedCmd,
		ecNewCmd,
		ecToPubCmd,
		ecToWifCmd,
		wifToEcCmd,
		wifToPubCmd,
		ecToAddrCmd,
		txEncodeCmd,
		txDecodeCmd,
		txSignCmd,
		msgSignCmd,
		msgVerifyCmd,
	}

	if len(os.Args) == 1 {
		usage()
	}
	switch os.Args[1] {
	case "help", "--help":
		usage()
	case "version", "--version":
		version()
	default:
		valid := false
		for _, cmd := range flagSet {
			if os.Args[1] == cmd.Name() {
				cmd.Parse(os.Args[2:])
				valid = true
				break
			}
		}
		if !valid {
			invalid := os.Args[1]
			if invalid[0] == '-' {
				fmt.Fprintf(os.Stderr, "unknown option: %q \n", invalid)
			} else {
				fmt.Fprintf(os.Stderr, "%q is not valid command\n", invalid)
			}
			os.Exit(1)
		}
	}
	// Handle base58check-encode
	if base58CheckEncodeCommand.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				base58CheckEncodeCommand.Usage()
			} else {
				base58CheckEncode(base58checkVersion.ver, base58checkVersion.mode, base58checkHasher, base58checkCksumSize, os.Args[len(os.Args)-1])
			}
		} else { //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			base58CheckEncode(base58checkVersion.ver, base58checkVersion.mode, base58checkHasher, base58checkCksumSize, str)
		}
	}

	// Handle base58check-decode
	if base58CheckDecodeCommand.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				base58CheckDecodeCommand.Usage()
			} else {
				base58CheckDecode(base58checkMode, base58checkHasher, base58checkVersionSize, base58checkCksumSize, os.Args[len(os.Args)-1])
			}
		} else { //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			base58CheckDecode(base58checkMode, base58checkHasher, base58checkVersionSize, base58checkCksumSize, str)
		}
	}

	// Handle base58-encode
	if base58EncodeCmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				base58EncodeCmd.Usage()
			} else {
				base58Encode(os.Args[len(os.Args)-1])
			}
		} else { //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			base58Encode(str)
		}
	}
	// Handle base58-decode
	if base58DecodeCmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				base58DecodeCmd.Usage()
			} else {
				base58Decode(os.Args[len(os.Args)-1])
			}
		} else { //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			base58Decode(str)
		}
	}

	if base64EncodeCmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				base64EncodeCmd.Usage()
			} else {
				base64Encode(os.Args[len(os.Args)-1])
			}
		} else { //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			base64Encode(str)
		}
	}

	if base64DecodeCmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				base64DecodeCmd.Usage()
			} else {
				base64Decode(os.Args[len(os.Args)-1])
			}
		} else { //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			base64Decode(str)
		}
	}

	if rlpEncodeCmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				rlpEncodeCmd.Usage()
			} else {
				rlpEncode(os.Args[len(os.Args)-1])
			}
		} else { //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			rlpEncode(str)
		}
	}
	if rlpDecodeCmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				rlpDecodeCmd.Usage()
			} else {
				rlpDecode(os.Args[len(os.Args)-1])
			}
		} else { //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			rlpDecode(str)
		}
	}

	if sha256cmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				sha256cmd.Usage()
			} else {
				sha256(os.Args[len(os.Args)-1])
			}
		} else { //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			sha256(str)
		}
	}

	if blake256cmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				blake256cmd.Usage()
			} else {
				blake256(os.Args[len(os.Args)-1])
			}
		} else { //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			blake256(str)
		}
	}

	if blake2b256cmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				blake2b256cmd.Usage()
			} else {
				blake2b256(os.Args[len(os.Args)-1])
			}
		} else { //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			blake2b256(str)
		}
	}

	if blake2b512cmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				blake2b512cmd.Usage()
			} else {
				blake2b512(os.Args[len(os.Args)-1])
			}
		} else { //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			blake2b512(str)
		}
	}

	if sha3_256cmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				sha3_256cmd.Usage()
			} else {
				sha3_256(os.Args[len(os.Args)-1])
			}
		} else { //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			sha3_256(str)
		}
	}

	if keccak256cmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				keccak256cmd.Usage()
			} else {
				keccak256(os.Args[len(os.Args)-1])
			}
		} else { //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			keccak256(str)
		}
	}

	if ripemd160Cmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				ripemd160Cmd.Usage()
			} else {
				ripemd160(os.Args[len(os.Args)-1])
			}
		} else { //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			ripemd160(str)
		}
	}

	if bitcion160Cmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				bitcion160Cmd.Usage()
			} else {
				bitcoin160(os.Args[len(os.Args)-1])
			}
		} else { //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			bitcoin160(str)
		}
	}

	if hash160Cmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				hash160Cmd.Usage()
			} else {
				hash160(os.Args[len(os.Args)-1])
			}
		} else { //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			hash160(str)
		}
	}

	if entropyCmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) > 2 && (os.Args[2] == "help" || os.Args[2] == "--help") {
				entropyCmd.Usage()
			} else {
				if seedSize%8 > 0 {
					errExit(fmt.Errorf("seed (entropy) length must be Must be divisible by 8"))
				}
				newEntropy(seedSize / 8)
			}
		} else {
			entropyCmd.Usage()
		}
	}

	if hdNewCmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				hdNewCmd.Usage()
			} else {
				hdNewMasterPrivateKey(hdVer.version, os.Args[len(os.Args)-1])
			}
		} else { //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			hdNewMasterPrivateKey(hdVer.version, str)
		}
	}

	if hdToPubCmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				hdToPubCmd.Usage()
			} else {
				hdPrivateKeyToHdPublicKey(hdVer.version, os.Args[len(os.Args)-1])
			}
		} else { //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			hdPrivateKeyToHdPublicKey(hdVer.version, str)
		}
	}

	if hdToEcCmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				hdToEcCmd.Usage()
			} else {
				hdKeyToEcKey(hdVer.version, os.Args[len(os.Args)-1])
			}
		} else { //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			hdKeyToEcKey(hdVer.version, str)
		}
	}

	if hdDecodeCmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				hdDecodeCmd.Usage()
			} else {
				hdDecode(os.Args[len(os.Args)-1])
			}
		} else { //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			hdDecode(str)
		}
	}

	if hdDeriveCmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				hdDeriveCmd.Usage()
			} else {
				hdDerive(hdHarden, uint32(hdIndex), derivePath.path, hdVer.version, os.Args[len(os.Args)-1])
			}
		} else { //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			hdDerive(hdHarden, uint32(hdIndex), derivePath.path, hdVer.version, str)
		}
	}

	if mnemonicNewCmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				mnemonicNewCmd.Usage()
			} else {
				mnemonicNew(os.Args[len(os.Args)-1])
			}
		} else { //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			mnemonicNew(str)
		}
	}

	if mnemonicToEntropyCmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				mnemonicToEntropyCmd.Usage()
			} else {
				mnemonicToEntropy(os.Args[len(os.Args)-1])
			}
		} else { //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			mnemonicToEntropy(str)
		}
	}

	if mnemonicToSeedCmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				mnemonicToSeedCmd.Usage()
			} else {
				mnemonicToSeed(mnemoicSeedPassphrase, os.Args[len(os.Args)-1])
			}
		} else { //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			mnemonicToSeed(mnemoicSeedPassphrase, str)
		}
	}

	if ecNewCmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				ecNewCmd.Usage()
			} else {
				ecNew(curve, os.Args[len(os.Args)-1])
			}
		} else { //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			ecNew(curve, str)
		}
	}

	if ecToPubCmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				ecToPubCmd.Usage()
			} else {
				ecPrivateKeyToEcPublicKey(uncompressedPKFormat, os.Args[len(os.Args)-1])
			}
		} else { //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			ecPrivateKeyToEcPublicKey(uncompressedPKFormat, str)
		}
	}

	if ecToWifCmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				ecToWifCmd.Usage()
			} else {
				ecPrivateKeyToWif(uncompressedPKFormat, os.Args[len(os.Args)-1])
			}
		} else { //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			ecPrivateKeyToWif(uncompressedPKFormat, str)
		}
	}

	if wifToEcCmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				wifToEcCmd.Usage()
			} else {
				wifToEcPrivateKey(os.Args[len(os.Args)-1])
			}
		} else { //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			wifToEcPrivateKey(str)
		}
	}

	if wifToPubCmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				wifToPubCmd.Usage()
			} else {
				wifToEcPubkey(uncompressedPKFormat, os.Args[len(os.Args)-1])
			}
		} else { //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			wifToEcPubkey(uncompressedPKFormat, str)
		}
	}

	if ecToAddrCmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				ecToAddrCmd.Usage()
			} else {
				ecPubKeyToAddress(base58checkVersion.ver, os.Args[len(os.Args)-1])
			}
		} else { //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			ecPubKeyToAddress(base58checkVersion.ver, str)
		}
	}

	if txDecodeCmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				txDecodeCmd.Usage()
			} else {
				txDecode(network, os.Args[len(os.Args)-1])
			}
		} else { //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			txDecode(network, str)
		}
	}

	if txEncodeCmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				txEncodeCmd.Usage()
			} else {
				txEncode(txVersion, txLockTime, txInputs, txOutputs)
			}
		}
	}

	if txSignCmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				txSignCmd.Usage()
			} else {
				txSign(privateKey, os.Args[len(os.Args)-1])
			}
		} else { //try from STDIN
			src, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				errExit(err)
			}
			str := strings.TrimSpace(string(src))
			txSign(privateKey, str)
		}
	}

	if msgSignCmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				msgSignCmd.Usage()
			} else {
				msgSign(msgSignatureMode, showDetails, os.Args[len(os.Args)-2], os.Args[len(os.Args)-1])
			}
		}
	}

	if msgVerifyCmd.Parsed() {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			if len(os.Args) == 2 || os.Args[2] == "help" || os.Args[2] == "--help" {
				msgVerifyCmd.Usage()
			} else {
				verifyMsgSignature(msgSignatureMode, os.Args[len(os.Args)-3], os.Args[len(os.Args)-2], os.Args[len(os.Args)-1])
			}
		}
	}
}
