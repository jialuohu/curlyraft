// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.29.3
// source: raftcomm/requestvote.proto

package raftcomm

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

type RequestVoteRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Candidate’s current term.
	Term uint32 `protobuf:"varint,1,opt,name=Term,proto3" json:"Term,omitempty"`
	// Candidate’s ID requesting vote.
	CandidateId string `protobuf:"bytes,2,opt,name=CandidateId,proto3" json:"CandidateId,omitempty"`
	// Index of candidate’s last log entry.
	LastLogIndex uint32 `protobuf:"varint,3,opt,name=LastLogIndex,proto3" json:"LastLogIndex,omitempty"`
	// Term of candidate’s last log entry.
	LastLogTerm   uint32 `protobuf:"varint,4,opt,name=LastLogTerm,proto3" json:"LastLogTerm,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RequestVoteRequest) Reset() {
	*x = RequestVoteRequest{}
	mi := &file_raftcomm_requestvote_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RequestVoteRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RequestVoteRequest) ProtoMessage() {}

func (x *RequestVoteRequest) ProtoReflect() protoreflect.Message {
	mi := &file_raftcomm_requestvote_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RequestVoteRequest.ProtoReflect.Descriptor instead.
func (*RequestVoteRequest) Descriptor() ([]byte, []int) {
	return file_raftcomm_requestvote_proto_rawDescGZIP(), []int{0}
}

func (x *RequestVoteRequest) GetTerm() uint32 {
	if x != nil {
		return x.Term
	}
	return 0
}

func (x *RequestVoteRequest) GetCandidateId() string {
	if x != nil {
		return x.CandidateId
	}
	return ""
}

func (x *RequestVoteRequest) GetLastLogIndex() uint32 {
	if x != nil {
		return x.LastLogIndex
	}
	return 0
}

func (x *RequestVoteRequest) GetLastLogTerm() uint32 {
	if x != nil {
		return x.LastLogTerm
	}
	return 0
}

type RequestVoteResponse struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Follower’s current term (so candidate can update itself if needed).
	Term uint32 `protobuf:"varint,1,opt,name=Term,proto3" json:"Term,omitempty"`
	// True means follower received and granted its vote to the candidate.
	VoteGranted   bool `protobuf:"varint,2,opt,name=VoteGranted,proto3" json:"VoteGranted,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RequestVoteResponse) Reset() {
	*x = RequestVoteResponse{}
	mi := &file_raftcomm_requestvote_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RequestVoteResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RequestVoteResponse) ProtoMessage() {}

func (x *RequestVoteResponse) ProtoReflect() protoreflect.Message {
	mi := &file_raftcomm_requestvote_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RequestVoteResponse.ProtoReflect.Descriptor instead.
func (*RequestVoteResponse) Descriptor() ([]byte, []int) {
	return file_raftcomm_requestvote_proto_rawDescGZIP(), []int{1}
}

func (x *RequestVoteResponse) GetTerm() uint32 {
	if x != nil {
		return x.Term
	}
	return 0
}

func (x *RequestVoteResponse) GetVoteGranted() bool {
	if x != nil {
		return x.VoteGranted
	}
	return false
}

var File_raftcomm_requestvote_proto protoreflect.FileDescriptor

const file_raftcomm_requestvote_proto_rawDesc = "" +
	"\n" +
	"\x1araftcomm/requestvote.proto\x12\braftcomm\"\x90\x01\n" +
	"\x12RequestVoteRequest\x12\x12\n" +
	"\x04Term\x18\x01 \x01(\rR\x04Term\x12 \n" +
	"\vCandidateId\x18\x02 \x01(\tR\vCandidateId\x12\"\n" +
	"\fLastLogIndex\x18\x03 \x01(\rR\fLastLogIndex\x12 \n" +
	"\vLastLogTerm\x18\x04 \x01(\rR\vLastLogTerm\"K\n" +
	"\x13RequestVoteResponse\x12\x12\n" +
	"\x04Term\x18\x01 \x01(\rR\x04Term\x12 \n" +
	"\vVoteGranted\x18\x02 \x01(\bR\vVoteGrantedB@Z>github.com/jialuohu/curlyraft/internal/proto/raftcomm;raftcommb\x06proto3"

var (
	file_raftcomm_requestvote_proto_rawDescOnce sync.Once
	file_raftcomm_requestvote_proto_rawDescData []byte
)

func file_raftcomm_requestvote_proto_rawDescGZIP() []byte {
	file_raftcomm_requestvote_proto_rawDescOnce.Do(func() {
		file_raftcomm_requestvote_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_raftcomm_requestvote_proto_rawDesc), len(file_raftcomm_requestvote_proto_rawDesc)))
	})
	return file_raftcomm_requestvote_proto_rawDescData
}

var file_raftcomm_requestvote_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_raftcomm_requestvote_proto_goTypes = []any{
	(*RequestVoteRequest)(nil),  // 0: raftcomm.RequestVoteRequest
	(*RequestVoteResponse)(nil), // 1: raftcomm.RequestVoteResponse
}
var file_raftcomm_requestvote_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_raftcomm_requestvote_proto_init() }
func file_raftcomm_requestvote_proto_init() {
	if File_raftcomm_requestvote_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_raftcomm_requestvote_proto_rawDesc), len(file_raftcomm_requestvote_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_raftcomm_requestvote_proto_goTypes,
		DependencyIndexes: file_raftcomm_requestvote_proto_depIdxs,
		MessageInfos:      file_raftcomm_requestvote_proto_msgTypes,
	}.Build()
	File_raftcomm_requestvote_proto = out.File
	file_raftcomm_requestvote_proto_goTypes = nil
	file_raftcomm_requestvote_proto_depIdxs = nil
}
