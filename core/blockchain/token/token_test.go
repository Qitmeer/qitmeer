package token

import (
	"github.com/Qitmeer/qitmeer/core/types"
	"testing"
)

var (
	// private key
	privateKey []byte = []byte {
		0x9a,0xf3,0xb7,0xc0,0xb4,0xf1,0x96,0x35,
		0xf9,0x0a,0x5f,0xc7,0x22,0xde,0xfb,0x96,
		0x1a,0xc4,0x35,0x08,0xc6,0x6f,0xfe,0x5d,
		0xf9,0x92,0xe9,0x31,0x4f,0x2a,0x29,0x48,
	}
	// compressed public key
	pubkey []byte = []byte{
		0x02,0xab,0xb1,0x3c,0xd5,0x26,0x0d,0x3e,
		0x9f,0x8b,0xc3,0xdb,0x86,0x87,0x14,0x7a,
		0xce,0x7b,0x6e,0x5b,0x63,0xb0,0x61,0xaf,
		0xe3,0x7d,0x09,0xa8,0xe4,0x55,0x0c,0xd1,
		0x74,
	}
	// schnorr signature for hash.HashB([]byte("qitmeer"))
	signature []byte = []byte{
		0xb2,0xcb,0x95,0xbb,0x27,0x32,0xac,0xb9,
		0xcc,0x14,0x5f,0xe8,0x78,0xc8,0x99,0xc8,
		0xd0,0xf6,0x19,0x0a,0x3b,0x97,0xcd,0x44,
		0xf1,0x20,0xaa,0x78,0x17,0xc8,0x08,0x6d,
		0x43,0xc1,0x6d,0x61,0x1d,0xa6,0x40,0x1d,
		0xd1,0x72,0x3b,0x4d,0x9f,0x6e,0xc1,0x76,
		0xd8,0x4b,0x23,0xaa,0x82,0xc2,0xca,0x44,
		0xf9,0x4a,0x9a,0x24,0xd2,0x7e,0x80,0x7b,
	}
)


func TestCheckTokenMint(t *testing.T) {
	tests := []struct {
		name     string
		createTx func() *types.Transaction
		expected bool
	}{
		{
			"1. invalid empty tx",
			func() *types.Transaction {
				tx := &types.Transaction{}
				return tx
			},
			false,
		},
		{
			name:"2. token mint ",
			expected: false,
			createTx: func() *types.Transaction {
				tx := types.NewTransaction()

				return tx
			},
		},
	}

	for i, test:= range tests {
		if got := IsTokenMint(test.createTx()); got != test.expected {
			_,_,_,err := CheckTokenMint(test.createTx())
			t.Errorf("failed test[%d]:%v, expect [%v] but [%v], error:[%v]",i, test.name, test.expected, got, err)
		}
	}

}


//func generateKeys() {
//	printKey := func(key []byte){
//		for i, v := range key {
//			if i != 0 && i%8 == 0 {
//				fmt.Printf("\n")
//			}
//			fmt.Printf("0x%02x,", v)
//		}
//		fmt.Println()
//	}
//	// private key
//	priKeyStr := "9af3b7c0b4f19635f90a5fc722defb961ac43508c66ffe5df992e9314f2a2948"
//	priKey,_ :=hex.DecodeString(priKeyStr)
//	key,_ := secp256k1.PrivKeyFromBytes(priKey)
//	fmt.Printf("Private key : %x\n",string(privateKey))
//	printKey(priKey)
//
//	// public key
//	pubKey := key.PubKey()
//	fmt.Printf("Public key len(0x%x): %x\n", len(pubKey.SerializeCompressed()),
//		pubKey.SerializeCompressed())
//	printKey(pubKey.SerializeCompressed())
//
//	// sigature
//	message := "qitmeer"
//	messageHash := hash.HashB([]byte(message))
//	r, s, err := schnorr.Sign(key, messageHash)
//	signature := schnorr.Signature{r,s}
//	if err != nil {
//		panic(err)
//	}
//	sig := signature.Serialize()
//	fmt.Printf("Signature len(0x%x): %x\n", len(sig), sig)
//	printKey(sig)
//
//}
//
//func init() {
//	generateKeys()
//}
