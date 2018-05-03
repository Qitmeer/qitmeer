package crypto

import (
	"testing"
	"encoding/hex"
)

func TestCreateEd25519PrivateKey(t *testing.T) {
	_, err := GenerateKeyEd25519()
	if err != nil {
		t.Fatal(err)
	}
}

func TestEd25519PrivateKeyFromByte(t *testing.T) {
	privKeyBytes, _ := hex.DecodeString("162e84a5ffd1a131ac7d9fa127b66d1bf203b161f16b32ab74643bc6f1158fd7")
	privkey, _ := parsePrivKeyFromBytes(privKeyBytes)
	if privkey == nil {
		t.Errorf("error when parsing privkey from bytes")
	}
}
