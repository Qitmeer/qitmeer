package bls

import (
	"testing"
	"github.com/dedis/kyber/util/random"
)

func TestBLSSig(t *testing.T) {
	pairing := NewPairingFp382_1()
	sk, pk := NewKeyPair(pairing, random.New())
	msg := []byte("hello world")

	sig := Sign(pairing, sk, msg)
	err := Verify(pairing, pk, msg, sig)
	if err !=nil {
		t.Error(err)
	}

	wrongMsg := []byte("evil message")
	err = Verify(pairing, pk, msg, wrongMsg)
	if err ==nil {
		t.Error("should not verified")
	}
}
