module github.com/HalalChain/qitmeer

go 1.12

require (
	github.com/AndreasBriese/bbloom v0.0.0-20190306092124-e2d15f34fcf9 // indirect
	github.com/HalalChain/qitmeer-lib v0.0.0-20190625033329-0a024ba36652
	github.com/coreos/bbolt v1.3.2
	github.com/davecgh/go-spew v1.1.1
	github.com/dchest/blake256 v1.0.0
	github.com/deckarep/golang-set v1.7.1
	github.com/dgraph-io/badger v1.5.4
	github.com/dgryski/go-farm v0.0.0-20190423205320-6a90982ecee2 // indirect
	github.com/etcd-io/bbolt v1.3.2
	github.com/go-stack/stack v1.8.0
	github.com/golang-collections/collections v0.0.0-20130729185459-604e922904d3
	github.com/jessevdk/go-flags v1.4.0
	github.com/jrick/logrotate v1.0.0
	github.com/mattn/go-colorable v0.1.1
	github.com/pkg/errors v0.8.1
	github.com/rcrowley/go-metrics v0.0.0-20181016184325-3113b8401b8a
	github.com/stretchr/testify v1.3.0
	github.com/syndtr/goleveldb v1.0.0
	golang.org/x/crypto v0.0.0-20190621222207-cc06ce4a13d4
	golang.org/x/net v0.0.0-20190503192946-f4e77d36d62c
	golang.org/x/tools v0.0.0-20190511041617-99f201b6807e
	gonum.org/v1/gonum v0.0.0-20190608115022-c5f01565d866
)

replace (
	golang.org/x/crypto v0.0.0-20181001203147-e3636079e1a4 => github.com/golang/crypto v0.0.0-20181001203147-e3636079e1a4
	golang.org/x/net v0.0.0-20180906233101-161cd47e91fd => github.com/golang/net v0.0.0-20180906233101-161cd47e91fd
	golang.org/x/net v0.0.0-20181005035420-146acd28ed58 => github.com/golang/net v0.0.0-20181005035420-146acd28ed58
	golang.org/x/tools v0.0.0-20181006002542-f60d9635b16a => github.com/golang/tools v0.0.0-20181006002542-f60d9635b16a
)
