module github.com/Qitmeer/qitmeer

go 1.12

require (
	github.com/Qitmeer/qitmeer-lib v0.0.0-20190828083637-18d335b214c8
	github.com/davecgh/go-spew v1.1.1
	github.com/golang-collections/collections v0.0.0-20130729185459-604e922904d3
	github.com/jessevdk/go-flags v1.4.0
	github.com/jrick/logrotate v1.0.0
	github.com/mattn/go-colorable v0.1.1
	github.com/rcrowley/go-metrics v0.0.0-20181016184325-3113b8401b8a
	github.com/syndtr/goleveldb v1.0.0
	golang.org/x/tools v0.0.0-20190511041617-99f201b6807e
	gonum.org/v1/gonum v0.0.0-20190608115022-c5f01565d866
)

replace (
	golang.org/x/crypto v0.0.0-20181001203147-e3636079e1a4 => github.com/golang/crypto v0.0.0-20181001203147-e3636079e1a4
	golang.org/x/net v0.0.0-20180906233101-161cd47e91fd => github.com/golang/net v0.0.0-20180906233101-161cd47e91fd
	golang.org/x/net v0.0.0-20181005035420-146acd28ed58 => github.com/golang/net v0.0.0-20181005035420-146acd28ed58
	golang.org/x/tools v0.0.0-20181006002542-f60d9635b16a => github.com/golang/tools v0.0.0-20181006002542-f60d9635b16a
)
