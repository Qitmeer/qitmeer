// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: syncqnr.proto

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

type SyncQNR struct {
	Qnr                  []byte   `protobuf:"bytes,1,opt,name=qnr,proto3" json:"qnr,omitempty" ssz-max:"300"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SyncQNR) Reset()         { *m = SyncQNR{} }
func (m *SyncQNR) String() string { return proto.CompactTextString(m) }
func (*SyncQNR) ProtoMessage()    {}
func (*SyncQNR) Descriptor() ([]byte, []int) {
	return fileDescriptor_eb432e0b84d45ff6, []int{0}
}
func (m *SyncQNR) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SyncQNR) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SyncQNR.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *SyncQNR) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SyncQNR.Merge(m, src)
}
func (m *SyncQNR) XXX_Size() int {
	return m.Size()
}
func (m *SyncQNR) XXX_DiscardUnknown() {
	xxx_messageInfo_SyncQNR.DiscardUnknown(m)
}

var xxx_messageInfo_SyncQNR proto.InternalMessageInfo

func (m *SyncQNR) GetQnr() []byte {
	if m != nil {
		return m.Qnr
	}
	return nil
}

func init() {
	proto.RegisterType((*SyncQNR)(nil), "qitmeer.p2p.v1.SyncQNR")
}

func init() { proto.RegisterFile("syncqnr.proto", fileDescriptor_eb432e0b84d45ff6) }

var fileDescriptor_eb432e0b84d45ff6 = []byte{
	// 156 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2d, 0xae, 0xcc, 0x4b,
	0x2e, 0xcc, 0x2b, 0xd2, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x2b, 0xcc, 0x2c, 0xc9, 0x4d,
	0x4d, 0x2d, 0xd2, 0x2b, 0x30, 0x2a, 0xd0, 0x2b, 0x33, 0x94, 0xd2, 0x4d, 0xcf, 0x2c, 0xc9, 0x28,
	0x4d, 0xd2, 0x4b, 0xce, 0xcf, 0xd5, 0x4f, 0xcf, 0x4f, 0xcf, 0xd7, 0x07, 0x2b, 0x4b, 0x2a, 0x4d,
	0x03, 0xf3, 0xc0, 0x1c, 0x30, 0x0b, 0xa2, 0x5d, 0x49, 0x8f, 0x8b, 0x3d, 0xb8, 0x32, 0x2f, 0x39,
	0xd0, 0x2f, 0x48, 0x48, 0x99, 0x8b, 0xb9, 0x30, 0xaf, 0x48, 0x82, 0x51, 0x81, 0x51, 0x83, 0xc7,
	0x49, 0xf0, 0xd3, 0x3d, 0x79, 0xde, 0xe2, 0xe2, 0x2a, 0xdd, 0xdc, 0xc4, 0x0a, 0x2b, 0x25, 0x63,
	0x03, 0x03, 0xa5, 0x20, 0x90, 0xac, 0x93, 0xc0, 0x89, 0x47, 0x72, 0x8c, 0x17, 0x1e, 0xc9, 0x31,
	0x3e, 0x78, 0x24, 0xc7, 0x38, 0xe3, 0xb1, 0x1c, 0x43, 0x12, 0x1b, 0xd8, 0x20, 0x63, 0x40, 0x00,
	0x00, 0x00, 0xff, 0xff, 0xea, 0x73, 0xfa, 0x0d, 0x98, 0x00, 0x00, 0x00,
}

func (m *SyncQNR) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SyncQNR) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *SyncQNR) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if len(m.Qnr) > 0 {
		i -= len(m.Qnr)
		copy(dAtA[i:], m.Qnr)
		i = encodeVarintSyncqnr(dAtA, i, uint64(len(m.Qnr)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintSyncqnr(dAtA []byte, offset int, v uint64) int {
	offset -= sovSyncqnr(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *SyncQNR) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Qnr)
	if l > 0 {
		n += 1 + l + sovSyncqnr(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovSyncqnr(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozSyncqnr(x uint64) (n int) {
	return sovSyncqnr(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *SyncQNR) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSyncqnr
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
			return fmt.Errorf("proto: SyncQNR: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SyncQNR: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Qnr", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSyncqnr
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
				return ErrInvalidLengthSyncqnr
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthSyncqnr
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Qnr = append(m.Qnr[:0], dAtA[iNdEx:postIndex]...)
			if m.Qnr == nil {
				m.Qnr = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipSyncqnr(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthSyncqnr
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthSyncqnr
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
func skipSyncqnr(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowSyncqnr
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
					return 0, ErrIntOverflowSyncqnr
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
					return 0, ErrIntOverflowSyncqnr
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
				return 0, ErrInvalidLengthSyncqnr
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupSyncqnr
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthSyncqnr
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthSyncqnr        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowSyncqnr          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupSyncqnr = fmt.Errorf("proto: unexpected end of group")
)