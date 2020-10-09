// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: getblocks.proto

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

type GetBlocks struct {
	Locator              []*Hash  `protobuf:"bytes,1,rep,name=locator,proto3" json:"locator,omitempty" ssz-max:"500"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetBlocks) Reset()         { *m = GetBlocks{} }
func (m *GetBlocks) String() string { return proto.CompactTextString(m) }
func (*GetBlocks) ProtoMessage()    {}
func (*GetBlocks) Descriptor() ([]byte, []int) {
	return fileDescriptor_ed45aea3b866f95b, []int{0}
}
func (m *GetBlocks) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GetBlocks) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GetBlocks.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GetBlocks) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetBlocks.Merge(m, src)
}
func (m *GetBlocks) XXX_Size() int {
	return m.Size()
}
func (m *GetBlocks) XXX_DiscardUnknown() {
	xxx_messageInfo_GetBlocks.DiscardUnknown(m)
}

var xxx_messageInfo_GetBlocks proto.InternalMessageInfo

func (m *GetBlocks) GetLocator() []*Hash {
	if m != nil {
		return m.Locator
	}
	return nil
}

type Blocks struct {
	Blocks               []*Block `protobuf:"bytes,1,rep,name=blocks,proto3" json:"blocks,omitempty" ssz-max:"500"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Blocks) Reset()         { *m = Blocks{} }
func (m *Blocks) String() string { return proto.CompactTextString(m) }
func (*Blocks) ProtoMessage()    {}
func (*Blocks) Descriptor() ([]byte, []int) {
	return fileDescriptor_ed45aea3b866f95b, []int{1}
}
func (m *Blocks) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Blocks) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Blocks.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Blocks) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Blocks.Merge(m, src)
}
func (m *Blocks) XXX_Size() int {
	return m.Size()
}
func (m *Blocks) XXX_DiscardUnknown() {
	xxx_messageInfo_Blocks.DiscardUnknown(m)
}

var xxx_messageInfo_Blocks proto.InternalMessageInfo

func (m *Blocks) GetBlocks() []*Block {
	if m != nil {
		return m.Blocks
	}
	return nil
}

