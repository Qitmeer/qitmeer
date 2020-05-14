package keccak256

import (
	"encoding/binary"
)

var RC = [...]uint64{
	0x0000000000000001,
	0x0000000000008082,
	0x800000000000808a,
	0x8000000080008000,
	0x000000000000808b,
	0x0000000080000001,
	0x8000000080008081,
	0x8000000000008009,
	0x000000000000008a,
	0x0000000000000088,
	0x0000000080008009,
	0x000000008000000a,
	0x000000008000808b,
	0x800000000000008b,
	0x8000000000008089,
	0x8000000000008003,
	0x8000000000008002,
	0x8000000000000080,
	0x000000000000800a,
	0x800000008000000a,
	0x8000000080008081,
	0x8000000000008080,
	0x0000000080000001,
	0x8000000080008008}

func ROL(a uint64, offset uint) uint64 {
	return ((a << offset) | (a >> (64 - offset)))
}

var Aba, Abe, Abi, Abo, Abu, Aga, Age, Agi, Ago, Agu, Aka, Ake, Aki, Ako, Aku, Ama, Ame, Ami, Amo, Amu, Asa, Ase, Asi, Aso, Asu uint64
var BCa, BCe, BCi, BCo, BCu uint64
var Da, De, Di, Do, Du uint64
var Eba, Ebe, Ebi, Ebo, Ebu, Ega, Ege, Egi, Ego, Egu, Eka, Eke, Eki, Eko, Eku, Ema, Eme, Emi, Emo, Emu, Esa, Ese, Esi, Eso, Esu uint64
var round uint32

func keccak_core() {
	for round := 0; round < 24; round += 1 {
		BCa = Aba ^ Aga ^ Aka ^ Ama ^ Asa
		BCe = Abe ^ Age ^ Ake ^ Ame ^ Ase
		BCi = Abi ^ Agi ^ Aki ^ Ami ^ Asi
		BCo = Abo ^ Ago ^ Ako ^ Amo ^ Aso
		BCu = Abu ^ Agu ^ Aku ^ Amu ^ Asu
		Da = BCu ^ ROL(BCe, 1)
		De = BCa ^ ROL(BCi, 1)
		Di = BCe ^ ROL(BCo, 1)
		Do = BCi ^ ROL(BCu, 1)
		Du = BCo ^ ROL(BCa, 1)
		Aba ^= Da
		BCa = Aba
		Age ^= De
		BCe = ROL(Age, 44)
		Aki ^= Di
		BCi = ROL(Aki, 43)
		Amo ^= Do
		BCo = ROL(Amo, 21)
		Asu ^= Du
		BCu = ROL(Asu, 14)
		Eba = BCa ^ ((^BCe) & BCi)
		Eba ^= RC[round]
		Ebe = BCe ^ ((^BCi) & BCo)
		Ebi = BCi ^ ((^BCo) & BCu)
		Ebo = BCo ^ ((^BCu) & BCa)
		Ebu = BCu ^ ((^BCa) & BCe)
		Abo ^= Do
		BCa = ROL(Abo, 28)
		Agu ^= Du
		BCe = ROL(Agu, 20)
		Aka ^= Da
		BCi = ROL(Aka, 3)
		Ame ^= De
		BCo = ROL(Ame, 45)
		Asi ^= Di
		BCu = ROL(Asi, 61)
		Ega = BCa ^ ((^BCe) & BCi)
		Ege = BCe ^ ((^BCi) & BCo)
		Egi = BCi ^ ((^BCo) & BCu)
		Ego = BCo ^ ((^BCu) & BCa)
		Egu = BCu ^ ((^BCa) & BCe)
		Abe ^= De
		BCa = ROL(Abe, 1)
		Agi ^= Di
		BCe = ROL(Agi, 6)
		Ako ^= Do
		BCi = ROL(Ako, 25)
		Amu ^= Du
		BCo = ROL(Amu, 8)
		Asa ^= Da
		BCu = ROL(Asa, 18)
		Eka = BCa ^ ((^BCe) & BCi)
		Eke = BCe ^ ((^BCi) & BCo)
		Eki = BCi ^ ((^BCo) & BCu)
		Eko = BCo ^ ((^BCu) & BCa)
		Eku = BCu ^ ((^BCa) & BCe)
		Abu ^= Du
		BCa = ROL(Abu, 27)
		Aga ^= Da
		BCe = ROL(Aga, 36)
		Ake ^= De
		BCi = ROL(Ake, 10)
		Ami ^= Di
		BCo = ROL(Ami, 15)
		Aso ^= Do
		BCu = ROL(Aso, 56)
		Ema = BCa ^ ((^BCe) & BCi)
		Eme = BCe ^ ((^BCi) & BCo)
		Emi = BCi ^ ((^BCo) & BCu)
		Emo = BCo ^ ((^BCu) & BCa)
		Emu = BCu ^ ((^BCa) & BCe)
		Abi ^= Di
		BCa = ROL(Abi, 62)
		Ago ^= Do
		BCe = ROL(Ago, 55)
		Aku ^= Du
		BCi = ROL(Aku, 39)
		Ama ^= Da
		BCo = ROL(Ama, 41)
		Ase ^= De
		BCu = ROL(Ase, 2)
		Esa = BCa ^ ((^BCe) & BCi)
		Ese = BCe ^ ((^BCi) & BCo)
		Esi = BCi ^ ((^BCo) & BCu)
		Eso = BCo ^ ((^BCu) & BCa)
		Esu = BCu ^ ((^BCa) & BCe)
		Aba = Eba
		Abe = Ebe
		Abi = Ebi
		Abo = Ebo
		Abu = Ebu
		Aga = Ega
		Age = Ege
		Agi = Egi
		Ago = Ego
		Agu = Egu
		Aka = Eka
		Ake = Eke
		Aki = Eki
		Ako = Eko
		Aku = Eku
		Ama = Ema
		Ame = Eme
		Ami = Emi
		Amo = Emo
		Amu = Emu
		Asa = Esa
		Ase = Ese
		Asi = Esi
		Aso = Eso
		Asu = Esu
	}
}

