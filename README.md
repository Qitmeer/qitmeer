# Qitmeer

[![Build Status](https://travis-ci.com/Qitmeer/qitmeer.svg?token=DzCFNC6nhEqPc89sq1nd&branch=master)](https://travis-ci.com/Qitmeer/qitmeer) [![Go Report Card](https://goreportcard.com/badge/github.com/Qitmeer/qitmeer)](https://goreportcard.com/report/github.com/Qitmeer/qitmeer)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FQitmeer%2Fqitmeer.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2FQitmeer%2Fqitmeer?ref=badge_shield)

The guardian of trust. The core backend of the qitmeer network.

## Qitmeer Testnet Notice

 The Qitmeer Official Public Testnet is **OPEN** Now. The Code name is the *Medina Network*. Please Join the Medina Network !

| Latest Testnet            | Compatible Qitmeer Vesion | Start Date | Type            |
| ------------------------- |-------------------------- | ---------- | --------------- |
|[Medina Network 2.0](TESTNET.md#v090-20200624-medina20)| v0.9.x    | 2020/06/24 | Official Public Testnet |


Please know more details from the [Qitmeer Testnet](TESTNET.md)

## Installation
### Binary archives
[![Releases](https://img.shields.io/github/downloads/Qitmeer/qitmeer/total.svg)][Releases]

Standalone installation archive for *Linux*, *macOS* and *Windows* are provided in
the [Releases] section.
Please download an archive for your operating system and unpack the content to a place
accessible from command line.

| Builds | Release | Date |
| ------ | ------- | ---- |
| Last   | [![GitHub release](https://img.shields.io/github/release/Qitmeer/qitmeer/all.svg)][Releases] | [![GitHub Release Date](https://img.shields.io/github/release-date-pre/Qitmeer/qitmeer.svg)][Releases] |
| Stable | [![GitHub release](https://img.shields.io/github/release/Qitmeer/qitmeer.svg)][Latest] | [![GitHub Release Date](https://img.shields.io/github/release-date/Qitmeer/qitmeer.svg)][Latest] |

[Releases]: https://github.com/Qitmeer/qitmeer/releases
[Latest]: https://github.com/Qitmeer/qitmeer/releases/latest

### How to build

`Golang` at least 1.12 (required >= **1.12**)

Check your golang version.

```bash
~ go version
go version go1.16 darwin/amd64
```

#### Build

```bash
~ mkdir -p /tmp/work
~ cd /tmp/work
~ git clone https://github.com/Qitmeer/qitmeer.git
~ cd qitmeer
~ make
~ make
Done building.
  qitmeer version 0.9.2+dev-1f5defd (Go version go1.16))
Run "./build/bin/qitmeer" to launch.
```

### How to Run

Memory Requirements
  * Recommended: 2GB

Make sure to have at least 2GB free memory to run qitmeer. Insufficient memory
may lead to the process being killed unexpectedly when its running.
See [FAQ #3](#Qitmeer-is-killed-unexpectedly) for details.

#### Getting Started
```
./build/bin/qitmeer --testnet
```
Please make sure use `--testnet` to connect to the correct network.
currently, only `testnet` network is officially supported.

Several configuration options available to tweak how it runs. Please see details by
using the `help` command
```
./build/bin/qitmeer --help
```

#### Running with Docker

You can also run `qitmeer` by using docker

```
docker run qitmeer/qitmeerd
```

## Other useful qitmeer repository

### [qitmeer-wallet](https://github.com/Qitmeer/qitmeer-wallet)

The command-line wallet of the Qitmeer network

### [qitmeer-miner](https://github.com/Qitmeer/qitmeer-miner)

The GPU miner for the Qitmeer netowrk.

### [qitmeer-cli](https://github.com/Qitmeer/qitmeer-cli)

The command line utility of Qitmeer

## How to Work with ZeroMQ
[Block and Transaction Broadcasting with ZeroMQ](services/zmq/README.md) for details on how this works.

## FAQ

### How to exit qitmeer properly.

You can use `Ctrl+C` to exit in the foreground, or `kill` or `kill -2` if in the backgroud.
Please don't use `kill -9` to kill the `qitmeer` process , this terminates the process abruptly
and may leave database files improperly closed. may result corrupt data files.
In the worst case, you might need to do a refresh block synchronization.

### How to clean up corrupt data
```
qitmeer --cleanup --testnet
```
***Please be careful! the command results your data to be removed!***


### How to fix `golang.org unrecognized` Issue

If you got trouble to download the `golang.org` depends automatically

```
go: golang.org/x/crypto@v0.0.0-20181001203147-e3636079e1a4: unrecognized import path "golang.org/x/crypto" (https fetch: Get https://golang.org/x/crypto?go-get=1: dial tcp 216.239.37.1:443: i/o timeout)
go: golang.org/x/net@v0.0.0-20181005035420-146acd28ed58: unrecognized import path "golang.org/x/net" (https fetch: Get https://golang.org/x/net?go-get=1: dial tcp 216.239.37.1:443: i/o timeout)
go: golang.org/x/net@v0.0.0-20180906233101-161cd47e91fd: unrecognized import path "golang.org/x/net" (https fetch: Get https://golang.org/x/net?go-get=1: dial tcp 216.239.37.1:443: i/o timeout)
```

you might need to `replace` the download url (ex: using a mirror site like github.com) on your `go.mod`

```
replace (
	golang.org/x/crypto v0.0.0-20181001203147-e3636079e1a4 => github.com/golang/crypto v0.0.0-20181001203147-e3636079e1a4
	golang.org/x/net v0.0.0-20180906233101-161cd47e91fd => github.com/golang/net v0.0.0-20180906233101-161cd47e91fd
	golang.org/x/net v0.0.0-20181005035420-146acd28ed58 => github.com/golang/net v0.0.0-20181005035420-146acd28ed58
)
```

### Qitmeer is killed unexpectedly

`qitmeer` is killed unexpectedly probably due to the `Out of Memory`, If you're using `Ubuntu` linux,
you can verify it by following command.
```sh
dmesg -T| grep -E -i -B100 'killed process'
```
If you find similar output as follows, then that maybe the case
```sh
[Tue Mar  9 11:34:26 2021] Out of memory: Killed process 140587 (qitmeer) total-vm:1403144kB,
anon-rss:675828kB, file-rss:0kB, shmem-rss:0kB, UID:1001 pgtables:1532kB oom_score_adj:0
```
The minimum memory requirement is 1GB, and we strongly recommend upgrading the memory to 2GB.

1. Update golang

If the memory resource restrictions is do your case, You might try to upgrade your `Golang` to the latest version
and re-compile `Qitmeer` and try yourself. We have received feedbacks from community that the
newly golang compiler have better memory optimizations, might work better wth low memory
requirement, and especially for `Ubuntu 20.04`.

If you're using `Ubuntu 18.04/20.04`, then you can use the `longsleep/golang-backports` PPA
and update to latest `golang`. then re-compile your qitmeer and try if works for you.
```sh
sudo add-apt-repository ppa:longsleep/golang-backports
sudo apt update
sudo apt install golang-go
```
Please note, it does not guarantee that compiling with the latest `golang` might work.
Adding more computer memory is always the recommended way.

2. Mount Swap Memory

Swap Memory may exploit your external memory, typically hard disk, to simulate physical memory.
If your swap memory is not mounted, or the allocation is insufficient, 
you may follow this [tutorial](https://qitmeer.github.io/docs/en/tutorials/swap-memory/). 

### Compliing failed by missing the `go.sum` entries

If your `golang` version is **1.16 or above**, you might see similar error as follows when compiling.

```shell
go: github.com/Qitmeer/crypto@v0.0.0-20200516043559-dd457edff06c: missing go.sum entry; to add it:
        go mod download github.com/Qitmeer/crypto
make: *** [Makefile:41: qitmeer-build] Error 1
```
It's due to `go1.16` changes default mod rules by disabling auto fixing missing entry in `go.sum` file.
see [details here](https://blog.golang.org/go116-module-changes).

It's an known issue and fixed by the latest code, Please update `Qitmeer` to the latest version
and do the compiling again.

If you somehow need to stick to the current version, please make sure to execute following command before you compile qitmeer:
```shell
go mod tidy
```
### Qitmeer gets stuck on synchronization.
> Note: reproduced in 0.9.x only so far, not found in 0.10.x 

#### Problem
If qitmeer gets stuck on synchronization, and you find qitmeer keeping iterating logs similar as follows:
```shell
 [INFO ] Syncing to state (824024,761687,761687,824025,1) from peer 47.116.50.38:18130 cur graph state:(235978,228535,228535,235979,1) module=blkmanager
```

#### Reason
Due to the node's computation performance is insufficient to proceed synchronization on time, peers will be disconnected gradually
due to timeout until no peers for you to synchronize from, specifically, in the case that your memory is lacking, 
or you connectivity is poor.

Please see [issue 444](https://github.com/Qitmeer/qitmeer/issues/444) to get more details

#### Solution
1. Ensure memory is not less than 2G  
Currently, it is only reproduced on machines with 1G or less, please make sure that your have 2G memory as recommended.
2. Publish your ip if you have a public ip  
Please add such entry in your launching parameters to improve connectivity 

```sh
externalip=YOUR_PUBLIC_IP:18130
```


## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FQitmeer%2Fqitmeer.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2FQitmeer%2Fqitmeer?ref=badge_large)
