// Copyright 2017-2018 The qitmeer developers
// Copyright 2015 The Decred developers
// Copyright 2013, 2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package types

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strconv"
)

// AmountUnit describes a method of converting an Amount to something
// other than the base unit of a coin.  The value of the AmountUnit
// is the exponent component of the decadic multiple to convert from
// an amount in coins to an amount counted in atomic units.
type AmountUnit int

// These constants define various units used when describing a coin
// monetary amount.
const (
	AmountMegaCoin  AmountUnit = 6
	AmountKiloCoin  AmountUnit = 3
	AmountCoin      AmountUnit = 0
	AmountMilliCoin AmountUnit = -3
	AmountMicroCoin AmountUnit = -6
	AmountAtom      AmountUnit = -8
)

// String returns the unit as a string.  For recognized units, the SI
// prefix is used, or "Atom" for the base unit.  For all unrecognized
// units, "1eN MEER" is returned, where N is the AmountUnit.
func (u AmountUnit) String() string {
	switch u {
	case AmountMegaCoin:
		return "M"
	case AmountKiloCoin:
		return "k"
	case AmountCoin:
		return ""
	case AmountMilliCoin:
		return "m"
	case AmountMicroCoin:
		return "μ"
	case AmountAtom:
		return "atom"
	default:
		return "1e" + strconv.FormatInt(int64(u), 10) + " "
	}
}

// from 0 ~ 65535
// 0 ~ 255 : Qitmeer reserved
type CoinID uint16

const (
	MEERID CoinID = 0
	QITID  CoinID = 1
)

func (c CoinID) Name() string {
	switch c {
	case MEERID:
		return "MEER"
	case QITID:
		return "QIT"
	default:
		return "Unknown-CoinID:" + strconv.FormatInt(int64(c), 10)
	}
}
func (c CoinID) Bytes() []byte {
	b := [2]byte{}
	binary.LittleEndian.PutUint16(b[:],uint16(c))
	return b[:]
}
var CoinIDList = []CoinID{
	MEERID,QITID,
}

// Check if a valid coinId, current only check if the coinId is known.
func CheckCoinID(id CoinID) error{
	unknownCoin := true
	for _, coinId := range CoinIDList {
		if id == coinId {
			unknownCoin = false
			break
		}
	}
	if unknownCoin {
		return fmt.Errorf("unknown coin id %s", id.Name())
	}
	return nil
}


// Amount represents the base coin monetary unit (colloquially referred
// to as an `Atom').  A single Amount is equal to 1e-8 of a coin.
// size is 10 bytes ( value -> 8 byte , coinId -> 2 byte)
type Amount struct {
	Value int64
	Id    CoinID
}

func checkMaxAmount(x *Amount) error {
	if x.Value > MaxAmount {
		return errors.New("x+y, x exceeds max amount of " + x.Id.Name())
	}
	return nil
}
func checkSameCoinId(x, y CoinID) error {
	if x != y {
		err := errors.New("invalid amount add, unmatched " +
			x.Name() + " with " + y.Name())
		return err
	}
	return nil
}

func (z *Amount) Add(x, y *Amount) (*Amount, error) {
	err := checkSameCoinId(z.Id, x.Id)
	if err != nil {
		return z, err
	}
	err = checkSameCoinId(x.Id, y.Id)
	if err != nil {
		return z, err
	}
	err = checkMaxAmount(x)
	if err != nil {
		return z, err
	}
	err = checkMaxAmount(y)
	if err != nil {
		return z, err
	}
	sum := x.Value + y.Value
	if x.Value > 0 && y.Value > 0 && sum < 0 {
		err := errors.New("add overflow")
		return z, err
	}
	if x.Value < 0 && y.Value < 0 && sum > 0 {
		err := errors.New("add overflow")
		return z, err
	}
	z.Value = sum
	return z, nil
}

// AmountGroup represents a group of multiple Amount,
type AmountGroup []Amount

// round converts a floating point number, which may or may not be representable
// as an integer, to the Amount integer type by rounding to the nearest integer.
// This is performed by adding or subtracting 0.5 depending on the sign, and
// relying on integer truncation to round the value to the nearest Amount.
func round(f float64) int64 {
	if f < 0 {
		return int64(f - 0.5)
	}
	return int64(f + 0.5)
}
// NewAmount creates an Amount from a floating point value representing
// some value in the currency.  NewAmount errors if f is NaN or +-Infinity,
// but does not check that the amount is within the total amount of coins
// producible as f may not refer to an amount at a single moment in time.
//
// NewAmount is for specifically for converting qitmeer to Atoms (atomic units).
// For creating a new Amount with an int64 value which denotes a quantity of
// Atoms, do a simple type conversion from type int64 to Amount.
func NewAmount(f float64) (*Amount, error) {
	// The amount is only considered invalid if it cannot be represented
	// as an integer type.  This may happen if f is NaN or +-Infinity.
	switch {
	case math.IsNaN(f):
		fallthrough
	case math.IsInf(f, 1):
		fallthrough
	case math.IsInf(f, -1):
		return &Amount{0, MEERID}, errors.New("invalid coin amount")
	}

	return &Amount{round(f * AtomsPerCoin), MEERID}, nil
}

func NewMeer(a uint64) (*Amount, error) {
	amt := Amount{int64(a), MEERID}
	err := checkMaxAmount(&amt)
	if err != nil {
		zero := Amount{0, MEERID}
		return &zero, err
	}
	return &amt, nil
}

func NewQit(a uint64) (*Amount, error) {
	amt := Amount{int64(a), QITID}
	err := checkMaxAmount(&amt)
	if err != nil {
		zero := Amount{0, QITID}
		return &zero, err
	}
	return &amt, nil
}

// ToUnit converts a monetary amount counted in coin base units to a
// floating point value representing an amount of coins.
func (a *Amount) ToUnit(u AmountUnit) float64 {
	return float64(a.Value) / math.Pow10(int(u+8))
}

// ToCoin is the equivalent of calling ToUnit with AmountCoin.
func (a *Amount) ToCoin() float64 {
	return a.ToUnit(AmountCoin)
}

// Format formats a monetary amount counted in coin base units as a
// string for a given unit.  The conversion will succeed for any unit,
// however, known units will be formated with an appended label describing
// the units with SI notation, or "atom" for the base unit.
func (a *Amount) Format(u AmountUnit) string {
	units := " " + u.String() + a.Id.Name()
	return strconv.FormatFloat(a.ToUnit(u), 'f', -int(u+8), 64) + units
}

// String is the equivalent of calling Format with AmountCoin.
func (a *Amount) String() string {
	return a.Format(AmountCoin)
}

// MulF64 multiplies an Amount by a floating point value.  While this is not
// an operation that must typically be done by a full node or wallet, it is
// useful for services that build on top of qitmeer (for example, calculating
// a fee by multiplying by a percentage).
func (a *Amount) MulF64(f float64) *Amount {
	return &Amount{round(float64(a.Value) * f), a.Id}
}

// AmountSorter implements sort.Interface to allow a slice of Amounts to
// be sorted.
type AmountSorter []Amount

// Len returns the number of Amounts in the slice.  It is part of the
// sort.Interface implementation.
func (s AmountSorter) Len() int {
	return len(s)
}

// Swap swaps the Amounts at the passed indices.  It is part of the
// sort.Interface implementation.
func (s AmountSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less returns whether the Amount with index i should sort before the
// Amount with index j.  It is part of the sort.Interface
// implementation.
func (s AmountSorter) Less(i, j int) bool {
	return s[i].Value < s[j].Value
}
