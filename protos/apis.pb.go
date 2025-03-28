// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        v5.26.1
// source: apis.proto

package protos

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// (옵션) 클라이언트가 강제로 동기화를 요청할 때 사용 (필요 없으면 빈 메시지로 대체 가능)
type SyncFoldersInfoRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Force bool `protobuf:"varint,1,opt,name=force,proto3" json:"force,omitempty"` // force update flag, 기본값 false
}

func (x *SyncFoldersInfoRequest) Reset() {
	*x = SyncFoldersInfoRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_apis_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SyncFoldersInfoRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SyncFoldersInfoRequest) ProtoMessage() {}

func (x *SyncFoldersInfoRequest) ProtoReflect() protoreflect.Message {
	mi := &file_apis_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SyncFoldersInfoRequest.ProtoReflect.Descriptor instead.
func (*SyncFoldersInfoRequest) Descriptor() ([]byte, []int) {
	return file_apis_proto_rawDescGZIP(), []int{0}
}

func (x *SyncFoldersInfoRequest) GetForce() bool {
	if x != nil {
		return x.Force
	}
	return false
}

// 동기화 작업 결과를 응답
type SyncFoldersInfoResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// 업데이트가 이루어졌으면 true, 그렇지 않으면 false
	Updated bool `protobuf:"varint,1,opt,name=updated,proto3" json:"updated,omitempty"`
}

func (x *SyncFoldersInfoResponse) Reset() {
	*x = SyncFoldersInfoResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_apis_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SyncFoldersInfoResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SyncFoldersInfoResponse) ProtoMessage() {}

func (x *SyncFoldersInfoResponse) ProtoReflect() protoreflect.Message {
	mi := &file_apis_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SyncFoldersInfoResponse.ProtoReflect.Descriptor instead.
func (*SyncFoldersInfoResponse) Descriptor() ([]byte, []int) {
	return file_apis_proto_rawDescGZIP(), []int{1}
}

func (x *SyncFoldersInfoResponse) GetUpdated() bool {
	if x != nil {
		return x.Updated
	}
	return false
}

// 단일 파일 블럭을 나타내는 메시지
type FileBlock struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BlockId       string   `protobuf:"bytes,1,opt,name=block_id,json=blockId,proto3" json:"block_id,omitempty"`                   // 블록을 구분하기 위한 고유 ID (예: 파일 경로)
	ColumnHeaders []string `protobuf:"bytes,2,rep,name=column_headers,json=columnHeaders,proto3" json:"column_headers,omitempty"` // 컬럼 이름들
	Rows          []*Row   `protobuf:"bytes,3,rep,name=rows,proto3" json:"rows,omitempty"`                                        // 행 데이터
}

func (x *FileBlock) Reset() {
	*x = FileBlock{}
	if protoimpl.UnsafeEnabled {
		mi := &file_apis_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FileBlock) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FileBlock) ProtoMessage() {}

func (x *FileBlock) ProtoReflect() protoreflect.Message {
	mi := &file_apis_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FileBlock.ProtoReflect.Descriptor instead.
func (*FileBlock) Descriptor() ([]byte, []int) {
	return file_apis_proto_rawDescGZIP(), []int{2}
}

func (x *FileBlock) GetBlockId() string {
	if x != nil {
		return x.BlockId
	}
	return ""
}

func (x *FileBlock) GetColumnHeaders() []string {
	if x != nil {
		return x.ColumnHeaders
	}
	return nil
}

func (x *FileBlock) GetRows() []*Row {
	if x != nil {
		return x.Rows
	}
	return nil
}

// 하나의 행(row)을 나타내며, 행 번호와 헤더-값 매핑을 포함
type Row struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RowNumber   int32             `protobuf:"varint,1,opt,name=row_number,json=rowNumber,proto3" json:"row_number,omitempty"`                                                                                              // 행 번호
	CellColumns map[string]string `protobuf:"bytes,2,rep,name=cell_columns,json=cellColumns,proto3" json:"cell_columns,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"` // 헤더 이름과 셀 값의 매핑
}

