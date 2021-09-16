module github.com/Qitmeer/qitmeer

go 1.12

require (
	github.com/Qitmeer/crypto v0.0.0-20200516043559-dd457edff06c
	github.com/Qitmeer/crypto/cryptonight v0.0.0-20201028030128-6ed4040ca34a
	github.com/aristanetworks/goarista v0.0.0-20200812190859-4cb0e71f3c0e
	github.com/cloudflare/roughtime v0.0.0-20200829152512-a9bb6267a4f5
	github.com/davecgh/go-spew v1.1.1
	github.com/davidlazar/go-crypto v0.0.0-20190912175916-7055855a373f // indirect
	github.com/dchest/blake256 v1.0.0
	github.com/deckarep/golang-set v1.7.1
	github.com/dgraph-io/ristretto v0.0.2
	github.com/ferranbt/fastssz v0.0.0-20200514094935-99fccaf93472
	github.com/go-stack/stack v1.8.0
	github.com/gogo/protobuf v1.3.2
	github.com/golang-collections/collections v0.0.0-20130729185459-604e922904d3
	github.com/golang/protobuf v1.4.0
	github.com/golang/snappy v0.0.1
	github.com/hashicorp/golang-lru v0.5.4
	github.com/ipfs/go-ds-leveldb v0.4.2
	github.com/ipfs/go-ipfs-addr v0.0.1
	github.com/jessevdk/go-flags v1.4.0
	github.com/jrick/logrotate v1.0.0
	github.com/libp2p/go-libp2p v0.11.0
	github.com/libp2p/go-libp2p-circuit v0.3.1
	github.com/libp2p/go-libp2p-core v0.6.1
	github.com/libp2p/go-libp2p-discovery v0.5.0
	github.com/libp2p/go-libp2p-kad-dht v0.5.0
	github.com/libp2p/go-libp2p-noise v0.1.1
	github.com/libp2p/go-libp2p-peerstore v0.2.6
	github.com/libp2p/go-libp2p-pubsub v0.3.2
	github.com/libp2p/go-libp2p-secio v0.2.2
	github.com/libp2p/go-sockaddr v0.1.0 // indirect
	github.com/magiconair/properties v1.8.1
	github.com/mattn/go-colorable v0.1.7
	github.com/minio/highwayhash v1.0.0
	github.com/multiformats/go-multiaddr v0.3.1
	github.com/multiformats/go-multiaddr-net v0.2.0
	github.com/multiformats/go-multistream v0.1.2
	github.com/pkg/errors v0.9.1
	github.com/prysmaticlabs/go-bitfield v0.0.0-20200618145306-2ae0807bef65
	github.com/prysmaticlabs/go-ssz v0.0.0-20200101200214-e24db4d9e963
	github.com/rcrowley/go-metrics v0.0.0-20190826022208-cac0b30c2563
	github.com/schollz/progressbar/v3 v3.8.3
	github.com/stretchr/testify v1.6.1
	github.com/syndtr/goleveldb v1.0.0
	github.com/urfave/cli/v2 v2.2.0
	github.com/zeromq/goczmq v4.1.0+incompatible
	go.opencensus.io v0.22.4
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5
	golang.org/x/net v0.0.0-20210226172049-e18ecbb05110
	golang.org/x/sys v0.0.0-20210910150752-751e447fb3d0
	golang.org/x/tools v0.0.0-20210106214847-113979e3529a
	gonum.org/v1/gonum v0.0.0-20190608115022-c5f01565d866
	gopkg.in/yaml.v2 v2.2.8
)

replace (
	golang.org/x/crypto v0.0.0-20181001203147-e3636079e1a4 => github.com/golang/crypto v0.0.0-20181001203147-e3636079e1a4
	golang.org/x/net v0.0.0-20180906233101-161cd47e91fd => github.com/golang/net v0.0.0-20180906233101-161cd47e91fd
	golang.org/x/net v0.0.0-20181005035420-146acd28ed58 => github.com/golang/net v0.0.0-20181005035420-146acd28ed58
	golang.org/x/tools v0.0.0-20181006002542-f60d9635b16a => github.com/golang/tools v0.0.0-20181006002542-f60d9635b16a
)
