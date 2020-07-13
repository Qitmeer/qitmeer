package p2p

import (
	"github.com/prysmaticlabs/go-bitfield"
)

type MetaData struct {
	SeqNumber uint64
	Attnets   bitfield.Bitvector64
}

func (m *MetaData) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MetaData) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MetaData) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if len(m.Attnets) > 0 {
		i -= len(m.Attnets)
		copy(dAtA[i:], m.Attnets)
		i = encodeVarintMessages(dAtA, i, uint64(len(m.Attnets)))
		i--
		dAtA[i] = 0x12
	}
	if m.SeqNumber != 0 {
		i = encodeVarintMessages(dAtA, i, uint64(m.SeqNumber))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}