func keccak_init() {
	Aba = 0
	Abe = 0
	Abi = 0
	Abo = 0
	Abu = 0
	Aga = 0
	Age = 0
	Agi = 0
	Ago = 0
	Agu = 0
	Aka = 0
	Ake = 0
	Aki = 0
	Ako = 0
	Aku = 0
	Ama = 0
	Ame = 0
	Ami = 0
	Amo = 0
	Amu = 0
	Asa = 0
	Ase = 0
	Asi = 0
	Aso = 0
	Asu = 0
}

func keccak_update(inbuf []uint64) {
	Aba ^= inbuf[0]
	Abe ^= inbuf[1]
	Abi ^= inbuf[2]
	Abo ^= inbuf[3]
	Abu ^= inbuf[4]
	Aga ^= inbuf[5]
	Age ^= inbuf[6]
	Agi ^= inbuf[7]
	Ago ^= inbuf[8]
	Agu ^= inbuf[9]
	Aka ^= inbuf[10]
	Ake ^= inbuf[11]
	Aki ^= inbuf[12]
	Ako ^= inbuf[13]
	Aku ^= inbuf[14]
	Ama ^= inbuf[15]
	Ame ^= inbuf[16]
}

func keccak256_32(inraw []uint8, inlen int, out []uint8) {
	var in = make([]uint8, 18*8)
	var in64 = make([]uint64, 18)

	var lastblock bool

	for !lastblock {
		if inlen >= 136 {
			for i := 0; i < 136; i++ {
				in[i] = inraw[i]
			}
			inlen -= 136
			inraw = inraw[136:]
		} else {
			for i := 0; i < inlen; i++ {
				in[i] = inraw[i]
			}
			in[inlen] = 0x80
			for i := 0; i < (136 - inlen - 1); i++ {
				in[inlen+1+i] = 0
			}

			in[136-1] = uint8(inlen)
			lastblock = true
		}
		for i := 0; i < len(in); i += 8 {
			in64[i/8] = binary.LittleEndian.Uint64(in[i:])
		}

		keccak_update(in64)
		keccak_core()
	}

	binary.LittleEndian.PutUint64(out[0:8], Aba)
	binary.LittleEndian.PutUint64(out[8:16], Abe)
	binary.LittleEndian.PutUint64(out[16:24], Abi)
	binary.LittleEndian.PutUint64(out[24:32], Abo)

}

func Sph_keccak256_process(data []uint8, dst []uint8, length int) {
	keccak_init()
	keccak256_32(data, length, dst)
}

func Sum256(in []byte) [32]byte {
	out := make([]byte, 32)
	Sph_keccak256_process(in, out, len(in))
	var sum [32]byte
	copy(sum[:], out[:32])
	return sum
}
