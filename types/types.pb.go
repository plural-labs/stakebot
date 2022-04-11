// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.18.1
// source: types.proto

package types

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Frequency int32

const (
	Frequency_UNKNOWN    Frequency = 0
	Frequency_QUARTERDAY Frequency = 1
	Frequency_DAILY      Frequency = 2
	Frequency_WEEKLY     Frequency = 3
	Frequency_MONTHLY    Frequency = 4
)

// Enum value maps for Frequency.
var (
	Frequency_name = map[int32]string{
		0: "UNKNOWN",
		1: "QUARTERDAY",
		2: "DAILY",
		3: "WEEKLY",
		4: "MONTHLY",
	}
	Frequency_value = map[string]int32{
		"UNKNOWN":    0,
		"QUARTERDAY": 1,
		"DAILY":      2,
		"WEEKLY":     3,
		"MONTHLY":    4,
	}
)

func (x Frequency) Enum() *Frequency {
	p := new(Frequency)
	*p = x
	return p
}

func (x Frequency) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Frequency) Descriptor() protoreflect.EnumDescriptor {
	return file_types_proto_enumTypes[0].Descriptor()
}

func (Frequency) Type() protoreflect.EnumType {
	return &file_types_proto_enumTypes[0]
}

func (x Frequency) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Frequency.Descriptor instead.
func (Frequency) EnumDescriptor() ([]byte, []int) {
	return file_types_proto_rawDescGZIP(), []int{0}
}

type Record struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Address                string    `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Frequency              Frequency `protobuf:"varint,2,opt,name=frequency,proto3,enum=Frequency" json:"frequency,omitempty"`
	Tolerance              int64     `protobuf:"varint,3,opt,name=tolerance,proto3" json:"tolerance,omitempty"`
	LastUpdatedUnixTime    int64     `protobuf:"varint,4,opt,name=last_updated_unix_time,json=lastUpdatedUnixTime,proto3" json:"last_updated_unix_time,omitempty"`
	TotalAutostakedRewards int64     `protobuf:"varint,5,opt,name=total_autostaked_rewards,json=totalAutostakedRewards,proto3" json:"total_autostaked_rewards,omitempty"`
	ErrorLogs              string    `protobuf:"bytes,6,opt,name=error_logs,json=errorLogs,proto3" json:"error_logs,omitempty"`
}

func (x *Record) Reset() {
	*x = Record{}
	if protoimpl.UnsafeEnabled {
		mi := &file_types_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Record) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Record) ProtoMessage() {}

func (x *Record) ProtoReflect() protoreflect.Message {
	mi := &file_types_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Record.ProtoReflect.Descriptor instead.
func (*Record) Descriptor() ([]byte, []int) {
	return file_types_proto_rawDescGZIP(), []int{0}
}

func (x *Record) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *Record) GetFrequency() Frequency {
	if x != nil {
		return x.Frequency
	}
	return Frequency_UNKNOWN
}

func (x *Record) GetTolerance() int64 {
	if x != nil {
		return x.Tolerance
	}
	return 0
}

func (x *Record) GetLastUpdatedUnixTime() int64 {
	if x != nil {
		return x.LastUpdatedUnixTime
	}
	return 0
}

func (x *Record) GetTotalAutostakedRewards() int64 {
	if x != nil {
		return x.TotalAutostakedRewards
	}
	return 0
}

func (x *Record) GetErrorLogs() string {
	if x != nil {
		return x.ErrorLogs
	}
	return ""
}

type Job struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id        int64     `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Frequency Frequency `protobuf:"varint,2,opt,name=frequency,proto3,enum=Frequency" json:"frequency,omitempty"`
}

func (x *Job) Reset() {
	*x = Job{}
	if protoimpl.UnsafeEnabled {
		mi := &file_types_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Job) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Job) ProtoMessage() {}

func (x *Job) ProtoReflect() protoreflect.Message {
	mi := &file_types_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Job.ProtoReflect.Descriptor instead.
func (*Job) Descriptor() ([]byte, []int) {
	return file_types_proto_rawDescGZIP(), []int{1}
}

func (x *Job) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *Job) GetFrequency() Frequency {
	if x != nil {
		return x.Frequency
	}
	return Frequency_UNKNOWN
}

var File_types_proto protoreflect.FileDescriptor

