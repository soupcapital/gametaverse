// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.6.1
// source: proxy.proto

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

type Chain int32

const (
	Chain_UNKNOWN Chain = 0
	Chain_BSC     Chain = 1
	Chain_POLYGON Chain = 2
	Chain_ETH     Chain = 3
	Chain_AVAX    Chain = 4
	Chain_WAX     Chain = 5
	Chain_SOLANA  Chain = 6
)

// Enum value maps for Chain.
var (
	Chain_name = map[int32]string{
		0: "UNKNOWN",
		1: "BSC",
		2: "POLYGON",
		3: "ETH",
		4: "AVAX",
		5: "WAX",
		6: "SOLANA",
	}
	Chain_value = map[string]int32{
		"UNKNOWN": 0,
		"BSC":     1,
		"POLYGON": 2,
		"ETH":     3,
		"AVAX":    4,
		"WAX":     5,
		"SOLANA":  6,
	}
)

func (x Chain) Enum() *Chain {
	p := new(Chain)
	*p = x
	return p
}

func (x Chain) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Chain) Descriptor() protoreflect.EnumDescriptor {
	return file_proxy_proto_enumTypes[0].Descriptor()
}

func (Chain) Type() protoreflect.EnumType {
	return &file_proxy_proto_enumTypes[0]
}

func (x Chain) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Chain.Descriptor instead.
func (Chain) EnumDescriptor() ([]byte, []int) {
	return file_proxy_proto_rawDescGZIP(), []int{0}
}

