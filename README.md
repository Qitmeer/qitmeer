# Qitmeer

[![Build Status](https://travis-ci.com/HalalChain/qitmeer.svg?token=DzCFNC6nhEqPc89sq1nd&branch=master)](https://travis-ci.com/HalalChain/qitmeer)


The guardian of trust. The core network of the Halachain

##  Prerequisites

- Update Go to version at least 1.12 (required >= **1.12**)

Check your golang version

```bash
~ go version
go version go1.12 darwin/amd64
```

## How to build

```bash
~ mkdir -p /tmp/work
~ cd /tmp/work
~ git clone https://github.com/HalalChain/qitmeer.git
~ cd qitmeer
~ go build
~ ./nox --version
nox version 0.3.0+dev (Go version go1.12)
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

[qitmeer rpc tools](https://github.com/HalalChain/qitmeer-cli)

**happy hacking!**
