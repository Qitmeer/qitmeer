module qitmeer

go 1.12

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/dchest/blake256 v1.0.0
	github.com/deckarep/golang-set v1.7.1
	github.com/go-stack/stack v1.8.0
	github.com/golang-collections/collections v0.0.0-20130729185459-604e922904d3
	github.com/jessevdk/go-flags v1.4.0
	github.com/jrick/logrotate v1.0.0
	github.com/mattn/go-colorable v0.1.1
	github.com/syndtr/goleveldb v1.0.0
	golang.org/x/crypto v0.0.0-20181001203147-e3636079e1a4
	golang.org/x/net v0.0.0-20180906233101-161cd47e91fd
	golang.org/x/sync v0.0.0-20180314180146-1d60e4601c6f
	golang.org/x/sys v0.0.0-20190222072716-a9d3bda3a223
	golang.org/x/text v0.3.0
)

replace (
	golang.org/x/crypto v0.0.0-20181001203147-e3636079e1a4 => github.com/golang/crypto v0.0.0-20181001203147-e3636079e1a4
	golang.org/x/net v0.0.0-20180906233101-161cd47e91fd => github.com/golang/net v0.0.0-20180906233101-161cd47e91fd
	golang.org/x/sync v0.0.0-20180314180146-1d60e4601c6f => github.com/golang/sync v0.0.0-20180314180146-1d60e4601c6f
	golang.org/x/sys v0.0.0-20180909124046-d0be0721c37e => github.com/golang/sys v0.0.0-20180909124046-d0be0721c37e
	golang.org/x/sys v0.0.0-20190222072716-a9d3bda3a223 => github.com/golang/sys v0.0.0-20190222072716-a9d3bda3a223
	golang.org/x/text v0.3.0 => github.com/golang/text v0.3.0
)
