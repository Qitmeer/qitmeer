// Copyright (c) 2017-2020 The qitmeer developers
// license that can be found in the LICENSE file.
// Reference resources of rust bitVector
package pow

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/log"
	"math/big"
)

var (
	// bigOne is 1 represented as a big.Int.  It is defined here to avoid
	// the overhead of creating it multiple times.
	bigOne = big.NewInt(1)

	// oneLsh256 is 1 shifted left 256 bits.  It is defined here to avoid
	// the overhead of creating it multiple times.
	OneLsh256 = new(big.Int).Lsh(bigOne, 256)
)

// HashToBig converts a hash.Hash into a big.Int that can be used to
// perform math comparisons.
func HashToBig(hash *hash.Hash) *big.Int {
	// A Hash is in little-endian, but the big package wants the bytes in
	// big-endian, so reverse them.
	buf := *hash
	blen := len(buf)
	for i := 0; i < blen/2; i++ {
		buf[i], buf[blen-1-i] = buf[blen-1-i], buf[i]
	}

	return new(big.Int).SetBytes(buf[:])
}

// CompactToBig converts a compact representation of a whole number N to an
// unsigned 32-bit number.  The representation is similar to IEEE754 floating
// point numbers.
//
// Like IEEE754 floating point, there are three basic components: the sign,
// the exponent, and the mantissa.  They are broken out as follows:
//
//	* the most significant 8 bits represent the unsigned base 256 exponent
// 	* bit 23 (the 24th bit) represents the sign bit
//	* the least significant 23 bits represent the mantissa
//
//	-------------------------------------------------
//	|   Exponent     |    Sign    |    Mantissa     |
//	-------------------------------------------------
//	| 8 bits [31-24] | 1 bit [23] | 23 bits [22-00] |
//	-------------------------------------------------
//
// The formula to calculate N is:
// 	N = (-1^sign) * mantissa * 256^(exponent-3)
//
// This compact form is only used to encode unsigned 256-bit numbers which
// represent difficulty targets, thus there really is not a need for a sign
// bit, but it is implemented here to stay consistent with bitcoind.
// TODO, revisit the compact difficulty form design
func CompactToBig(compact uint32) *big.Int {
	// Extract the mantissa, sign bit, and exponent.
	mantissa := compact & 0x007fffff
	isNegative := compact&0x00800000 != 0
	exponent := uint(compact >> 24)

	// Since the base for the exponent is 256, the exponent can be treated
	// as the number of bytes to represent the full 256-bit number.  So,
	// treat the exponent as the number of bytes and shift the mantissa
	// right or left accordingly.  This is equivalent to:
	// N = mantissa * 256^(exponent-3)
	var bn *big.Int
	if exponent <= 3 {
		mantissa >>= 8 * (3 - exponent)
		bn = big.NewInt(int64(mantissa))
	} else {
		bn = big.NewInt(int64(mantissa))
		bn.Lsh(bn, 8*(exponent-3))
	}

	// Make it negative if the sign bit is set.
	if isNegative {
		bn = bn.Neg(bn)
	}

	return bn
}

// BigToCompact converts a whole number N to a compact representation using
// an unsigned 32-bit number.  The compact representation only provides 23 bits
// of precision, so values larger than (2^23 - 1) only encode the most
// significant digits of the number.  See CompactToBig for details.
func BigToCompact(n *big.Int) uint32 {
	// No need to do any work if it's zero.
	if n.Sign() == 0 {
		return 0
	}

	// Since the base for the exponent is 256, the exponent can be treated
	// as the number of bytes.  So, shift the number right or left
	// accordingly.  This is equivalent to:
	// mantissa = mantissa / 256^(exponent-3)
	var mantissa uint32
	exponent := uint(len(n.Bytes()))
	if exponent <= 3 {
		mantissa = uint32(n.Bits()[0])
		mantissa <<= 8 * (3 - exponent)
	} else {
		// Use a copy to avoid modifying the caller's original number.
		tn := new(big.Int).Set(n)
		mantissa = uint32(tn.Rsh(tn, 8*(exponent-3)).Bits()[0])
	}

	// When the mantissa already has the sign bit set, the number is too
	// large to fit into the available 23-bits, so divide the number by 256
	// and increment the exponent accordingly.
	if mantissa&0x00800000 != 0 {
		mantissa >>= 8
		exponent++
	}

	// Pack the exponent, sign bit, and mantissa into an unsigned 32-bit
	// int and return it.
	compact := uint32(exponent<<24) | mantissa
	if n.Sign() < 0 {
		compact |= 0x00800000
	}
	return compact
}

