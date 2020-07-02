// Code generated by protoc-gen-go. DO NOT EDIT.
// source: pdxsafe_op.proto

package protos

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
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

type PDXS int32

const (
	PDXS_CREATE_DOMAIN PDXS = 0
	PDXS_UPDATE_DOMAIN PDXS = 1
	PDXS_DELETE_DOMAIN PDXS = 2
	PDXS_GET_DOMAIN    PDXS = 3
	PDXS_CREATE_KEY    PDXS = 4
	PDXS_UPDATE_KEY    PDXS = 5
	PDXS_DELETE_KEY    PDXS = 6
	PDXS_GET_KEY       PDXS = 7
)

var PDXS_name = map[int32]string{
	0: "CREATE_DOMAIN",
	1: "UPDATE_DOMAIN",
	2: "DELETE_DOMAIN",
	3: "GET_DOMAIN",
	4: "CREATE_KEY",
	5: "UPDATE_KEY",
	6: "DELETE_KEY",
	7: "GET_KEY",
}

var PDXS_value = map[string]int32{
	"CREATE_DOMAIN": 0,
	"UPDATE_DOMAIN": 1,
	"DELETE_DOMAIN": 2,
	"GET_DOMAIN":    3,
	"CREATE_KEY":    4,
	"UPDATE_KEY":    5,
	"DELETE_KEY":    6,
	"GET_KEY":       7,
}

func (x PDXS) String() string {
	return proto.EnumName(PDXS_name, int32(x))
}

func (PDXS) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_a967d8cb84ed25a9, []int{0}
}

type DomainItem struct {
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Type                 string   `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`
	Desc                 string   `protobuf:"bytes,3,opt,name=desc,proto3" json:"desc,omitempty"`
	Cauth                [][]byte `protobuf:"bytes,4,rep,name=Cauth,proto3" json:"Cauth,omitempty"`
	Uauth                [][]byte `protobuf:"bytes,5,rep,name=Uauth,proto3" json:"Uauth,omitempty"`
	Dauth                [][]byte `protobuf:"bytes,6,rep,name=Dauth,proto3" json:"Dauth,omitempty"`
	Keycount             uint64   `protobuf:"varint,7,opt,name=keycount,proto3" json:"keycount,omitempty"`
	Version              uint64   `protobuf:"varint,8,opt,name=version,proto3" json:"version,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DomainItem) Reset()         { *m = DomainItem{} }
func (m *DomainItem) String() string { return proto.CompactTextString(m) }
func (*DomainItem) ProtoMessage()    {}
func (*DomainItem) Descriptor() ([]byte, []int) {
	return fileDescriptor_a967d8cb84ed25a9, []int{0}
}

func (m *DomainItem) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DomainItem.Unmarshal(m, b)
}
func (m *DomainItem) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DomainItem.Marshal(b, m, deterministic)
}
func (m *DomainItem) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DomainItem.Merge(m, src)
}
func (m *DomainItem) XXX_Size() int {
	return xxx_messageInfo_DomainItem.Size(m)
}
func (m *DomainItem) XXX_DiscardUnknown() {
	xxx_messageInfo_DomainItem.DiscardUnknown(m)
}

var xxx_messageInfo_DomainItem proto.InternalMessageInfo

func (m *DomainItem) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *DomainItem) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

func (m *DomainItem) GetDesc() string {
	if m != nil {
		return m.Desc
	}
	return ""
}

func (m *DomainItem) GetCauth() [][]byte {
	if m != nil {
		return m.Cauth
	}
	return nil
}

func (m *DomainItem) GetUauth() [][]byte {
	if m != nil {
		return m.Uauth
	}
	return nil
}

func (m *DomainItem) GetDauth() [][]byte {
	if m != nil {
		return m.Dauth
	}
	return nil
}

func (m *DomainItem) GetKeycount() uint64 {
	if m != nil {
		return m.Keycount
	}
	return 0
}

func (m *DomainItem) GetVersion() uint64 {
	if m != nil {
		return m.Version
	}
	return 0
}

type KeyItem struct {
	Key                  string   `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value                []byte   `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	Desc                 string   `protobuf:"bytes,3,opt,name=desc,proto3" json:"desc,omitempty"`
	Uauth                [][]byte `protobuf:"bytes,4,rep,name=Uauth,proto3" json:"Uauth,omitempty"`
	Dauth                [][]byte `protobuf:"bytes,5,rep,name=Dauth,proto3" json:"Dauth,omitempty"`
	Dname                string   `protobuf:"bytes,6,opt,name=dname,proto3" json:"dname,omitempty"`
	Version              uint64   `protobuf:"varint,7,opt,name=version,proto3" json:"version,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *KeyItem) Reset()         { *m = KeyItem{} }
func (m *KeyItem) String() string { return proto.CompactTextString(m) }
func (*KeyItem) ProtoMessage()    {}
func (*KeyItem) Descriptor() ([]byte, []int) {
	return fileDescriptor_a967d8cb84ed25a9, []int{1}
}

func (m *KeyItem) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KeyItem.Unmarshal(m, b)
}
func (m *KeyItem) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KeyItem.Marshal(b, m, deterministic)
}
func (m *KeyItem) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KeyItem.Merge(m, src)
}
func (m *KeyItem) XXX_Size() int {
	return xxx_messageInfo_KeyItem.Size(m)
}
func (m *KeyItem) XXX_DiscardUnknown() {
	xxx_messageInfo_KeyItem.DiscardUnknown(m)
}