func (x *Row) Reset() {
	*x = Row{}
	if protoimpl.UnsafeEnabled {
		mi := &file_apis_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Row) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Row) ProtoMessage() {}

func (x *Row) ProtoReflect() protoreflect.Message {
	mi := &file_apis_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Row.ProtoReflect.Descriptor instead.
func (*Row) Descriptor() ([]byte, []int) {
	return file_apis_proto_rawDescGZIP(), []int{3}
}

func (x *Row) GetRowNumber() int32 {
	if x != nil {
		return x.RowNumber
	}
	return 0
}

func (x *Row) GetCellColumns() map[string]string {
	if x != nil {
		return x.CellColumns
	}
	return nil
}

// 여러 파일 블럭을 묶어서 나타내는 메시지
type DataBlock struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UpdatedAt *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"` // 최종 업데이트 시간
	Blocks    []*FileBlock           `protobuf:"bytes,2,rep,name=blocks,proto3" json:"blocks,omitempty"`                        // 파일 블럭 리스트
}

func (x *DataBlock) Reset() {
	*x = DataBlock{}
	if protoimpl.UnsafeEnabled {
		mi := &file_apis_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DataBlock) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DataBlock) ProtoMessage() {}

func (x *DataBlock) ProtoReflect() protoreflect.Message {
	mi := &file_apis_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DataBlock.ProtoReflect.Descriptor instead.
func (*DataBlock) Descriptor() ([]byte, []int) {
	return file_apis_proto_rawDescGZIP(), []int{4}
}

func (x *DataBlock) GetUpdatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.UpdatedAt
	}
	return nil
}

func (x *DataBlock) GetBlocks() []*FileBlock {
	if x != nil {
		return x.Blocks
	}
	return nil
}

// 클라이언트가 현재 가지고 있는 데이터의 업데이트 타임스탬프를 포함하는 요청 메시지
type GetDataBlockRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// 클라이언트가 마지막으로 받은 데이터의 updated_at 값
	CurrentUpdatedAt *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=current_updated_at,json=currentUpdatedAt,proto3" json:"current_updated_at,omitempty"`
}

func (x *GetDataBlockRequest) Reset() {
	*x = GetDataBlockRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_apis_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetDataBlockRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetDataBlockRequest) ProtoMessage() {}

func (x *GetDataBlockRequest) ProtoReflect() protoreflect.Message {
	mi := &file_apis_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetDataBlockRequest.ProtoReflect.Descriptor instead.
func (*GetDataBlockRequest) Descriptor() ([]byte, []int) {
	return file_apis_proto_rawDescGZIP(), []int{5}
}

func (x *GetDataBlockRequest) GetCurrentUpdatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.CurrentUpdatedAt
	}
	return nil
}

