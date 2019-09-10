package qx

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTxSign(t *testing.T) {
	k := "c39fb9103419af8be42385f3d6390b4c0c8f2cb67cf24dd43a059c4045d1a409"
	tx := "0100000001db69c95b7e472c56714435f84a9f5c8e6da1b759414c1951d1b919df066d4cfe00000000ffffffff0100943577000000001976a914c50b62be2f7c23cf0b9d904fa9984efbdb75859888ac00000000000000000100"
	net := "testnet"
	rs, _ := TxSign(k, tx, net)
	fmt.Println(rs)
	assert.Equal(t, rs, "0100000001db69c95b7e472c56714435f84a9f5c8e6da1b759414c1951d1b919df066d4cfe00000000ffffffff0100943577000000001976a914c50b62be2f7c23cf0b9d904fa9984efbdb75859888ac0000000000000000016a473044022078deab29a47651c393ef6801e4920a46673a051b5b0416bda6478c114caafce50220497c051c20d054fad4829a5511c417a4baf78de8a4d8f5377e1bf8fe7dee49f1012102b3e7c21a906433171cad38589335002c34a6928e19b7798224077c30f03e835e")
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
