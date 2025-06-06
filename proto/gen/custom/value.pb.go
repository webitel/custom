// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.27.3
// source: custom/value.proto

package custompb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	structpb "google.golang.org/protobuf/types/known/structpb"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Lookup struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Required. Unique Identifier.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// Readonly. Display name.
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	// Optional. Reference type.
	Type string `protobuf:"bytes,3,opt,name=type,proto3" json:"type,omitempty"`
}

func (x *Lookup) Reset() {
	*x = Lookup{}
	if protoimpl.UnsafeEnabled {
		mi := &file_custom_value_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Lookup) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Lookup) ProtoMessage() {}

func (x *Lookup) ProtoReflect() protoreflect.Message {
	mi := &file_custom_value_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Lookup.ProtoReflect.Descriptor instead.
func (*Lookup) Descriptor() ([]byte, []int) {
	return file_custom_value_proto_rawDescGZIP(), []int{0}
}

func (x *Lookup) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Lookup) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Lookup) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

// `List` is a wrapper around a repeated field of values.
//
// The JSON representation for `List` is JSON array.
type List struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Repeated field of dynamically typed values.
	Values []*Value `protobuf:"bytes,1,rep,name=values,proto3" json:"values,omitempty"`
}

func (x *List) Reset() {
	*x = List{}
	if protoimpl.UnsafeEnabled {
		mi := &file_custom_value_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *List) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*List) ProtoMessage() {}

func (x *List) ProtoReflect() protoreflect.Message {
	mi := &file_custom_value_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use List.ProtoReflect.Descriptor instead.
func (*List) Descriptor() ([]byte, []int) {
	return file_custom_value_proto_rawDescGZIP(), []int{1}
}

func (x *List) GetValues() []*Value {
	if x != nil {
		return x.Values
	}
	return nil
}

// `Value` represents a dynamically typed value which can be either
// null, a number, a string, a boolean, a recursive struct value, or a
// list of values. A producer of value is expected to set one of these
// variants. Absence of any variant indicates an error.
//
// The JSON representation for `Value` is JSON value.
type Value struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The kind of value.
	//
	// Types that are assignable to Kind:
	//
	//	*Value_Null
	//	*Value_Bool
	//	*Value_Int32
	//	*Value_Int64
	//	*Value_Uint32
	//	*Value_Uint64
	//	*Value_Float32
	//	*Value_Float64
	//	*Value_Datetime
	//	*Value_String_
	//	*Value_Binary
	//	*Value_Lookup
	//	*Value_List
	Kind isValue_Kind `protobuf_oneof:"kind"`
}

func (x *Value) Reset() {
	*x = Value{}
	if protoimpl.UnsafeEnabled {
		mi := &file_custom_value_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Value) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Value) ProtoMessage() {}