var xxx_messageInfo_KeyItem proto.InternalMessageInfo

func (m *KeyItem) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *KeyItem) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *KeyItem) GetDesc() string {
	if m != nil {
		return m.Desc
	}
	return ""
}

func (m *KeyItem) GetUauth() [][]byte {
	if m != nil {
		return m.Uauth
	}
	return nil
}

func (m *KeyItem) GetDauth() [][]byte {
	if m != nil {
		return m.Dauth
	}
	return nil
}

func (m *KeyItem) GetDname() string {
	if m != nil {
		return m.Dname
	}
	return ""
}

func (m *KeyItem) GetVersion() uint64 {
	if m != nil {
		return m.Version
	}
	return 0
}

func init() {
	proto.RegisterEnum("protos.PDXS", PDXS_name, PDXS_value)
	proto.RegisterType((*DomainItem)(nil), "protos.DomainItem")
	proto.RegisterType((*KeyItem)(nil), "protos.KeyItem")
}

func init() { proto.RegisterFile("pdxsafe_op.proto", fileDescriptor_a967d8cb84ed25a9) }

var fileDescriptor_a967d8cb84ed25a9 = []byte{
	// 316 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x91, 0xdd, 0x4e, 0xc2, 0x30,
	0x14, 0x80, 0x2d, 0xfb, 0x29, 0x1e, 0xd1, 0xcc, 0x86, 0x8b, 0xc6, 0x2b, 0xc2, 0x15, 0xf1, 0xc2,
	0x1b, 0x9f, 0x00, 0x69, 0x63, 0x08, 0xfe, 0x90, 0xc9, 0x12, 0xf5, 0x86, 0x4c, 0xa8, 0x91, 0x20,
	0xeb, 0xc2, 0x36, 0xe2, 0xde, 0xc1, 0x97, 0xf0, 0x4d, 0x7c, 0x34, 0x73, 0xda, 0x4d, 0xa7, 0x89,
	0x57, 0x3d, 0xdf, 0xb7, 0xec, 0xec, 0xcb, 0x0a, 0x41, 0xba, 0x7c, 0xcb, 0xe2, 0x67, 0x35, 0xd7,
	0xe9, 0x59, 0xba, 0xd5, 0xb9, 0x66, 0xbe, 0x39, 0xb2, 0xfe, 0x27, 0x01, 0x10, 0x7a, 0x13, 0xaf,
	0x92, 0x71, 0xae, 0x36, 0x8c, 0x81, 0x9b, 0xc4, 0x1b, 0xc5, 0x49, 0x8f, 0x0c, 0xf6, 0x43, 0x33,
	0xa3, 0xcb, 0xcb, 0x54, 0xf1, 0x96, 0x75, 0x38, 0xa3, 0x5b, 0xaa, 0x6c, 0xc1, 0x1d, 0xeb, 0x70,
	0x66, 0x5d, 0xf0, 0x46, 0x71, 0x91, 0xbf, 0x70, 0xb7, 0xe7, 0x0c, 0x3a, 0xa1, 0x05, 0xb4, 0x91,
	0xb1, 0x9e, 0xb5, 0x51, 0x6d, 0x85, 0xb1, 0xbe, 0xb5, 0x06, 0xd8, 0x09, 0xb4, 0xd7, 0xaa, 0x5c,
	0xe8, 0x22, 0xc9, 0x39, 0xed, 0x91, 0x81, 0x1b, 0x7e, 0x33, 0xe3, 0x40, 0x77, 0x6a, 0x9b, 0xad,
	0x74, 0xc2, 0xdb, 0xe6, 0x51, 0x8d, 0xfd, 0x0f, 0x02, 0x74, 0xa2, 0x4a, 0xd3, 0x1f, 0x80, 0xb3,
	0x56, 0x65, 0x95, 0x8f, 0x23, 0x7e, 0x69, 0x17, 0xbf, 0x16, 0x36, 0xbf, 0x13, 0x5a, 0xf8, 0xaf,
	0x3f, 0x6a, 0xf6, 0xff, 0x29, 0xf5, 0x9a, 0xa5, 0x5d, 0xf0, 0x96, 0xe6, 0x47, 0xf9, 0x66, 0x81,
	0x85, 0x66, 0x23, 0xfd, 0xd5, 0x78, 0xfa, 0x4e, 0xc0, 0x9d, 0x8a, 0xfb, 0x3b, 0x76, 0x0c, 0x87,
	0xa3, 0x50, 0x0e, 0x67, 0x72, 0x2e, 0x6e, 0xaf, 0x87, 0xe3, 0x9b, 0x60, 0x0f, 0x55, 0x34, 0x15,
	0x0d, 0x45, 0x50, 0x09, 0x79, 0x25, 0x7f, 0x54, 0x8b, 0x1d, 0x01, 0x5c, 0xca, 0x59, 0xcd, 0x0e,
	0x72, 0xb5, 0x68, 0x22, 0x1f, 0x02, 0x17, 0xb9, 0xda, 0x82, 0xec, 0x21, 0x57, 0x2b, 0x90, 0x7d,
	0x76, 0x00, 0x14, 0xdf, 0x47, 0xa0, 0x17, 0xed, 0xc7, 0xea, 0xfe, 0x9f, 0xec, 0x79, 0xfe, 0x15,
	0x00, 0x00, 0xff, 0xff, 0x40, 0xd2, 0xcc, 0x19, 0x22, 0x02, 0x00, 0x00,
}