type Contract struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Chain   Chain  `protobuf:"varint,1,opt,name=chain,proto3,enum=Chain" json:"chain,omitempty"`
	Address string `protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
}

func (x *Contract) Reset() {
	*x = Contract{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proxy_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Contract) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Contract) ProtoMessage() {}

func (x *Contract) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Contract.ProtoReflect.Descriptor instead.
func (*Contract) Descriptor() ([]byte, []int) {
	return file_proxy_proto_rawDescGZIP(), []int{0}
}

func (x *Contract) GetChain() Chain {
	if x != nil {
		return x.Chain
	}
	return Chain_UNKNOWN
}

func (x *Contract) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

type GameReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Start     int64       `protobuf:"varint,2,opt,name=start,proto3" json:"start,omitempty"`
	End       int64       `protobuf:"varint,3,opt,name=end,proto3" json:"end,omitempty"`
	Contracts []*Contract `protobuf:"bytes,4,rep,name=contracts,proto3" json:"contracts,omitempty"`
}

func (x *GameReq) Reset() {
	*x = GameReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proxy_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GameReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GameReq) ProtoMessage() {}

func (x *GameReq) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GameReq.ProtoReflect.Descriptor instead.
func (*GameReq) Descriptor() ([]byte, []int) {
	return file_proxy_proto_rawDescGZIP(), []int{1}
}

func (x *GameReq) GetStart() int64 {
	if x != nil {
		return x.Start
	}
	return 0
}

func (x *GameReq) GetEnd() int64 {
	if x != nil {
		return x.End
	}
	return 0
}

func (x *GameReq) GetContracts() []*Contract {
	if x != nil {
		return x.Contracts
	}
	return nil
}

type DauRsp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Dau uint64 `protobuf:"varint,1,opt,name=dau,proto3" json:"dau,omitempty"`
}

func (x *DauRsp) Reset() {
	*x = DauRsp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proxy_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DauRsp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DauRsp) ProtoMessage() {}

func (x *DauRsp) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DauRsp.ProtoReflect.Descriptor instead.
func (*DauRsp) Descriptor() ([]byte, []int) {
	return file_proxy_proto_rawDescGZIP(), []int{2}
}

func (x *DauRsp) GetDau() uint64 {
	if x != nil {
		return x.Dau
	}
	return 0
}

type TxCountRsp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Count uint64 `protobuf:"varint,1,opt,name=count,proto3" json:"count,omitempty"`
}

func (x *TxCountRsp) Reset() {
	*x = TxCountRsp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proxy_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TxCountRsp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TxCountRsp) ProtoMessage() {}

func (x *TxCountRsp) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TxCountRsp.ProtoReflect.Descriptor instead.
func (*TxCountRsp) Descriptor() ([]byte, []int) {
	return file_proxy_proto_rawDescGZIP(), []int{3}
}

func (x *TxCountRsp) GetCount() uint64 {
	if x != nil {
		return x.Count
	}
	return 0
}

type ChainGameReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Start  int64   `protobuf:"varint,2,opt,name=start,proto3" json:"start,omitempty"`
	End    int64   `protobuf:"varint,3,opt,name=end,proto3" json:"end,omitempty"`
	Chains []Chain `protobuf:"varint,4,rep,packed,name=chains,proto3,enum=Chain" json:"chains,omitempty"`
}

func (x *ChainGameReq) Reset() {
	*x = ChainGameReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proxy_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ChainGameReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChainGameReq) ProtoMessage() {}

func (x *ChainGameReq) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChainGameReq.ProtoReflect.Descriptor instead.
func (*ChainGameReq) Descriptor() ([]byte, []int) {
	return file_proxy_proto_rawDescGZIP(), []int{4}
}

func (x *ChainGameReq) GetStart() int64 {
	if x != nil {
		return x.Start
	}
	return 0
}

func (x *ChainGameReq) GetEnd() int64 {
	if x != nil {
		return x.End
	}
	return 0
}

func (x *ChainGameReq) GetChains() []Chain {
	if x != nil {
		return x.Chains
	}
	return nil
}

var File_proxy_proto protoreflect.FileDescriptor

var file_proxy_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x42, 0x0a,
	0x08, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x12, 0x1c, 0x0a, 0x05, 0x63, 0x68, 0x61,
	0x69, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x06, 0x2e, 0x43, 0x68, 0x61, 0x69, 0x6e,
	0x52, 0x05, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65,
	0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73,
	0x73, 0x22, 0x5a, 0x0a, 0x07, 0x47, 0x61, 0x6d, 0x65, 0x52, 0x65, 0x71, 0x12, 0x14, 0x0a, 0x05,
	0x73, 0x74, 0x61, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x73, 0x74, 0x61,
	0x72, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x65, 0x6e, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x03, 0x65, 0x6e, 0x64, 0x12, 0x27, 0x0a, 0x09, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74,
	0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x09, 0x2e, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x61,
	0x63, 0x74, 0x52, 0x09, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x73, 0x22, 0x1a, 0x0a,
	0x06, 0x44, 0x61, 0x75, 0x52, 0x73, 0x70, 0x12, 0x10, 0x0a, 0x03, 0x64, 0x61, 0x75, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x04, 0x52, 0x03, 0x64, 0x61, 0x75, 0x22, 0x22, 0x0a, 0x0a, 0x54, 0x78, 0x43,
	0x6f, 0x75, 0x6e, 0x74, 0x52, 0x73, 0x70, 0x12, 0x14, 0x0a, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x22, 0x56, 0x0a,
	0x0c, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x47, 0x61, 0x6d, 0x65, 0x52, 0x65, 0x71, 0x12, 0x14, 0x0a,
	0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x73, 0x74,
	0x61, 0x72, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x65, 0x6e, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x03, 0x65, 0x6e, 0x64, 0x12, 0x1e, 0x0a, 0x06, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x73, 0x18,
	0x04, 0x20, 0x03, 0x28, 0x0e, 0x32, 0x06, 0x2e, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x52, 0x06, 0x63,
	0x68, 0x61, 0x69, 0x6e, 0x73, 0x2a, 0x52, 0x0a, 0x05, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x12, 0x0b,
	0x0a, 0x07, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x00, 0x12, 0x07, 0x0a, 0x03, 0x42,
	0x53, 0x43, 0x10, 0x01, 0x12, 0x0b, 0x0a, 0x07, 0x50, 0x4f, 0x4c, 0x59, 0x47, 0x4f, 0x4e, 0x10,
	0x02, 0x12, 0x07, 0x0a, 0x03, 0x45, 0x54, 0x48, 0x10, 0x03, 0x12, 0x08, 0x0a, 0x04, 0x41, 0x56,
	0x41, 0x58, 0x10, 0x04, 0x12, 0x07, 0x0a, 0x03, 0x57, 0x41, 0x58, 0x10, 0x05, 0x12, 0x0a, 0x0a,
	0x06, 0x53, 0x4f, 0x4c, 0x41, 0x4e, 0x41, 0x10, 0x06, 0x32, 0x9d, 0x01, 0x0a, 0x07, 0x44, 0x42,
	0x50, 0x72, 0x6f, 0x78, 0x79, 0x12, 0x1a, 0x0a, 0x03, 0x44, 0x61, 0x75, 0x12, 0x08, 0x2e, 0x47,
	0x61, 0x6d, 0x65, 0x52, 0x65, 0x71, 0x1a, 0x07, 0x2e, 0x44, 0x61, 0x75, 0x52, 0x73, 0x70, 0x22,
	0x00, 0x12, 0x22, 0x0a, 0x07, 0x54, 0x78, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x08, 0x2e, 0x47,
	0x61, 0x6d, 0x65, 0x52, 0x65, 0x71, 0x1a, 0x0b, 0x2e, 0x54, 0x78, 0x43, 0x6f, 0x75, 0x6e, 0x74,
	0x52, 0x73, 0x70, 0x22, 0x00, 0x12, 0x24, 0x0a, 0x08, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x44, 0x61,
	0x75, 0x12, 0x0d, 0x2e, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x47, 0x61, 0x6d, 0x65, 0x52, 0x65, 0x71,
	0x1a, 0x07, 0x2e, 0x44, 0x61, 0x75, 0x52, 0x73, 0x70, 0x22, 0x00, 0x12, 0x2c, 0x0a, 0x0c, 0x43,
	0x68, 0x61, 0x69, 0x6e, 0x54, 0x78, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x0d, 0x2e, 0x43, 0x68,
	0x61, 0x69, 0x6e, 0x47, 0x61, 0x6d, 0x65, 0x52, 0x65, 0x71, 0x1a, 0x0b, 0x2e, 0x54, 0x78, 0x43,
	0x6f, 0x75, 0x6e, 0x74, 0x52, 0x73, 0x70, 0x22, 0x00, 0x42, 0x24, 0x5a, 0x22, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x67, 0x61, 0x6d, 0x65, 0x74, 0x61, 0x76, 0x65,
	0x72, 0x73, 0x65, 0x2f, 0x67, 0x66, 0x64, 0x70, 0x2f, 0x72, 0x70, 0x63, 0x2f, 0x70, 0x62, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proxy_proto_rawDescOnce sync.Once
	file_proxy_proto_rawDescData = file_proxy_proto_rawDesc
)

func file_proxy_proto_rawDescGZIP() []byte {
	file_proxy_proto_rawDescOnce.Do(func() {
		file_proxy_proto_rawDescData = protoimpl.X.CompressGZIP(file_proxy_proto_rawDescData)
	})
	return file_proxy_proto_rawDescData
}

var file_proxy_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_proxy_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_proxy_proto_goTypes = []interface{}{
	(Chain)(0),           // 0: Chain
	(*Contract)(nil),     // 1: Contract
	(*GameReq)(nil),      // 2: GameReq
	(*DauRsp)(nil),       // 3: DauRsp
	(*TxCountRsp)(nil),   // 4: TxCountRsp
	(*ChainGameReq)(nil), // 5: ChainGameReq
}
var file_proxy_proto_depIdxs = []int32{
	0, // 0: Contract.chain:type_name -> Chain
	1, // 1: GameReq.contracts:type_name -> Contract
	0, // 2: ChainGameReq.chains:type_name -> Chain
	2, // 3: DBProxy.Dau:input_type -> GameReq
	2, // 4: DBProxy.TxCount:input_type -> GameReq
	5, // 5: DBProxy.ChainDau:input_type -> ChainGameReq
	5, // 6: DBProxy.ChainTxCount:input_type -> ChainGameReq
	3, // 7: DBProxy.Dau:output_type -> DauRsp
	4, // 8: DBProxy.TxCount:output_type -> TxCountRsp
	3, // 9: DBProxy.ChainDau:output_type -> DauRsp
	4, // 10: DBProxy.ChainTxCount:output_type -> TxCountRsp
	7, // [7:11] is the sub-list for method output_type
	3, // [3:7] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_proxy_proto_init() }
func file_proxy_proto_init() {
	if File_proxy_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proxy_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Contract); i {
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
		file_proxy_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GameReq); i {
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
		file_proxy_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DauRsp); i {
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
		file_proxy_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TxCountRsp); i {
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
		file_proxy_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ChainGameReq); i {
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
			RawDescriptor: file_proxy_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proxy_proto_goTypes,
		DependencyIndexes: file_proxy_proto_depIdxs,
		EnumInfos:         file_proxy_proto_enumTypes,
		MessageInfos:      file_proxy_proto_msgTypes,
	}.Build()
	File_proxy_proto = out.File
	file_proxy_proto_rawDesc = nil
	file_proxy_proto_goTypes = nil
	file_proxy_proto_depIdxs = nil
}
