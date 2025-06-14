// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.29.3
// source: raftcomm/appendentries.proto

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

type AppendEntriesRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Leader’s term.
	Term uint32 `protobuf:"varint,1,opt,name=Term,proto3" json:"Term,omitempty"`
	// Leader’s ID so follower can redirect clients if necessary.
	LeaderId string `protobuf:"bytes,2,opt,name=LeaderId,proto3" json:"LeaderId,omitempty"`
	// Index of log entry immediately preceding new ones.
	PrevLogIndex uint32 `protobuf:"varint,3,opt,name=PrevLogIndex,proto3" json:"PrevLogIndex,omitempty"`
	// Term of the entry at PrevLogIndex.
	PrevLogTerm uint32 `protobuf:"varint,4,opt,name=PrevLogTerm,proto3" json:"PrevLogTerm,omitempty"`
	// Log entries to store on the follower.
	// - Empty for heartbeat
	// - May contain multiple entries for efficiency.
	Entries []*LogEntry `protobuf:"bytes,5,rep,name=Entries,proto3" json:"Entries,omitempty"`
	// Leader’s commitIndex: once follower’s log is updated, it can advance
	// its commit index up to this value.
	LeaderCommit  uint32 `protobuf:"varint,6,opt,name=LeaderCommit,proto3" json:"LeaderCommit,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AppendEntriesRequest) Reset() {
	*x = AppendEntriesRequest{}
	mi := &file_raftcomm_appendentries_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AppendEntriesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AppendEntriesRequest) ProtoMessage() {}

func (x *AppendEntriesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_raftcomm_appendentries_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AppendEntriesRequest.ProtoReflect.Descriptor instead.
func (*AppendEntriesRequest) Descriptor() ([]byte, []int) {
	return file_raftcomm_appendentries_proto_rawDescGZIP(), []int{0}
}

func (x *AppendEntriesRequest) GetTerm() uint32 {
	if x != nil {
		return x.Term
	}
	return 0
}

func (x *AppendEntriesRequest) GetLeaderId() string {
	if x != nil {
		return x.LeaderId
	}
	return ""
}

func (x *AppendEntriesRequest) GetPrevLogIndex() uint32 {
	if x != nil {
		return x.PrevLogIndex
	}
	return 0
}

func (x *AppendEntriesRequest) GetPrevLogTerm() uint32 {
	if x != nil {
		return x.PrevLogTerm
	}
	return 0
}

func (x *AppendEntriesRequest) GetEntries() []*LogEntry {
	if x != nil {
		return x.Entries
	}
	return nil
}

func (x *AppendEntriesRequest) GetLeaderCommit() uint32 {
	if x != nil {
		return x.LeaderCommit
	}
	return 0
}

type AppendEntriesResponse struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Follower’s currentTerm (so leader can update itself if needed).
	Term uint32 `protobuf:"varint,1,opt,name=Term,proto3" json:"Term,omitempty"`
	// True if follower’s log contained an entry at PrevLogIndex whose
	// term matched PrevLogTerm. In that case, the follower appends any new
	// entries and advances its commit index; otherwise, it rejects the append.
	Success       bool `protobuf:"varint,2,opt,name=Success,proto3" json:"Success,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AppendEntriesResponse) Reset() {
	*x = AppendEntriesResponse{}
	mi := &file_raftcomm_appendentries_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AppendEntriesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AppendEntriesResponse) ProtoMessage() {}

