// Code generated by protoc-gen-go. DO NOT EDIT.
// source: test_schema/user.proto

package test_schema

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

type User struct {
	UserId               *int32   `protobuf:"varint,1,opt,name=user_id,json=userId" json:"user_id,omitempty"`
	FirstName            *string  `protobuf:"bytes,2,opt,name=first_name,json=firstName" json:"first_name,omitempty"`
	LastName             *string  `protobuf:"bytes,3,opt,name=last_name,json=lastName" json:"last_name,omitempty"`
	Email                *string  `protobuf:"bytes,4,opt,name=email" json:"email,omitempty"`
	UserToken            *string  `protobuf:"bytes,5,opt,name=user_token,json=userToken" json:"user_token,omitempty"`
	Enabled              *bool    `protobuf:"varint,6,opt,name=enabled" json:"enabled,omitempty"`
	AkaId                *int32   `protobuf:"varint,7,opt,name=aka_id,json=akaId" json:"aka_id,omitempty"`
	Lat                  *float64 `protobuf:"fixed64,8,opt,name=lat" json:"lat,omitempty"`
	Lon                  *float64 `protobuf:"fixed64,9,opt,name=lon" json:"lon,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *User) Reset()         { *m = User{} }
func (m *User) String() string { return proto.CompactTextString(m) }
func (*User) ProtoMessage()    {}
func (*User) Descriptor() ([]byte, []int) {
	return fileDescriptor_43feef3e3d9881c0, []int{0}
}

func (m *User) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_User.Unmarshal(m, b)
}
func (m *User) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_User.Marshal(b, m, deterministic)
}
func (m *User) XXX_Merge(src proto.Message) {
	xxx_messageInfo_User.Merge(m, src)
}
func (m *User) XXX_Size() int {
	return xxx_messageInfo_User.Size(m)
}
func (m *User) XXX_DiscardUnknown() {
	xxx_messageInfo_User.DiscardUnknown(m)
}

var xxx_messageInfo_User proto.InternalMessageInfo

func (m *User) GetUserId() int32 {
	if m != nil && m.UserId != nil {
		return *m.UserId
	}
	return 0
}

func (m *User) GetFirstName() string {
	if m != nil && m.FirstName != nil {
		return *m.FirstName
	}
	return ""
}

func (m *User) GetLastName() string {
	if m != nil && m.LastName != nil {
		return *m.LastName
	}
	return ""
}

func (m *User) GetEmail() string {
	if m != nil && m.Email != nil {
		return *m.Email
	}
	return ""
}

func (m *User) GetUserToken() string {
	if m != nil && m.UserToken != nil {
		return *m.UserToken
	}
	return ""
}

func (m *User) GetEnabled() bool {
	if m != nil && m.Enabled != nil {
		return *m.Enabled
	}
	return false
}

func (m *User) GetAkaId() int32 {
	if m != nil && m.AkaId != nil {
		return *m.AkaId
	}
	return 0
}

func (m *User) GetLat() float64 {
	if m != nil && m.Lat != nil {
		return *m.Lat
	}
	return 0
}

func (m *User) GetLon() float64 {
	if m != nil && m.Lon != nil {
		return *m.Lon
	}
	return 0
}

func init() {
	proto.RegisterType((*User)(nil), "test_schema.User")
}

func init() { proto.RegisterFile("test_schema/user.proto", fileDescriptor_43feef3e3d9881c0) }

var fileDescriptor_43feef3e3d9881c0 = []byte{
	// 243 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x4c, 0x90, 0xcf, 0x4a, 0x03, 0x31,
	0x10, 0x87, 0x49, 0xb7, 0xfb, 0x27, 0xe3, 0x45, 0x82, 0xda, 0x80, 0x08, 0xe2, 0xc9, 0xd3, 0xf6,
	0x1d, 0x8a, 0x1e, 0x3c, 0x28, 0x65, 0xd1, 0x73, 0x19, 0xbb, 0x23, 0x2e, 0xcd, 0x26, 0x65, 0x37,
	0x82, 0xcf, 0xe3, 0xbb, 0xf9, 0x0e, 0x1e, 0x9d, 0x49, 0x5b, 0xd8, 0x5b, 0x7e, 0xdf, 0x97, 0xcc,
	0x4c, 0x06, 0xae, 0x22, 0x8d, 0x71, 0x33, 0x6e, 0x3f, 0xa9, 0xc7, 0xe5, 0xd7, 0x48, 0x43, 0xbd,
	0x1f, 0x42, 0x0c, 0xe6, 0x6c, 0xc2, 0xef, 0x7e, 0x15, 0xcc, 0xdf, 0xd8, 0x99, 0x05, 0x94, 0x72,
	0x67, 0xd3, 0xb5, 0x56, 0xdd, 0xaa, 0xfb, 0xbc, 0x29, 0x24, 0x3e, 0xb5, 0xe6, 0x06, 0xe0, 0xa3,
	0x1b, 0xf8, 0x85, 0xc7, 0x9e, 0xec, 0x8c, 0x9d, 0x6e, 0x74, 0x22, 0x2f, 0x0c, 0xcc, 0x35, 0x68,
	0x87, 0x27, 0x9b, 0x25, 0x5b, 0x09, 0x48, 0xf2, 0x02, 0x72, 0x6e, 0xd2, 0x39, 0x3b, 0x4f, 0xe2,
	0x10, 0xa4, 0x62, 0x6a, 0x15, 0xc3, 0x8e, 0xbc, 0xcd, 0x0f, 0x15, 0x85, 0xbc, 0x0a, 0x30, 0x16,
	0x4a, 0xf2, 0xf8, 0xee, 0xa8, 0xb5, 0x05, 0xbb, 0xaa, 0x39, 0x45, 0x73, 0x09, 0x05, 0xee, 0x50,
	0x46, 0x2c, 0xd3, 0x88, 0x39, 0x27, 0x9e, 0xf0, 0x1c, 0x32, 0x87, 0xd1, 0x56, 0xcc, 0x54, 0x23,
	0xc7, 0x44, 0x82, 0xb7, 0xfa, 0x48, 0x82, 0x5f, 0x2d, 0x61, 0xb1, 0x0d, 0x7d, 0x4d, 0xdf, 0xd8,
	0xef, 0x1d, 0xd5, 0x93, 0x15, 0xac, 0xb4, 0xfc, 0x7f, 0x2d, 0xab, 0x59, 0xab, 0x3f, 0xa5, 0x7e,
	0x66, 0xd9, 0xe3, 0xc3, 0xf3, 0x7f, 0x00, 0x00, 0x00, 0xff, 0xff, 0xd6, 0x2c, 0xc5, 0xb0, 0x3e,
	0x01, 0x00, 0x00,
}
