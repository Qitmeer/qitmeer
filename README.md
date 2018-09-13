# nox


###  Prerequisites

Update Go to latest version 1.11

```
$ go version
go version go1.11 darwin/amd64
```

Install vgo 

```
$ go get -u golang.org/x/vgo
$ vgo version
go version go1.11 darwin/amd64 vgo:devel +b0a1c5df98
```

## How to build

```
$ mkdir -p /tmp/work
$ cd /tmp/work
$ git clone https://github.com/noxproject/nox 
$ cd /tmp/work/nox/nox
$ vgo build
$ ./nox --version
nox version 0.1.0+dev (Go version go1.11)
```

happy hacking!

