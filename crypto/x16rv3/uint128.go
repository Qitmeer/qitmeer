package x16rv3

import (
	"encoding/binary"
	"encoding/hex"

	"github.com/pkg/errors"
)

// Uint128 is a big-endian 128 bit unsigned integer which wraps two uint64s.
type Uint128 struct {
	V0, V1 uint64
}

// GetBytes returns a big-endian byte representation.
func (u Uint128) GetBytes() []byte {
	buf := make([]byte, 16)
	binary.LittleEndian.PutUint64(buf[:8], u.V0)
	binary.LittleEndian.PutUint64(buf[8:], u.V1)
	return buf
}

// String returns a hexadecimal string representation.
func (u Uint128) String() string {
	return hex.EncodeToString(u.GetBytes())
}

// Equal returns whether or not the Uint128 are equivalent.
func (u Uint128) Equal(o Uint128) bool {
	return u.V0 == o.V0 && u.V1 == o.V1
}

// Compare compares the two Uint128.
func (u Uint128) Compare(o Uint128) int {
	if u.V0 > o.V0 {
		return 1
	} else if u.V0 < o.V0 {
		return -1
	} else if u.V1 > o.V1 {
		return 1
	} else if u.V1 < o.V1 {
		return -1
	}
	return 0
}

// Add returns a new Uint128 incremented by n.
func (u Uint128) Add(n uint64) Uint128 {
	V1 := u.V1 + n
	hi := u.V0
	if u.V1 > V1 {
		hi++
	}
	return Uint128{hi, V1}
}

// Sub returns a new Uint128 decremented by n.
func (u Uint128) Sub(n uint64) Uint128 {
	V1 := u.V1 - n
	hi := u.V0
	if u.V1 < V1 {
		hi--
	}
	return Uint128{hi, V1}
}

// And returns a new Uint128 that is the bitwise AND of two Uint128 values.
func (u Uint128) And(o Uint128) Uint128 {
	return Uint128{u.V0 & o.V0, u.V1 & o.V1}
}

// Or returns a new Uint128 that is the bitwise OR of two Uint128 values.
func (u Uint128) Or(o Uint128) Uint128 {
	return Uint128{u.V0 | o.V0, u.V1 | o.V1}
}

// Xor returns a new Uint128 that is the bitwise XOR of two Uint128 values.
func (u Uint128) Xor(o Uint128) Uint128 {
	return Uint128{u.V0 ^ o.V0, u.V1 ^ o.V1}
}

// FromBytes parses the byte slice as a 128 bit big-endian unsigned integer.
// The caller is responsible for ensuring the byte slice contains 16 bytes.
func FromBytes(b []byte) Uint128 {
	hi := binary.LittleEndian.Uint64(b[:8])
	V1 := binary.LittleEndian.Uint64(b[8:])
	return Uint128{hi, V1}
}

// FromString parses a hexadecimal string as a 128-bit big-endian unsigned integer.
func FromString(s string) (Uint128, error) {
	if len(s) > 32 {
		return Uint128{}, errors.Errorf("input string %s too large for uint128", s)
	}
	bytes, err := hex.DecodeString(s)
	if err != nil {
		return Uint128{}, errors.Wrapf(err, "could not decode %s as hex", s)
	}

	// Grow the byte slice if it's smaller than 16 bytes, by prepending 0s
	if len(bytes) < 16 {
		bytesCopy := make([]byte, 16)
		copy(bytesCopy[(16-len(bytes)):], bytes)
		bytes = bytesCopy
	}

	return FromBytes(bytes), nil
}

// FromInts takes in two unsigned 64-bit integers and constructs a Uint128.
func FromInts(hi uint64, V1 uint64) Uint128 {
	return Uint128{hi, V1}
}

// FromInts takes in two unsigned 64-bit integers and constructs a Uint128.
func FromIntsArray(arr []uint64) Uint128 {
	return Uint128{arr[0], arr[1]}
}

func (u Uint128) ToUint64() []uint64 {
	return []uint64{u.V0, u.V1}
}

func Ur128_5xor(in0, in1, in2, in3, in4 Uint128) Uint128 {
	var out = Uint128{}
	out.V0 = in0.V0 ^ in1.V0 ^ in2.V0 ^ in3.V0 ^ in4.V0
	out.V1 = in0.V1 ^ in1.V1 ^ in2.V1 ^ in3.V1 ^ in4.V1
	return out
}

func Xor128(in0, in1 Uint128) Uint128 {
	var out = Uint128{}
	out.V0 = in0.V0 ^ in1.V0
	out.V1 = in0.V1 ^ in1.V1
	return out
}

func ArrayToBytes(a []Uint128) []byte {
	b := make([]byte, 0)
	for i := 0; i < len(a); i++ {
		b = append(b, a[i].GetBytes()...)
	}
	return b
}
