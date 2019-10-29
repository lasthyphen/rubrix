// Code generated by protoc-gen-go. DO NOT EDIT.
// source: services/collection/collection.proto

package collection

import (
	context "context"
	fmt "fmt"
	entities "github.com/dapperlabs/flow-go/proto/sdk/entities"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

type PingRequest struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PingRequest) Reset()         { *m = PingRequest{} }
func (m *PingRequest) String() string { return proto.CompactTextString(m) }
func (*PingRequest) ProtoMessage()    {}
func (*PingRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_0c274ed31f9d8079, []int{0}
}

func (m *PingRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PingRequest.Unmarshal(m, b)
}
func (m *PingRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PingRequest.Marshal(b, m, deterministic)
}
func (m *PingRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PingRequest.Merge(m, src)
}
func (m *PingRequest) XXX_Size() int {
	return xxx_messageInfo_PingRequest.Size(m)
}
func (m *PingRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_PingRequest.DiscardUnknown(m)
}

var xxx_messageInfo_PingRequest proto.InternalMessageInfo

type PingResponse struct {
	Address              []byte   `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PingResponse) Reset()         { *m = PingResponse{} }
func (m *PingResponse) String() string { return proto.CompactTextString(m) }
func (*PingResponse) ProtoMessage()    {}
func (*PingResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_0c274ed31f9d8079, []int{1}
}

func (m *PingResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PingResponse.Unmarshal(m, b)
}
func (m *PingResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PingResponse.Marshal(b, m, deterministic)
}
func (m *PingResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PingResponse.Merge(m, src)
}
func (m *PingResponse) XXX_Size() int {
	return xxx_messageInfo_PingResponse.Size(m)
}
func (m *PingResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_PingResponse.DiscardUnknown(m)
}

var xxx_messageInfo_PingResponse proto.InternalMessageInfo

func (m *PingResponse) GetAddress() []byte {
	if m != nil {
		return m.Address
	}
	return nil
}

type SubmitTransactionRequest struct {
	Transaction          *entities.Transaction `protobuf:"bytes,1,opt,name=transaction,proto3" json:"transaction,omitempty"`
	XXX_NoUnkeyedLiteral struct{}              `json:"-"`
	XXX_unrecognized     []byte                `json:"-"`
	XXX_sizecache        int32                 `json:"-"`
}

func (m *SubmitTransactionRequest) Reset()         { *m = SubmitTransactionRequest{} }
func (m *SubmitTransactionRequest) String() string { return proto.CompactTextString(m) }
func (*SubmitTransactionRequest) ProtoMessage()    {}
func (*SubmitTransactionRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_0c274ed31f9d8079, []int{2}
}

func (m *SubmitTransactionRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SubmitTransactionRequest.Unmarshal(m, b)
}
func (m *SubmitTransactionRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SubmitTransactionRequest.Marshal(b, m, deterministic)
}
func (m *SubmitTransactionRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SubmitTransactionRequest.Merge(m, src)
}
func (m *SubmitTransactionRequest) XXX_Size() int {
	return xxx_messageInfo_SubmitTransactionRequest.Size(m)
}
func (m *SubmitTransactionRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_SubmitTransactionRequest.DiscardUnknown(m)
}

var xxx_messageInfo_SubmitTransactionRequest proto.InternalMessageInfo

func (m *SubmitTransactionRequest) GetTransaction() *entities.Transaction {
	if m != nil {
		return m.Transaction
	}
	return nil
}

type SubmitTransactionResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SubmitTransactionResponse) Reset()         { *m = SubmitTransactionResponse{} }
func (m *SubmitTransactionResponse) String() string { return proto.CompactTextString(m) }
func (*SubmitTransactionResponse) ProtoMessage()    {}
func (*SubmitTransactionResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_0c274ed31f9d8079, []int{3}
}

func (m *SubmitTransactionResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SubmitTransactionResponse.Unmarshal(m, b)
}
func (m *SubmitTransactionResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SubmitTransactionResponse.Marshal(b, m, deterministic)
}
func (m *SubmitTransactionResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SubmitTransactionResponse.Merge(m, src)
}
func (m *SubmitTransactionResponse) XXX_Size() int {
	return xxx_messageInfo_SubmitTransactionResponse.Size(m)
}
func (m *SubmitTransactionResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_SubmitTransactionResponse.DiscardUnknown(m)
}

var xxx_messageInfo_SubmitTransactionResponse proto.InternalMessageInfo

type SubmitCollectionRequest struct {
	Collection           *entities.Collection `protobuf:"bytes,1,opt,name=collection,proto3" json:"collection,omitempty"`
	AccessSignature      []byte               `protobuf:"bytes,2,opt,name=accessSignature,proto3" json:"accessSignature,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *SubmitCollectionRequest) Reset()         { *m = SubmitCollectionRequest{} }
func (m *SubmitCollectionRequest) String() string { return proto.CompactTextString(m) }
func (*SubmitCollectionRequest) ProtoMessage()    {}
func (*SubmitCollectionRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_0c274ed31f9d8079, []int{4}
}

func (m *SubmitCollectionRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SubmitCollectionRequest.Unmarshal(m, b)
}
func (m *SubmitCollectionRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SubmitCollectionRequest.Marshal(b, m, deterministic)
}
func (m *SubmitCollectionRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SubmitCollectionRequest.Merge(m, src)
}
func (m *SubmitCollectionRequest) XXX_Size() int {
	return xxx_messageInfo_SubmitCollectionRequest.Size(m)
}
func (m *SubmitCollectionRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_SubmitCollectionRequest.DiscardUnknown(m)
}

var xxx_messageInfo_SubmitCollectionRequest proto.InternalMessageInfo

func (m *SubmitCollectionRequest) GetCollection() *entities.Collection {
	if m != nil {
		return m.Collection
	}
	return nil
}

func (m *SubmitCollectionRequest) GetAccessSignature() []byte {
	if m != nil {
		return m.AccessSignature
	}
	return nil
}

type GetTransactionRequest struct {
	Hash                 []byte   `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetTransactionRequest) Reset()         { *m = GetTransactionRequest{} }
func (m *GetTransactionRequest) String() string { return proto.CompactTextString(m) }
func (*GetTransactionRequest) ProtoMessage()    {}
func (*GetTransactionRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_0c274ed31f9d8079, []int{5}
}

func (m *GetTransactionRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetTransactionRequest.Unmarshal(m, b)
}
func (m *GetTransactionRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetTransactionRequest.Marshal(b, m, deterministic)
}
func (m *GetTransactionRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetTransactionRequest.Merge(m, src)
}
func (m *GetTransactionRequest) XXX_Size() int {
	return xxx_messageInfo_GetTransactionRequest.Size(m)
}
func (m *GetTransactionRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetTransactionRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetTransactionRequest proto.InternalMessageInfo

func (m *GetTransactionRequest) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

type GetTransactionResponse struct {
	Transaction          *entities.Transaction `protobuf:"bytes,1,opt,name=transaction,proto3" json:"transaction,omitempty"`
	XXX_NoUnkeyedLiteral struct{}              `json:"-"`
	XXX_unrecognized     []byte                `json:"-"`
	XXX_sizecache        int32                 `json:"-"`
}

func (m *GetTransactionResponse) Reset()         { *m = GetTransactionResponse{} }
func (m *GetTransactionResponse) String() string { return proto.CompactTextString(m) }
func (*GetTransactionResponse) ProtoMessage()    {}
func (*GetTransactionResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_0c274ed31f9d8079, []int{6}
}

func (m *GetTransactionResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetTransactionResponse.Unmarshal(m, b)
}
func (m *GetTransactionResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetTransactionResponse.Marshal(b, m, deterministic)
}
func (m *GetTransactionResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetTransactionResponse.Merge(m, src)
}
func (m *GetTransactionResponse) XXX_Size() int {
	return xxx_messageInfo_GetTransactionResponse.Size(m)
}
func (m *GetTransactionResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetTransactionResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetTransactionResponse proto.InternalMessageInfo

func (m *GetTransactionResponse) GetTransaction() *entities.Transaction {
	if m != nil {
		return m.Transaction
	}
	return nil
}

type SubmitCollectionResponse struct {
	AccessSignature      []byte   `protobuf:"bytes,1,opt,name=accessSignature,proto3" json:"accessSignature,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SubmitCollectionResponse) Reset()         { *m = SubmitCollectionResponse{} }
func (m *SubmitCollectionResponse) String() string { return proto.CompactTextString(m) }
func (*SubmitCollectionResponse) ProtoMessage()    {}
func (*SubmitCollectionResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_0c274ed31f9d8079, []int{7}
}

func (m *SubmitCollectionResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SubmitCollectionResponse.Unmarshal(m, b)
}
func (m *SubmitCollectionResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SubmitCollectionResponse.Marshal(b, m, deterministic)
}
func (m *SubmitCollectionResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SubmitCollectionResponse.Merge(m, src)
}
func (m *SubmitCollectionResponse) XXX_Size() int {
	return xxx_messageInfo_SubmitCollectionResponse.Size(m)
}
func (m *SubmitCollectionResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_SubmitCollectionResponse.DiscardUnknown(m)
}

var xxx_messageInfo_SubmitCollectionResponse proto.InternalMessageInfo

func (m *SubmitCollectionResponse) GetAccessSignature() []byte {
	if m != nil {
		return m.AccessSignature
	}
	return nil
}

type GetCollectionRequest struct {
	CollectionHash       []byte   `protobuf:"bytes,1,opt,name=collectionHash,proto3" json:"collectionHash,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetCollectionRequest) Reset()         { *m = GetCollectionRequest{} }
func (m *GetCollectionRequest) String() string { return proto.CompactTextString(m) }
func (*GetCollectionRequest) ProtoMessage()    {}
func (*GetCollectionRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_0c274ed31f9d8079, []int{8}
}

func (m *GetCollectionRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetCollectionRequest.Unmarshal(m, b)
}
func (m *GetCollectionRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetCollectionRequest.Marshal(b, m, deterministic)
}
func (m *GetCollectionRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetCollectionRequest.Merge(m, src)
}
func (m *GetCollectionRequest) XXX_Size() int {
	return xxx_messageInfo_GetCollectionRequest.Size(m)
}
func (m *GetCollectionRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetCollectionRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetCollectionRequest proto.InternalMessageInfo

func (m *GetCollectionRequest) GetCollectionHash() []byte {
	if m != nil {
		return m.CollectionHash
	}
	return nil
}

type GetCollectionResponse struct {
	Collection           *entities.Collection `protobuf:"bytes,1,opt,name=collection,proto3" json:"collection,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *GetCollectionResponse) Reset()         { *m = GetCollectionResponse{} }
func (m *GetCollectionResponse) String() string { return proto.CompactTextString(m) }
func (*GetCollectionResponse) ProtoMessage()    {}
func (*GetCollectionResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_0c274ed31f9d8079, []int{9}
}

func (m *GetCollectionResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetCollectionResponse.Unmarshal(m, b)
}
func (m *GetCollectionResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetCollectionResponse.Marshal(b, m, deterministic)
}
func (m *GetCollectionResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetCollectionResponse.Merge(m, src)
}
func (m *GetCollectionResponse) XXX_Size() int {
	return xxx_messageInfo_GetCollectionResponse.Size(m)
}
func (m *GetCollectionResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetCollectionResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetCollectionResponse proto.InternalMessageInfo

func (m *GetCollectionResponse) GetCollection() *entities.Collection {
	if m != nil {
		return m.Collection
	}
	return nil
}

func init() {
	proto.RegisterType((*PingRequest)(nil), "flow.services.collection.PingRequest")
	proto.RegisterType((*PingResponse)(nil), "flow.services.collection.PingResponse")
	proto.RegisterType((*SubmitTransactionRequest)(nil), "flow.services.collection.SubmitTransactionRequest")
	proto.RegisterType((*SubmitTransactionResponse)(nil), "flow.services.collection.SubmitTransactionResponse")
	proto.RegisterType((*SubmitCollectionRequest)(nil), "flow.services.collection.SubmitCollectionRequest")
	proto.RegisterType((*GetTransactionRequest)(nil), "flow.services.collection.GetTransactionRequest")
	proto.RegisterType((*GetTransactionResponse)(nil), "flow.services.collection.GetTransactionResponse")
	proto.RegisterType((*SubmitCollectionResponse)(nil), "flow.services.collection.SubmitCollectionResponse")
	proto.RegisterType((*GetCollectionRequest)(nil), "flow.services.collection.GetCollectionRequest")
	proto.RegisterType((*GetCollectionResponse)(nil), "flow.services.collection.GetCollectionResponse")
}

func init() {
	proto.RegisterFile("services/collection/collection.proto", fileDescriptor_0c274ed31f9d8079)
}

var fileDescriptor_0c274ed31f9d8079 = []byte{
	// 413 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x54, 0x5f, 0x4b, 0xfb, 0x30,
	0x14, 0xa5, 0x3f, 0xc6, 0x4f, 0xb8, 0xfb, 0xa3, 0x06, 0xff, 0xd4, 0x0e, 0x45, 0x82, 0x8e, 0x81,
	0x90, 0x6a, 0xf7, 0xac, 0x88, 0x0a, 0xf3, 0x51, 0x36, 0xf5, 0x61, 0xf8, 0xd2, 0xb5, 0x71, 0x2b,
	0x9b, 0xed, 0x6c, 0x32, 0x45, 0xf0, 0xc9, 0x2f, 0xe7, 0xd7, 0x12, 0xdb, 0x74, 0xcd, 0xba, 0xd4,
	0x6d, 0xec, 0x2d, 0x4d, 0xee, 0xb9, 0xe7, 0x9e, 0x73, 0x0f, 0x85, 0x23, 0x46, 0xc3, 0x37, 0xcf,
	0xa1, 0xcc, 0x74, 0x82, 0xe1, 0x90, 0x3a, 0xdc, 0x0b, 0x7c, 0xe9, 0x48, 0x46, 0x61, 0xc0, 0x03,
	0xa4, 0x3f, 0x0f, 0x83, 0x77, 0x92, 0x94, 0x92, 0xf4, 0xdd, 0xa8, 0x32, 0x77, 0x60, 0x52, 0x9f,
	0x7b, 0xdc, 0xa3, 0x6c, 0x72, 0x88, 0x61, 0xb8, 0x0c, 0xc5, 0x3b, 0xcf, 0xef, 0xb5, 0xe8, 0xeb,
	0x98, 0x32, 0x8e, 0xeb, 0x50, 0x8a, 0x3f, 0xd9, 0x28, 0xf0, 0x19, 0x45, 0x3a, 0xac, 0xd9, 0xae,
	0x1b, 0x52, 0xc6, 0x74, 0xed, 0x50, 0xab, 0x97, 0x5a, 0xc9, 0x27, 0x7e, 0x02, 0xbd, 0x3d, 0xee,
	0xbe, 0x78, 0xfc, 0x3e, 0xb4, 0x7d, 0x66, 0x47, 0x54, 0xa2, 0x0b, 0xba, 0x84, 0x22, 0x4f, 0x6f,
	0x23, 0x64, 0xd1, 0x3a, 0x20, 0xf1, 0x84, 0xee, 0x80, 0x4c, 0x66, 0x90, 0xb1, 0x32, 0x04, 0x57,
	0x61, 0x4f, 0xd1, 0x3d, 0x1e, 0x0a, 0x7f, 0x69, 0xb0, 0x1b, 0xbf, 0x5e, 0x4f, 0x54, 0x26, 0xd4,
	0xe7, 0x00, 0xa9, 0x74, 0xc1, 0xbc, 0xaf, 0x60, 0x96, 0x90, 0x12, 0x00, 0xd5, 0x61, 0xdd, 0x76,
	0x1c, 0xca, 0x58, 0xdb, 0xeb, 0xf9, 0x36, 0x1f, 0x87, 0x54, 0xff, 0x17, 0xe9, 0xce, 0x5e, 0xe3,
	0x13, 0xd8, 0x6e, 0x52, 0x95, 0x78, 0x04, 0x85, 0xbe, 0xcd, 0xfa, 0xc2, 0xaf, 0xe8, 0x8c, 0x3b,
	0xb0, 0x93, 0x2d, 0x16, 0x06, 0xaf, 0x6e, 0xd5, 0x4d, 0xb2, 0x08, 0xd9, 0x0c, 0xd1, 0x5d, 0x21,
	0x47, 0x53, 0xcb, 0xb9, 0x80, 0xad, 0x26, 0x55, 0xf8, 0x59, 0x83, 0x4a, 0x6a, 0xcf, 0x6d, 0xaa,
	0x2b, 0x73, 0x8b, 0x1f, 0x23, 0x3b, 0x14, 0x23, 0xac, 0xb6, 0x10, 0xeb, 0xbb, 0x00, 0x15, 0xf1,
	0xd4, 0x8e, 0xb3, 0x8d, 0x1e, 0xa0, 0xf0, 0x9b, 0x51, 0x74, 0x4c, 0xf2, 0x22, 0x4f, 0xa4, 0x48,
	0x1b, 0xb5, 0x79, 0x65, 0x62, 0xd0, 0x4f, 0xd8, 0x9c, 0x89, 0x1c, 0xb2, 0xf2, 0xc1, 0x79, 0xe9,
	0x37, 0x1a, 0x4b, 0x61, 0x04, 0xfb, 0x07, 0x6c, 0x64, 0xb7, 0x88, 0xce, 0xe6, 0x35, 0x9a, 0x59,
	0x97, 0x61, 0x2d, 0x03, 0x11, 0xd4, 0x0c, 0x2a, 0xd3, 0xe1, 0x44, 0x66, 0x7e, 0x17, 0x65, 0xe6,
	0x8d, 0xd3, 0xc5, 0x01, 0x82, 0x74, 0x04, 0xe5, 0xa9, 0xbc, 0x20, 0xf2, 0x67, 0x8b, 0x59, 0xa5,
	0xe6, 0xc2, 0xf5, 0x31, 0xe3, 0x55, 0xa9, 0x23, 0xe5, 0xaa, 0xfb, 0x3f, 0xfa, 0xfd, 0x35, 0x7e,
	0x02, 0x00, 0x00, 0xff, 0xff, 0xd2, 0x98, 0x70, 0xe5, 0x5d, 0x05, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// CollectServiceClient is the client API for CollectService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type CollectServiceClient interface {
	Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error)
	SubmitTransaction(ctx context.Context, in *SubmitTransactionRequest, opts ...grpc.CallOption) (*SubmitTransactionResponse, error)
	SubmitCollection(ctx context.Context, in *SubmitCollectionRequest, opts ...grpc.CallOption) (*SubmitCollectionResponse, error)
	GetTransaction(ctx context.Context, in *GetTransactionRequest, opts ...grpc.CallOption) (*GetTransactionResponse, error)
	GetCollection(ctx context.Context, in *GetCollectionRequest, opts ...grpc.CallOption) (*GetCollectionResponse, error)
}

type collectServiceClient struct {
	cc *grpc.ClientConn
}

func NewCollectServiceClient(cc *grpc.ClientConn) CollectServiceClient {
	return &collectServiceClient{cc}
}

func (c *collectServiceClient) Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error) {
	out := new(PingResponse)
	err := c.cc.Invoke(ctx, "/flow.services.collection.CollectService/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *collectServiceClient) SubmitTransaction(ctx context.Context, in *SubmitTransactionRequest, opts ...grpc.CallOption) (*SubmitTransactionResponse, error) {
	out := new(SubmitTransactionResponse)
	err := c.cc.Invoke(ctx, "/flow.services.collection.CollectService/SubmitTransaction", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *collectServiceClient) SubmitCollection(ctx context.Context, in *SubmitCollectionRequest, opts ...grpc.CallOption) (*SubmitCollectionResponse, error) {
	out := new(SubmitCollectionResponse)
	err := c.cc.Invoke(ctx, "/flow.services.collection.CollectService/SubmitCollection", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *collectServiceClient) GetTransaction(ctx context.Context, in *GetTransactionRequest, opts ...grpc.CallOption) (*GetTransactionResponse, error) {
	out := new(GetTransactionResponse)
	err := c.cc.Invoke(ctx, "/flow.services.collection.CollectService/GetTransaction", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *collectServiceClient) GetCollection(ctx context.Context, in *GetCollectionRequest, opts ...grpc.CallOption) (*GetCollectionResponse, error) {
	out := new(GetCollectionResponse)
	err := c.cc.Invoke(ctx, "/flow.services.collection.CollectService/GetCollection", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CollectServiceServer is the server API for CollectService service.
type CollectServiceServer interface {
	Ping(context.Context, *PingRequest) (*PingResponse, error)
	SubmitTransaction(context.Context, *SubmitTransactionRequest) (*SubmitTransactionResponse, error)
	SubmitCollection(context.Context, *SubmitCollectionRequest) (*SubmitCollectionResponse, error)
	GetTransaction(context.Context, *GetTransactionRequest) (*GetTransactionResponse, error)
	GetCollection(context.Context, *GetCollectionRequest) (*GetCollectionResponse, error)
}

// UnimplementedCollectServiceServer can be embedded to have forward compatible implementations.
type UnimplementedCollectServiceServer struct {
}

func (*UnimplementedCollectServiceServer) Ping(ctx context.Context, req *PingRequest) (*PingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (*UnimplementedCollectServiceServer) SubmitTransaction(ctx context.Context, req *SubmitTransactionRequest) (*SubmitTransactionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitTransaction not implemented")
}
func (*UnimplementedCollectServiceServer) SubmitCollection(ctx context.Context, req *SubmitCollectionRequest) (*SubmitCollectionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitCollection not implemented")
}
func (*UnimplementedCollectServiceServer) GetTransaction(ctx context.Context, req *GetTransactionRequest) (*GetTransactionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTransaction not implemented")
}
func (*UnimplementedCollectServiceServer) GetCollection(ctx context.Context, req *GetCollectionRequest) (*GetCollectionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCollection not implemented")
}

func RegisterCollectServiceServer(s *grpc.Server, srv CollectServiceServer) {
	s.RegisterService(&_CollectService_serviceDesc, srv)
}

func _CollectService_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CollectServiceServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/flow.services.collection.CollectService/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CollectServiceServer).Ping(ctx, req.(*PingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CollectService_SubmitTransaction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SubmitTransactionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CollectServiceServer).SubmitTransaction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/flow.services.collection.CollectService/SubmitTransaction",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CollectServiceServer).SubmitTransaction(ctx, req.(*SubmitTransactionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CollectService_SubmitCollection_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SubmitCollectionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CollectServiceServer).SubmitCollection(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/flow.services.collection.CollectService/SubmitCollection",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CollectServiceServer).SubmitCollection(ctx, req.(*SubmitCollectionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CollectService_GetTransaction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTransactionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CollectServiceServer).GetTransaction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/flow.services.collection.CollectService/GetTransaction",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CollectServiceServer).GetTransaction(ctx, req.(*GetTransactionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CollectService_GetCollection_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCollectionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CollectServiceServer).GetCollection(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/flow.services.collection.CollectService/GetCollection",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CollectServiceServer).GetCollection(ctx, req.(*GetCollectionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _CollectService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "flow.services.collection.CollectService",
	HandlerType: (*CollectServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _CollectService_Ping_Handler,
		},
		{
			MethodName: "SubmitTransaction",
			Handler:    _CollectService_SubmitTransaction_Handler,
		},
		{
			MethodName: "SubmitCollection",
			Handler:    _CollectService_SubmitCollection_Handler,
		},
		{
			MethodName: "GetTransaction",
			Handler:    _CollectService_GetTransaction_Handler,
		},
		{
			MethodName: "GetCollection",
			Handler:    _CollectService_GetCollection_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "services/collection/collection.proto",
}