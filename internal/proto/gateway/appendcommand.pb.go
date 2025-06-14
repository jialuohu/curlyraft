// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.29.3
// source: gateway/appendcommand.proto

package gateway

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type AppendCommandRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Command       []byte                 `protobuf:"bytes,1,opt,name=Command,proto3" json:"Command,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AppendCommandRequest) Reset() {
	*x = AppendCommandRequest{}
	mi := &file_gateway_appendcommand_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AppendCommandRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AppendCommandRequest) ProtoMessage() {}

func (x *AppendCommandRequest) ProtoReflect() protoreflect.Message {
	mi := &file_gateway_appendcommand_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AppendCommandRequest.ProtoReflect.Descriptor instead.
func (*AppendCommandRequest) Descriptor() ([]byte, []int) {
	return file_gateway_appendcommand_proto_rawDescGZIP(), []int{0}
}

func (x *AppendCommandRequest) GetCommand() []byte {
	if x != nil {
		return x.Command
	}
	return nil
}

type AppendCommandResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Applied       bool                   `protobuf:"varint,1,opt,name=Applied,proto3" json:"Applied,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AppendCommandResponse) Reset() {
	*x = AppendCommandResponse{}
	mi := &file_gateway_appendcommand_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AppendCommandResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AppendCommandResponse) ProtoMessage() {}

func (x *AppendCommandResponse) ProtoReflect() protoreflect.Message {
	mi := &file_gateway_appendcommand_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AppendCommandResponse.ProtoReflect.Descriptor instead.
func (*AppendCommandResponse) Descriptor() ([]byte, []int) {
	return file_gateway_appendcommand_proto_rawDescGZIP(), []int{1}
}

func (x *AppendCommandResponse) GetApplied() bool {
	if x != nil {
		return x.Applied
	}
	return false
}

var File_gateway_appendcommand_proto protoreflect.FileDescriptor

const file_gateway_appendcommand_proto_rawDesc = "" +
	"\n" +
	"\x1bgateway/appendcommand.proto\x12\agateway\"0\n" +
	"\x14AppendCommandRequest\x12\x18\n" +
	"\aCommand\x18\x01 \x01(\fR\aCommand\"1\n" +
	"\x15AppendCommandResponse\x12\x18\n" +
	"\aApplied\x18\x01 \x01(\bR\aAppliedB>Z<github.com/jialuohu/curlyraft/internal/proto/gateway;gatewayb\x06proto3"

var (
	file_gateway_appendcommand_proto_rawDescOnce sync.Once
	file_gateway_appendcommand_proto_rawDescData []byte
)

func file_gateway_appendcommand_proto_rawDescGZIP() []byte {
	file_gateway_appendcommand_proto_rawDescOnce.Do(func() {
		file_gateway_appendcommand_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_gateway_appendcommand_proto_rawDesc), len(file_gateway_appendcommand_proto_rawDesc)))
	})
	return file_gateway_appendcommand_proto_rawDescData
}

var file_gateway_appendcommand_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_gateway_appendcommand_proto_goTypes = []any{
	(*AppendCommandRequest)(nil),  // 0: gateway.AppendCommandRequest
	(*AppendCommandResponse)(nil), // 1: gateway.AppendCommandResponse
}
var file_gateway_appendcommand_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_gateway_appendcommand_proto_init() }
func file_gateway_appendcommand_proto_init() {
	if File_gateway_appendcommand_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_gateway_appendcommand_proto_rawDesc), len(file_gateway_appendcommand_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_gateway_appendcommand_proto_goTypes,
		DependencyIndexes: file_gateway_appendcommand_proto_depIdxs,
		MessageInfos:      file_gateway_appendcommand_proto_msgTypes,
	}.Build()
	File_gateway_appendcommand_proto = out.File
	file_gateway_appendcommand_proto_goTypes = nil
	file_gateway_appendcommand_proto_depIdxs = nil
}
