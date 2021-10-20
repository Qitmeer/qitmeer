# Building from source

## Table of Contents

* [Prequisite](#prequisite)
    * [Common](#common)
* [Build](#build)
    
### Prequisite

### Common

1. [Git](https://git-scm.com/downloads) 
2. [Go](https://golang.org/dl/) version >= 1.12
3. [Rust/Cargo](https://www.rust-lang.org/tools/install) >= 1.38.0

## Build 

### 1. Get Source code

```bash
$ git clone git@github.com:Qitmeer/qitmeer.git
```

[Build Step]

### 2. Build qitmeer-miner  

```bash
//# cpu
$ make
//# asic
$ apt-get install gcc-arm*
$ CGO_ENABLED=1 GOOS=linux GOARCH=arm CC=arm-linux-gnueabihf-gcc go build -a --tags asic
```

### 3. Verify Build OK

```bash
$ ./build/bin/qitmeer-miner --version
```