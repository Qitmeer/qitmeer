// Copyright 2017-2018 The qitmeer developers

package common

import (
	"fmt"
	"math/big"
)

var (
	Big0 = big.NewInt(0)

	// bigOne is 1 represented as a big.Int.  It is defined here to avoid
	// the overhead of creating it multiple times.
	Big1 = big.NewInt(1)

	Big2   = big.NewInt(2)
	Big256 = big.NewInt(0xff)

	tt256     = new(big.Int).Lsh(big.NewInt(1), 256)   //2^256
	tt256m1   = new(big.Int).Sub(tt256, big.NewInt(1)) //2^256-1
	MaxBig256 = new(big.Int).Set(tt256m1)

	// to remove
	tt255    = new(big.Int).Lsh(big.NewInt(1), 255)
	tt255_   = BigPow(2, 255)
	tt256_   = BigPow(2, 256)
	tt256m1_ = new(big.Int).Sub(tt256, big.NewInt(1))
	tt63     = BigPow(2, 63)
	MaxBig63 = new(big.Int).Sub(tt63, big.NewInt(1))
)

const (
	// number of bits in a big.Word
	wordBits = 32 << (uint64(^big.Word(0)) >> 63)
	// number of bytes in a big.Word
	wordBytes = wordBits / 8
)

// BigPow returns a ** b as a big integer.
func BigPow(a, b int) *big.Int {
	r := big.NewInt(int64(a))
	return r.Exp(r, big.NewInt(int64(b)), nil)
}

// Compute x^n by using the binary powering algorithm (aka. the repeated square-and-multiply algorithm)
// Reference: Knuth's The Art of Computer Programming, Volume 2, The Seminumerical Algorithms
// 4.6.3. Evaluation of Powers
// Suppose, for example, that we need to compute x^16; we could simply start
// with x and multiply by x fifteen times. But it is possible to obtain the
// same answer with only four multiplications, if we repeatedly take the square
// of each partial result, successively forming x^2, x^4, x^8, x^16.
// The same idea applies, in general, to any value of n, in the following way:
// Write n in the binary number system (suppressing zeros at the left). Then replace each “1”
// by the pair of letters SX, replace each “0” by S, and cross off the “SX” that now appears
// at the left. The result is a rule for computing x^n, if “S” is interpreted as the operation
// of squaring, and if “X” is interpreted as the operation of multiplying by x.
// For example, if n = 23, its binary representation is 10111; so we form the sequence SX S SX SX SX
// and remove the leading SX to obtain the rule SSXSXSX. This rule states that we should “square,
// square, multiply by x, square, multiply by x, square, and multiply by x”;
//
// TODO : use Montgomery's ladder against side-channel attack
// Reference https://en.wikipedia.org/wiki/Exponentiation_by_squaring
// https://en.wikipedia.org/wiki/Exponentiation_by_squaring#Montgomery's_ladder_technique
// Montgomery, Peter L. (1987). "Speeding the Pollard and Elliptic Curve Methods of Factorization"
//
func Pow(x, n int) (uint64, error) {
	input_x := x
	input_n := n
	result := uint64(1)
	for n != 0 {
		if n&1 != 0 {
			//tmp := result*uint64(x)
			if x != 0 && result < ((1<<64)-1)/uint64(x) {
				result *= uint64(x) //odd, multiple
			} else {
				if x == 0 {
					const MaxUint = 18446744073709551616
					return 0, fmt.Errorf("Pow(%d,%d) overfollow to do %v * %v", input_x, input_n, result, 0)
				}
				return 0, fmt.Errorf("Pow(%d,%d) overfollow to do %d * %d", input_x, input_n, result, x)
			}
		}
		n >>= 1 //halve n
		x *= x  //even, square
	}
	return result, nil
}

// only for showing algorithm
// compare to use BigPow
// 1000000	      1621 ns/op   BigPow
// 300000	      4200 ns/op   PowBig
func PowBig(x, n int) *big.Int {
	tmp := big.NewInt(int64(x))
	res := Big1
	for n != 0 {
		temp := new(big.Int)
		if n&1 != 0 {

			temp.Mul(res, tmp)
			res = temp
		}
		n >>= 1
		temp = new(big.Int)
		temp.Mul(tmp, tmp)
		tmp = temp
	}
	return res
}

/*
func Pow(a, b int) int {
	result := 1
	for b > 0 {
		if b&1 != 0 {
			result *= a
		}
		b >>= 1
		a *= a
	}
	return result
}
*/

// Compute x^n mod m by using the binary powering algorithm
// panic when m == 0
func PowMod(x, n, m int) int {
	result := 1 % m
	x = x % m
	for n != 0 {
		if n&1 != 0 {
			result = (result * x) % m
		}
		n >>= 1
		x = (x * x) % m
	}
	return result
}

// ReadBits encodes the absolute value of bigint as big-endian bytes. Callers must ensure
// that buf has enough space. If buf is too short the result will be incomplete.
func ReadBits(bigint *big.Int, buf []byte) {
	i := len(buf)
	for _, d := range bigint.Bits() {
		for j := 0; j < wordBytes && i > 0; j++ {
			i--
			buf[i] = byte(d)
			d >>= 8
		}
	}
}
