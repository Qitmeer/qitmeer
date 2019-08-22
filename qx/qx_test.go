package qx

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTxSign(t *testing.T) {
	k := "c39fb9103419af8be42385f3d6390b4c0c8f2cb67cf24dd43a059c4045d1a409"
	tx := "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff145108f7f5221cb995f41b092f7169746d6565722fffffffff0100f90295000000001976a914c1777151516afe2b9f59bbd1479231aa2f250d2888ac0000000000000000"
	net := "testnet"
	rs, _ := TxSign(k, tx, net)
	assert.Equal(t, rs, "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff6a4730440220191a9bb2a7456995c2fc652b44d8ea02554e7d015e6eb54ab3d4959829b4255902201c1ca54b85b495682726d3d6566ffd34b88f7f8ccb0484af04ca41590d8f3fbe012102b3e7c21a906433171cad38589335002c34a6928e19b7798224077c30f03e835effffffff0100f90295000000001976a914c1777151516afe2b9f59bbd1479231aa2f250d2888ac0000000000000000")
}

func TestNewEntropy(t *testing.T) {
	s, _ := NewEntropy(32)
	fmt.Printf("%s\n", s)
	assert.Equal(t, len(s), 64)

}

func TestEcNew(t *testing.T) {
	s, _ := EcNew("secp256k1", "7686a4df8171ebf04ede968167d0593fd4fbd8ee9feb07d453e768e06cc5e51d")
	assert.Equal(t, s, "dbae6e0b3174330ad24be8d952307e95106eb8d573defdc1f393ef2abf2e7b9c")
}

func TestEcPrivateKeyToEcPublicKey(t *testing.T) {
	s, _ := EcPrivateKeyToEcPublicKey(false, "dbae6e0b3174330ad24be8d952307e95106eb8d573defdc1f393ef2abf2e7b9c")
	assert.Equal(t, s, "02addd806e8813f85fad05b97541915eb3a1f27528d3156f2ef8166823d6722b58")
}

func TestEcPubKeyToAddress(t *testing.T) {
	s, _ := EcPubKeyToAddress("testnet", "02addd806e8813f85fad05b97541915eb3a1f27528d3156f2ef8166823d6722b58")
	assert.Equal(t, s, "TmgMiXziDuFiyLc159zagcCnmVxhReojytr")
}

func TestCreateAddress(t *testing.T) {
	s, _ := NewEntropy(32)
	k, _ := EcNew("secp256k1", s)
	p, _ := EcPrivateKeyToEcPublicKey(false, k)
	a, _ := EcPubKeyToAddress("testnet", p)
	fmt.Printf("%s\n%s\n%s\n%s\n", s, k, p, a)
	assert.Contains(t, a, "Tm")
}
