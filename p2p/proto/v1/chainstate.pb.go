// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: chainstate.proto

package qitmeer_p2p_v1

import (
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/golang/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type ChainState struct {
	GenesisHash          *Hash       `protobuf:"bytes,1,opt,name=genesisHash,proto3" json:"genesisHash,omitempty"`
	ProtocolVersion      uint32      `protobuf:"varint,2,opt,name=protocolVersion,proto3" json:"protocolVersion,omitempty"`
	Timestamp            uint64      `protobuf:"varint,3,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Services             uint64      `protobuf:"varint,4,opt,name=services,proto3" json:"services,omitempty"`
	DisableRelayTx       bool        `protobuf:"varint,5,opt,name=disableRelayTx,proto3" json:"disableRelayTx,omitempty"`
	GraphState           *GraphState `protobuf:"bytes,6,opt,name=graphState,proto3" json:"graphState,omitempty"`
	UserAgent            []byte      `protobuf:"bytes,7,opt,name=userAgent,proto3" json:"userAgent,omitempty" ssz-max:"256"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *ChainState) Reset()         { *m = ChainState{} }
func (m *ChainState) String() string { return proto.CompactTextString(m) }
func (*ChainState) ProtoMessage()    {}
func (*ChainState) Descriptor() ([]byte, []int) {
	return fileDescriptor_42d14f768699ae63, []int{0}
}
func (m *ChainState) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ChainState) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ChainState.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ChainState) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ChainState.Merge(m, src)
}
func (m *ChainState) XXX_Size() int {
	return m.Size()
}
func (m *ChainState) XXX_DiscardUnknown() {
	xxx_messageInfo_ChainState.DiscardUnknown(m)
}

var xxx_messageInfo_ChainState proto.InternalMessageInfo

func (m *ChainState) GetGenesisHash() *Hash {
	if m != nil {
		return m.GenesisHash
	}
	return nil
}

func (m *ChainState) GetProtocolVersion() uint32 {
	if m != nil {
		return m.ProtocolVersion
	}
	return 0
}

func (m *ChainState) GetTimestamp() uint64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *ChainState) GetServices() uint64 {
	if m != nil {
		return m.Services
	}
	return 0
}

func (m *ChainState) GetDisableRelayTx() bool {
	if m != nil {
		return m.DisableRelayTx
	}
	return false
}

func (m *ChainState) GetGraphState() *GraphState {
	if m != nil {
		return m.GraphState
	}
	return nil
}

func (m *ChainState) GetUserAgent() []byte {
	if m != nil {
		return m.UserAgent
	}
	return nil
}

func init() {
	proto.RegisterType((*ChainState)(nil), "qitmeer.p2p.v1.ChainState")
}

func init() { proto.RegisterFile("chainstate.proto", fileDescriptor_42d14f768699ae63) }

var fileDescriptor_42d14f768699ae63 = []byte{
	// 321 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x5c, 0x8f, 0xc1, 0x4e, 0x2a, 0x31,
	0x14, 0x86, 0x6f, 0xb9, 0x5c, 0x2e, 0x14, 0x41, 0x6c, 0x5c, 0x34, 0x13, 0x33, 0x4e, 0x58, 0x98,
	0xd9, 0x30, 0x44, 0x8c, 0x2c, 0xd8, 0x89, 0x0b, 0x5d, 0x57, 0xe3, 0xbe, 0x33, 0x1e, 0x3b, 0x4d,
	0x98, 0xe9, 0x38, 0xa7, 0x10, 0xf4, 0x49, 0x7c, 0x19, 0xf7, 0x2e, 0x7d, 0x02, 0x63, 0xf0, 0x0d,
	0x7c, 0x02, 0x43, 0x35, 0x80, 0xec, 0xfa, 0xff, 0xe7, 0x6b, 0x4f, 0x3f, 0xda, 0x49, 0x52, 0xa9,
	0x73, 0xb4, 0xd2, 0x42, 0x54, 0x94, 0xc6, 0x1a, 0xd6, 0xbe, 0xd7, 0x36, 0x03, 0x28, 0xa3, 0x62,
	0x50, 0x44, 0xb3, 0x63, 0xaf, 0xa7, 0xb4, 0x4d, 0xa7, 0x71, 0x94, 0x98, 0xac, 0xaf, 0x8c, 0x32,
	0x7d, 0x87, 0xc5, 0xd3, 0x3b, 0x97, 0x5c, 0x70, 0xa7, 0xef, 0xeb, 0x5e, 0x3b, 0x03, 0x44, 0xa9,
	0x00, 0x7f, 0x72, 0x47, 0x95, 0xb2, 0x48, 0x37, 0x16, 0x74, 0x9f, 0x2b, 0x94, 0x9e, 0x2f, 0xb7,
	0x5e, 0x2d, 0x4b, 0x36, 0xa4, 0x4d, 0x05, 0x39, 0xa0, 0xc6, 0x4b, 0x89, 0x29, 0x27, 0x01, 0x09,
	0x9b, 0x83, 0xfd, 0xe8, 0xf7, 0x2f, 0xa2, 0xe5, 0x4c, 0x6c, 0x82, 0x2c, 0xa4, 0xbb, 0xee, 0xbd,
	0xc4, 0x4c, 0x6e, 0xa0, 0x44, 0x6d, 0x72, 0x5e, 0x09, 0x48, 0xd8, 0x12, 0xdb, 0x35, 0x3b, 0xa0,
	0x0d, 0xab, 0x33, 0x40, 0x2b, 0xb3, 0x82, 0xff, 0x0d, 0x48, 0x58, 0x15, 0xeb, 0x82, 0x79, 0xb4,
	0x8e, 0x50, 0xce, 0x74, 0x02, 0xc8, 0xab, 0x6e, 0xb8, 0xca, 0xec, 0x88, 0xb6, 0x6f, 0x35, 0xca,
	0x78, 0x02, 0x02, 0x26, 0xf2, 0xe1, 0x7a, 0xce, 0xff, 0x05, 0x24, 0xac, 0x8b, 0xad, 0x96, 0x8d,
	0x28, 0x75, 0x9a, 0xce, 0x88, 0xd7, 0x9c, 0x82, 0xb7, 0xad, 0x70, 0xb1, 0x22, 0xc4, 0x06, 0xcd,
	0xfa, 0xb4, 0x31, 0x45, 0x28, 0xcf, 0x14, 0xe4, 0x96, 0xff, 0x0f, 0x48, 0xb8, 0x33, 0xde, 0xfb,
	0x7c, 0x3b, 0x6c, 0x21, 0x3e, 0xf6, 0x32, 0x39, 0x1f, 0x75, 0x07, 0xa7, 0xc3, 0xae, 0x58, 0x33,
	0xe3, 0xce, 0xcb, 0xc2, 0x27, 0xaf, 0x0b, 0x9f, 0xbc, 0x2f, 0x7c, 0xf2, 0xf4, 0xe1, 0xff, 0x89,
	0x6b, 0xce, 0xf8, 0xe4, 0x2b, 0x00, 0x00, 0xff, 0xff, 0xa7, 0x7c, 0xb7, 0xe1, 0xcd, 0x01, 0x00,
	0x00,
}

