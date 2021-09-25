package params

import (
	"bytes"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/encode/base58"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"github.com/stretchr/testify/assert"
	"math/big"
	"strings"
	"testing"
)

func TestPowLimitToBits(t *testing.T) {
	compact := pow.BigToCompact(testMixNetPowLimit)
	assert.Equal(t, fmt.Sprintf("0x%064x", testMixNetPowLimit), "0x0000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	assert.Equal(t, fmt.Sprintf("0x%x", compact), "0x1c00ffff")
}

//test blake2bd percent params
func TestPercent(t *testing.T) {
	types := []pow.PowType{pow.BLAKE2BD, pow.CUCKAROO, pow.CUCKATOO, pow.CUCKAROOM}
	for _, powType := range types {
		instance := pow.GetInstance(powType, 0, []byte{})
		instance.SetParams(PrivNetParam.PowConfig)
		percent := new(big.Int)
		for mheight, pi := range PrivNetParam.PowConfig.Percent {
			instance.SetMainHeight(pow.MainHeight(mheight + 1))
			percent.SetInt64(int64(pi[powType]))

			percent.Lsh(percent, 32)
			assert.Equal(t, percent.Uint64(), instance.PowPercent().Uint64())
		}
	}

}

// test address prefixes for all qitmeer network params
func TestQitmeerAddressPerfixes(t *testing.T) {
	checkAddressPrefixesAreConsistent(t, "Pm", &MainNetParams)
	checkAddressPrefixesAreConsistent(t, "Pt", &TestNetParams)
	checkAddressPrefixesAreConsistent(t, "Px", &MixNetParams)
	checkAddressPrefixesAreConsistent(t, "Pr", &PrivNetParams)
}

func checkAddressPrefixesAreConsistent(t *testing.T, privateKeyPrefix string, params *Params) {
	P := params.NetworkAddressPrefix

	Pk := P + "k"
	Pm := P + "m"
	if P == "T" {
		// Since 0.10.0 release Testnet PubKeyHashAddrID using Tn, Tm using by 0.9.x testnet
		Pm = P + "n"
	}
	Pe := P + "e"
	Pr := P + "r"
	PS := P + "S"
	pk := privateKeyPrefix

	checkEncodedKeys(t, Pk, 33, params.Name, params.PubKeyAddrID)
	checkEncodedKeys(t, Pm, 20, params.Name, params.PubKeyHashAddrID)
	checkEncodedKeys(t, Pe, 20, params.Name, params.PKHEdwardsAddrID)
	checkEncodedKeys(t, Pr, 20, params.Name, params.PKHSchnorrAddrID)
	checkEncodedKeys(t, PS, 20, params.Name, params.ScriptHashAddrID)
	checkEncodedKeys(t, pk, 33, params.Name, params.PrivateKeyID)
}

// check for the min and max possible keys
// Then prefixes are checked for mismatch.
func checkEncodedKeys(t *testing.T, desiredPrefix string, keySize int, networkName string, magic [2]byte) {
	// min and max possible keys
	// all zeroes
	minKey := bytes.Repeat([]byte{0x00}, keySize)
	// all ones
	maxKey := bytes.Repeat([]byte{0xff}, keySize)

	encodedMinKey, _ := base58.QitmeerCheckEncode(minKey, magic[:])
	encodedMaxKey, _ := base58.QitmeerCheckEncode(maxKey, magic[:])

	checkPrefix(t, desiredPrefix, string(encodedMinKey), networkName, magic)
	checkPrefix(t, desiredPrefix, string(encodedMaxKey), networkName, magic)
}

// checkPrefix checks if targetString starts with the given prefix
func checkPrefix(t *testing.T, prefix string, targetString, networkName string, magic [2]byte) {
	if strings.Index(targetString, prefix) != 0 {
		t.Logf("Address prefix mismatch for <%s>: expected <%s> received <%s>, magic <%x>",
			networkName, prefix, targetString, magic)
		t.FailNow()
	}
}