func (x *Value) ProtoReflect() protoreflect.Message {
	mi := &file_custom_value_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Value.ProtoReflect.Descriptor instead.
func (*Value) Descriptor() ([]byte, []int) {
	return file_custom_value_proto_rawDescGZIP(), []int{2}
}

func (m *Value) GetKind() isValue_Kind {
	if m != nil {
		return m.Kind
	}
	return nil
}

func (x *Value) GetNull() structpb.NullValue {
	if x, ok := x.GetKind().(*Value_Null); ok {
		return x.Null
	}
	return structpb.NullValue(0)
}

func (x *Value) GetBool() *wrapperspb.BoolValue {
	if x, ok := x.GetKind().(*Value_Bool); ok {
		return x.Bool
	}
	return nil
}

func (x *Value) GetInt32() *wrapperspb.Int32Value {
	if x, ok := x.GetKind().(*Value_Int32); ok {
		return x.Int32
	}
	return nil
}

func (x *Value) GetInt64() *wrapperspb.Int64Value {
	if x, ok := x.GetKind().(*Value_Int64); ok {
		return x.Int64
	}
	return nil
}

func (x *Value) GetUint32() *wrapperspb.UInt32Value {
	if x, ok := x.GetKind().(*Value_Uint32); ok {
		return x.Uint32
	}
	return nil
}

func (x *Value) GetUint64() *wrapperspb.UInt64Value {
	if x, ok := x.GetKind().(*Value_Uint64); ok {
		return x.Uint64
	}
	return nil
}

func (x *Value) GetFloat32() *wrapperspb.FloatValue {
	if x, ok := x.GetKind().(*Value_Float32); ok {
		return x.Float32
	}
	return nil
}

func (x *Value) GetFloat64() *wrapperspb.DoubleValue {
	if x, ok := x.GetKind().(*Value_Float64); ok {
		return x.Float64
	}
	return nil
}

func (x *Value) GetDatetime() *wrapperspb.DoubleValue {
	if x, ok := x.GetKind().(*Value_Datetime); ok {
		return x.Datetime
	}
	return nil
}

func (x *Value) GetString_() *wrapperspb.StringValue {
	if x, ok := x.GetKind().(*Value_String_); ok {
		return x.String_
	}
	return nil
}

func (x *Value) GetBinary() *wrapperspb.BytesValue {
	if x, ok := x.GetKind().(*Value_Binary); ok {
		return x.Binary
	}
	return nil
}

func (x *Value) GetLookup() *Lookup {
	if x, ok := x.GetKind().(*Value_Lookup); ok {
		return x.Lookup
	}
	return nil
}

func (x *Value) GetList() *List {
	if x, ok := x.GetKind().(*Value_List); ok {
		return x.List
	}
	return nil
}

type isValue_Kind interface {
	isValue_Kind()
}

type Value_Null struct {
	// Represents a null value.
	Null structpb.NullValue `protobuf:"varint,1,opt,name=null,proto3,enum=google.protobuf.NullValue,oneof"`
}

type Value_Bool struct {
	// Represents a boolean value.
	Bool *wrapperspb.BoolValue `protobuf:"bytes,2,opt,name=bool,proto3,oneof"`
}

type Value_Int32 struct {
	// Represents a signed integer value.
	Int32 *wrapperspb.Int32Value `protobuf:"bytes,3,opt,name=int32,proto3,oneof"`
}

type Value_Int64 struct {
	Int64 *wrapperspb.Int64Value `protobuf:"bytes,4,opt,name=int64,proto3,oneof"`
}

type Value_Uint32 struct {
	Uint32 *wrapperspb.UInt32Value `protobuf:"bytes,5,opt,name=uint32,proto3,oneof"`
}

type Value_Uint64 struct {
	Uint64 *wrapperspb.UInt64Value `protobuf:"bytes,6,opt,name=uint64,proto3,oneof"`
}

type Value_Float32 struct {
	Float32 *wrapperspb.FloatValue `protobuf:"bytes,7,opt,name=float32,proto3,oneof"`
}

type Value_Float64 struct {
	Float64 *wrapperspb.DoubleValue `protobuf:"bytes,8,opt,name=float64,proto3,oneof"`
}

type Value_Datetime struct {
	Datetime *wrapperspb.DoubleValue `protobuf:"bytes,9,opt,name=datetime,proto3,oneof"`
}

type Value_String_ struct {
	String_ *wrapperspb.StringValue `protobuf:"bytes,10,opt,name=string,proto3,oneof"`
}

type Value_Binary struct {
	Binary *wrapperspb.BytesValue `protobuf:"bytes,11,opt,name=binary,proto3,oneof"`
}

type Value_Lookup struct {
	Lookup *Lookup `protobuf:"bytes,12,opt,name=lookup,proto3,oneof"`
}

type Value_List struct {
	List *List `protobuf:"bytes,13,opt,name=list,proto3,oneof"`
}

func (*Value_Null) isValue_Kind() {}

func (*Value_Bool) isValue_Kind() {}

func (*Value_Int32) isValue_Kind() {}

func (*Value_Int64) isValue_Kind() {}

func (*Value_Uint32) isValue_Kind() {}

func (*Value_Uint64) isValue_Kind() {}

func (*Value_Float32) isValue_Kind() {}

func (*Value_Float64) isValue_Kind() {}

func (*Value_Datetime) isValue_Kind() {}

func (*Value_String_) isValue_Kind() {}

func (*Value_Binary) isValue_Kind() {}

func (*Value_Lookup) isValue_Kind() {}

func (*Value_List) isValue_Kind() {}

var File_custom_value_proto protoreflect.FileDescriptor

var file_custom_value_proto_rawDesc = []byte{
	0x0a, 0x12, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x2f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0e, 0x77, 0x65, 0x62, 0x69, 0x74, 0x65, 0x6c, 0x2e, 0x63, 0x75,
	0x73, 0x74, 0x6f, 0x6d, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2f, 0x77, 0x72, 0x61, 0x70, 0x70, 0x65, 0x72, 0x73, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x22, 0x40, 0x0a, 0x06, 0x4c, 0x6f, 0x6f, 0x6b, 0x75, 0x70, 0x12, 0x0e, 0x0a, 0x02,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x12, 0x12, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x74, 0x79, 0x70, 0x65, 0x22, 0x35, 0x0a, 0x04, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x2d, 0x0a, 0x06,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x77,
	0x65, 0x62, 0x69, 0x74, 0x65, 0x6c, 0x2e, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x2e, 0x56, 0x61,
	0x6c, 0x75, 0x65, 0x52, 0x06, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x22, 0xc9, 0x05, 0x0a, 0x05,
	0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x30, 0x0a, 0x04, 0x6e, 0x75, 0x6c, 0x6c, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0e, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x4e, 0x75, 0x6c, 0x6c, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x48,
	0x00, 0x52, 0x04, 0x6e, 0x75, 0x6c, 0x6c, 0x12, 0x30, 0x0a, 0x04, 0x62, 0x6f, 0x6f, 0x6c, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x42, 0x6f, 0x6f, 0x6c, 0x56, 0x61, 0x6c, 0x75,
	0x65, 0x48, 0x00, 0x52, 0x04, 0x62, 0x6f, 0x6f, 0x6c, 0x12, 0x33, 0x0a, 0x05, 0x69, 0x6e, 0x74,
	0x33, 0x32, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x49, 0x6e, 0x74, 0x33, 0x32,
	0x56, 0x61, 0x6c, 0x75, 0x65, 0x48, 0x00, 0x52, 0x05, 0x69, 0x6e, 0x74, 0x33, 0x32, 0x12, 0x33,
	0x0a, 0x05, 0x69, 0x6e, 0x74, 0x36, 0x34, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e,
	0x49, 0x6e, 0x74, 0x36, 0x34, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x48, 0x00, 0x52, 0x05, 0x69, 0x6e,
	0x74, 0x36, 0x34, 0x12, 0x36, 0x0a, 0x06, 0x75, 0x69, 0x6e, 0x74, 0x33, 0x32, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x55, 0x49, 0x6e, 0x74, 0x33, 0x32, 0x56, 0x61, 0x6c, 0x75,
	0x65, 0x48, 0x00, 0x52, 0x06, 0x75, 0x69, 0x6e, 0x74, 0x33, 0x32, 0x12, 0x36, 0x0a, 0x06, 0x75,
	0x69, 0x6e, 0x74, 0x36, 0x34, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x55, 0x49,
	0x6e, 0x74, 0x36, 0x34, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x48, 0x00, 0x52, 0x06, 0x75, 0x69, 0x6e,
	0x74, 0x36, 0x34, 0x12, 0x37, 0x0a, 0x07, 0x66, 0x6c, 0x6f, 0x61, 0x74, 0x33, 0x32, 0x18, 0x07,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x46, 0x6c, 0x6f, 0x61, 0x74, 0x56, 0x61, 0x6c, 0x75,
	0x65, 0x48, 0x00, 0x52, 0x07, 0x66, 0x6c, 0x6f, 0x61, 0x74, 0x33, 0x32, 0x12, 0x38, 0x0a, 0x07,
	0x66, 0x6c, 0x6f, 0x61, 0x74, 0x36, 0x34, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e,
	0x44, 0x6f, 0x75, 0x62, 0x6c, 0x65, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x48, 0x00, 0x52, 0x07, 0x66,
	0x6c, 0x6f, 0x61, 0x74, 0x36, 0x34, 0x12, 0x3a, 0x0a, 0x08, 0x64, 0x61, 0x74, 0x65, 0x74, 0x69,
	0x6d, 0x65, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44, 0x6f, 0x75, 0x62, 0x6c,
	0x65, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x48, 0x00, 0x52, 0x08, 0x64, 0x61, 0x74, 0x65, 0x74, 0x69,
	0x6d, 0x65, 0x12, 0x36, 0x0a, 0x06, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x18, 0x0a, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x56, 0x61, 0x6c, 0x75, 0x65,
	0x48, 0x00, 0x52, 0x06, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x12, 0x35, 0x0a, 0x06, 0x62, 0x69,
	0x6e, 0x61, 0x72, 0x79, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x42, 0x79, 0x74,
	0x65, 0x73, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x48, 0x00, 0x52, 0x06, 0x62, 0x69, 0x6e, 0x61, 0x72,
	0x79, 0x12, 0x30, 0x0a, 0x06, 0x6c, 0x6f, 0x6f, 0x6b, 0x75, 0x70, 0x18, 0x0c, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x16, 0x2e, 0x77, 0x65, 0x62, 0x69, 0x74, 0x65, 0x6c, 0x2e, 0x63, 0x75, 0x73, 0x74,
	0x6f, 0x6d, 0x2e, 0x4c, 0x6f, 0x6f, 0x6b, 0x75, 0x70, 0x48, 0x00, 0x52, 0x06, 0x6c, 0x6f, 0x6f,
	0x6b, 0x75, 0x70, 0x12, 0x2a, 0x0a, 0x04, 0x6c, 0x69, 0x73, 0x74, 0x18, 0x0d, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x14, 0x2e, 0x77, 0x65, 0x62, 0x69, 0x74, 0x65, 0x6c, 0x2e, 0x63, 0x75, 0x73, 0x74,
	0x6f, 0x6d, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x48, 0x00, 0x52, 0x04, 0x6c, 0x69, 0x73, 0x74, 0x42,
	0x06, 0x0a, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x42, 0x2e, 0x5a, 0x2c, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x77, 0x65, 0x62, 0x69, 0x74, 0x65, 0x6c, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x3b, 0x63,
	0x75, 0x73, 0x74, 0x6f, 0x6d, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_custom_value_proto_rawDescOnce sync.Once
	file_custom_value_proto_rawDescData = file_custom_value_proto_rawDesc
)

func file_custom_value_proto_rawDescGZIP() []byte {
	file_custom_value_proto_rawDescOnce.Do(func() {
		file_custom_value_proto_rawDescData = protoimpl.X.CompressGZIP(file_custom_value_proto_rawDescData)
	})
	return file_custom_value_proto_rawDescData
}

var file_custom_value_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_custom_value_proto_goTypes = []any{
	(*Lookup)(nil),                 // 0: webitel.custom.Lookup
	(*List)(nil),                   // 1: webitel.custom.List
	(*Value)(nil),                  // 2: webitel.custom.Value
	(structpb.NullValue)(0),        // 3: google.protobuf.NullValue
	(*wrapperspb.BoolValue)(nil),   // 4: google.protobuf.BoolValue
	(*wrapperspb.Int32Value)(nil),  // 5: google.protobuf.Int32Value
	(*wrapperspb.Int64Value)(nil),  // 6: google.protobuf.Int64Value
	(*wrapperspb.UInt32Value)(nil), // 7: google.protobuf.UInt32Value
	(*wrapperspb.UInt64Value)(nil), // 8: google.protobuf.UInt64Value
	(*wrapperspb.FloatValue)(nil),  // 9: google.protobuf.FloatValue
	(*wrapperspb.DoubleValue)(nil), // 10: google.protobuf.DoubleValue
	(*wrapperspb.StringValue)(nil), // 11: google.protobuf.StringValue
	(*wrapperspb.BytesValue)(nil),  // 12: google.protobuf.BytesValue
}
var file_custom_value_proto_depIdxs = []int32{
	2,  // 0: webitel.custom.List.values:type_name -> webitel.custom.Value
	3,  // 1: webitel.custom.Value.null:type_name -> google.protobuf.NullValue
	4,  // 2: webitel.custom.Value.bool:type_name -> google.protobuf.BoolValue
	5,  // 3: webitel.custom.Value.int32:type_name -> google.protobuf.Int32Value
	6,  // 4: webitel.custom.Value.int64:type_name -> google.protobuf.Int64Value
	7,  // 5: webitel.custom.Value.uint32:type_name -> google.protobuf.UInt32Value
	8,  // 6: webitel.custom.Value.uint64:type_name -> google.protobuf.UInt64Value
	9,  // 7: webitel.custom.Value.float32:type_name -> google.protobuf.FloatValue
	10, // 8: webitel.custom.Value.float64:type_name -> google.protobuf.DoubleValue
	10, // 9: webitel.custom.Value.datetime:type_name -> google.protobuf.DoubleValue
	11, // 10: webitel.custom.Value.string:type_name -> google.protobuf.StringValue
	12, // 11: webitel.custom.Value.binary:type_name -> google.protobuf.BytesValue
	0,  // 12: webitel.custom.Value.lookup:type_name -> webitel.custom.Lookup
	1,  // 13: webitel.custom.Value.list:type_name -> webitel.custom.List
	14, // [14:14] is the sub-list for method output_type
	14, // [14:14] is the sub-list for method input_type
	14, // [14:14] is the sub-list for extension type_name
	14, // [14:14] is the sub-list for extension extendee
	0,  // [0:14] is the sub-list for field type_name
}

func init() { file_custom_value_proto_init() }
func file_custom_value_proto_init() {
	if File_custom_value_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_custom_value_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*Lookup); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_custom_value_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*List); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_custom_value_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*Value); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_custom_value_proto_msgTypes[2].OneofWrappers = []any{
		(*Value_Null)(nil),
		(*Value_Bool)(nil),
		(*Value_Int32)(nil),
		(*Value_Int64)(nil),
		(*Value_Uint32)(nil),
		(*Value_Uint64)(nil),
		(*Value_Float32)(nil),
		(*Value_Float64)(nil),
		(*Value_Datetime)(nil),
		(*Value_String_)(nil),
		(*Value_Binary)(nil),
		(*Value_Lookup)(nil),
		(*Value_List)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_custom_value_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_custom_value_proto_goTypes,
		DependencyIndexes: file_custom_value_proto_depIdxs,
		MessageInfos:      file_custom_value_proto_msgTypes,
	}.Build()
	File_custom_value_proto = out.File
	file_custom_value_proto_rawDesc = nil
	file_custom_value_proto_goTypes = nil
	file_custom_value_proto_depIdxs = nil
}
