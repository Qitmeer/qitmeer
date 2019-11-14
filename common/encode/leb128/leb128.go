// Copyright (c) 2017-2018 The qitmeer developers

package leb128

import (
	"math/big"
)

// The code originally copied from https://github.com/filecoin-project/go-leb128
//
// LEB128 or Little Endian Base 128 is a form of variable-length code compression
// used to store an arbitrarily large integer in a small number of bytes.
// see :
//    https://en.wikipedia.org/wiki/LEB128.
// see also
//    https://en.wikipedia.org/wiki/Variable-length_quantity
//
// Unsigned LEB128
//   unsigned number 624485
//		 MSB ----------------- LSB
//		      10011000011101100101  In raw binary
//		     010011000011101100101  Padded to a multiple of 7 bits
//		 0100110  0001110  1100101  Split into 7-bit groups
//		00100110 10001110 11100101  Add high 1 bits on all but last (most significant) group to form bytes
//		    0x26     0x8E     0xE5  In hexadecimal
//
//	     0xE5 0x8E 0x26             Output stream (LSB to MSB)
//
// Signed LEB128
//   signed number -624485 (0xFFF6789B)
//		 MSB ------------------ LSB
//		      01100111100010011011  In raw two's complement binary
//		     101100111100010011011  Sign extended to a multiple of 7 bits
//		 1011001  1110001  0011011  Split into 7-bit groups
//		01011001 11110001 10011011  Add high 1 bits on all but last (most significant) group to form bytes
//		    0x59     0xF1     0x9B  In hexadecimal
//
//		→ 0x9B 0xF1 0x59            Output stream (LSB to MSB)
//
// TODO revisit & refactor the impl & refactor the current core serialization
//
// Reference:
//
// 1.) Varint define in Git (https://github.com/git/git/blob/master/varint.c)
//
//		uintmax_t decode_varint(const unsigned char **bufp)
//		{
//			const unsigned char *buf = *bufp;
//			unsigned char c = *buf++;
//			uintmax_t val = c & 127;
//			while (c & 128) {
//				val += 1;
//				if (!val || MSB(val, 7))
//					return 0; /* overflow */
//				c = *buf++;
//				val = (val << 7) + (c & 127);
//			}
//			*bufp = buf;
//			return val;
//		}
//
//		int encode_varint(uintmax_t value, unsigned char *buf)
//		{
//			unsigned char varint[16];
//			unsigned pos = sizeof(varint) - 1;
//			varint[pos] = value & 127;
//			while (value >>= 7)
//				varint[--pos] = 128 | (--value & 127);
//			if (buf)
//				memcpy(buf, varint + pos, sizeof(varint) - pos);
//			return sizeof(varint) - pos;
//		}
//
// 2.) VarInt Define in the Bitcoin Core (https://github.com/bitcoin/bitcoin/blob/master/src/serialize.h)
//
//   Variable-length integers: bytes are a MSB base-128 encoding of the number.
//   The high bit in each byte signifies whether another digit follows. To make
//   sure the encoding is one-to-one, one is subtracted from all but the last digit.
//   Thus, the byte sequence a[] with length len, where all but the last byte
//   has bit 128 set, encodes the number:
//
//    (a[len-1] & 0x7F) + sum(i=1..len-1, 128^i*((a[len-i-1] & 0x7F)+1))
//
//   Properties:
//   * Very small (0-127: 1 byte, 128-16511: 2 bytes, 16512-2113663: 3 bytes)
//   * Every integer has exactly one encoding
//   * Encoding does not depend on size of original integer type
//   * No redundancy: every (infinite) byte sequence corresponds to a list
//     of encoded integers.
//
//   0:         [0x00]  256:        [0x81 0x00]
//   1:         [0x01]  16383:      [0xFE 0x7F]
//   127:       [0x7F]  16384:      [0xFF 0x00]
//   128:  [0x80 0x00]  16511:      [0xFF 0x7F]
//   255:  [0x80 0x7F]  65535: [0x82 0xFE 0x7F]
//   2^32:           [0x8E 0xFE 0xFE 0xFF 0x00]
//
//   Currently there is no support for signed encodings. The default mode will not
//   compile with signed values, and the legacy "nonnegative signed" mode will
//   accept signed values, but improperly encode and decode them if they are
//   negative. In the future, the DEFAULT mode could be extended to support
//   negative numbers in a backwards compatible way, and additional modes could be
//   added to support different varint formats (e.g. zigzag encoding).
//
//   	template<VarIntMode Mode, typename I>
//		inline unsigned int GetSizeOfVarInt(I n)
//		{
//		    CheckVarIntMode<Mode, I>();
//		    int nRet = 0;
//		    while(true) {
//		        nRet++;
//		        if (n <= 0x7F)
//		            break;
//		        n = (n >> 7) - 1;
//		    }
//		    return nRet;
//		}
//
//		template<typename Stream, VarIntMode Mode, typename I>
//		void WriteVarInt(Stream& os, I n)
//		{
//		    CheckVarIntMode<Mode, I>();
//		    unsigned char tmp[(sizeof(n)*8+6)/7];
//		    int len=0;
//		    while(true) {
//		        tmp[len] = (n & 0x7F) | (len ? 0x80 : 0x00);
//		        if (n <= 0x7F)
//		            break;
//		        n = (n >> 7) - 1;
//		        len++;
//		    }
//		    do {
//		        ser_writedata8(os, tmp[len]);
//		    } while(len--);
//		}
//
//		template<typename Stream, VarIntMode Mode, typename I>
//		I ReadVarInt(Stream& is)
//		{
//		    CheckVarIntMode<Mode, I>();
//		    I n = 0;
//		    while(true) {
//		        unsigned char chData = ser_readdata8(is);
//		        if (n > (std::numeric_limits<I>::max() >> 7)) {
//		           throw std::ios_base::failure("ReadVarInt(): size too large");
//		        }
//		        n = (n << 7) | (chData & 0x7F);
//		        if (chData & 0x80) {
//		            if (n == std::numeric_limits<I>::max()) {
//		                throw std::ios_base::failure("ReadVarInt(): size too large");
//		            }
//		            n++;
//		        } else {
//		            return n;
//        		}
//    		}
//		}
//

