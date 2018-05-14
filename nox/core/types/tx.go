// Copyright 2017-2018 The nox developers

package types

// The abstract of Tx, serializable
type Tx []byte

// It's generally a payment transaction, but also can be abstract as
// the instructions for state-transit
type Transaction struct {
	// sender
	from AccountId

	// receiver
	to AccountId

	// how much to send
	Amount uint64

	// How many transactions sender already sent.
	Nonce uint64
}