// 서버가 응답으로 DataBlockData 를 포함하여 보내는 메시지
type GetDataBlockResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data *DataBlock `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
	// 예를 들어, 데이터가 최신이면 no_update 플래그를 true 로 설정할 수도 있음
	NoUpdate bool `protobuf:"varint,2,opt,name=no_update,json=noUpdate,proto3" json:"no_update,omitempty"`
}

func (x *GetDataBlockResponse) Reset() {
	*x = GetDataBlockResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_apis_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetDataBlockResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetDataBlockResponse) ProtoMessage() {}

func (x *GetDataBlockResponse) ProtoReflect() protoreflect.Message {
	mi := &file_apis_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetDataBlockResponse.ProtoReflect.Descriptor instead.
func (*GetDataBlockResponse) Descriptor() ([]byte, []int) {
	return file_apis_proto_rawDescGZIP(), []int{6}
}

func (x *GetDataBlockResponse) GetData() *DataBlock {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *GetDataBlockResponse) GetNoUpdate() bool {
	if x != nil {
		return x.NoUpdate
	}
	return false
}

var File_apis_proto protoreflect.FileDescriptor

var file_apis_proto_rawDesc = []byte{
	0x0a, 0x0a, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x73, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x2e, 0x0a, 0x16, 0x53, 0x79, 0x6e, 0x63, 0x46, 0x6f, 0x6c,
	0x64, 0x65, 0x72, 0x73, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x14, 0x0a, 0x05, 0x66, 0x6f, 0x72, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x05,
	0x66, 0x6f, 0x72, 0x63, 0x65, 0x22, 0x33, 0x0a, 0x17, 0x53, 0x79, 0x6e, 0x63, 0x46, 0x6f, 0x6c,
	0x64, 0x65, 0x72, 0x73, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x18, 0x0a, 0x07, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x07, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x22, 0x6e, 0x0a, 0x09, 0x46, 0x69,
	0x6c, 0x65, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x19, 0x0a, 0x08, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
	0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
	0x49, 0x64, 0x12, 0x25, 0x0a, 0x0e, 0x63, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x5f, 0x68, 0x65, 0x61,
	0x64, 0x65, 0x72, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0d, 0x63, 0x6f, 0x6c, 0x75,
	0x6d, 0x6e, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x12, 0x1f, 0x0a, 0x04, 0x72, 0x6f, 0x77,
	0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73,
	0x2e, 0x52, 0x6f, 0x77, 0x52, 0x04, 0x72, 0x6f, 0x77, 0x73, 0x22, 0xa5, 0x01, 0x0a, 0x03, 0x52,
	0x6f, 0x77, 0x12, 0x1d, 0x0a, 0x0a, 0x72, 0x6f, 0x77, 0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x09, 0x72, 0x6f, 0x77, 0x4e, 0x75, 0x6d, 0x62, 0x65,
	0x72, 0x12, 0x3f, 0x0a, 0x0c, 0x63, 0x65, 0x6c, 0x6c, 0x5f, 0x63, 0x6f, 0x6c, 0x75, 0x6d, 0x6e,
	0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73,
	0x2e, 0x52, 0x6f, 0x77, 0x2e, 0x43, 0x65, 0x6c, 0x6c, 0x43, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x73,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0b, 0x63, 0x65, 0x6c, 0x6c, 0x43, 0x6f, 0x6c, 0x75, 0x6d,
	0x6e, 0x73, 0x1a, 0x3e, 0x0a, 0x10, 0x43, 0x65, 0x6c, 0x6c, 0x43, 0x6f, 0x6c, 0x75, 0x6d, 0x6e,
	0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02,
	0x38, 0x01, 0x22, 0x71, 0x0a, 0x09, 0x44, 0x61, 0x74, 0x61, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12,
	0x39, 0x0a, 0x0a, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52,
	0x09, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12, 0x29, 0x0a, 0x06, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x73, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x06, 0x62,
	0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x22, 0x5f, 0x0a, 0x13, 0x47, 0x65, 0x74, 0x44, 0x61, 0x74, 0x61,
	0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x48, 0x0a, 0x12,
	0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x5f, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x5f,
	0x61, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x52, 0x10, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x55, 0x70, 0x64,
	0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x22, 0x5a, 0x0a, 0x14, 0x47, 0x65, 0x74, 0x44, 0x61, 0x74,
	0x61, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x25,
	0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52,
	0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x1b, 0x0a, 0x09, 0x6e, 0x6f, 0x5f, 0x75, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x6e, 0x6f, 0x55, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x32, 0x63, 0x0a, 0x0d, 0x44, 0x42, 0x41, 0x70, 0x69, 0x73, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x12, 0x52, 0x0a, 0x0f, 0x53, 0x79, 0x6e, 0x63, 0x46, 0x6f, 0x6c, 0x64, 0x65,
	0x72, 0x73, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x1e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e,
	0x53, 0x79, 0x6e, 0x63, 0x46, 0x6f, 0x6c, 0x64, 0x65, 0x72, 0x73, 0x49, 0x6e, 0x66, 0x6f, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e,
	0x53, 0x79, 0x6e, 0x63, 0x46, 0x6f, 0x6c, 0x64, 0x65, 0x72, 0x73, 0x49, 0x6e, 0x66, 0x6f, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x32, 0x5d, 0x0a, 0x10, 0x44, 0x61, 0x74, 0x61, 0x42,
	0x6c, 0x6f, 0x63, 0x6b, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x49, 0x0a, 0x0c, 0x47,
	0x65, 0x74, 0x44, 0x61, 0x74, 0x61, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x1b, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x47, 0x65, 0x74, 0x44, 0x61, 0x74, 0x61, 0x42, 0x6c, 0x6f, 0x63,
	0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x73, 0x2e, 0x47, 0x65, 0x74, 0x44, 0x61, 0x74, 0x61, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x22, 0x5a, 0x20, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x65, 0x6f, 0x79, 0x68, 0x61, 0x65, 0x69, 0x6e, 0x2f, 0x74,
	0x6f, 0x72, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_apis_proto_rawDescOnce sync.Once
	file_apis_proto_rawDescData = file_apis_proto_rawDesc
)

func file_apis_proto_rawDescGZIP() []byte {
	file_apis_proto_rawDescOnce.Do(func() {
		file_apis_proto_rawDescData = protoimpl.X.CompressGZIP(file_apis_proto_rawDescData)
	})
	return file_apis_proto_rawDescData
}

var file_apis_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_apis_proto_goTypes = []interface{}{
	(*SyncFoldersInfoRequest)(nil),  // 0: protos.SyncFoldersInfoRequest
	(*SyncFoldersInfoResponse)(nil), // 1: protos.SyncFoldersInfoResponse
	(*FileBlock)(nil),               // 2: protos.FileBlock
	(*Row)(nil),                     // 3: protos.Row
	(*DataBlock)(nil),               // 4: protos.DataBlock
	(*GetDataBlockRequest)(nil),     // 5: protos.GetDataBlockRequest
	(*GetDataBlockResponse)(nil),    // 6: protos.GetDataBlockResponse
	nil,                             // 7: protos.Row.CellColumnsEntry
	(*timestamppb.Timestamp)(nil),   // 8: google.protobuf.Timestamp
}
var file_apis_proto_depIdxs = []int32{
	3, // 0: protos.FileBlock.rows:type_name -> protos.Row
	7, // 1: protos.Row.cell_columns:type_name -> protos.Row.CellColumnsEntry
	8, // 2: protos.DataBlock.updated_at:type_name -> google.protobuf.Timestamp
	2, // 3: protos.DataBlock.blocks:type_name -> protos.FileBlock
	8, // 4: protos.GetDataBlockRequest.current_updated_at:type_name -> google.protobuf.Timestamp
	4, // 5: protos.GetDataBlockResponse.data:type_name -> protos.DataBlock
	0, // 6: protos.DBApisService.SyncFoldersInfo:input_type -> protos.SyncFoldersInfoRequest
	5, // 7: protos.DataBlockService.GetDataBlock:input_type -> protos.GetDataBlockRequest
	1, // 8: protos.DBApisService.SyncFoldersInfo:output_type -> protos.SyncFoldersInfoResponse
	6, // 9: protos.DataBlockService.GetDataBlock:output_type -> protos.GetDataBlockResponse
	8, // [8:10] is the sub-list for method output_type
	6, // [6:8] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_apis_proto_init() }
func file_apis_proto_init() {
	if File_apis_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_apis_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SyncFoldersInfoRequest); i {
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
		file_apis_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SyncFoldersInfoResponse); i {
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
		file_apis_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FileBlock); i {
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
		file_apis_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Row); i {
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
		file_apis_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DataBlock); i {
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
		file_apis_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetDataBlockRequest); i {
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
		file_apis_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetDataBlockResponse); i {
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
			RawDescriptor: file_apis_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   2,
		},
		GoTypes:           file_apis_proto_goTypes,
		DependencyIndexes: file_apis_proto_depIdxs,
		MessageInfos:      file_apis_proto_msgTypes,
	}.Build()
	File_apis_proto = out.File
	file_apis_proto_rawDesc = nil
	file_apis_proto_goTypes = nil
	file_apis_proto_depIdxs = nil
}
