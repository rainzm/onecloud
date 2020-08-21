// Code generated by protoc-gen-go. DO NOT EDIT.
// source: send_server.proto

package apis

import (
	context "context"
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

type SendParams struct {
	Contact              string   `protobuf:"bytes,1,opt,name=Contact,proto3" json:"Contact,omitempty"`
	Topic                string   `protobuf:"bytes,2,opt,name=Topic,proto3" json:"Topic,omitempty"`
	Title                string   `protobuf:"bytes,3,opt,name=Title,proto3" json:"Title,omitempty"`
	Message              string   `protobuf:"bytes,4,opt,name=Message,proto3" json:"Message,omitempty"`
	Priority             string   `protobuf:"bytes,5,opt,name=Priority,proto3" json:"Priority,omitempty"`
	RemoteTemplate       string   `protobuf:"bytes,6,opt,name=RemoteTemplate,proto3" json:"RemoteTemplate,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SendParams) Reset()         { *m = SendParams{} }
func (m *SendParams) String() string { return proto.CompactTextString(m) }
func (*SendParams) ProtoMessage()    {}
func (*SendParams) Descriptor() ([]byte, []int) {
	return fileDescriptor_63fdd68f7eb311f9, []int{0}
}

func (m *SendParams) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SendParams.Unmarshal(m, b)
}
func (m *SendParams) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SendParams.Marshal(b, m, deterministic)
}
func (m *SendParams) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SendParams.Merge(m, src)
}
func (m *SendParams) XXX_Size() int {
	return xxx_messageInfo_SendParams.Size(m)
}
func (m *SendParams) XXX_DiscardUnknown() {
	xxx_messageInfo_SendParams.DiscardUnknown(m)
}

var xxx_messageInfo_SendParams proto.InternalMessageInfo

func (m *SendParams) GetContact() string {
	if m != nil {
		return m.Contact
	}
	return ""
}

func (m *SendParams) GetTopic() string {
	if m != nil {
		return m.Topic
	}
	return ""
}

func (m *SendParams) GetTitle() string {
	if m != nil {
		return m.Title
	}
	return ""
}

func (m *SendParams) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func (m *SendParams) GetPriority() string {
	if m != nil {
		return m.Priority
	}
	return ""
}

func (m *SendParams) GetRemoteTemplate() string {
	if m != nil {
		return m.RemoteTemplate
	}
	return ""
}

type UpdateConfigParams struct {
	Configs              map[string]string `protobuf:"bytes,1,rep,name=configs,proto3" json:"configs,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *UpdateConfigParams) Reset()         { *m = UpdateConfigParams{} }
func (m *UpdateConfigParams) String() string { return proto.CompactTextString(m) }
func (*UpdateConfigParams) ProtoMessage()    {}
func (*UpdateConfigParams) Descriptor() ([]byte, []int) {
	return fileDescriptor_63fdd68f7eb311f9, []int{1}
}

func (m *UpdateConfigParams) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UpdateConfigParams.Unmarshal(m, b)
}
func (m *UpdateConfigParams) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UpdateConfigParams.Marshal(b, m, deterministic)
}
func (m *UpdateConfigParams) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UpdateConfigParams.Merge(m, src)
}
func (m *UpdateConfigParams) XXX_Size() int {
	return xxx_messageInfo_UpdateConfigParams.Size(m)
}
func (m *UpdateConfigParams) XXX_DiscardUnknown() {
	xxx_messageInfo_UpdateConfigParams.DiscardUnknown(m)
}

var xxx_messageInfo_UpdateConfigParams proto.InternalMessageInfo

func (m *UpdateConfigParams) GetConfigs() map[string]string {
	if m != nil {
		return m.Configs
	}
	return nil
}