func (m *ChainState) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ChainState) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ChainState) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if len(m.UserAgent) > 0 {
		i -= len(m.UserAgent)
		copy(dAtA[i:], m.UserAgent)
		i = encodeVarintChainstate(dAtA, i, uint64(len(m.UserAgent)))
		i--
		dAtA[i] = 0x3a
	}
	if m.GraphState != nil {
		{
			size, err := m.GraphState.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintChainstate(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x32
	}
	if m.DisableRelayTx {
		i--
		if m.DisableRelayTx {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x28
	}
	if m.Services != 0 {
		i = encodeVarintChainstate(dAtA, i, uint64(m.Services))
		i--
		dAtA[i] = 0x20
	}
	if m.Timestamp != 0 {
		i = encodeVarintChainstate(dAtA, i, uint64(m.Timestamp))
		i--
		dAtA[i] = 0x18
	}
	if m.ProtocolVersion != 0 {
		i = encodeVarintChainstate(dAtA, i, uint64(m.ProtocolVersion))
		i--
		dAtA[i] = 0x10
	}
	if m.GenesisHash != nil {
		{
			size, err := m.GenesisHash.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintChainstate(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintChainstate(dAtA []byte, offset int, v uint64) int {
	offset -= sovChainstate(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *ChainState) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.GenesisHash != nil {
		l = m.GenesisHash.Size()
		n += 1 + l + sovChainstate(uint64(l))
	}
	if m.ProtocolVersion != 0 {
		n += 1 + sovChainstate(uint64(m.ProtocolVersion))
	}
	if m.Timestamp != 0 {
		n += 1 + sovChainstate(uint64(m.Timestamp))
	}
	if m.Services != 0 {
		n += 1 + sovChainstate(uint64(m.Services))
	}
	if m.DisableRelayTx {
		n += 2
	}
	if m.GraphState != nil {
		l = m.GraphState.Size()
		n += 1 + l + sovChainstate(uint64(l))
	}
	l = len(m.UserAgent)
	if l > 0 {
		n += 1 + l + sovChainstate(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovChainstate(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozChainstate(x uint64) (n int) {
	return sovChainstate(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *ChainState) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowChainstate
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: ChainState: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ChainState: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field GenesisHash", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowChainstate
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthChainstate
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthChainstate
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.GenesisHash == nil {
				m.GenesisHash = &Hash{}
			}
			if err := m.GenesisHash.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ProtocolVersion", wireType)
			}
			m.ProtocolVersion = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowChainstate
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ProtocolVersion |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Timestamp", wireType)
			}
			m.Timestamp = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowChainstate
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Timestamp |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Services", wireType)
			}
			m.Services = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowChainstate
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Services |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field DisableRelayTx", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowChainstate
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.DisableRelayTx = bool(v != 0)
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field GraphState", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowChainstate
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthChainstate
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthChainstate
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.GraphState == nil {
				m.GraphState = &GraphState{}
			}
			if err := m.GraphState.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field UserAgent", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowChainstate
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthChainstate
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthChainstate
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.UserAgent = append(m.UserAgent[:0], dAtA[iNdEx:postIndex]...)
			if m.UserAgent == nil {
				m.UserAgent = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipChainstate(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthChainstate
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthChainstate
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipChainstate(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowChainstate
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowChainstate
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowChainstate
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthChainstate
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupChainstate
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthChainstate
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthChainstate        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowChainstate          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupChainstate = fmt.Errorf("proto: unexpected end of group")
)
