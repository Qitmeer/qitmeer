# nox


###  Prerequisites

Update Go to latest version 1.11

```
$ go version
go version go1.11 darwin/amd64
```

## How to build

```
$ mkdir -p /tmp/work
$ cd /tmp/work
$ git clone https://github.com/noxproject/nox 
$ git checkout cleanup 
$ go build
$ ./nox --version
nox version 0.1.0+dev (Go version go1.11)
```

### Go Mod
```
$ go mod graph
github.com/noxproject/nox github.com/AndreasBriese/bbloom@v0.0.0-20170702084017-28f7e881ca57
github.com/noxproject/nox github.com/boltdb/bolt@v1.3.1
github.com/noxproject/nox github.com/btcsuite/btcd@v0.0.0-20180903232927-cff30e1d23fc
github.com/noxproject/nox github.com/coreos/bbolt@v1.3.0
github.com/noxproject/nox github.com/davecgh/go-spew@v1.1.1
github.com/noxproject/nox github.com/dchest/blake256@v1.0.0
github.com/noxproject/nox github.com/deckarep/golang-set@v0.0.0-20171013212420-1d4478f51bed
github.com/noxproject/nox github.com/dgraph-io/badger@v1.5.3
github.com/noxproject/nox github.com/dgryski/go-farm@v0.0.0-20180109070241-2de33835d102
github.com/noxproject/nox github.com/ethereum/go-ethereum@v1.8.15
github.com/noxproject/nox github.com/go-stack/stack@v1.8.0
github.com/noxproject/nox github.com/golang/snappy@v0.0.0-20180518054509-2e65f85255db
github.com/noxproject/nox github.com/jessevdk/go-flags@v1.4.0
github.com/noxproject/nox github.com/jrick/logrotate@v1.0.0
github.com/noxproject/nox github.com/kr/pretty@v0.1.0
github.com/noxproject/nox github.com/mattn/go-colorable@v0.0.9
github.com/noxproject/nox github.com/mattn/go-isatty@v0.0.4
github.com/noxproject/nox github.com/onsi/gomega@v1.4.2
github.com/noxproject/nox github.com/pkg/errors@v0.8.0
github.com/noxproject/nox github.com/pmezard/go-difflib@v1.0.0
github.com/noxproject/nox github.com/rcrowley/go-metrics@v0.0.0-20180503174638-e2704e165165
github.com/noxproject/nox github.com/stretchr/testify@v1.2.2
github.com/noxproject/nox github.com/syndtr/goleveldb@v0.0.0-20180815032940-ae2bd5eed72d
github.com/noxproject/nox golang.org/x/crypto@v0.0.0-20180830192347-182538f80094
github.com/noxproject/nox golang.org/x/net@v0.0.0-20180906233101-161cd47e91fd
github.com/noxproject/nox golang.org/x/sys@v0.0.0-20180909124046-d0be0721c37e
github.com/noxproject/nox golang.org/x/tools@v0.0.0-20180831211245-7ca132754999
github.com/noxproject/nox gopkg.in/check.v1@v1.0.0-20180628173108-788fd7840127
github.com/onsi/gomega@v1.4.2 github.com/fsnotify/fsnotify@v1.4.7
github.com/onsi/gomega@v1.4.2 github.com/golang/protobuf@v1.2.0
github.com/onsi/gomega@v1.4.2 github.com/hpcloud/tail@v1.0.0
github.com/onsi/gomega@v1.4.2 github.com/onsi/ginkgo@v1.6.0
github.com/onsi/gomega@v1.4.2 golang.org/x/net@v0.0.0-20180906233101-161cd47e91fd
github.com/onsi/gomega@v1.4.2 golang.org/x/sync@v0.0.0-20180314180146-1d60e4601c6f
github.com/onsi/gomega@v1.4.2 golang.org/x/sys@v0.0.0-20180909124046-d0be0721c37e
github.com/onsi/gomega@v1.4.2 golang.org/x/text@v0.3.0
github.com/onsi/gomega@v1.4.2 gopkg.in/fsnotify.v1@v1.4.7
github.com/onsi/gomega@v1.4.2 gopkg.in/tomb.v1@v1.0.0-20141024135613-dd632973f1e7
github.com/onsi/gomega@v1.4.2 gopkg.in/yaml.v2@v2.2.1
github.com/kr/pretty@v0.1.0 github.com/kr/text@v0.1.0
gopkg.in/yaml.v2@v2.2.1 gopkg.in/check.v1@v0.0.0-20161208181325-20d25e280405
github.com/kr/text@v0.1.0 github.com/kr/pty@v1.1.1
$ go mod verify
all modules verified
```

happy hacking!