type UseridByMobileParams struct {
	Mobile               string   `protobuf:"bytes,1,opt,name=mobile,proto3" json:"mobile,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *UseridByMobileParams) Reset()         { *m = UseridByMobileParams{} }
func (m *UseridByMobileParams) String() string { return proto.CompactTextString(m) }
func (*UseridByMobileParams) ProtoMessage()    {}
func (*UseridByMobileParams) Descriptor() ([]byte, []int) {
	return fileDescriptor_63fdd68f7eb311f9, []int{2}
}

func (m *UseridByMobileParams) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UseridByMobileParams.Unmarshal(m, b)
}
func (m *UseridByMobileParams) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UseridByMobileParams.Marshal(b, m, deterministic)
}
func (m *UseridByMobileParams) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UseridByMobileParams.Merge(m, src)
}
func (m *UseridByMobileParams) XXX_Size() int {
	return xxx_messageInfo_UseridByMobileParams.Size(m)
}
func (m *UseridByMobileParams) XXX_DiscardUnknown() {
	xxx_messageInfo_UseridByMobileParams.DiscardUnknown(m)
}

var xxx_messageInfo_UseridByMobileParams proto.InternalMessageInfo

func (m *UseridByMobileParams) GetMobile() string {
	if m != nil {
		return m.Mobile
	}
	return ""
}

type Empty struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Empty) Reset()         { *m = Empty{} }
func (m *Empty) String() string { return proto.CompactTextString(m) }
func (*Empty) ProtoMessage()    {}
func (*Empty) Descriptor() ([]byte, []int) {
	return fileDescriptor_63fdd68f7eb311f9, []int{3}
}

func (m *Empty) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Empty.Unmarshal(m, b)
}
func (m *Empty) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Empty.Marshal(b, m, deterministic)
}
func (m *Empty) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Empty.Merge(m, src)
}
func (m *Empty) XXX_Size() int {
	return xxx_messageInfo_Empty.Size(m)
}
func (m *Empty) XXX_DiscardUnknown() {
	xxx_messageInfo_Empty.DiscardUnknown(m)
}

var xxx_messageInfo_Empty proto.InternalMessageInfo

type UseridByMobileReply struct {
	Userid               string   `protobuf:"bytes,1,opt,name=userid,proto3" json:"userid,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *UseridByMobileReply) Reset()         { *m = UseridByMobileReply{} }
func (m *UseridByMobileReply) String() string { return proto.CompactTextString(m) }
func (*UseridByMobileReply) ProtoMessage()    {}
func (*UseridByMobileReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_63fdd68f7eb311f9, []int{4}
}

func (m *UseridByMobileReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UseridByMobileReply.Unmarshal(m, b)
}
func (m *UseridByMobileReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UseridByMobileReply.Marshal(b, m, deterministic)
}
func (m *UseridByMobileReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UseridByMobileReply.Merge(m, src)
}
func (m *UseridByMobileReply) XXX_Size() int {
	return xxx_messageInfo_UseridByMobileReply.Size(m)
}
func (m *UseridByMobileReply) XXX_DiscardUnknown() {
	xxx_messageInfo_UseridByMobileReply.DiscardUnknown(m)
}

var xxx_messageInfo_UseridByMobileReply proto.InternalMessageInfo

func (m *UseridByMobileReply) GetUserid() string {
	if m != nil {
		return m.Userid
	}
	return ""
}

type ValidateConfigReply struct {
	IsValid              bool     `protobuf:"varint,1,opt,name=isValid,proto3" json:"isValid,omitempty"`
	Msg                  string   `protobuf:"bytes,2,opt,name=msg,proto3" json:"msg,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ValidateConfigReply) Reset()         { *m = ValidateConfigReply{} }
func (m *ValidateConfigReply) String() string { return proto.CompactTextString(m) }
func (*ValidateConfigReply) ProtoMessage()    {}
func (*ValidateConfigReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_63fdd68f7eb311f9, []int{5}
}

func (m *ValidateConfigReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ValidateConfigReply.Unmarshal(m, b)
}
func (m *ValidateConfigReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ValidateConfigReply.Marshal(b, m, deterministic)
}
func (m *ValidateConfigReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ValidateConfigReply.Merge(m, src)
}
func (m *ValidateConfigReply) XXX_Size() int {
	return xxx_messageInfo_ValidateConfigReply.Size(m)
}
func (m *ValidateConfigReply) XXX_DiscardUnknown() {
	xxx_messageInfo_ValidateConfigReply.DiscardUnknown(m)
}

var xxx_messageInfo_ValidateConfigReply proto.InternalMessageInfo

func (m *ValidateConfigReply) GetIsValid() bool {
	if m != nil {
		return m.IsValid
	}
	return false
}

func (m *ValidateConfigReply) GetMsg() string {
	if m != nil {
		return m.Msg
	}
	return ""
}

type BatchSendParams struct {
	Contacts             []string `protobuf:"bytes,1,rep,name=Contacts,proto3" json:"Contacts,omitempty"`
	Title                string   `protobuf:"bytes,2,opt,name=Title,proto3" json:"Title,omitempty"`
	Message              string   `protobuf:"bytes,3,opt,name=Message,proto3" json:"Message,omitempty"`
	Priority             string   `protobuf:"bytes,4,opt,name=Priority,proto3" json:"Priority,omitempty"`
	RemoteTemplate       string   `protobuf:"bytes,5,opt,name=RemoteTemplate,proto3" json:"RemoteTemplate,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BatchSendParams) Reset()         { *m = BatchSendParams{} }
func (m *BatchSendParams) String() string { return proto.CompactTextString(m) }
func (*BatchSendParams) ProtoMessage()    {}
func (*BatchSendParams) Descriptor() ([]byte, []int) {
	return fileDescriptor_63fdd68f7eb311f9, []int{6}
}

func (m *BatchSendParams) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BatchSendParams.Unmarshal(m, b)
}
func (m *BatchSendParams) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BatchSendParams.Marshal(b, m, deterministic)
}
func (m *BatchSendParams) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BatchSendParams.Merge(m, src)
}
func (m *BatchSendParams) XXX_Size() int {
	return xxx_messageInfo_BatchSendParams.Size(m)
}
func (m *BatchSendParams) XXX_DiscardUnknown() {
	xxx_messageInfo_BatchSendParams.DiscardUnknown(m)
}