func (x *AppendEntriesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_raftcomm_appendentries_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AppendEntriesResponse.ProtoReflect.Descriptor instead.
func (*AppendEntriesResponse) Descriptor() ([]byte, []int) {
	return file_raftcomm_appendentries_proto_rawDescGZIP(), []int{1}
}

func (x *AppendEntriesResponse) GetTerm() uint32 {
	if x != nil {
		return x.Term
	}
	return 0
}

func (x *AppendEntriesResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

type LogEntry struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Command       []byte                 `protobuf:"bytes,1,opt,name=Command,proto3" json:"Command,omitempty"`
	LogTerm       uint32                 `protobuf:"varint,2,opt,name=LogTerm,proto3" json:"LogTerm,omitempty"`
	LogIndex      uint32                 `protobuf:"varint,3,opt,name=LogIndex,proto3" json:"LogIndex,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *LogEntry) Reset() {
	*x = LogEntry{}
	mi := &file_raftcomm_appendentries_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *LogEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LogEntry) ProtoMessage() {}

func (x *LogEntry) ProtoReflect() protoreflect.Message {
	mi := &file_raftcomm_appendentries_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LogEntry.ProtoReflect.Descriptor instead.
func (*LogEntry) Descriptor() ([]byte, []int) {
	return file_raftcomm_appendentries_proto_rawDescGZIP(), []int{2}
}

func (x *LogEntry) GetCommand() []byte {
	if x != nil {
		return x.Command
	}
	return nil
}

func (x *LogEntry) GetLogTerm() uint32 {
	if x != nil {
		return x.LogTerm
	}
	return 0
}

func (x *LogEntry) GetLogIndex() uint32 {
	if x != nil {
		return x.LogIndex
	}
	return 0
}

var File_raftcomm_appendentries_proto protoreflect.FileDescriptor

const file_raftcomm_appendentries_proto_rawDesc = "" +
	"\n" +
	"\x1craftcomm/appendentries.proto\x12\braftcomm\"\xde\x01\n" +
	"\x14AppendEntriesRequest\x12\x12\n" +
	"\x04Term\x18\x01 \x01(\rR\x04Term\x12\x1a\n" +
	"\bLeaderId\x18\x02 \x01(\tR\bLeaderId\x12\"\n" +
	"\fPrevLogIndex\x18\x03 \x01(\rR\fPrevLogIndex\x12 \n" +
	"\vPrevLogTerm\x18\x04 \x01(\rR\vPrevLogTerm\x12,\n" +
	"\aEntries\x18\x05 \x03(\v2\x12.raftcomm.LogEntryR\aEntries\x12\"\n" +
	"\fLeaderCommit\x18\x06 \x01(\rR\fLeaderCommit\"E\n" +
	"\x15AppendEntriesResponse\x12\x12\n" +
	"\x04Term\x18\x01 \x01(\rR\x04Term\x12\x18\n" +
	"\aSuccess\x18\x02 \x01(\bR\aSuccess\"Z\n" +
	"\bLogEntry\x12\x18\n" +
	"\aCommand\x18\x01 \x01(\fR\aCommand\x12\x18\n" +
	"\aLogTerm\x18\x02 \x01(\rR\aLogTerm\x12\x1a\n" +
	"\bLogIndex\x18\x03 \x01(\rR\bLogIndexB@Z>github.com/jialuohu/curlyraft/internal/proto/raftcomm;raftcommb\x06proto3"

var (
	file_raftcomm_appendentries_proto_rawDescOnce sync.Once
	file_raftcomm_appendentries_proto_rawDescData []byte
)

func file_raftcomm_appendentries_proto_rawDescGZIP() []byte {
	file_raftcomm_appendentries_proto_rawDescOnce.Do(func() {
		file_raftcomm_appendentries_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_raftcomm_appendentries_proto_rawDesc), len(file_raftcomm_appendentries_proto_rawDesc)))
	})
	return file_raftcomm_appendentries_proto_rawDescData
}

var file_raftcomm_appendentries_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_raftcomm_appendentries_proto_goTypes = []any{
	(*AppendEntriesRequest)(nil),  // 0: raftcomm.AppendEntriesRequest
	(*AppendEntriesResponse)(nil), // 1: raftcomm.AppendEntriesResponse
	(*LogEntry)(nil),              // 2: raftcomm.LogEntry
}
var file_raftcomm_appendentries_proto_depIdxs = []int32{
	2, // 0: raftcomm.AppendEntriesRequest.Entries:type_name -> raftcomm.LogEntry
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_raftcomm_appendentries_proto_init() }
func file_raftcomm_appendentries_proto_init() {
	if File_raftcomm_appendentries_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_raftcomm_appendentries_proto_rawDesc), len(file_raftcomm_appendentries_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_raftcomm_appendentries_proto_goTypes,
		DependencyIndexes: file_raftcomm_appendentries_proto_depIdxs,
		MessageInfos:      file_raftcomm_appendentries_proto_msgTypes,
	}.Build()
	File_raftcomm_appendentries_proto = out.File
	file_raftcomm_appendentries_proto_goTypes = nil
	file_raftcomm_appendentries_proto_depIdxs = nil
}
