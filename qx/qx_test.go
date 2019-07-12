package qx

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTxSign(t *testing.T) {
	k := "c39fb9103419af8be42385f3d6390b4c0c8f2cb67cf24dd43a059c4045d1a409"
	tx := "01000000024b173cffa14e2ddc6098e0c75d460e8fe4ee648d61c7552d6019b4ffd44ff20602000000ffffffff3297666d28668de36433d5f4fe5362aa6b785f4269ee7ab52936b10914f36dfe02000000ffffffff02005ed0b2000000001976a914fb9ed2cb73355778c71a9436b4e45dddb16dda2b88ac002f6859000000001976a914864c051cdb39c31f21924a5ac88b4cf82124d2c188ac000000000000000002000000000000000000000000ffffffff00000000000000000000000000ffffffff00"
	net := "testnet"
	rs, _ := TxSign(k, tx, net)
	assert.Equal(t, rs, "01000000024b173cffa14e2ddc6098e0c75d460e8fe4ee648d61c7552d6019b4ffd44ff20602000000ffffffff3297666d28668de36433d5f4fe5362aa6b785f4269ee7ab52936b10914f36dfe02000000ffffffff02005ed0b2000000001976a914fb9ed2cb73355778c71a9436b4e45dddb16dda2b88ac002f6859000000001976a914864c051cdb39c31f21924a5ac88b4cf82124d2c188ac000000000000000002000000000000000000000000ffffffff6a47304402204d4de7de2fff7e7ef4c3657eabe11f67b3de2f83c5e241f662950139318bebcd02204a5e90b1437757fc8ebba34d9343c69724679a86e02b32c24175258d39495f03012102b3e7c21a906433171cad38589335002c34a6928e19b7798224077c30f03e835e000000000000000000000000ffffffff6a473044022060e981a4e091b4f5d8771bd386e292ecb68f0b0bf1f27b5b8c7d873a6a34cfa8022007c050ea4e148f01c0f5bbd4a80c9f25e3e58b700ce174e7084bc6dade9d12c7012102b3e7c21a906433171cad38589335002c34a6928e19b7798224077c30f03e835e")
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