type Block struct {
	BlockBytes           []byte   `protobuf:"bytes,100,opt,name=blockBytes,proto3" json:"blockBytes,omitempty" ssz-max:"1048576"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Block) Reset()         { *m = Block{} }
func (m *Block) String() string { return proto.CompactTextString(m) }
func (*Block) ProtoMessage()    {}
func (*Block) Descriptor() ([]byte, []int) {
	return fileDescriptor_ed45aea3b866f95b, []int{2}
}
func (m *Block) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Block) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Block.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Block) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Block.Merge(m, src)
}
func (m *Block) XXX_Size() int {
	return m.Size()
}
func (m *Block) XXX_DiscardUnknown() {
	xxx_messageInfo_Block.DiscardUnknown(m)
}

var xxx_messageInfo_Block proto.InternalMessageInfo

func (m *Block) GetBlockBytes() []byte {
	if m != nil {
		return m.BlockBytes
	}
	return nil
}

func init() {
	proto.RegisterType((*GetBlocks)(nil), "qitmeer.p2p.v1.GetBlocks")
	proto.RegisterType((*Blocks)(nil), "qitmeer.p2p.v1.Blocks")
	proto.RegisterType((*Block)(nil), "qitmeer.p2p.v1.Block")
}

func init() { proto.RegisterFile("getblocks.proto", fileDescriptor_ed45aea3b866f95b) }

var fileDescriptor_ed45aea3b866f95b = []byte{
	// 248 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x4f, 0x4f, 0x2d, 0x49,
	0xca, 0xc9, 0x4f, 0xce, 0x2e, 0xd6, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x2b, 0xcc, 0x2c,
	0xc9, 0x4d, 0x4d, 0x2d, 0xd2, 0x2b, 0x30, 0x2a, 0xd0, 0x2b, 0x33, 0x94, 0xd2, 0x4d, 0xcf, 0x2c,
	0xc9, 0x28, 0x4d, 0xd2, 0x4b, 0xce, 0xcf, 0xd5, 0x4f, 0xcf, 0x4f, 0xcf, 0xd7, 0x07, 0x2b, 0x4b,
	0x2a, 0x4d, 0x03, 0xf3, 0xc0, 0x1c, 0x30, 0x0b, 0xa2, 0x5d, 0x8a, 0x2f, 0x37, 0xb5, 0xb8, 0x38,
	0x31, 0x3d, 0x15, 0x6a, 0x9c, 0x92, 0x1f, 0x17, 0xa7, 0x7b, 0x6a, 0x89, 0x13, 0xd8, 0x06, 0x21,
	0x47, 0x2e, 0xf6, 0x9c, 0xfc, 0xe4, 0xc4, 0x92, 0xfc, 0x22, 0x09, 0x46, 0x05, 0x66, 0x0d, 0x6e,
	0x23, 0x11, 0x3d, 0x54, 0xdb, 0xf4, 0x3c, 0x12, 0x8b, 0x33, 0x9c, 0x04, 0x3f, 0xdd, 0x93, 0xe7,
	0x2d, 0x2e, 0xae, 0xd2, 0xcd, 0x4d, 0xac, 0xb0, 0x52, 0x32, 0x35, 0x30, 0x50, 0x0a, 0x82, 0xe9,
	0x53, 0xf2, 0xe2, 0x62, 0x83, 0x1a, 0xe6, 0xc0, 0xc5, 0x06, 0x71, 0x38, 0xd4, 0x2c, 0x51, 0x74,
	0xb3, 0xc0, 0xea, 0xb0, 0x19, 0x06, 0xd5, 0xa7, 0x64, 0xc7, 0xc5, 0x0a, 0x56, 0x23, 0x64, 0xca,
	0xc5, 0x05, 0x16, 0x72, 0xaa, 0x2c, 0x49, 0x2d, 0x96, 0x48, 0x51, 0x60, 0xd4, 0xe0, 0x71, 0x12,
	0xfd, 0x74, 0x4f, 0x5e, 0x10, 0xae, 0xcf, 0xd0, 0xc0, 0xc4, 0xc2, 0xd4, 0xdc, 0x4c, 0x29, 0x08,
	0x49, 0xa1, 0x93, 0xc0, 0x89, 0x47, 0x72, 0x8c, 0x17, 0x1e, 0xc9, 0x31, 0x3e, 0x78, 0x24, 0xc7,
	0x38, 0xe3, 0xb1, 0x1c, 0x43, 0x12, 0x1b, 0xd8, 0xd3, 0xc6, 0x80, 0x00, 0x00, 0x00, 0xff, 0xff,
	0xe7, 0x77, 0x32, 0x73, 0x56, 0x01, 0x00, 0x00,
}

func (m *GetBlocks) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GetBlocks) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GetBlocks) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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
				i = encodeVarintGetblocks(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *Blocks) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Blocks) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Blocks) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if len(m.Blocks) > 0 {
		for iNdEx := len(m.Blocks) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Blocks[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGetblocks(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *Block) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Block) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Block) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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
		i = encodeVarintGetblocks(dAtA, i, uint64(len(m.BlockBytes)))
		i--
		dAtA[i] = 0x6
		i--
		dAtA[i] = 0xa2
	}
	return len(dAtA) - i, nil
}

func encodeVarintGetblocks(dAtA []byte, offset int, v uint64) int {
	offset -= sovGetblocks(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *GetBlocks) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Locator) > 0 {
		for _, e := range m.Locator {
			l = e.Size()
			n += 1 + l + sovGetblocks(uint64(l))
		}
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *Blocks) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Blocks) > 0 {
		for _, e := range m.Blocks {
			l = e.Size()
			n += 1 + l + sovGetblocks(uint64(l))
		}
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *Block) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.BlockBytes)
	if l > 0 {
		n += 2 + l + sovGetblocks(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovGetblocks(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozGetblocks(x uint64) (n int) {
	return sovGetblocks(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *GetBlocks) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGetblocks
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
			return fmt.Errorf("proto: GetBlocks: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GetBlocks: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Locator", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGetblocks
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
				return ErrInvalidLengthGetblocks
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGetblocks
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
			skippy, err := skipGetblocks(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthGetblocks
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthGetblocks
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
func (m *Blocks) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGetblocks
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
			return fmt.Errorf("proto: Blocks: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Blocks: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Blocks", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGetblocks
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
				return ErrInvalidLengthGetblocks
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGetblocks
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Blocks = append(m.Blocks, &Block{})
			if err := m.Blocks[len(m.Blocks)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGetblocks(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthGetblocks
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthGetblocks
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
func (m *Block) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGetblocks
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
			return fmt.Errorf("proto: Block: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Block: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 100:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BlockBytes", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGetblocks
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
				return ErrInvalidLengthGetblocks
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthGetblocks
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
			skippy, err := skipGetblocks(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthGetblocks
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthGetblocks
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
func skipGetblocks(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowGetblocks
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
					return 0, ErrIntOverflowGetblocks
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
					return 0, ErrIntOverflowGetblocks
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
				return 0, ErrInvalidLengthGetblocks
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupGetblocks
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthGetblocks
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthGetblocks        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowGetblocks          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupGetblocks = fmt.Errorf("proto: unexpected end of group")
)