var xxx_messageInfo_BatchSendParams proto.InternalMessageInfo

func (m *BatchSendParams) GetContacts() []string {
	if m != nil {
		return m.Contacts
	}
	return nil
}

func (m *BatchSendParams) GetTitle() string {
	if m != nil {
		return m.Title
	}
	return ""
}

func (m *BatchSendParams) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func (m *BatchSendParams) GetPriority() string {
	if m != nil {
		return m.Priority
	}
	return ""
}

func (m *BatchSendParams) GetRemoteTemplate() string {
	if m != nil {
		return m.RemoteTemplate
	}
	return ""
}

type FailedRecord struct {
	Contact              string   `protobuf:"bytes,1,opt,name=Contact,proto3" json:"Contact,omitempty"`
	Reason               string   `protobuf:"bytes,2,opt,name=Reason,proto3" json:"Reason,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *FailedRecord) Reset()         { *m = FailedRecord{} }
func (m *FailedRecord) String() string { return proto.CompactTextString(m) }
func (*FailedRecord) ProtoMessage()    {}
func (*FailedRecord) Descriptor() ([]byte, []int) {
	return fileDescriptor_63fdd68f7eb311f9, []int{7}
}

func (m *FailedRecord) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FailedRecord.Unmarshal(m, b)
}
func (m *FailedRecord) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FailedRecord.Marshal(b, m, deterministic)
}
func (m *FailedRecord) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FailedRecord.Merge(m, src)
}
func (m *FailedRecord) XXX_Size() int {
	return xxx_messageInfo_FailedRecord.Size(m)
}
func (m *FailedRecord) XXX_DiscardUnknown() {
	xxx_messageInfo_FailedRecord.DiscardUnknown(m)
}

var xxx_messageInfo_FailedRecord proto.InternalMessageInfo

func (m *FailedRecord) GetContact() string {
	if m != nil {
		return m.Contact
	}
	return ""
}

func (m *FailedRecord) GetReason() string {
	if m != nil {
		return m.Reason
	}
	return ""
}

type BatchSendReply struct {
	FailedRecords        []*FailedRecord `protobuf:"bytes,1,rep,name=FailedRecords,proto3" json:"FailedRecords,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *BatchSendReply) Reset()         { *m = BatchSendReply{} }
func (m *BatchSendReply) String() string { return proto.CompactTextString(m) }
func (*BatchSendReply) ProtoMessage()    {}
func (*BatchSendReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_63fdd68f7eb311f9, []int{8}
}

func (m *BatchSendReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BatchSendReply.Unmarshal(m, b)
}
func (m *BatchSendReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BatchSendReply.Marshal(b, m, deterministic)
}
func (m *BatchSendReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BatchSendReply.Merge(m, src)
}
func (m *BatchSendReply) XXX_Size() int {
	return xxx_messageInfo_BatchSendReply.Size(m)
}
func (m *BatchSendReply) XXX_DiscardUnknown() {
	xxx_messageInfo_BatchSendReply.DiscardUnknown(m)
}

var xxx_messageInfo_BatchSendReply proto.InternalMessageInfo

func (m *BatchSendReply) GetFailedRecords() []*FailedRecord {
	if m != nil {
		return m.FailedRecords
	}
	return nil
}

func init() {
	proto.RegisterType((*SendParams)(nil), "apis.SendParams")
	proto.RegisterType((*UpdateConfigParams)(nil), "apis.UpdateConfigParams")
	proto.RegisterMapType((map[string]string)(nil), "apis.UpdateConfigParams.ConfigsEntry")
	proto.RegisterType((*UseridByMobileParams)(nil), "apis.UseridByMobileParams")
	proto.RegisterType((*Empty)(nil), "apis.Empty")
	proto.RegisterType((*UseridByMobileReply)(nil), "apis.UseridByMobileReply")
	proto.RegisterType((*ValidateConfigReply)(nil), "apis.ValidateConfigReply")
	proto.RegisterType((*BatchSendParams)(nil), "apis.BatchSendParams")
	proto.RegisterType((*FailedRecord)(nil), "apis.FailedRecord")
	proto.RegisterType((*BatchSendReply)(nil), "apis.BatchSendReply")
}

func init() { proto.RegisterFile("send_server.proto", fileDescriptor_63fdd68f7eb311f9) }

var fileDescriptor_63fdd68f7eb311f9 = []byte{
	// 497 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x94, 0xdd, 0x6a, 0xd4, 0x40,
	0x14, 0xc7, 0xc9, 0x7e, 0xef, 0xe9, 0xba, 0xd6, 0xe9, 0x5a, 0xc6, 0x5c, 0x95, 0x40, 0xa5, 0x37,
	0xe6, 0xa2, 0x22, 0x2c, 0xbd, 0xd1, 0xb6, 0xac, 0x82, 0x50, 0x28, 0xb1, 0xf5, 0x56, 0xa6, 0xc9,
	0x71, 0x1d, 0x4c, 0x32, 0x21, 0x33, 0x2d, 0xe4, 0x31, 0x7c, 0x04, 0xdf, 0x40, 0xf0, 0x05, 0x65,
	0x3e, 0xb2, 0x9b, 0xac, 0xbb, 0xed, 0x5d, 0x7e, 0xe7, 0x8b, 0x39, 0xe7, 0xff, 0x27, 0xf0, 0x42,
	0x62, 0x9e, 0x7c, 0x93, 0x58, 0x3e, 0x60, 0x19, 0x16, 0xa5, 0x50, 0x82, 0xf4, 0x58, 0xc1, 0x65,
	0xf0, 0xc7, 0x03, 0xf8, 0x82, 0x79, 0x72, 0xcd, 0x4a, 0x96, 0x49, 0x42, 0x61, 0x78, 0x29, 0x72,
	0xc5, 0x62, 0x45, 0xbd, 0x23, 0xef, 0x64, 0x1c, 0xd5, 0x48, 0x66, 0xd0, 0xbf, 0x11, 0x05, 0x8f,
	0x69, 0xc7, 0xc4, 0x2d, 0x98, 0x28, 0x57, 0x29, 0xd2, 0xae, 0x8b, 0x6a, 0xd0, 0x53, 0xae, 0x50,
	0x4a, 0xb6, 0x44, 0xda, 0xb3, 0x53, 0x1c, 0x12, 0x1f, 0x46, 0xd7, 0x25, 0x17, 0x25, 0x57, 0x15,
	0xed, 0x9b, 0xd4, 0x8a, 0xc9, 0x6b, 0x98, 0x46, 0x98, 0x09, 0x85, 0x37, 0x98, 0x15, 0x29, 0x53,
	0x48, 0x07, 0xa6, 0x62, 0x23, 0x1a, 0xfc, 0xf2, 0x80, 0xdc, 0x16, 0x09, 0x53, 0x78, 0x29, 0xf2,
	0xef, 0x7c, 0xe9, 0x9e, 0xfe, 0x1e, 0x86, 0xb1, 0x61, 0x49, 0xbd, 0xa3, 0xee, 0xc9, 0xde, 0xe9,
	0x71, 0xa8, 0x37, 0x0c, 0xff, 0x2f, 0x0d, 0x2d, 0xc8, 0x45, 0xae, 0xca, 0x2a, 0xaa, 0xbb, 0xfc,
	0x33, 0x98, 0x34, 0x13, 0x64, 0x1f, 0xba, 0x3f, 0xb1, 0x72, 0x77, 0xd0, 0x9f, 0x7a, 0xdb, 0x07,
	0x96, 0xde, 0x63, 0x7d, 0x03, 0x03, 0x67, 0x9d, 0xb9, 0x17, 0x84, 0x30, 0xbb, 0x95, 0x58, 0xf2,
	0xe4, 0xa2, 0xba, 0x12, 0x77, 0x3c, 0x45, 0xf7, 0xa8, 0x43, 0x18, 0x64, 0x86, 0xdd, 0x18, 0x47,
	0xc1, 0x10, 0xfa, 0x8b, 0xac, 0x50, 0x55, 0xf0, 0x06, 0x0e, 0xda, 0x8d, 0x11, 0x16, 0x69, 0xa5,
	0xfb, 0xee, 0x4d, 0xb8, 0xee, 0xb3, 0x14, 0x9c, 0xc3, 0xc1, 0x57, 0x96, 0xf2, 0xf5, 0x46, 0xb6,
	0x9c, 0xc2, 0x90, 0x4b, 0x93, 0x30, 0xf5, 0xa3, 0xa8, 0x46, 0xbd, 0x44, 0x26, 0x97, 0xee, 0xc1,
	0xfa, 0x33, 0xf8, 0xed, 0xc1, 0xf3, 0x0b, 0xa6, 0xe2, 0x1f, 0x0d, 0xd9, 0x7d, 0x18, 0x39, 0x9d,
	0xed, 0xf1, 0xc6, 0xd1, 0x8a, 0xd7, 0x12, 0x77, 0x76, 0x48, 0xdc, 0xdd, 0x2d, 0x71, 0xef, 0x49,
	0x89, 0xfb, 0x5b, 0x25, 0xfe, 0x00, 0x93, 0x8f, 0x8c, 0xa7, 0x98, 0x44, 0x18, 0x8b, 0x32, 0x79,
	0xc4, 0x96, 0x87, 0x30, 0x88, 0x90, 0x49, 0x91, 0xbb, 0xe7, 0x39, 0x0a, 0x3e, 0xc3, 0x74, 0xb5,
	0xa4, 0xbd, 0xd1, 0x1c, 0x9e, 0x35, 0x67, 0xd6, 0x2e, 0x21, 0xd6, 0x25, 0xcd, 0x54, 0xd4, 0x2e,
	0x3c, 0xfd, 0xdb, 0x81, 0xb1, 0x9e, 0x73, 0xbe, 0xc4, 0x5c, 0x91, 0x63, 0xe8, 0x69, 0x20, 0xfb,
	0xb6, 0x71, 0x7d, 0x45, 0x7f, 0xcf, 0x46, 0x8c, 0xb0, 0xe4, 0x1d, 0x4c, 0x9a, 0xce, 0x23, 0x74,
	0x97, 0x1b, 0xdb, 0x6d, 0x0b, 0x98, 0xb6, 0x05, 0x7e, 0xa4, 0xf1, 0x95, 0xcd, 0x6c, 0x33, 0xc4,
	0x27, 0x98, 0xb6, 0x6d, 0x45, 0x7c, 0x37, 0x66, 0x8b, 0x4b, 0xeb, 0x41, 0xdb, 0x8c, 0x38, 0x87,
	0xf1, 0xea, 0x8e, 0xe4, 0xa5, 0xad, 0xdb, 0x70, 0x8f, 0x3f, 0xdb, 0x08, 0x9b, 0xce, 0xbb, 0x81,
	0xf9, 0xcd, 0xbc, 0xfd, 0x17, 0x00, 0x00, 0xff, 0xff, 0x71, 0x6f, 0x12, 0x61, 0x7b, 0x04, 0x00,
	0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// SendAgentClient is the client API for SendAgent service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type SendAgentClient interface {
	Send(ctx context.Context, in *SendParams, opts ...grpc.CallOption) (*Empty, error)
	UpdateConfig(ctx context.Context, in *UpdateConfigParams, opts ...grpc.CallOption) (*Empty, error)
	ValidateConfig(ctx context.Context, in *UpdateConfigParams, opts ...grpc.CallOption) (*ValidateConfigReply, error)
	UseridByMobile(ctx context.Context, in *UseridByMobileParams, opts ...grpc.CallOption) (*UseridByMobileReply, error)
	BatchSend(ctx context.Context, in *BatchSendParams, opts ...grpc.CallOption) (*BatchSendReply, error)
}

type sendAgentClient struct {
	cc *grpc.ClientConn
}

func NewSendAgentClient(cc *grpc.ClientConn) SendAgentClient {
	return &sendAgentClient{cc}
}

func (c *sendAgentClient) Send(ctx context.Context, in *SendParams, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/apis.SendAgent/Send", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sendAgentClient) UpdateConfig(ctx context.Context, in *UpdateConfigParams, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/apis.SendAgent/UpdateConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sendAgentClient) ValidateConfig(ctx context.Context, in *UpdateConfigParams, opts ...grpc.CallOption) (*ValidateConfigReply, error) {
	out := new(ValidateConfigReply)
	err := c.cc.Invoke(ctx, "/apis.SendAgent/ValidateConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sendAgentClient) UseridByMobile(ctx context.Context, in *UseridByMobileParams, opts ...grpc.CallOption) (*UseridByMobileReply, error) {
	out := new(UseridByMobileReply)
	err := c.cc.Invoke(ctx, "/apis.SendAgent/UseridByMobile", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sendAgentClient) BatchSend(ctx context.Context, in *BatchSendParams, opts ...grpc.CallOption) (*BatchSendReply, error) {
	out := new(BatchSendReply)
	err := c.cc.Invoke(ctx, "/apis.SendAgent/BatchSend", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SendAgentServer is the server API for SendAgent service.
type SendAgentServer interface {
	Send(context.Context, *SendParams) (*Empty, error)
	UpdateConfig(context.Context, *UpdateConfigParams) (*Empty, error)
	ValidateConfig(context.Context, *UpdateConfigParams) (*ValidateConfigReply, error)
	UseridByMobile(context.Context, *UseridByMobileParams) (*UseridByMobileReply, error)
	BatchSend(context.Context, *BatchSendParams) (*BatchSendReply, error)
}

// UnimplementedSendAgentServer can be embedded to have forward compatible implementations.
type UnimplementedSendAgentServer struct {
}

func (*UnimplementedSendAgentServer) Send(ctx context.Context, req *SendParams) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Send not implemented")
}
func (*UnimplementedSendAgentServer) UpdateConfig(ctx context.Context, req *UpdateConfigParams) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateConfig not implemented")
}
func (*UnimplementedSendAgentServer) ValidateConfig(ctx context.Context, req *UpdateConfigParams) (*ValidateConfigReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidateConfig not implemented")
}
func (*UnimplementedSendAgentServer) UseridByMobile(ctx context.Context, req *UseridByMobileParams) (*UseridByMobileReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UseridByMobile not implemented")
}
func (*UnimplementedSendAgentServer) BatchSend(ctx context.Context, req *BatchSendParams) (*BatchSendReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BatchSend not implemented")
}

func RegisterSendAgentServer(s *grpc.Server, srv SendAgentServer) {
	s.RegisterService(&_SendAgent_serviceDesc, srv)
}

func _SendAgent_Send_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendParams)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SendAgentServer).Send(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/apis.SendAgent/Send",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SendAgentServer).Send(ctx, req.(*SendParams))
	}
	return interceptor(ctx, in, info, handler)
}