var file_types_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xf8, 0x01,
	0x0a, 0x06, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72,
	0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65,
	0x73, 0x73, 0x12, 0x28, 0x0a, 0x09, 0x66, 0x72, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x79, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0a, 0x2e, 0x46, 0x72, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63,
	0x79, 0x52, 0x09, 0x66, 0x72, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x79, 0x12, 0x1c, 0x0a, 0x09,
	0x74, 0x6f, 0x6c, 0x65, 0x72, 0x61, 0x6e, 0x63, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x09, 0x74, 0x6f, 0x6c, 0x65, 0x72, 0x61, 0x6e, 0x63, 0x65, 0x12, 0x33, 0x0a, 0x16, 0x6c, 0x61,
	0x73, 0x74, 0x5f, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x75, 0x6e, 0x69, 0x78, 0x5f,
	0x74, 0x69, 0x6d, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x13, 0x6c, 0x61, 0x73, 0x74,
	0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x55, 0x6e, 0x69, 0x78, 0x54, 0x69, 0x6d, 0x65, 0x12,
	0x38, 0x0a, 0x18, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x5f, 0x61, 0x75, 0x74, 0x6f, 0x73, 0x74, 0x61,
	0x6b, 0x65, 0x64, 0x5f, 0x72, 0x65, 0x77, 0x61, 0x72, 0x64, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x16, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x41, 0x75, 0x74, 0x6f, 0x73, 0x74, 0x61, 0x6b,
	0x65, 0x64, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x73, 0x12, 0x1d, 0x0a, 0x0a, 0x65, 0x72, 0x72,
	0x6f, 0x72, 0x5f, 0x6c, 0x6f, 0x67, 0x73, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x65,
	0x72, 0x72, 0x6f, 0x72, 0x4c, 0x6f, 0x67, 0x73, 0x22, 0x3f, 0x0a, 0x03, 0x4a, 0x6f, 0x62, 0x12,
	0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x69, 0x64, 0x12,
	0x28, 0x0a, 0x09, 0x66, 0x72, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x79, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0e, 0x32, 0x0a, 0x2e, 0x46, 0x72, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x79, 0x52, 0x09,
	0x66, 0x72, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x79, 0x2a, 0x4c, 0x0a, 0x09, 0x46, 0x72, 0x65,
	0x71, 0x75, 0x65, 0x6e, 0x63, 0x79, 0x12, 0x0b, 0x0a, 0x07, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57,
	0x4e, 0x10, 0x00, 0x12, 0x0e, 0x0a, 0x0a, 0x51, 0x55, 0x41, 0x52, 0x54, 0x45, 0x52, 0x44, 0x41,
	0x59, 0x10, 0x01, 0x12, 0x09, 0x0a, 0x05, 0x44, 0x41, 0x49, 0x4c, 0x59, 0x10, 0x02, 0x12, 0x0a,
	0x0a, 0x06, 0x57, 0x45, 0x45, 0x4b, 0x4c, 0x59, 0x10, 0x03, 0x12, 0x0b, 0x0a, 0x07, 0x4d, 0x4f,
	0x4e, 0x54, 0x48, 0x4c, 0x59, 0x10, 0x04, 0x42, 0x29, 0x5a, 0x27, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x6c, 0x75, 0x72, 0x61, 0x6c, 0x2d, 0x6c, 0x61, 0x62,
	0x73, 0x2f, 0x61, 0x75, 0x74, 0x6f, 0x73, 0x74, 0x61, 0x6b, 0x65, 0x72, 0x2f, 0x74, 0x79, 0x70,
	0x65, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_types_proto_rawDescOnce sync.Once
	file_types_proto_rawDescData = file_types_proto_rawDesc
)

func file_types_proto_rawDescGZIP() []byte {
	file_types_proto_rawDescOnce.Do(func() {
		file_types_proto_rawDescData = protoimpl.X.CompressGZIP(file_types_proto_rawDescData)
	})
	return file_types_proto_rawDescData
}

var file_types_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_types_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_types_proto_goTypes = []interface{}{
	(Frequency)(0), // 0: Frequency
	(*Record)(nil), // 1: Record
	(*Job)(nil),    // 2: Job
}
var file_types_proto_depIdxs = []int32{
	0, // 0: Record.frequency:type_name -> Frequency
	0, // 1: Job.frequency:type_name -> Frequency
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_types_proto_init() }
func file_types_proto_init() {
	if File_types_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_types_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Record); i {
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
		file_types_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Job); i {
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
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_types_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_types_proto_goTypes,
		DependencyIndexes: file_types_proto_depIdxs,
		EnumInfos:         file_types_proto_enumTypes,
		MessageInfos:      file_types_proto_msgTypes,
	}.Build()
	File_types_proto = out.File
	file_types_proto_rawDesc = nil
	file_types_proto_goTypes = nil
	file_types_proto_depIdxs = nil
}
