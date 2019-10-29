// Code generated by capnpc-go. DO NOT EDIT.

package captain

import (
	capnp "zombiezen.com/go/capnproto2"
	text "zombiezen.com/go/capnproto2/encoding/text"
	schemas "zombiezen.com/go/capnproto2/schemas"
)

type SnapshotRequest struct{ capnp.Struct }

// SnapshotRequest_TypeID is the unique identifier for the type SnapshotRequest.
const SnapshotRequest_TypeID = 0xa593161ba5bb470b

func NewSnapshotRequest(s *capnp.Segment) (SnapshotRequest, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return SnapshotRequest{st}, err
}

func NewRootSnapshotRequest(s *capnp.Segment) (SnapshotRequest, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return SnapshotRequest{st}, err
}

func ReadRootSnapshotRequest(msg *capnp.Message) (SnapshotRequest, error) {
	root, err := msg.RootPtr()
	return SnapshotRequest{root.Struct()}, err
}

func (s SnapshotRequest) String() string {
	str, _ := text.Marshal(0xa593161ba5bb470b, s.Struct)
	return str
}

func (s SnapshotRequest) Nonce() uint64 {
	return s.Struct.Uint64(0)
}

func (s SnapshotRequest) SetNonce(v uint64) {
	s.Struct.SetUint64(0, v)
}

func (s SnapshotRequest) MempoolHash() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return []byte(p.Data()), err
}

func (s SnapshotRequest) HasMempoolHash() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s SnapshotRequest) SetMempoolHash(v []byte) error {
	return s.Struct.SetData(0, v)
}

// SnapshotRequest_List is a list of SnapshotRequest.
type SnapshotRequest_List struct{ capnp.List }

// NewSnapshotRequest creates a new list of SnapshotRequest.
func NewSnapshotRequest_List(s *capnp.Segment, sz int32) (SnapshotRequest_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1}, sz)
	return SnapshotRequest_List{l}, err
}

func (s SnapshotRequest_List) At(i int) SnapshotRequest { return SnapshotRequest{s.List.Struct(i)} }

func (s SnapshotRequest_List) Set(i int, v SnapshotRequest) error {
	return s.List.SetStruct(i, v.Struct)
}

func (s SnapshotRequest_List) String() string {
	str, _ := text.MarshalList(0xa593161ba5bb470b, s.List)
	return str
}

// SnapshotRequest_Promise is a wrapper for a SnapshotRequest promised by a client call.
type SnapshotRequest_Promise struct{ *capnp.Pipeline }

func (p SnapshotRequest_Promise) Struct() (SnapshotRequest, error) {
	s, err := p.Pipeline.Struct()
	return SnapshotRequest{s}, err
}

type SnapshotResponse struct{ capnp.Struct }

// SnapshotResponse_TypeID is the unique identifier for the type SnapshotResponse.
const SnapshotResponse_TypeID = 0xec87db413a028924

func NewSnapshotResponse(s *capnp.Segment) (SnapshotResponse, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return SnapshotResponse{st}, err
}

func NewRootSnapshotResponse(s *capnp.Segment) (SnapshotResponse, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return SnapshotResponse{st}, err
}

func ReadRootSnapshotResponse(msg *capnp.Message) (SnapshotResponse, error) {
	root, err := msg.RootPtr()
	return SnapshotResponse{root.Struct()}, err
}

func (s SnapshotResponse) String() string {
	str, _ := text.Marshal(0xec87db413a028924, s.Struct)
	return str
}

func (s SnapshotResponse) Nonce() uint64 {
	return s.Struct.Uint64(0)
}

func (s SnapshotResponse) SetNonce(v uint64) {
	s.Struct.SetUint64(0, v)
}

func (s SnapshotResponse) MempoolHash() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return []byte(p.Data()), err
}

func (s SnapshotResponse) HasMempoolHash() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s SnapshotResponse) SetMempoolHash(v []byte) error {
	return s.Struct.SetData(0, v)
}

// SnapshotResponse_List is a list of SnapshotResponse.
type SnapshotResponse_List struct{ capnp.List }