func _SendAgent_UpdateConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateConfigParams)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SendAgentServer).UpdateConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/apis.SendAgent/UpdateConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SendAgentServer).UpdateConfig(ctx, req.(*UpdateConfigParams))
	}
	return interceptor(ctx, in, info, handler)
}

func _SendAgent_ValidateConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateConfigParams)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SendAgentServer).ValidateConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/apis.SendAgent/ValidateConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SendAgentServer).ValidateConfig(ctx, req.(*UpdateConfigParams))
	}
	return interceptor(ctx, in, info, handler)
}

func _SendAgent_UseridByMobile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UseridByMobileParams)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SendAgentServer).UseridByMobile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/apis.SendAgent/UseridByMobile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SendAgentServer).UseridByMobile(ctx, req.(*UseridByMobileParams))
	}
	return interceptor(ctx, in, info, handler)
}

func _SendAgent_BatchSend_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BatchSendParams)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SendAgentServer).BatchSend(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/apis.SendAgent/BatchSend",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SendAgentServer).BatchSend(ctx, req.(*BatchSendParams))
	}
	return interceptor(ctx, in, info, handler)
}

var _SendAgent_serviceDesc = grpc.ServiceDesc{
	ServiceName: "apis.SendAgent",
	HandlerType: (*SendAgentServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Send",
			Handler:    _SendAgent_Send_Handler,
		},
		{
			MethodName: "UpdateConfig",
			Handler:    _SendAgent_UpdateConfig_Handler,
		},
		{
			MethodName: "ValidateConfig",
			Handler:    _SendAgent_ValidateConfig_Handler,
		},
		{
			MethodName: "UseridByMobile",
			Handler:    _SendAgent_UseridByMobile_Handler,
		},
		{
			MethodName: "BatchSend",
			Handler:    _SendAgent_BatchSend_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "send_server.proto",
}
