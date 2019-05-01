# Qitmeer

The guardian of trust. The core network of the Halachain

##  Prerequisites

- Update Go to version at least 1.11 (required >= **1.11**)

Check your golang version

```bash
~ go version
go version go1.11.4 darwin/amd64
```

## How to build

```bash
~ mkdir -p /tmp/work
~ cd /tmp/work
~ git clone https://where/you/can/find/nox
~ cd nox
~ go build
~ ./nox --version
nox version 0.3.0+dev (Go version go1.11.4)
```

### How to fix `golang.org unrecognized` Issue

If you got trouble to download the `golang.org` depends automatically

```
go: golang.org/x/crypto@v0.0.0-20181001203147-e3636079e1a4: unrecognized import path "golang.org/x/crypto" (https fetch: Get https://golang.org/x/crypto?go-get=1: dial tcp 216.239.37.1:443: i/o timeout)
go: golang.org/x/tools@v0.0.0-20181006002542-f60d9635b16a: unrecognized import path "golang.org/x/tools" (https fetch: Get https://golang.org/x/tools?go-get=1: dial tcp 216.239.37.1:443: i/o timeout)
go: golang.org/x/sync@v0.0.0-20180314180146-1d60e4601c6f: unrecognized import path "golang.org/x/sync" (https fetch: Get https://golang.org/x/sync?go-get=1: dial tcp 216.239.37.1:443: i/o timeout)
go: golang.org/x/net@v0.0.0-20181005035420-146acd28ed58: unrecognized import path "golang.org/x/net" (https fetch: Get https://golang.org/x/net?go-get=1: dial tcp 216.239.37.1:443: i/o timeout)
go: golang.org/x/net@v0.0.0-20180906233101-161cd47e91fd: unrecognized import path "golang.org/x/net" (https fetch: Get https://golang.org/x/net?go-get=1: dial tcp 216.239.37.1:443: i/o timeout)
go: golang.org/x/text@v0.3.0: unrecognized import path "golang.org/x/text" (https fetch: Get https://golang.org/x/text?go-get=1: dial tcp 216.239.37.1:443: i/o timeout)
go: golang.org/x/sys@v0.0.0-20181005133103-4497e2df6f9e: unrecognized import path "golang.org/x/sys" (https fetch: Get https://golang.org/x/sys?go-get=1: dial tcp 216.239.37.1:443: i/o timeout)
go: golang.org/x/sys@v0.0.0-20180909124046-d0be0721c37e: unrecognized import path "golang.org/x/sys" (https fetch: Get https://golang.org/x/sys?go-get=1: dial tcp 216.239.37.1:443: i/o timeout)
```

you might need to `replace` the download url (ex: using a mirror site like github.com) on your `go.mod`

```
replace (
	golang.org/x/crypto v0.0.0-20181001203147-e3636079e1a4 => github.com/golang/crypto v0.0.0-20181001203147-e3636079e1a4
	golang.org/x/net v0.0.0-20180906233101-161cd47e91fd => github.com/golang/net v0.0.0-20180906233101-161cd47e91fd
	golang.org/x/net v0.0.0-20181005035420-146acd28ed58 => github.com/golang/net v0.0.0-20181005035420-146acd28ed58
	golang.org/x/sync v0.0.0-20180314180146-1d60e4601c6f => github.com/golang/sync v0.0.0-20180314180146-1d60e4601c6f
	golang.org/x/sys v0.0.0-20180909124046-d0be0721c37e => github.com/golang/sys v0.0.0-20180909124046-d0be0721c37e
	golang.org/x/sys v0.0.0-20181005133103-4497e2df6f9e => github.com/golang/sys v0.0.0-20181005133103-4497e2df6f9e
	golang.org/x/text v0.3.0 => github.com/golang/text v0.3.0
	golang.org/x/tools v0.0.0-20181006002542-f60d9635b16a => github.com/golang/tools v0.0.0-20181006002542-f60d9635b16a
)
```

**happy hacking!**
