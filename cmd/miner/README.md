# Qitmeer Miner

[![Build Status](https://travis-ci.com/Qitmeer/qitmeer/cmd/miner.svg?token=n9AoZUDqAJmhesf4MYUd&branch=master)](https://travis-ci.com/Qitmeer/qitmeer/cmd/miner)

> The official CPU and Asic miner of the Qitmeer network  

**Qitmeer-miner** is an CPU and Asic miner for the Qitmeer netowrk. It's the official reference implement maintained by the Qitmeer team.
Currently it support 3 Qitmeer POW algorithms including Cuckaroo, Cuckatoo and Blake2bd,MeerXKeccak.

## Table of Contents
* [Install](#install)
* [Usage](#usage)
   - [Run with config file](#run-with-config-file)
   - [Run by Command line options](#command-line-usage)
* [Build](#build)
   - [Building from source](#building-from-source)
* [Tutorial](#tutorial)    
* [FAQ](#faq)


## Install

[![Releases](https://img.shields.io/github/downloads/Qitmeer/qitmeer/cmd/miner/total.svg)][Releases]

Standalone installation archive for *Linux*, *macOS* and *Windows* are provided in
the [Releases] section. 
Please download an archive for your operating system and unpack the content to a place
accessible from command line. 

| Builds | Release | Date |
| ------ | ------- | ---- |
| Last   | [![GitHub release](https://img.shields.io/github/release/Qitmeer/qitmeer/cmd/miner/all.svg)][Releases] | [![GitHub Release Date](https://img.shields.io/github/release-date-pre/Qitmeer/qitmeer/cmd/miner.svg)][Releases] |
| Stable | [![GitHub release](https://img.shields.io/github/release/Qitmeer/qitmeer/cmd/miner.svg)][latest] | [![GitHub Release Date](https://img.shields.io/github/release-date/Qitmeer/qitmeer/cmd/miner.svg)][latest] |

## Usage

### Run with config file 
1. go to your 
2. create a new config file by copying from the example config file. 
```bash
$ cp example.solo.conf solo.conf
```
3. edit the config file which your create, you might need to change the `mineraddress`. 
you need to create a Qitmeer address if you don't have it. Please see [FAQ](#FAQ)  
4. run miner with the config file

```bash
$ ./qitmeer-miner -C solo.conf
```

### Command line usage

The qitmeer-miner is a command line program. This means you can also launch it by provided valid command line options. For a full list of available command optinos, please run:

```bash
$ ./qitmeer-miner --help 
Debug Command:
  -l, --listdevices    List number of devices.

The Config File Options:
  -C, --configfile=    Path to configuration file
      --minerlog=      Write miner log file

The Necessary Config Options:
  -P, --pow=           blake2bd|cuckaroo|cuckatoo (blake2bd)
  -S, --symbol=        Symbol (PMEER)
  -N, --network=       network privnet|testnet|mainnet (mainnet)

The Solo Config Option:
  -M, --mineraddress=  Miner Address
  -s, --rpcserver=     RPC server to connect to (127.0.0.1)
  -u, --rpcuser=       RPC username
  -p, --rpcpass=       RPC password
      --randstr=       Rand String,Your Unique Marking. (Come from Qitmeer!)
      --notls          Do not verify tls certificates (true)
      --rpccert=       RPC server certificate chain for validation

The pool Config Option:
  -o, --pool=          Pool to connect to (e.g.stratum+tcp://pool:port)
  -m, --pooluser=      Pool username
  -n, --poolpass=      Pool password

The Optional Config Option:
      --cpuminer       CPUMiner (false)
      --proxy=         Connect via SOCKS5 proxy (eg. 127.0.0.1:9050)
      --proxyuser=     Username for proxy server
      --proxypass=     Password for proxy server
      --trimmerTimes=  the cuckaroo trimmer times (40)
      --intensity=     Intensities (the work size is 2^intensity) per device. Single global value or a comma separated list. (24)
      --worksize=      The explicitly declared sizes of the work to do per device (overrides intensity). Single global value or a comma separated list. (256)
      --timeout=       rpc timeout. (60)
      --use_devices=   all gpu devices,you can use ./qitmeer-miner -l to see. examples:0,1 use the #0 device and #1 device
      --max_tx_count=  max pack tx count (1000)
      --max_sig_count= max sign tx count (5000)
      --log_level=     info|debug|error|warn|trace (debug)
      --stats_server=  stats web server (127.0.0.1:1235)
      --edge_bits=     edge bits (24)
      --local_size=    local size (4096)
      --group_size=    work group size (256)

Help Options:
  -h, --help           Show this help message
 
```
Please see [Qitmeer-Miner User References](https://qitmeer.github.io/docs/en/reference/qitmeer-miner/) for more details

## Build
### Building from source
See [BUILD.md](BUILD.md) for build/compilation details.

## Tutorial

### Community Tutorials

* Chinese [中文教程：windows系统qitmeer-miner编译及环境准备指导](https://github.com/Qitmeer/qitmeer/cmd/miner/issues/88)
* Chinese [Qitmeer挖矿终极指南 https://www.qitmeertalk.org/t/qitmeer-2019-11-02/906](https://www.qitmeertalk.org/t/qitmeer-2019-11-02/906)

## FAQ

### How to create Qitmeer adderss
There are several ways to create a Qitmeer address. you can use [qx][Qx] command , [qitmeer-wallet][Qitmeer-wallet], etc.
The most easy way to download the [kafh wallet][kafh.io], which provide a more user friendly GUI to create your address/wallet step by step. 

### Which POW algorithm I should choose to mine ?
Qitmeer test network support mixing minning, which means your can choice from `Cuckaroo`, `Cuckatoo` and `Blake2bd` anyone you like. 
But the start difficulty targets are quite different. For the most case you might use `Cuckaroo` as a safe choice at the beginning. 

### Where I can find more documentation ? 
Please find more documentation from the [Qitmeer doc site at https://qitmeer.github.io](https://qitmeer.github.io/docs/en/reference/qitmeer-miner/)

[Releases]: https://github.com/Qitmeer/qitmeer/cmd/miner/releases
[Latest]: https://github.com/Qitmeer/qitmeer/cmd/miner/releases/latest
[Qx]: https://qitmeer.github.io/docs/en/reference/qxtools/
[Qitmeer-wallet]: https://github.com/Qitmeer/qitmeer/cmd/wallet
[Kafh.io]:https://www.kahf.io/

