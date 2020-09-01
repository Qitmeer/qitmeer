# Qitmeer

[![Build Status](https://travis-ci.com/Qitmeer/qitmeer.svg?token=DzCFNC6nhEqPc89sq1nd&branch=master)](https://travis-ci.com/Qitmeer/qitmeer) [![Go Report Card](https://goreportcard.com/badge/github.com/Qitmeer/qitmeer)](https://goreportcard.com/report/github.com/Qitmeer/qitmeer)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FQitmeer%2Fqitmeer.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2FQitmeer%2Fqitmeer?ref=badge_shield)

The guardian of trust. The core backend of the qitmeer network.

## Qitmeer Testnet Notice

 The Qitmeer Official Public Testnet is **OPEN** Now. The Code name is the *Medina Network*. Please Join the Medina Network !

| Latest Testnet            | Compatible Qitmeer Vesion | Start Date | Type            |
| ------------------------- |-------------------------- | ---------- | --------------- |
|[Medina Network 2.0](TESTNET.md#v090-20200624-medina20)| v0.9.0    | 2020/06/24 | Official Public Testnet |


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
####  Prerequisites

- Update Go to version at least 1.12 (required >= **1.12**)

Check your golang version

```bash
$ go version
go version go1.13.4 darwin/amd64
```
```bash
~ mkdir -p /tmp/work
~ cd /tmp/work
~ git clone https://github.com/Qitmeer/qitmeer.git
~ cd qitmeer
~ make
~ ./build/bin/qitmeer --version
qitmeer version 0.8.2+dev-f45bcf8 (Go version go1.13.4)
```

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

### P.S.
* You must use ctrl+c ,kill(the default is 15) or kill -2 to close the qitmeer, otherwise, it may destroy the integrity of program data.

## qitmeer-cli

[qitmeer rpc tools](https://github.com/Qitmeer/qitmeer-cli)

## ZeroMQ
[Block and Transaction Broadcasting with ZeroMQ](services/zmq/README.md) for details on how this works.

**happy hacking!**


## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FQitmeer%2Fqitmeer.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2FQitmeer%2Fqitmeer?ref=badge_large)
