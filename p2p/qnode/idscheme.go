/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package qnode

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/Qitmeer/qitmeer/crypto/ecc/secp256k1"
	"io"

	"github.com/Qitmeer/qitmeer/common/encode/rlp"
	"github.com/Qitmeer/qitmeer/common/math"
	"github.com/Qitmeer/qitmeer/crypto"
	"github.com/Qitmeer/qitmeer/p2p/qnr"
	"golang.org/x/crypto/sha3"
)

// List of known secure identity schemes.
var ValidSchemes = qnr.SchemeMap{
	"v4": V4ID{},
}

var ValidSchemesForTesting = qnr.SchemeMap{
	"v4":   V4ID{},
	"null": NullID{},
}

// v4ID is the "v4" identity scheme.
type V4ID struct{}

// SignV4 signs a record using the v4 scheme.
func SignV4(r *qnr.Record, privkey *ecdsa.PrivateKey) error {
	// Copy r to avoid modifying it if signing fails.
	cpy := *r
	cpy.Set(qnr.ID("v4"))
	cpy.Set(Secp256k1(privkey.PublicKey))

	h := sha3.NewLegacyKeccak256()
	rlp.Encode(h, cpy.AppendElements(nil))
	sig, err := crypto.Sign(h.Sum(nil), privkey)
	if err != nil {
		return err
	}
	sig = sig[1:] // remove header
	if err = cpy.SetSig(V4ID{}, sig); err == nil {
		*r = cpy
	}
	return err
}

func (V4ID) Verify(r *qnr.Record, sig []byte) error {
	var entry s256raw
	if err := r.Load(&entry); err != nil {
		return err
	} else if len(entry) != 33 {
		return fmt.Errorf("invalid public key")
	}

	h := sha3.NewLegacyKeccak256()
	rlp.Encode(h, r.AppendElements(nil))

	pk, err := secp256k1.ParsePubKey(entry)
	if err != nil {
		return err
	}
	if !crypto.VerifySignature(pk.ToECDSA(), h.Sum(nil), sig) {
		return qnr.ErrInvalidSig
	}
	return nil
}

func (V4ID) NodeAddr(r *qnr.Record) []byte {
	var pubkey Secp256k1
	err := r.Load(&pubkey)
	if err != nil {
		return nil
	}
	buf := make([]byte, 64)
	math.ReadBits(pubkey.X, buf[:32])
	math.ReadBits(pubkey.Y, buf[32:])
	return crypto.Keccak256(buf)
}

// Secp256k1 is the "secp256k1" key, which holds a public key.
type Secp256k1 ecdsa.PublicKey

func (v Secp256k1) QNRKey() string { return "secp256k1" }

// EncodeRLP implements rlp.Encoder.
func (v Secp256k1) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, crypto.CompressPubkey((*ecdsa.PublicKey)(&v)))
}

// DecodeRLP implements rlp.Decoder.
func (v *Secp256k1) DecodeRLP(s *rlp.Stream) error {
	buf, err := s.Bytes()
	if err != nil {
		return err
	}
	pk, err := crypto.DecompressPubkey(buf)
	if err != nil {
		return err
	}
	*v = (Secp256k1)(*pk)
	return nil
}

// s256raw is an unparsed secp256k1 public key entry.
type s256raw []byte

func (s256raw) QNRKey() string { return "secp256k1" }

// v4CompatID is a weaker and insecure version of the "v4" scheme which only checks for the
// presence of a secp256k1 public key, but doesn't verify the signature.
type v4CompatID struct {
	V4ID
}

func (v4CompatID) Verify(r *qnr.Record, sig []byte) error {
	var pubkey Secp256k1
	return r.Load(&pubkey)
}

func signV4Compat(r *qnr.Record, pubkey *ecdsa.PublicKey) {
	r.Set((*Secp256k1)(pubkey))
	if err := r.SetSig(v4CompatID{}, []byte{}); err != nil {
		panic(err)
	}
}

// NullID is the "null" QNR identity scheme. This scheme stores the node
// ID in the record without any signature.
type NullID struct{}

func (NullID) Verify(r *qnr.Record, sig []byte) error {
	return nil
}

func (NullID) NodeAddr(r *qnr.Record) []byte {
	var id ID
	r.Load(qnr.WithEntry("nulladdr", &id))
	return id[:]
}

func SignNull(r *qnr.Record, id ID) *Node {
	r.Set(qnr.ID("null"))
	r.Set(qnr.WithEntry("nulladdr", id))
	if err := r.SetSig(NullID{}, []byte{}); err != nil {
		panic(err)
	}
	return &Node{r: *r, id: id}
}
