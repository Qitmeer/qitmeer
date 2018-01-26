package entropy

import (
	"testing"
	"fmt"
	"bytes"
	"encoding/hex"
	"crypto/elliptic"
	"github.com/dindinw/dagproject/crypto"
)

func TestRandSig(t *testing.T) {
	prvkey1, _ := crypto.GenerateKey();
	key1 := prvkey1.PublicKey
	pubkey1 := elliptic.Marshal(crypto.S256(), key1.X, key1.Y)

	msg := GetEntropyCSPRNG(32)
	const TestCount = 100
	for i := 0; i < TestCount; i++ {
		randSig := GetEntropyCSPRNG(65)
		randSig[32] &= 0x70
		randSig[64] %= 4
		key2, err := crypto.SigToPub(msg, randSig)
		if err == nil {
			//fmt.Printf("key2=%X,%X\n",key2.X,key2.Y)
			pubkey2 := elliptic.Marshal(crypto.S256(), key2.X, key2.Y)
			if bytes.Equal(pubkey1, pubkey2) {
				t.Fatalf("iteration: %d: pubkey mismatch: do NOT want %x: ", i, pubkey2)
			}
		}else{
			// the SigToPub might fail, when calculated Rx is larger than curve P
			//fmt.Printf("err=%s\n",err)
		}
	}
}

func ExampleGetEntropyCSPRNG() {
	e := GetEntropyCSPRNG(32)

	// The slice should now contain random bytes instead of only zeroes.
	fmt.Println(bytes.Equal(e, make([]byte, 32)))
	// 32 bytes -> 32*8 = 256bits

	var testKey, _ = crypto.ToECDSA(e) //to private key

	addr :=crypto.PubkeyToEthAddress(testKey.PublicKey)
	fmt.Printf("entropy length = %d\n",len(e))   // [32]bytes -> 32*8 -> 256bits
	fmt.Printf("entropy hexlen = %d\n",len(hex.EncodeToString(e))) //64*4-> 256bits
	fmt.Printf("address length = %d\n",len(addr))
	// Output:
	//false
	//entropy length = 32
	//entropy hexlen = 64
	//address length = 20
}
