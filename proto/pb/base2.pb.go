// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.5
// source: base2.proto

package pb

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

type ProtoBase2 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Code   string  `protobuf:"bytes,1,opt,name=code,proto3" json:"code,omitempty"`       // 代码
	Time   int64   `protobuf:"varint,2,opt,name=time,proto3" json:"time,omitempty"`      // 时间 202005011020
	Open   float64 `protobuf:"fixed64,3,opt,name=open,proto3" json:"open,omitempty"`     // 开盘价
	Close  float64 `protobuf:"fixed64,4,opt,name=close,proto3" json:"close,omitempty"`   // 收盘价
	High   float64 `protobuf:"fixed64,5,opt,name=high,proto3" json:"high,omitempty"`     // 最高价
	Low    float64 `protobuf:"fixed64,6,opt,name=low,proto3" json:"low,omitempty"`       // 最低价
	Volume float64 `protobuf:"fixed64,7,opt,name=volume,proto3" json:"volume,omitempty"` // 成交量
	Amount float64 `protobuf:"fixed64,8,opt,name=amount,proto3" json:"amount,omitempty"` // 成交额(￥)
}

func (x *ProtoBase2) Reset() {
	*x = ProtoBase2{}
	if protoimpl.UnsafeEnabled {
		mi := &file_base2_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProtoBase2) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoBase2) ProtoMessage() {}

func (x *ProtoBase2) ProtoReflect() protoreflect.Message {
	mi := &file_base2_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoBase2.ProtoReflect.Descriptor instead.
func (*ProtoBase2) Descriptor() ([]byte, []int) {
	return file_base2_proto_rawDescGZIP(), []int{0}
}

func (x *ProtoBase2) GetCode() string {
	if x != nil {
		return x.Code
	}
	return ""
}

func (x *ProtoBase2) GetTime() int64 {
	if x != nil {
		return x.Time
	}
	return 0
}

func (x *ProtoBase2) GetOpen() float64 {
	if x != nil {
		return x.Open
	}
	return 0
}

func (x *ProtoBase2) GetClose() float64 {
	if x != nil {
		return x.Close
	}
	return 0
}

func (x *ProtoBase2) GetHigh() float64 {
	if x != nil {
		return x.High
	}
	return 0
}

func (x *ProtoBase2) GetLow() float64 {
	if x != nil {
		return x.Low
	}
	return 0
}

func (x *ProtoBase2) GetVolume() float64 {
	if x != nil {
		return x.Volume
	}
	return 0
}

func (x *ProtoBase2) GetAmount() float64 {
	if x != nil {
		return x.Amount
	}
	return 0
}

var File_base2_proto protoreflect.FileDescriptor

var file_base2_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x62, 0x61, 0x73, 0x65, 0x32, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x70,
	0x62, 0x22, 0xb4, 0x01, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x42, 0x61, 0x73, 0x65, 0x32,
	0x12, 0x12, 0x0a, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x63, 0x6f, 0x64, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6f, 0x70, 0x65, 0x6e,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x01, 0x52, 0x04, 0x6f, 0x70, 0x65, 0x6e, 0x12, 0x14, 0x0a, 0x05,
	0x63, 0x6c, 0x6f, 0x73, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x01, 0x52, 0x05, 0x63, 0x6c, 0x6f,
	0x73, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x69, 0x67, 0x68, 0x18, 0x05, 0x20, 0x01, 0x28, 0x01,
	0x52, 0x04, 0x68, 0x69, 0x67, 0x68, 0x12, 0x10, 0x0a, 0x03, 0x6c, 0x6f, 0x77, 0x18, 0x06, 0x20,
	0x01, 0x28, 0x01, 0x52, 0x03, 0x6c, 0x6f, 0x77, 0x12, 0x16, 0x0a, 0x06, 0x76, 0x6f, 0x6c, 0x75,
	0x6d, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x01, 0x52, 0x06, 0x76, 0x6f, 0x6c, 0x75, 0x6d, 0x65,
	0x12, 0x16, 0x0a, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x08, 0x20, 0x01, 0x28, 0x01,
	0x52, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x42, 0x06, 0x5a, 0x04, 0x2e, 0x2f, 0x70, 0x62,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_base2_proto_rawDescOnce sync.Once
	file_base2_proto_rawDescData = file_base2_proto_rawDesc
)

func file_base2_proto_rawDescGZIP() []byte {
	file_base2_proto_rawDescOnce.Do(func() {
		file_base2_proto_rawDescData = protoimpl.X.CompressGZIP(file_base2_proto_rawDescData)
	})
	return file_base2_proto_rawDescData
}

var file_base2_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_base2_proto_goTypes = []interface{}{
	(*ProtoBase2)(nil), // 0: pb.protoBase2
}
var file_base2_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_base2_proto_init() }
func file_base2_proto_init() {
	if File_base2_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_base2_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ProtoBase2); i {
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
			RawDescriptor: file_base2_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_base2_proto_goTypes,
		DependencyIndexes: file_base2_proto_depIdxs,
		MessageInfos:      file_base2_proto_msgTypes,
	}.Build()
	File_base2_proto = out.File
	file_base2_proto_rawDesc = nil
	file_base2_proto_goTypes = nil
	file_base2_proto_depIdxs = nil
}
