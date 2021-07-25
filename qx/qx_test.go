package qx

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTxSign(t *testing.T) {
	k := "c39fb9103419af8be42385f3d6390b4c0c8f2cb67cf24dd43a059c4045d1a409"
	tx := "0100000001255fea249c9747f7f4a8c432ca6f6bbed20db023fa9101288cad1a4e8056a5f600000000ffffffff01000000943577000000001976a914c50b62be2f7c23cf0b9d904fa9984efbdb75859888ac0000000000000000a2b54c5e0100"
	net := "testnet"
	pks := make([]string, 0)
	script, _ := DecodePkString("OP_DUP OP_HASH160 864c051cdb39c31f21924a5ac88b4cf82124d2c1 OP_EQUALVERIFY OP_CHECKSIG")
	//
	pks = append(pks, script)
	rs, err := TxSign([]string{k}, tx, net, pks)
	if err != nil {
		t.Errorf("%v", err)
	}
	fmt.Println(rs)
	assert.Equal(t, rs, "0100000001255fea249c9747f7f4a8c432ca6f6bbed20db023fa9101288cad1a4e8056a5f600000000ffffffff01000000943577000000001976a914c50b62be2f7c23cf0b9d904fa9984efbdb75859888ac0000000000000000a2b54c5e016b483045022100fae2248ae8f68854f5f5a522caa326f14e6ed9fa55c1a6da0cc0ad7659ec0e100220784c0f325fdb53be1fbf6466dc8a26297696fcebf3dffbc57a3a1fa0f9a96d42012102b3e7c21a906433171cad38589335002c34a6928e19b7798224077c30f03e835e")
}

func TestTxEncode(t *testing.T) {
	inputs := make([]Input, 0)
	outputs := make(map[string]Amount)
	inputs = append(inputs, Input{
		TxID:     "25517e3b3759365e80a164a3d4d2db2462c5d6888e4bd874c5fbfbb6fb130b41",
		OutIndex: 0,
	})
	outputs["TnTf7hM9kzm7ssvQ7RAcrjni5jGQbVykd2w"] = Amount{
		0,
		2083509771,
		0,
	}
	outputs["TnU8gXq9xHFrfchwk2bjyGHR2HMswANsVU5"] = Amount{
		0,
		100000000,
		0,
	}
	timestamp, _ := time.Parse("2016-01-02 15:04:05", "2019-13-14 00:00:00")
	rs, _ := TxEncode(1, 0, &timestamp, inputs, outputs)

	fmt.Println(rs)
	assert.Equal(t, rs, "0100000001410b13fbb6fbfbc574d84b8e88d6c56224dbd2d4a364a1805e3659373b7e512500000000ffffffff0200000bd62f7c000000001976a914afda839fa515ffdbcbc8630b60909c64cfd73f7a88ac000000e1f505000000001976a914b51127b89f9b704e7cfbc69286f0de2e00e7196988ac000000000000000000096e880100")
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
	assert.Equal(t, s, "TnV2vWDJoKceiyUHFCqFKaTLLkyDK6cY5ka")
}

func TestCreateAddress(t *testing.T) {
	s, _ := NewEntropy(32)
	k, _ := EcNew("secp256k1", s)
	fmt.Println("[privateKey]", k)
	p, _ := EcPrivateKeyToEcPublicKey(false, k)
	a, _ := EcPubKeyToAddress("mixnet", p)
	fmt.Println("[address]", a)
	fmt.Printf("%s\n%s\n%s\n%s\n", s, k, p, a)
	assert.Contains(t, a, "Xm")
}

func TestCreateMixParamsAddressPublicKeyHash(t *testing.T) {
	times := 0
	for {
		if times > 20000 {
			break
		}
		s, _ := NewEntropy(32)
		k, _ := EcNew("secp256k1", s)
		p, _ := EcPrivateKeyToEcPublicKey(false, k)
		a, _ := EcPubKeyToAddress("mixnet", p)
		//fmt.Printf("%s\n%s\n%s\n%s\n", s, k, p, a)
		if !assert.Contains(t, a, "Xm") {
			break
		}
		times++
	}
}

func TestCreateMixParamsSciptToHashAddress(t *testing.T) {
	times := 0
	for {
		if times > 20000 {
			break
		}
		s, _ := NewEntropy(32)
		k, _ := EcNew("secp256k1", s)
		p, _ := EcPrivateKeyToEcPublicKey(false, k)
		a, _ := EcScriptKeyToAddress("mixnet", p)
		//fmt.Printf("%s\n%s\n%s\n%s\n", s, k, p, a)
		if !assert.Contains(t, a, "XS") {
			break
		}
		times++
	}

}

func TestHashCompactToHashrate(t *testing.T) {
	CompactToHashrate("471859199", "H", false, 100)
	CompactToHashrate("471859199", "M", true, 100)
	// output :
	// 343597547
	// 343.597547 MH/s
}

func TestHashrateToCompact(t *testing.T) {
	HashrateToCompact("343597547", 100)
	// output :
	// 471859199
}

func TestCompactToGPS(t *testing.T) {
	CompactToGPS("36284416", 43, 1856, false)
	CompactToGPS("36284416", 43, 1856, true)
	// output :
	// 6.681034 GPS
}

func TestGPSToCompact(t *testing.T) {
	GPSToCompact("6.681034", 43, 1856)
	// output :
	// 36284416
}
