package main

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
)

type DebugAddrInfo struct {
	id        int
	addresses string
	amount    int64
}

func decodeDebugAddrInfo(data []byte) ([]DebugAddrInfo, error) {
	dataLen := len(data)
	if dataLen <= 0 {
		return nil, fmt.Errorf("no DebugAddrInfo data")
	}
	result := []DebugAddrInfo{}

	offset := 0

	total := dbnamespace.ByteOrder.Uint32(data[offset : offset+4])
	offset += 4

	for i := uint32(0); i < total; {
		if (offset + 4 + 35 + 8) > dataLen {
			return nil, fmt.Errorf("data length not match")
		}
		id := dbnamespace.ByteOrder.Uint32(data[offset : offset+4])
		offset += 4

		address := [35]byte{}
		copy(address[:], data[offset:offset+35])
		offset += 35
		amount := dbnamespace.ByteOrder.Uint64(data[offset : offset+8])
		offset += 8

		di := DebugAddrInfo{}
		di.id = int(id)
		di.addresses = string(address[:])
		di.amount = int64(amount)

		result = append(result, di)
	}

	return result, nil
}
