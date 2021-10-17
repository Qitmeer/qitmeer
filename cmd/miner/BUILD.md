# Building from source

## Table of Contents

* [Prequisite](#prequisite)
    * [Common](#common)
    * [Linux](#linux)
        * [Ubuntu](#ubuntu)
        * [Centos](#centos)
    * [macOS](#macos)
    * [Windows](#windows)
* [Build](#build)
    * [Windows-additional step](#windows-additional-step)
    * [Linux-additional step](#Linux-additional step)
    
### Prequisite

### Common

1. [Git](https://git-scm.com/downloads) 
2. [Go](https://golang.org/dl/) version >= 1.12
3. [Rust/Cargo](https://www.rust-lang.org/tools/install) >= 1.38.0

### Linux


#### Ubuntu

```bash
$ sudo apt-get install beignet-dev nvidia-cuda-dev nvidia-cuda-toolkit
```
        
#### Centos 

```bash
$ sudo yum install opencl-headers
$ sudo yum install ocl-icd
$ sudo ln -s /usr/lib64/libOpenCL.so.1 /usr/lib/libOpenCL.so
```  
### MacOS

### Windows

Install [**Build Tools for Visual Studio**](https://visualstudio.microsoft.com/thank-you-downloading-visual-studio/?sku=BuildTools&rel=16)
    
## Build 

### 1. Get Source code

```bash
$ git clone git@github.com:Qitmeer/qitmeer/cmd/miner.git
```

### 2. Build the cuckatoo library 

```bash
$ cd qitmeer-miner 
$ sh installLibrary.sh
```

### 3. Build the cudacuckaroom library 

[Build Step](lib/cuda/cuckaroom/README.md)

### 3. Build qitmeer-miner  

```bash
//# mac
$ go build --tags opencl
//# linux apt install musl-tools g++ -y
$ apt-get install gcc-arm*
$ CGO_ENABLED=1 GOOS=linux GOARCH=arm CC=arm-linux-gnueabihf-gcc go build -a --tags asic
//# windows 
$ CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -tags cuda -o win-miner.exe main.go
```

### 4. Verify Build OK

```bash
$ ./qitmeer-miner --version
```

### Windows-additional step

Before step 3, do following 
```bash
$ copy lib/cuckoo/target/release/x86_64-pc-windows-gnu/cuckoo.lib to C:/mingw64/lib
$ copy lib/opencl/windows/libOpenCL.a to C:/mingw64/lib
```
### Linux-additional step

Before step 3, do following 
```bash
$ sudo copy lib/cuckoo/target/x86_64-unknown-linux-musl/release/libcuckoo.a /usr/lib/x86_64-linux-musl
$ sudo copy lib/opencl/linux/libOpenCL.a /usr/lib/x86_64-linux-musl
```