// CalcWork calculates a work value from difficulty bits. it increases the difficulty
// for generating a block by decreasing the value which the generated hash must be
// less than.
//
// This difficulty target is stored in each block header using a compact
// representation as described in the documentation for CompactToBig.
//
// The main chain is selected by choosing the chain that has the most proof of
// work (highest difficulty).
//
// Since a lower target difficulty value equates to higher actual difficulty, the
// work value which will be accumulated must be the inverse of the difficulty.
// Also, in order to avoid potential division by zero and really small floating
// point numbers, the result adds 1 to the denominator and multiplies the numerator
// by 2^256.
func CalcWork(bits uint32, powType PowType) *big.Int {
	// Return a work value of zero if the passed difficulty bits represent
	// a negative number. Note this should not happen in practice with valid
	// blocks, but an invalid block could trigger it.
	difficultyNum := CompactToBig(bits)
	if difficultyNum.Sign() <= 0 {
		return big.NewInt(0)
	}

	if powType != BLAKE2BD {
		//cuckoo work sum
		allDiff := big.NewInt(1)
		allDiff = allDiff.Lsh(allDiff, 64)
		allDiff = allDiff.Mul(allDiff, big.NewInt(int64(1865)))
		return allDiff.Div(allDiff, difficultyNum)
	}

	// (1 << 256) / (difficultyNum + 1)
	denominator := new(big.Int).Add(difficultyNum, bigOne)
	return new(big.Int).Div(OneLsh256, denominator)
}

// mergeDifficulty takes an original stake difficulty and two new, scaled
// stake difficulties, merges the new difficulties, and outputs a new
// merged stake difficulty.
func mergeDifficulty(oldDiff int64, newDiff1 int64, newDiff2 int64) int64 {
	newDiff1Big := big.NewInt(newDiff1)
	newDiff2Big := big.NewInt(newDiff2)
	newDiff2Big.Lsh(newDiff2Big, 32)

	oldDiffBig := big.NewInt(oldDiff)
	oldDiffBigLSH := big.NewInt(oldDiff)
	oldDiffBigLSH.Lsh(oldDiffBig, 32)

	newDiff1Big.Div(oldDiffBigLSH, newDiff1Big)
	newDiff2Big.Div(newDiff2Big, oldDiffBig)

	// Combine the two changes in difficulty.
	summedChange := big.NewInt(0)
	summedChange.Set(newDiff2Big)
	summedChange.Lsh(summedChange, 32)
	summedChange.Div(summedChange, newDiff1Big)
	summedChange.Mul(summedChange, oldDiffBig)
	summedChange.Rsh(summedChange, 32)

	return summedChange.Int64()
}

//calc cuckoo diff
func CalcCuckooDiff(scale uint64, blockHash hash.Hash) *big.Int {
	c := HashToBig(&blockHash)
	a := &big.Int{}
	a.SetUint64(scale)
	d := big.NewInt(1)
	d.Lsh(d, 256)
	a.Mul(a, d)
	e := a.Div(a, c)
	log.Debug(fmt.Sprintf("solution difficulty:%d", e.Uint64()))
	return e
}

//calc cuckoo diff convert to target hash like 7fff000000000000000000000000000000000000000000000000000000000000
func CuckooDiffToTarget(scale uint64, diff *big.Int) string {
	a := &big.Int{}
	a.SetUint64(scale)
	d := big.NewInt(1)
	d.Lsh(d, 256)
	a.Mul(a, d)
	a.Div(a, diff)
	b := a.Bytes()
	c := make([]byte,32)
	copy(c[:],b)
	return fmt.Sprintf("%x",c)
}

//calc scale
//the edge_bits is bigger ,then scale is bigger
//Reference resources https://eprint.iacr.org/2014/059.pdf 9. Difficulty control page 6
//while the average number of cycles found increases slowly with size; from 2 at 2^20 to 3 at 2^30
//Less times of hash calculation with the same difficulty
// 24 => 48 25 => 100 26 => 208 27 => 432 28 => 896 29 => 1856 30 => 3840 31 => 7936
//assume init difficulty is 1000
//24 target is 0c49ba5e353f7ced000000000000000000000000000000000000000000000000
//（The meaning of difficulty needs to be found 1000/48 * 50 ≈ 1000 times in edge_bits 24, and the answer may be obtained once.）
// why * 50 , because the when edge_count/nodes = 1/2,to find 42 cycles the probality is 2.2%
//29 target is db22d0e560418937000000000000000000000000000000000000000000000000
//（The difficulty needs to be found 1000/1856 * 50 ≈ 26 times in edge_bits 29, and the answer may be obtained once.）
//so In order to ensure the fairness of different edge indexes, the mining difficulty is different.
func GraphWeight(edge_bits uint32) uint64 {
	return (2 << (edge_bits - MIN_CUCKAROOEDGEBITS)) * uint64(edge_bits)
}
