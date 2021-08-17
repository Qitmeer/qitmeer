package p2p

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer/crypto/ecc/secp256k1"
	"github.com/Qitmeer/qitmeer/p2p/common"
	"github.com/Qitmeer/qitmeer/p2p/iputils"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
	"github.com/Qitmeer/qitmeer/p2p/qnr"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/version"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/prysmaticlabs/go-bitfield"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"
	"time"
)

const keyPath = "network.key"
const metaDataPath = "metaData"
const PeerStore = "peerstore"

const dialTimeout = 1 * time.Second

// Retrieves an external ipv4 address and converts into a libp2p formatted value.
func IpAddr() net.IP {
	ip, err := iputils.ExternalIPv4()
	if err != nil {
		log.Error(fmt.Sprintf("Could not get IPv4 address: %v", err))
		return nil
	}
	return net.ParseIP(ip)
}

// Determines a private key for p2p networking from the p2p service's
// configuration struct. If no key is found, it generates a new one.
func privKey(cfg *common.Config) (*ecdsa.PrivateKey, error) {
	return PrivateKey(cfg.DataDir, cfg.PrivateKey, cfg.ReadWritePermissions)
}

// Determines a private key for p2p networking from the p2p service's
// configuration struct. If no key is found, it generates a new one.
func PrivateKey(dataDir string, privateKeyPath string, readWritePermissions os.FileMode) (*ecdsa.PrivateKey, error) {
	defaultKeyPath := path.Join(dataDir, keyPath)

	_, err := os.Stat(defaultKeyPath)
	defaultKeysExist := !os.IsNotExist(err)
	if err != nil && defaultKeysExist {
		return nil, err
	}

	if privateKeyPath == "" && !defaultKeysExist {
		priv, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
		if err != nil {
			return nil, err
		}
		rawbytes, err := priv.Raw()
		if err != nil {
			return nil, err
		}
		dst := make([]byte, hex.EncodedLen(len(rawbytes)))
		hex.Encode(dst, rawbytes)
		if err = ioutil.WriteFile(defaultKeyPath, dst, readWritePermissions); err != nil {
			return nil, err
		}
		convertedKey := convertFromInterfacePrivKey(priv)
		return convertedKey, nil
	}
	if defaultKeysExist && privateKeyPath == "" {
		privateKeyPath = defaultKeyPath
	}
	return retrievePrivKeyFromFile(privateKeyPath)
}

func convertFromInterfacePrivKey(privkey crypto.PrivKey) *ecdsa.PrivateKey {
	typeAssertedKey := (*ecdsa.PrivateKey)((*secp256k1.PrivateKey)(privkey.(*crypto.Secp256k1PrivateKey)))
	return typeAssertedKey
}

func ConvertToInterfacePrivkey(privkey *ecdsa.PrivateKey) crypto.PrivKey {
	typeAssertedKey := crypto.PrivKey((*crypto.Secp256k1PrivateKey)((*secp256k1.PrivateKey)(privkey)))
	return typeAssertedKey
}

func convertToInterfacePubkey(pubkey *ecdsa.PublicKey) crypto.PubKey {
	typeAssertedKey := crypto.PubKey((*crypto.Secp256k1PublicKey)((*secp256k1.PublicKey)(pubkey)))
	return typeAssertedKey
}

// SerializeQNR takes the qnr record in its key-value form and serializes it.
func SerializeQNR(record *qnr.Record) (string, error) {
	buf := bytes.NewBuffer([]byte{})
	if err := record.EncodeRLP(buf); err != nil {
		return "", fmt.Errorf("could not encode ENR record to bytes:%w", err)
	}
	enrString := base64.URLEncoding.EncodeToString(buf.Bytes())
	return enrString, nil
}

// Retrieves a p2p networking private key from a file path.
func retrievePrivKeyFromFile(path string) (*ecdsa.PrivateKey, error) {
	src, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error(fmt.Sprintf("Error reading private key from file:%v", err))
		return nil, err
	}
	dst := make([]byte, hex.DecodedLen(len(src)))
	_, err = hex.Decode(dst, src)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex string:%w", err)
	}
	unmarshalledKey, err := crypto.UnmarshalSecp256k1PrivateKey(dst)
	if err != nil {
		return nil, err
	}
	return convertFromInterfacePrivKey(unmarshalledKey), nil
}

// Retrieves node p2p metadata from a set of configuration values
// from the p2p service.
func metaDataFromConfig(cfg *common.Config) (*pb.MetaData, error) {
	defaultKeyPath := path.Join(cfg.DataDir, metaDataPath)
	metaDataPath := cfg.MetaDataDir

	_, err := os.Stat(defaultKeyPath)
	defaultMetadataExist := !os.IsNotExist(err)
	if err != nil && defaultMetadataExist {
		return nil, err
	}
	if metaDataPath == "" && !defaultMetadataExist {
		metaData := &pb.MetaData{
			SeqNumber: 0,
			Subnets:   bitfield.NewBitvector64(),
		}
		dst, err := metaData.Marshal()
		if err != nil {
			return nil, err
		}
		if err = ioutil.WriteFile(defaultKeyPath, dst, cfg.ReadWritePermissions); err != nil {
			return nil, err
		}
		return metaData, nil
	}
	if defaultMetadataExist && metaDataPath == "" {
		metaDataPath = defaultKeyPath
	}
	src, err := ioutil.ReadFile(metaDataPath)
	if err != nil {
		log.Error(fmt.Sprintf("Error reading metadata from file:%s", err.Error()))
		return nil, err
	}
	metaData := &pb.MetaData{}
	if err := metaData.Unmarshal(src); err != nil {
		return nil, err
	}
	return metaData, nil
}

// Attempt to dial an address to verify its connectivity
func verifyConnectivity(addr string, port uint, protocol string) {
	if addr != "" {
		a := fmt.Sprintf("%s:%d", addr, port)
		conn, err := net.DialTimeout(protocol, a, dialTimeout)
		if err != nil {
			log.Warn(fmt.Sprintf("IP address is not accessible:protocol=%s address=%s error=%s", protocol, a, err))
			return
		}
		if err := conn.Close(); err != nil {
			log.Debug(fmt.Sprintf("Could not close connection:protocol=%s address=%s error=%s", protocol, a, err))
		}
	}
}

func filterBootStrapAddrs(hostID string, addrs []string) []string {
	result := []string{}
	for _, addr := range addrs {
		if strings.HasSuffix(addr, hostID) {
			continue
		}
		result = append(result, addr)
	}
	return result
}

func BuildUserAgent(name string) string {
	return fmt.Sprintf("%s|%s|%s", name, version.String(), params.ActiveNetParams.Name)
}
