// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: getblockdatas.proto

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

type GetBlockDatas struct {
	Locator              []*Hash  `protobuf:"bytes,1,rep,name=locator,proto3" json:"locator,omitempty" ssz-max:"2000"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetBlockDatas) Reset()         { *m = GetBlockDatas{} }
func (m *GetBlockDatas) String() string { return proto.CompactTextString(m) }
func (*GetBlockDatas) ProtoMessage()    {}
func (*GetBlockDatas) Descriptor() ([]byte, []int) {
	return fileDescriptor_5c509316b8da1820, []int{0}
}
func (m *GetBlockDatas) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GetBlockDatas) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GetBlockDatas.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GetBlockDatas) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetBlockDatas.Merge(m, src)
}
func (m *GetBlockDatas) XXX_Size() int {
	return m.Size()
}
func (m *GetBlockDatas) XXX_DiscardUnknown() {
	xxx_messageInfo_GetBlockDatas.DiscardUnknown(m)
}

var xxx_messageInfo_GetBlockDatas proto.InternalMessageInfo

func (m *GetBlockDatas) GetLocator() []*Hash {
	if m != nil {
		return m.Locator
	}
	return nil
}

type BlockDatas struct {
	Locator              []*BlockData `protobuf:"bytes,1,rep,name=locator,proto3" json:"locator,omitempty" ssz-max:"2000"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *BlockDatas) Reset()         { *m = BlockDatas{} }
func (m *BlockDatas) String() string { return proto.CompactTextString(m) }
func (*BlockDatas) ProtoMessage()    {}
func (*BlockDatas) Descriptor() ([]byte, []int) {
	return fileDescriptor_5c509316b8da1820, []int{1}
}
func (m *BlockDatas) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *BlockDatas) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_BlockDatas.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *BlockDatas) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BlockDatas.Merge(m, src)
}
func (m *BlockDatas) XXX_Size() int {
	return m.Size()
}
func (m *BlockDatas) XXX_DiscardUnknown() {
	xxx_messageInfo_BlockDatas.DiscardUnknown(m)
}

var xxx_messageInfo_BlockDatas proto.InternalMessageInfo

func (m *BlockDatas) GetLocator() []*BlockData {
	if m != nil {
		return m.Locator
	}
	return nil
}

