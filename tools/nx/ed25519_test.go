// Copyright 2017-2018 The nox developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import (
	"testing"
	"qitmeer/crypto/ecc/ed25519"
	"fmt"
	"encoding/hex"
	"qitmeer/common/hash"
	"qitmeer/common/encode/base58"
	"qitmeer/params"
	"github.com/minio/blake2b-simd"
	"log"
	"github.com/stretchr/testify/assert"
)

//test create address
func TestCreateAddressByEd25519(t *testing.T) {
	masterKey,err := edwards.CreatePrivateKey()
	if err!=nil {
		errExit(err)
	}
	log.Println("[ed25519 private key]",masterKey)
	data,err := hex.DecodeString(masterKey)
	if err != nil{
		errExit(err)
	}
	_, pubKey,err := edwards.FromPrivateKeyByte(data)
	if err != nil{
		errExit(err)
	}
	log.Println(fmt.Sprintf("[ed25519 public key]%x",pubKey.SerializeUncompressed()))
	if err != nil {
		errExit(err)
	}
	h := hash.Hash160(pubKey.SerializeUncompressed())
	p := params.PrivNetParams
	address := base58.NoxCheckEncode(h, p.PKHEdwardsAddrID[:])
	log.Printf("%s\n",address)
}

//test sign and verify sign
func TestSignByEd25519(t *testing.T) {
	masterKey,err := edwards.CreatePrivateKey()
	if err!=nil {
		errExit(err)
	}
	log.Println("[ed25519 private key]",masterKey)
	data,err := hex.DecodeString(masterKey)
	if err != nil{
		errExit(err)
	}
	privKey, pubKey,err := edwards.FromPrivateKeyByte(data)
	if err != nil{
		errExit(err)
	}
	c := edwards.TwistedEdwardsCurve{}
	c.InitParam25519()
	content := "hello world"
	h := blake2b.Sum256([]byte(content))
	r,s,err := edwards.Sign(&c,&privKey,h[:])

	if err != nil{
		errExit(err)
	}

	if edwards.Verify(&pubKey, h[:], r, s){
		log.Println("verify success")
	} else{
		log.Println("verify failed")
	}
}

//test encrypt and decrypt
func TestEncodeByEd25519(t *testing.T) {
	masterKey,err := edwards.CreatePrivateKey()
	if err!=nil {
		errExit(err)
	}
	log.Println("[ed25519 private key]",masterKey)
	data,err := hex.DecodeString(masterKey)
	if err != nil{
		errExit(err)
	}
	privKey, pubKey,err := edwards.FromPrivateKeyByte(data)
	if err != nil{
		errExit(err)
	}
	c := edwards.TwistedEdwardsCurve{}
	c.InitParam25519()
	content := []byte("hello world")
	r,err := edwards.Encrypt(&c,&pubKey,content[:])

	if err != nil{
		errExit(err)
	}
	log.Println(fmt.Sprintf("secret result:%x",r))
	r2,err := edwards.Decrypt(&c,&privKey, r)
	if err != nil{
		errExit(err)
	}
	log.Println(fmt.Sprintf("decred result:%s",r2))
	assert.Equal(t, content, r2)
}