// NewSnapshotResponse creates a new list of SnapshotResponse.
func NewSnapshotResponse_List(s *capnp.Segment, sz int32) (SnapshotResponse_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1}, sz)
	return SnapshotResponse_List{l}, err
}

func (s SnapshotResponse_List) At(i int) SnapshotResponse { return SnapshotResponse{s.List.Struct(i)} }

func (s SnapshotResponse_List) Set(i int, v SnapshotResponse) error {
	return s.List.SetStruct(i, v.Struct)
}

func (s SnapshotResponse_List) String() string {
	str, _ := text.MarshalList(0xec87db413a028924, s.List)
	return str
}

// SnapshotResponse_Promise is a wrapper for a SnapshotResponse promised by a client call.
type SnapshotResponse_Promise struct{ *capnp.Pipeline }

func (p SnapshotResponse_Promise) Struct() (SnapshotResponse, error) {
	s, err := p.Pipeline.Struct()
	return SnapshotResponse{s}, err
}

type MempoolRequest struct{ capnp.Struct }

// MempoolRequest_TypeID is the unique identifier for the type MempoolRequest.
const MempoolRequest_TypeID = 0xa06d8562fac872e6

func NewMempoolRequest(s *capnp.Segment) (MempoolRequest, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return MempoolRequest{st}, err
}

func NewRootMempoolRequest(s *capnp.Segment) (MempoolRequest, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return MempoolRequest{st}, err
}

func ReadRootMempoolRequest(msg *capnp.Message) (MempoolRequest, error) {
	root, err := msg.RootPtr()
	return MempoolRequest{root.Struct()}, err
}

func (s MempoolRequest) String() string {
	str, _ := text.Marshal(0xa06d8562fac872e6, s.Struct)
	return str
}

func (s MempoolRequest) Nonce() uint64 {
	return s.Struct.Uint64(0)
}

func (s MempoolRequest) SetNonce(v uint64) {
	s.Struct.SetUint64(0, v)
}

// MempoolRequest_List is a list of MempoolRequest.
type MempoolRequest_List struct{ capnp.List }

// NewMempoolRequest creates a new list of MempoolRequest.
func NewMempoolRequest_List(s *capnp.Segment, sz int32) (MempoolRequest_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0}, sz)
	return MempoolRequest_List{l}, err
}

func (s MempoolRequest_List) At(i int) MempoolRequest { return MempoolRequest{s.List.Struct(i)} }

func (s MempoolRequest_List) Set(i int, v MempoolRequest) error { return s.List.SetStruct(i, v.Struct) }

func (s MempoolRequest_List) String() string {
	str, _ := text.MarshalList(0xa06d8562fac872e6, s.List)
	return str
}

// MempoolRequest_Promise is a wrapper for a MempoolRequest promised by a client call.
type MempoolRequest_Promise struct{ *capnp.Pipeline }

func (p MempoolRequest_Promise) Struct() (MempoolRequest, error) {
	s, err := p.Pipeline.Struct()
	return MempoolRequest{s}, err
}

type MempoolResponse struct{ capnp.Struct }

// MempoolResponse_TypeID is the unique identifier for the type MempoolResponse.
const MempoolResponse_TypeID = 0xb2741449df5f6582

func NewMempoolResponse(s *capnp.Segment) (MempoolResponse, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return MempoolResponse{st}, err
}

func NewRootMempoolResponse(s *capnp.Segment) (MempoolResponse, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return MempoolResponse{st}, err
}

func ReadRootMempoolResponse(msg *capnp.Message) (MempoolResponse, error) {
	root, err := msg.RootPtr()
	return MempoolResponse{root.Struct()}, err
}

func (s MempoolResponse) String() string {
	str, _ := text.Marshal(0xb2741449df5f6582, s.Struct)
	return str
}

func (s MempoolResponse) Nonce() uint64 {
	return s.Struct.Uint64(0)
}

func (s MempoolResponse) SetNonce(v uint64) {
	s.Struct.SetUint64(0, v)
}

func (s MempoolResponse) Collections() (GuaranteedCollection_List, error) {
	p, err := s.Struct.Ptr(0)
	return GuaranteedCollection_List{List: p.List()}, err
}