// FromUInt64 encodes n with LEB128 and returns the encoded bytes.
func FromUInt64(n uint64) (out []byte) {
	more := true
	for more {
		b := byte(n & 0x7F)
		n >>= 7
		if n == 0 {
			more = false
		} else {
			b = b | 0x80
		}
		out = append(out, b)
	}
	return
}

// TODO, not check the input overflow
// ToUInt64 decodes LEB128-encoded bytes into a uint64.
func ToUInt64(encoded []byte) uint64 {
	var result uint64
	var shift, i uint
	for {
		b := encoded[i]
		result |= (uint64(0x7F & b)) << shift
		if b&0x80 == 0 {
			break
		}
		shift += 7
		i++
	}
	return result
}

// FromBigInt encodes the signed big integer n in two's complement,
// LEB128-encodes it, and returns the encoded bytes.
func FromBigInt(n *big.Int) (out []byte) {
	size := n.BitLen()
	negative := n.Sign() < 0
	if negative {
		// big.Int stores integers using sign and magnitude. Returns a copy
		// as the code below is destructive.
		n = twosComplementBigInt(n)
	} else {
		// The code below is destructive so make a copy.
		n = big.NewInt(0).Set(n)
	}

	more := true
	for more {
		bBigInt := big.NewInt(0)
		n.DivMod(n, big.NewInt(128), bBigInt) // This does the mask and the shift.
		b := uint8(bBigInt.Int64())

		// We just logically right-shifted the bits of n so we need to sign extend
		// if n is negative (this simulates an arithmetic shift).
		if negative {
			signExtend(n, size)
		}

		if (n.Sign() == 0 && b&0x40 == 0) ||
			(negative && equalsNegativeOne(n, size) && b&0x40 > 0) {
			more = false
		} else {
			b = b | 0x80
		}
		out = append(out, b)
	}
	return
}

// ToBigInt decodes the signed big integer found in the given bytes.
func ToBigInt(encoded []byte) *big.Int {
	result := big.NewInt(0)
	var shift, i int
	var b byte
	size := len(encoded) * 8

	for {
		b = encoded[i]
		for bitPos := uint(0); bitPos < 7; bitPos++ {
			result.SetBit(result, 7*i+int(bitPos), uint((b>>bitPos)&0x01))
		}
		shift += 7
		if b&0x80 == 0 {
			break
		}
		i++
	}

	if b&0x40 > 0 {
		// Sign extend.
		for ; shift < size; shift++ {
			result.SetBit(result, shift, 1)
		}
		result = twosComplementBigInt(result)
		result.Neg(result)
	}
	return result
}

func twosComplementBigInt(n *big.Int) *big.Int {
	absValBytes := n.Bytes()
	for i, b := range absValBytes {
		absValBytes[i] = ^b
	}
	bitsFlipped := big.NewInt(0).SetBytes(absValBytes)
	return bitsFlipped.Add(bitsFlipped, big.NewInt(1))
}

func signExtend(n *big.Int, size int) {
	bitPos := size - 7
	max := size
	if bitPos < 0 {
		bitPos = 0
		max = 7
	}
	for ; bitPos < max; bitPos++ {
		n.SetBit(n, bitPos, 1)
	}
}

// equalsNegativeOne is a poor man's check that n, which
// is encoded in two's complement, is all 1's.
func equalsNegativeOne(n *big.Int, size int) bool {
	for i := 0; i < size; i++ {
		if !(n.Bit(i) == 1) {
			return false
		}
	}
	return true
}
