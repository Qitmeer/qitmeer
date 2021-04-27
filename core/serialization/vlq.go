/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package serialization

// currentCompressionVersion is the current script compression version of the
// database.
const CurrentCompressionVersion = 1

// -----------------------------------------------------------------------------
// A variable length quantity (VLQ) is an encoding that uses an arbitrary number
// of binary octets to represent an arbitrarily large integer.  The scheme
// employs a most significant byte (MSB) base-128 encoding where the high bit in
// each byte indicates whether or not the byte is the final one.  In addition,
// to ensure there are no redundant encodings, an offset is subtracted every
// time a group of 7 bits is shifted out.  Therefore each integer can be
// represented in exactly one way, and each representation stands for exactly
// one integer.
//
// Another nice property of this encoding is that it provides a compact
// representation of values that are typically used to indicate sizes.  For
// example, the values 0 - 127 are represented with a single byte, 128 - 16511
// with two bytes, and 16512 - 2113663 with three bytes.
//
// While the encoding allows arbitrarily large integers, it is artificially
// limited in this code to an unsigned 64-bit integer for efficiency purposes.
//
// Example encodings:
//           0 -> [0x00]
//         127 -> [0x7f]                 * Max 1-byte value
//         128 -> [0x80 0x00]
//         129 -> [0x80 0x01]
//         255 -> [0x80 0x7f]
//         256 -> [0x81 0x00]
//       16511 -> [0xff 0x7f]            * Max 2-byte value
//       16512 -> [0x80 0x80 0x00]
//       32895 -> [0x80 0xff 0x7f]
//     2113663 -> [0xff 0xff 0x7f]       * Max 3-byte value
//   270549119 -> [0xff 0xff 0xff 0x7f]  * Max 4-byte value
//      2^64-1 -> [0x80 0xfe 0xfe 0xfe 0xfe 0xfe 0xfe 0xfe 0xfe 0x7f]
//
// References:
//   https://en.wikipedia.org/wiki/Variable-length_quantity
//   http://www.codecodex.com/wiki/Variable-Length_Integers
// -----------------------------------------------------------------------------

// SerializeSizeVLQ returns the number of bytes it would take to serialize the
// passed number as a variable-length quantity according to the format described
// above.
func SerializeSizeVLQ(n uint64) int {
	size := 1
	for ; n > 0x7f; n = (n >> 7) - 1 {
		size++
	}

	return size
}

// PutVLQ serializes the provided number to a variable-length quantity according
// to the format described above and returns the number of bytes of the encoded
// value.  The result is placed directly into the passed byte slice which must
// be at least large enough to handle the number of bytes returned by the
// serializeSizeVLQ function or it will panic.
func PutVLQ(target []byte, n uint64) int {
	offset := 0
	for ; ; offset++ {
		// The high bit is set when another byte follows.
		highBitMask := byte(0x80)
		if offset == 0 {
			highBitMask = 0x00
		}

		target[offset] = byte(n&0x7f) | highBitMask
		if n <= 0x7f {
			break
		}
		n = (n >> 7) - 1
	}

	// Reverse the bytes so it is MSB-encoded.
	for i, j := 0, offset; i < j; i, j = i+1, j-1 {
		target[i], target[j] = target[j], target[i]
	}

	return offset + 1
}

// DeserializeVLQ deserializes the provided variable-length quantity according
// to the format described above.  It also returns the number of bytes
// deserialized.
func DeserializeVLQ(serialized []byte) (uint64, int) {
	var n uint64
	var size int
	for _, val := range serialized {
		size++
		n = (n << 7) | uint64(val&0x7f)
		if val&0x80 != 0x80 {
			break
		}
		n++
	}

	return n, size
}

// maxUint32VLQSerializeSize is the maximum number of bytes a max uint32 takes
// to serialize as a VLQ.
var MaxUint32VLQSerializeSize = SerializeSizeVLQ(1<<32 - 1)