type BlockData struct {
	BlockBytes           []byte   `protobuf:"bytes,100,opt,name=blockBytes,proto3" json:"blockBytes,omitempty" ssz-max:"1048576"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BlockData) Reset()         { *m = BlockData{} }
func (m *BlockData) String() string { return proto.CompactTextString(m) }
func (*BlockData) ProtoMessage()    {}
func (*BlockData) Descriptor() ([]byte, []int) {
	return fileDescriptor_5c509316b8da1820, []int{2}
}
func (m *BlockData) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *BlockData) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_BlockData.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *BlockData) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BlockData.Merge(m, src)
}
func (m *BlockData) XXX_Size() int {
	return m.Size()
}
func (m *BlockData) XXX_DiscardUnknown() {
	xxx_messageInfo_BlockData.DiscardUnknown(m)
}

var xxx_messageInfo_BlockData proto.InternalMessageInfo

func (m *BlockData) GetBlockBytes() []byte {
	if m != nil {
		return m.BlockBytes
	}
	return nil
}

func init() {
	proto.RegisterType((*GetBlockDatas)(nil), "qitmeer.p2p.v1.GetBlockDatas")
	proto.RegisterType((*BlockDatas)(nil), "qitmeer.p2p.v1.BlockDatas")
	proto.RegisterType((*BlockData)(nil), "qitmeer.p2p.v1.BlockData")
}

func init() { proto.RegisterFile("getblockdatas.proto", fileDescriptor_5c509316b8da1820) }

var fileDescriptor_5c509316b8da1820 = []byte{
	// 249 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x4e, 0x4f, 0x2d, 0x49,
	0xca, 0xc9, 0x4f, 0xce, 0x4e, 0x49, 0x2c, 0x49, 0x2c, 0xd6, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17,
	0xe2, 0x2b, 0xcc, 0x2c, 0xc9, 0x4d, 0x4d, 0x2d, 0xd2, 0x2b, 0x30, 0x2a, 0xd0, 0x2b, 0x33, 0x94,
	0xd2, 0x4d, 0xcf, 0x2c, 0xc9, 0x28, 0x4d, 0xd2, 0x4b, 0xce, 0xcf, 0xd5, 0x4f, 0xcf, 0x4f, 0xcf,
	0xd7, 0x07, 0x2b, 0x4b, 0x2a, 0x4d, 0x03, 0xf3, 0xc0, 0x1c, 0x30, 0x0b, 0xa2, 0x5d, 0x8a, 0x2f,
	0x37, 0xb5, 0xb8, 0x38, 0x31, 0x3d, 0x15, 0x6a, 0x9c, 0x52, 0x30, 0x17, 0xaf, 0x7b, 0x6a, 0x89,
	0x13, 0xc8, 0x16, 0x17, 0x90, 0x2d, 0x42, 0x4e, 0x5c, 0xec, 0x39, 0xf9, 0xc9, 0x89, 0x25, 0xf9,
	0x45, 0x12, 0x8c, 0x0a, 0xcc, 0x1a, 0xdc, 0x46, 0x22, 0x7a, 0xa8, 0x36, 0xea, 0x79, 0x24, 0x16,
	0x67, 0x38, 0x09, 0x7d, 0xba, 0x27, 0xcf, 0x57, 0x5c, 0x5c, 0xa5, 0x9b, 0x9b, 0x58, 0x61, 0xa5,
	0x64, 0x64, 0x60, 0x60, 0xa0, 0x14, 0x04, 0xd3, 0xa8, 0x14, 0xca, 0xc5, 0x85, 0x64, 0xa2, 0x3b,
	0xba, 0x89, 0x92, 0xe8, 0x26, 0xc2, 0x15, 0xe3, 0x37, 0xd6, 0x89, 0x8b, 0x13, 0xae, 0x52, 0xc8,
	0x94, 0x8b, 0x0b, 0x1c, 0x36, 0x4e, 0x95, 0x25, 0xa9, 0xc5, 0x12, 0x29, 0x0a, 0x8c, 0x1a, 0x3c,
	0x4e, 0xa2, 0x9f, 0xee, 0xc9, 0x0b, 0xc2, 0x75, 0x1b, 0x1a, 0x98, 0x58, 0x98, 0x9a, 0x9b, 0x29,
	0x05, 0x21, 0x29, 0x74, 0x12, 0x38, 0xf1, 0x48, 0x8e, 0xf1, 0xc2, 0x23, 0x39, 0xc6, 0x07, 0x8f,
	0xe4, 0x18, 0x67, 0x3c, 0x96, 0x63, 0x48, 0x62, 0x03, 0x07, 0x84, 0x31, 0x20, 0x00, 0x00, 0xff,
	0xff, 0xbd, 0x93, 0x10, 0x94, 0x6e, 0x01, 0x00, 0x00,
}

func (m *GetBlockDatas) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GetBlockDatas) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GetBlockDatas) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if len(m.Locator) > 0 {
		for iNdEx := len(m.Locator) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Locator[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGetblockdatas(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *BlockDatas) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *BlockDatas) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *BlockDatas) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if len(m.Locator) > 0 {
		for iNdEx := len(m.Locator) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Locator[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGetblockdatas(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *BlockData) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *BlockData) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *BlockData) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if len(m.BlockBytes) > 0 {
		i -= len(m.BlockBytes)
		copy(dAtA[i:], m.BlockBytes)
		i = encodeVarintGetblockdatas(dAtA, i, uint64(len(m.BlockBytes)))
		i--
		dAtA[i] = 0x6
		i--
		dAtA[i] = 0xa2
	}
	return len(dAtA) - i, nil
}

func encodeVarintGetblockdatas(dAtA []byte, offset int, v uint64) int {
	offset -= sovGetblockdatas(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *GetBlockDatas) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Locator) > 0 {
		for _, e := range m.Locator {
			l = e.Size()
			n += 1 + l + sovGetblockdatas(uint64(l))
		}
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *BlockDatas) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Locator) > 0 {
		for _, e := range m.Locator {
			l = e.Size()
			n += 1 + l + sovGetblockdatas(uint64(l))
		}
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *BlockData) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.BlockBytes)
	if l > 0 {
		n += 2 + l + sovGetblockdatas(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovGetblockdatas(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozGetblockdatas(x uint64) (n int) {
	return sovGetblockdatas(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *GetBlockDatas) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGetblockdatas
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
			return fmt.Errorf("proto: GetBlockDatas: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GetBlockDatas: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Locator", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGetblockdatas
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
				return ErrInvalidLengthGetblockdatas
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGetblockdatas
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Locator = append(m.Locator, &Hash{})
			if err := m.Locator[len(m.Locator)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGetblockdatas(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthGetblockdatas
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthGetblockdatas
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
func (m *BlockDatas) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGetblockdatas
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
			return fmt.Errorf("proto: BlockDatas: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: BlockDatas: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Locator", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGetblockdatas
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
				return ErrInvalidLengthGetblockdatas
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGetblockdatas
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Locator = append(m.Locator, &BlockData{})
			if err := m.Locator[len(m.Locator)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGetblockdatas(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthGetblockdatas
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthGetblockdatas
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
func (m *BlockData) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGetblockdatas
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
			return fmt.Errorf("proto: BlockData: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: BlockData: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 100:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BlockBytes", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGetblockdatas
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
				return ErrInvalidLengthGetblockdatas
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthGetblockdatas
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.BlockBytes = append(m.BlockBytes[:0], dAtA[iNdEx:postIndex]...)
			if m.BlockBytes == nil {
				m.BlockBytes = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGetblockdatas(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthGetblockdatas
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthGetblockdatas
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
func skipGetblockdatas(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowGetblockdatas
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
					return 0, ErrIntOverflowGetblockdatas
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
					return 0, ErrIntOverflowGetblockdatas
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
				return 0, ErrInvalidLengthGetblockdatas
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupGetblockdatas
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthGetblockdatas
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthGetblockdatas        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowGetblockdatas          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupGetblockdatas = fmt.Errorf("proto: unexpected end of group")
)