func (s MempoolResponse) HasCollections() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s MempoolResponse) SetCollections(v GuaranteedCollection_List) error {
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewCollections sets the collections field to a newly
// allocated GuaranteedCollection_List, preferring placement in s's segment.
func (s MempoolResponse) NewCollections(n int32) (GuaranteedCollection_List, error) {
	l, err := NewGuaranteedCollection_List(s.Struct.Segment(), n)
	if err != nil {
		return GuaranteedCollection_List{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

// MempoolResponse_List is a list of MempoolResponse.
type MempoolResponse_List struct{ capnp.List }

// NewMempoolResponse creates a new list of MempoolResponse.
func NewMempoolResponse_List(s *capnp.Segment, sz int32) (MempoolResponse_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1}, sz)
	return MempoolResponse_List{l}, err
}

func (s MempoolResponse_List) At(i int) MempoolResponse { return MempoolResponse{s.List.Struct(i)} }

func (s MempoolResponse_List) Set(i int, v MempoolResponse) error {
	return s.List.SetStruct(i, v.Struct)
}

func (s MempoolResponse_List) String() string {
	str, _ := text.MarshalList(0xb2741449df5f6582, s.List)
	return str
}

// MempoolResponse_Promise is a wrapper for a MempoolResponse promised by a client call.
type MempoolResponse_Promise struct{ *capnp.Pipeline }

func (p MempoolResponse_Promise) Struct() (MempoolResponse, error) {
	s, err := p.Pipeline.Struct()
	return MempoolResponse{s}, err
}

const schema_9691a973e8a27b6b = "x\xda2\x88ct`2d\xedgc`\x08\xf4a" +
	"e\xfb\xff\xac\xe8\xc4\xaf\xa4\xd6\xdc\x05\x0c\x81\x02\x8c\x8c" +
	"\xff\xb3\xab\x17\xbd(^9q\x1a\x03\x0b;\x03\x83\xb0" +
	".\xf3/aKf\x10\xcb\x94\xd9\x9e\x81\xf1?\xb7\xfb" +
	"\xee\xa5\xd2b\x93\x97\xa2\xa9eedg`0\x0ee" +
	"fb\x14N\x04\xab\x8ee.g`\xfc\xdf\x94\x1a\x7f" +
	"\xdfS\xa4d\x13V\xd5{A\xaaO\x82U\x1f\x05\xab" +
	"V\xe9d\xb2r\xbc\xdd\xfe\x06\xabjS\x16.Fa" +
	"W\xb0\x9b\x1cY\xca\x19\x96\xfdO\xce\xcf+N\xcd+" +
	".e*\xd6KN,\xc8+\xb0\xf2M\xcd-\xc8\xcf" +
	"\xcf\x09\xb2O-,M-.\x09`d\x0cdaf" +
	"a``ad`\x10\xe45b`\x08\xe4`f\x0c" +
	"\x14ab\x94\xcf\xcb\xcfKNe\xe4d`b\xe4d" +
	"`\x84\x1b\xc4\x0c3(8/\xb1\xa08#\xbf$\x08" +
	"b\x10\x03\xc8$\x0e\xb8I\x9a \x93T\x98\x19\x03\x0d" +
	"\x98\x18\x19\x19E\x18Ab\xbaI\x0c\x0c\x81:\xcc\x8c" +
	"\x81\x16\x18\xa6\xe7B\\\xe5\xc1\xc0\x9eX\x9c\xc1\xc8\xcb" +
	"\xc0\xc4\xc8\x8b\xcdN\x98\xe3S\x8b\x0b@R\xa4\xd8\xe9" +
	"\x81\xc5G99\xa9\xc9%\x99\x0c\xec\xf9y\xc5\x8c|" +
	"\x0c\x8c\x01\xcc\x8c\x8c\x02\xff\xa5\xe7\xb04\xad]\x15w" +
	"\x84\x81\x81\x11$\x88\xcf\xe3PW\xd0\xc0\xeb\x80\x00\x00" +
	"\x00\xff\xff2&\x9a\xe6"

func init() {
	schemas.Register(schema_9691a973e8a27b6b,
		0xa06d8562fac872e6,
		0xa593161ba5bb470b,
		0xb2741449df5f6582,
		0xec87db413a028924)
}