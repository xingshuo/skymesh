// Code generated by protoc-gen-go. DO NOT EDIT.
// source: ssproto.proto

package smproto

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
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

type SSCmd int32

const (
	SSCmd_INVALID_CMD SSCmd = 0
	//agent && nameserver 1000 ~ 1999
	SSCmd_REQ_REGISTER_APP            SSCmd = 1000
	SSCmd_RSP_REGISTER_APP            SSCmd = 1001
	SSCmd_REQ_REGISTER_SERVICE        SSCmd = 1002
	SSCmd_RSP_REGISTER_SERVICE        SSCmd = 1003
	SSCmd_NOTIFY_SERVICE_ONLINE       SSCmd = 1004
	SSCmd_REQ_UNREGISTER_SERVICE      SSCmd = 1005
	SSCmd_NOTIFY_NAMESERVER_HEARTBEAT SSCmd = 1006
	//agent && agent 2000 ~ 2999
	SSCmd_NOTIFY_SERVICE_MESSAGE SSCmd = 2000
	SSCmd_REQ_PING_SERVICE       SSCmd = 2001
	SSCmd_RSP_PING_SERVICE       SSCmd = 2002
)

var SSCmd_name = map[int32]string{
	0:    "INVALID_CMD",
	1000: "REQ_REGISTER_APP",
	1001: "RSP_REGISTER_APP",
	1002: "REQ_REGISTER_SERVICE",
	1003: "RSP_REGISTER_SERVICE",
	1004: "NOTIFY_SERVICE_ONLINE",
	1005: "REQ_UNREGISTER_SERVICE",
	1006: "NOTIFY_NAMESERVER_HEARTBEAT",
	2000: "NOTIFY_SERVICE_MESSAGE",
	2001: "REQ_PING_SERVICE",
	2002: "RSP_PING_SERVICE",
}

var SSCmd_value = map[string]int32{
	"INVALID_CMD":                 0,
	"REQ_REGISTER_APP":            1000,
	"RSP_REGISTER_APP":            1001,
	"REQ_REGISTER_SERVICE":        1002,
	"RSP_REGISTER_SERVICE":        1003,
	"NOTIFY_SERVICE_ONLINE":       1004,
	"REQ_UNREGISTER_SERVICE":      1005,
	"NOTIFY_NAMESERVER_HEARTBEAT": 1006,
	"NOTIFY_SERVICE_MESSAGE":      2000,
	"REQ_PING_SERVICE":            2001,
	"RSP_PING_SERVICE":            2002,
}

func (x SSCmd) String() string {
	return proto.EnumName(SSCmd_name, int32(x))
}

func (SSCmd) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_38a3ff22a5566682, []int{0}
}

type SSError int32

const (
	SSError_OK                         SSError = 0
	SSError_ERR_APP_NOT_REGISTER       SSError = -1000
	SSError_ERR_SERVICE_REGISTER_AGAIN SSError = -1001
	SSError_ERR_APP_SESSION_LOST       SSError = -1002
)

var SSError_name = map[int32]string{
	0:     "OK",
	-1000: "ERR_APP_NOT_REGISTER",
	-1001: "ERR_SERVICE_REGISTER_AGAIN",
	-1002: "ERR_APP_SESSION_LOST",
}

var SSError_value = map[string]int32{
	"OK":                         0,
	"ERR_APP_NOT_REGISTER":       -1000,
	"ERR_SERVICE_REGISTER_AGAIN": -1001,
	"ERR_APP_SESSION_LOST":       -1002,
}

func (x SSError) String() string {
	return proto.EnumName(SSError_name, int32(x))
}

func (SSError) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_38a3ff22a5566682, []int{1}
}

//注册App相关
type ReqRegisterApp struct {
	ServerAddr           string   `protobuf:"bytes,1,opt,name=ServerAddr,proto3" json:"ServerAddr,omitempty"`
	AppID                string   `protobuf:"bytes,2,opt,name=AppID,proto3" json:"AppID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ReqRegisterApp) Reset()         { *m = ReqRegisterApp{} }
func (m *ReqRegisterApp) String() string { return proto.CompactTextString(m) }
func (*ReqRegisterApp) ProtoMessage()    {}
func (*ReqRegisterApp) Descriptor() ([]byte, []int) {
	return fileDescriptor_38a3ff22a5566682, []int{0}
}

func (m *ReqRegisterApp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReqRegisterApp.Unmarshal(m, b)
}
func (m *ReqRegisterApp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReqRegisterApp.Marshal(b, m, deterministic)
}
func (m *ReqRegisterApp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReqRegisterApp.Merge(m, src)
}
func (m *ReqRegisterApp) XXX_Size() int {
	return xxx_messageInfo_ReqRegisterApp.Size(m)
}
func (m *ReqRegisterApp) XXX_DiscardUnknown() {
	xxx_messageInfo_ReqRegisterApp.DiscardUnknown(m)
}

var xxx_messageInfo_ReqRegisterApp proto.InternalMessageInfo

func (m *ReqRegisterApp) GetServerAddr() string {
	if m != nil {
		return m.ServerAddr
	}
	return ""
}

func (m *ReqRegisterApp) GetAppID() string {
	if m != nil {
		return m.AppID
	}
	return ""
}

type RspRegisterApp struct {
	Result               int32    `protobuf:"varint,2,opt,name=Result,proto3" json:"Result,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RspRegisterApp) Reset()         { *m = RspRegisterApp{} }
func (m *RspRegisterApp) String() string { return proto.CompactTextString(m) }
func (*RspRegisterApp) ProtoMessage()    {}
func (*RspRegisterApp) Descriptor() ([]byte, []int) {
	return fileDescriptor_38a3ff22a5566682, []int{1}
}

func (m *RspRegisterApp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RspRegisterApp.Unmarshal(m, b)
}
func (m *RspRegisterApp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RspRegisterApp.Marshal(b, m, deterministic)
}
func (m *RspRegisterApp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RspRegisterApp.Merge(m, src)
}
func (m *RspRegisterApp) XXX_Size() int {
	return xxx_messageInfo_RspRegisterApp.Size(m)
}
func (m *RspRegisterApp) XXX_DiscardUnknown() {
	xxx_messageInfo_RspRegisterApp.DiscardUnknown(m)
}

var xxx_messageInfo_RspRegisterApp proto.InternalMessageInfo

func (m *RspRegisterApp) GetResult() int32 {
	if m != nil {
		return m.Result
	}
	return 0
}

//注册/注销服务相关
type ServiceInfo struct {
	ServiceName          string   `protobuf:"bytes,1,opt,name=ServiceName,proto3" json:"ServiceName,omitempty"`
	ServiceId            uint64   `protobuf:"varint,2,opt,name=ServiceId,proto3" json:"ServiceId,omitempty"`
	AddrHandle           uint64   `protobuf:"varint,3,opt,name=AddrHandle,proto3" json:"AddrHandle,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ServiceInfo) Reset()         { *m = ServiceInfo{} }
func (m *ServiceInfo) String() string { return proto.CompactTextString(m) }
func (*ServiceInfo) ProtoMessage()    {}
func (*ServiceInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_38a3ff22a5566682, []int{2}
}

func (m *ServiceInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ServiceInfo.Unmarshal(m, b)
}
func (m *ServiceInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ServiceInfo.Marshal(b, m, deterministic)
}
func (m *ServiceInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ServiceInfo.Merge(m, src)
}
func (m *ServiceInfo) XXX_Size() int {
	return xxx_messageInfo_ServiceInfo.Size(m)
}
func (m *ServiceInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_ServiceInfo.DiscardUnknown(m)
}

var xxx_messageInfo_ServiceInfo proto.InternalMessageInfo

func (m *ServiceInfo) GetServiceName() string {
	if m != nil {
		return m.ServiceName
	}
	return ""
}

func (m *ServiceInfo) GetServiceId() uint64 {
	if m != nil {
		return m.ServiceId
	}
	return 0
}

func (m *ServiceInfo) GetAddrHandle() uint64 {
	if m != nil {
		return m.AddrHandle
	}
	return 0
}

type ReqRegisterService struct {
	ServiceInfo          *ServiceInfo `protobuf:"bytes,1,opt,name=service_info,json=serviceInfo,proto3" json:"service_info,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *ReqRegisterService) Reset()         { *m = ReqRegisterService{} }
func (m *ReqRegisterService) String() string { return proto.CompactTextString(m) }
func (*ReqRegisterService) ProtoMessage()    {}
func (*ReqRegisterService) Descriptor() ([]byte, []int) {
	return fileDescriptor_38a3ff22a5566682, []int{3}
}

func (m *ReqRegisterService) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReqRegisterService.Unmarshal(m, b)
}
func (m *ReqRegisterService) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReqRegisterService.Marshal(b, m, deterministic)
}
func (m *ReqRegisterService) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReqRegisterService.Merge(m, src)
}
func (m *ReqRegisterService) XXX_Size() int {
	return xxx_messageInfo_ReqRegisterService.Size(m)
}
func (m *ReqRegisterService) XXX_DiscardUnknown() {
	xxx_messageInfo_ReqRegisterService.DiscardUnknown(m)
}

var xxx_messageInfo_ReqRegisterService proto.InternalMessageInfo

func (m *ReqRegisterService) GetServiceInfo() *ServiceInfo {
	if m != nil {
		return m.ServiceInfo
	}
	return nil
}

type RspRegisterService struct {
	AddrHandle           uint64   `protobuf:"varint,1,opt,name=AddrHandle,proto3" json:"AddrHandle,omitempty"`
	Result               int32    `protobuf:"varint,2,opt,name=Result,proto3" json:"Result,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RspRegisterService) Reset()         { *m = RspRegisterService{} }
func (m *RspRegisterService) String() string { return proto.CompactTextString(m) }
func (*RspRegisterService) ProtoMessage()    {}
func (*RspRegisterService) Descriptor() ([]byte, []int) {
	return fileDescriptor_38a3ff22a5566682, []int{4}
}

func (m *RspRegisterService) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RspRegisterService.Unmarshal(m, b)
}
func (m *RspRegisterService) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RspRegisterService.Marshal(b, m, deterministic)
}
func (m *RspRegisterService) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RspRegisterService.Merge(m, src)
}
func (m *RspRegisterService) XXX_Size() int {
	return xxx_messageInfo_RspRegisterService.Size(m)
}
func (m *RspRegisterService) XXX_DiscardUnknown() {
	xxx_messageInfo_RspRegisterService.DiscardUnknown(m)
}

var xxx_messageInfo_RspRegisterService proto.InternalMessageInfo

func (m *RspRegisterService) GetAddrHandle() uint64 {
	if m != nil {
		return m.AddrHandle
	}
	return 0
}

func (m *RspRegisterService) GetResult() int32 {
	if m != nil {
		return m.Result
	}
	return 0
}

type ReqUnRegisterService struct {
	AddrHandle           uint64   `protobuf:"varint,1,opt,name=AddrHandle,proto3" json:"AddrHandle,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ReqUnRegisterService) Reset()         { *m = ReqUnRegisterService{} }
func (m *ReqUnRegisterService) String() string { return proto.CompactTextString(m) }
func (*ReqUnRegisterService) ProtoMessage()    {}
func (*ReqUnRegisterService) Descriptor() ([]byte, []int) {
	return fileDescriptor_38a3ff22a5566682, []int{5}
}

func (m *ReqUnRegisterService) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReqUnRegisterService.Unmarshal(m, b)
}
func (m *ReqUnRegisterService) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReqUnRegisterService.Marshal(b, m, deterministic)
}
func (m *ReqUnRegisterService) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReqUnRegisterService.Merge(m, src)
}
func (m *ReqUnRegisterService) XXX_Size() int {
	return xxx_messageInfo_ReqUnRegisterService.Size(m)
}
func (m *ReqUnRegisterService) XXX_DiscardUnknown() {
	xxx_messageInfo_ReqUnRegisterService.DiscardUnknown(m)
}

var xxx_messageInfo_ReqUnRegisterService proto.InternalMessageInfo

func (m *ReqUnRegisterService) GetAddrHandle() uint64 {
	if m != nil {
		return m.AddrHandle
	}
	return 0
}

//通知服务上下线相关
type NotifyServiceOnline struct {
	ServerAddr           string       `protobuf:"bytes,1,opt,name=ServerAddr,proto3" json:"ServerAddr,omitempty"`
	ServiceInfo          *ServiceInfo `protobuf:"bytes,2,opt,name=service_info,json=serviceInfo,proto3" json:"service_info,omitempty"`
	IsOnline             bool         `protobuf:"varint,3,opt,name=is_online,json=isOnline,proto3" json:"is_online,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *NotifyServiceOnline) Reset()         { *m = NotifyServiceOnline{} }
func (m *NotifyServiceOnline) String() string { return proto.CompactTextString(m) }
func (*NotifyServiceOnline) ProtoMessage()    {}
func (*NotifyServiceOnline) Descriptor() ([]byte, []int) {
	return fileDescriptor_38a3ff22a5566682, []int{6}
}

func (m *NotifyServiceOnline) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NotifyServiceOnline.Unmarshal(m, b)
}
func (m *NotifyServiceOnline) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NotifyServiceOnline.Marshal(b, m, deterministic)
}
func (m *NotifyServiceOnline) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NotifyServiceOnline.Merge(m, src)
}
func (m *NotifyServiceOnline) XXX_Size() int {
	return xxx_messageInfo_NotifyServiceOnline.Size(m)
}
func (m *NotifyServiceOnline) XXX_DiscardUnknown() {
	xxx_messageInfo_NotifyServiceOnline.DiscardUnknown(m)
}

var xxx_messageInfo_NotifyServiceOnline proto.InternalMessageInfo

func (m *NotifyServiceOnline) GetServerAddr() string {
	if m != nil {
		return m.ServerAddr
	}
	return ""
}

func (m *NotifyServiceOnline) GetServiceInfo() *ServiceInfo {
	if m != nil {
		return m.ServiceInfo
	}
	return nil
}

func (m *NotifyServiceOnline) GetIsOnline() bool {
	if m != nil {
		return m.IsOnline
	}
	return false
}

//跨进程服务消息通知
type NotifyServiceMessage struct {
	SrcService           *ServiceInfo `protobuf:"bytes,1,opt,name=src_service,json=srcService,proto3" json:"src_service,omitempty"`
	DstHandle            uint64       `protobuf:"varint,2,opt,name=dst_handle,json=dstHandle,proto3" json:"dst_handle,omitempty"`
	Data                 []byte       `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *NotifyServiceMessage) Reset()         { *m = NotifyServiceMessage{} }
func (m *NotifyServiceMessage) String() string { return proto.CompactTextString(m) }
func (*NotifyServiceMessage) ProtoMessage()    {}
func (*NotifyServiceMessage) Descriptor() ([]byte, []int) {
	return fileDescriptor_38a3ff22a5566682, []int{7}
}

func (m *NotifyServiceMessage) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NotifyServiceMessage.Unmarshal(m, b)
}
func (m *NotifyServiceMessage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NotifyServiceMessage.Marshal(b, m, deterministic)
}
func (m *NotifyServiceMessage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NotifyServiceMessage.Merge(m, src)
}
func (m *NotifyServiceMessage) XXX_Size() int {
	return xxx_messageInfo_NotifyServiceMessage.Size(m)
}
func (m *NotifyServiceMessage) XXX_DiscardUnknown() {
	xxx_messageInfo_NotifyServiceMessage.DiscardUnknown(m)
}

var xxx_messageInfo_NotifyServiceMessage proto.InternalMessageInfo

func (m *NotifyServiceMessage) GetSrcService() *ServiceInfo {
	if m != nil {
		return m.SrcService
	}
	return nil
}

func (m *NotifyServiceMessage) GetDstHandle() uint64 {
	if m != nil {
		return m.DstHandle
	}
	return 0
}

func (m *NotifyServiceMessage) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

//Agent服务与NameServer心跳
type NotifyNameServerHeartBeat struct {
	SrcHandle            uint64   `protobuf:"varint,1,opt,name=SrcHandle,proto3" json:"SrcHandle,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *NotifyNameServerHeartBeat) Reset()         { *m = NotifyNameServerHeartBeat{} }
func (m *NotifyNameServerHeartBeat) String() string { return proto.CompactTextString(m) }
func (*NotifyNameServerHeartBeat) ProtoMessage()    {}
func (*NotifyNameServerHeartBeat) Descriptor() ([]byte, []int) {
	return fileDescriptor_38a3ff22a5566682, []int{8}
}

func (m *NotifyNameServerHeartBeat) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NotifyNameServerHeartBeat.Unmarshal(m, b)
}
func (m *NotifyNameServerHeartBeat) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NotifyNameServerHeartBeat.Marshal(b, m, deterministic)
}
func (m *NotifyNameServerHeartBeat) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NotifyNameServerHeartBeat.Merge(m, src)
}
func (m *NotifyNameServerHeartBeat) XXX_Size() int {
	return xxx_messageInfo_NotifyNameServerHeartBeat.Size(m)
}
func (m *NotifyNameServerHeartBeat) XXX_DiscardUnknown() {
	xxx_messageInfo_NotifyNameServerHeartBeat.DiscardUnknown(m)
}

var xxx_messageInfo_NotifyNameServerHeartBeat proto.InternalMessageInfo

func (m *NotifyNameServerHeartBeat) GetSrcHandle() uint64 {
	if m != nil {
		return m.SrcHandle
	}
	return 0
}

//Agent服务间Ping包
type ReqServicePing struct {
	Seq                  uint64   `protobuf:"varint,1,opt,name=Seq,proto3" json:"Seq,omitempty"`
	DstHandle            uint64   `protobuf:"varint,2,opt,name=DstHandle,proto3" json:"DstHandle,omitempty"`
	SrcServerAddr        string   `protobuf:"bytes,3,opt,name=SrcServerAddr,proto3" json:"SrcServerAddr,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ReqServicePing) Reset()         { *m = ReqServicePing{} }
func (m *ReqServicePing) String() string { return proto.CompactTextString(m) }
func (*ReqServicePing) ProtoMessage()    {}
func (*ReqServicePing) Descriptor() ([]byte, []int) {
	return fileDescriptor_38a3ff22a5566682, []int{9}
}

func (m *ReqServicePing) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReqServicePing.Unmarshal(m, b)
}
func (m *ReqServicePing) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReqServicePing.Marshal(b, m, deterministic)
}
func (m *ReqServicePing) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReqServicePing.Merge(m, src)
}
func (m *ReqServicePing) XXX_Size() int {
	return xxx_messageInfo_ReqServicePing.Size(m)
}
func (m *ReqServicePing) XXX_DiscardUnknown() {
	xxx_messageInfo_ReqServicePing.DiscardUnknown(m)
}

var xxx_messageInfo_ReqServicePing proto.InternalMessageInfo

func (m *ReqServicePing) GetSeq() uint64 {
	if m != nil {
		return m.Seq
	}
	return 0
}

func (m *ReqServicePing) GetDstHandle() uint64 {
	if m != nil {
		return m.DstHandle
	}
	return 0
}

func (m *ReqServicePing) GetSrcServerAddr() string {
	if m != nil {
		return m.SrcServerAddr
	}
	return ""
}

type RspServicePing struct {
	Seq                  uint64   `protobuf:"varint,1,opt,name=Seq,proto3" json:"Seq,omitempty"`
	SrcHandle            uint64   `protobuf:"varint,2,opt,name=SrcHandle,proto3" json:"SrcHandle,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RspServicePing) Reset()         { *m = RspServicePing{} }
func (m *RspServicePing) String() string { return proto.CompactTextString(m) }
func (*RspServicePing) ProtoMessage()    {}
func (*RspServicePing) Descriptor() ([]byte, []int) {
	return fileDescriptor_38a3ff22a5566682, []int{10}
}

func (m *RspServicePing) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RspServicePing.Unmarshal(m, b)
}
func (m *RspServicePing) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RspServicePing.Marshal(b, m, deterministic)
}
func (m *RspServicePing) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RspServicePing.Merge(m, src)
}
func (m *RspServicePing) XXX_Size() int {
	return xxx_messageInfo_RspServicePing.Size(m)
}
func (m *RspServicePing) XXX_DiscardUnknown() {
	xxx_messageInfo_RspServicePing.DiscardUnknown(m)
}

var xxx_messageInfo_RspServicePing proto.InternalMessageInfo

func (m *RspServicePing) GetSeq() uint64 {
	if m != nil {
		return m.Seq
	}
	return 0
}

func (m *RspServicePing) GetSrcHandle() uint64 {
	if m != nil {
		return m.SrcHandle
	}
	return 0
}

type SSMsg struct {
	Cmd SSCmd `protobuf:"varint,1,opt,name=cmd,proto3,enum=smproto.SSCmd" json:"cmd,omitempty"`
	// Types that are valid to be assigned to Msg:
	//	*SSMsg_RegisterAppReq
	//	*SSMsg_RegisterAppRsp
	//	*SSMsg_RegisterServiceReq
	//	*SSMsg_RegisterServiceRsp
	//	*SSMsg_NotifyServiceOnline
	//	*SSMsg_UnregisterServiceReq
	//	*SSMsg_NotifyNameserverHb
	//	*SSMsg_NotifyServiceMessage
	//	*SSMsg_ServicePingReq
	//	*SSMsg_ServicePingRsp
	Msg                  isSSMsg_Msg `protobuf_oneof:"Msg"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *SSMsg) Reset()         { *m = SSMsg{} }
func (m *SSMsg) String() string { return proto.CompactTextString(m) }
func (*SSMsg) ProtoMessage()    {}
func (*SSMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_38a3ff22a5566682, []int{11}
}

func (m *SSMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SSMsg.Unmarshal(m, b)
}
func (m *SSMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SSMsg.Marshal(b, m, deterministic)
}
func (m *SSMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SSMsg.Merge(m, src)
}
func (m *SSMsg) XXX_Size() int {
	return xxx_messageInfo_SSMsg.Size(m)
}
func (m *SSMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_SSMsg.DiscardUnknown(m)
}

var xxx_messageInfo_SSMsg proto.InternalMessageInfo

func (m *SSMsg) GetCmd() SSCmd {
	if m != nil {
		return m.Cmd
	}
	return SSCmd_INVALID_CMD
}

type isSSMsg_Msg interface {
	isSSMsg_Msg()
}

type SSMsg_RegisterAppReq struct {
	RegisterAppReq *ReqRegisterApp `protobuf:"bytes,100,opt,name=register_app_req,json=registerAppReq,proto3,oneof"`
}

type SSMsg_RegisterAppRsp struct {
	RegisterAppRsp *RspRegisterApp `protobuf:"bytes,101,opt,name=register_app_rsp,json=registerAppRsp,proto3,oneof"`
}

type SSMsg_RegisterServiceReq struct {
	RegisterServiceReq *ReqRegisterService `protobuf:"bytes,102,opt,name=register_service_req,json=registerServiceReq,proto3,oneof"`
}

type SSMsg_RegisterServiceRsp struct {
	RegisterServiceRsp *RspRegisterService `protobuf:"bytes,103,opt,name=register_service_rsp,json=registerServiceRsp,proto3,oneof"`
}

type SSMsg_NotifyServiceOnline struct {
	NotifyServiceOnline *NotifyServiceOnline `protobuf:"bytes,104,opt,name=notify_service_online,json=notifyServiceOnline,proto3,oneof"`
}

type SSMsg_UnregisterServiceReq struct {
	UnregisterServiceReq *ReqUnRegisterService `protobuf:"bytes,105,opt,name=unregister_service_req,json=unregisterServiceReq,proto3,oneof"`
}

type SSMsg_NotifyNameserverHb struct {
	NotifyNameserverHb *NotifyNameServerHeartBeat `protobuf:"bytes,106,opt,name=notify_nameserver_hb,json=notifyNameserverHb,proto3,oneof"`
}

type SSMsg_NotifyServiceMessage struct {
	NotifyServiceMessage *NotifyServiceMessage `protobuf:"bytes,200,opt,name=notify_service_message,json=notifyServiceMessage,proto3,oneof"`
}

type SSMsg_ServicePingReq struct {
	ServicePingReq *ReqServicePing `protobuf:"bytes,201,opt,name=service_ping_req,json=servicePingReq,proto3,oneof"`
}

type SSMsg_ServicePingRsp struct {
	ServicePingRsp *RspServicePing `protobuf:"bytes,202,opt,name=service_ping_rsp,json=servicePingRsp,proto3,oneof"`
}

func (*SSMsg_RegisterAppReq) isSSMsg_Msg() {}

func (*SSMsg_RegisterAppRsp) isSSMsg_Msg() {}

func (*SSMsg_RegisterServiceReq) isSSMsg_Msg() {}

func (*SSMsg_RegisterServiceRsp) isSSMsg_Msg() {}

func (*SSMsg_NotifyServiceOnline) isSSMsg_Msg() {}

func (*SSMsg_UnregisterServiceReq) isSSMsg_Msg() {}

func (*SSMsg_NotifyNameserverHb) isSSMsg_Msg() {}

func (*SSMsg_NotifyServiceMessage) isSSMsg_Msg() {}

func (*SSMsg_ServicePingReq) isSSMsg_Msg() {}

func (*SSMsg_ServicePingRsp) isSSMsg_Msg() {}

func (m *SSMsg) GetMsg() isSSMsg_Msg {
	if m != nil {
		return m.Msg
	}
	return nil
}

func (m *SSMsg) GetRegisterAppReq() *ReqRegisterApp {
	if x, ok := m.GetMsg().(*SSMsg_RegisterAppReq); ok {
		return x.RegisterAppReq
	}
	return nil
}

func (m *SSMsg) GetRegisterAppRsp() *RspRegisterApp {
	if x, ok := m.GetMsg().(*SSMsg_RegisterAppRsp); ok {
		return x.RegisterAppRsp
	}
	return nil
}

func (m *SSMsg) GetRegisterServiceReq() *ReqRegisterService {
	if x, ok := m.GetMsg().(*SSMsg_RegisterServiceReq); ok {
		return x.RegisterServiceReq
	}
	return nil
}

func (m *SSMsg) GetRegisterServiceRsp() *RspRegisterService {
	if x, ok := m.GetMsg().(*SSMsg_RegisterServiceRsp); ok {
		return x.RegisterServiceRsp
	}
	return nil
}

func (m *SSMsg) GetNotifyServiceOnline() *NotifyServiceOnline {
	if x, ok := m.GetMsg().(*SSMsg_NotifyServiceOnline); ok {
		return x.NotifyServiceOnline
	}
	return nil
}

func (m *SSMsg) GetUnregisterServiceReq() *ReqUnRegisterService {
	if x, ok := m.GetMsg().(*SSMsg_UnregisterServiceReq); ok {
		return x.UnregisterServiceReq
	}
	return nil
}

func (m *SSMsg) GetNotifyNameserverHb() *NotifyNameServerHeartBeat {
	if x, ok := m.GetMsg().(*SSMsg_NotifyNameserverHb); ok {
		return x.NotifyNameserverHb
	}
	return nil
}

func (m *SSMsg) GetNotifyServiceMessage() *NotifyServiceMessage {
	if x, ok := m.GetMsg().(*SSMsg_NotifyServiceMessage); ok {
		return x.NotifyServiceMessage
	}
	return nil
}

func (m *SSMsg) GetServicePingReq() *ReqServicePing {
	if x, ok := m.GetMsg().(*SSMsg_ServicePingReq); ok {
		return x.ServicePingReq
	}
	return nil
}

func (m *SSMsg) GetServicePingRsp() *RspServicePing {
	if x, ok := m.GetMsg().(*SSMsg_ServicePingRsp); ok {
		return x.ServicePingRsp
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*SSMsg) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*SSMsg_RegisterAppReq)(nil),
		(*SSMsg_RegisterAppRsp)(nil),
		(*SSMsg_RegisterServiceReq)(nil),
		(*SSMsg_RegisterServiceRsp)(nil),
		(*SSMsg_NotifyServiceOnline)(nil),
		(*SSMsg_UnregisterServiceReq)(nil),
		(*SSMsg_NotifyNameserverHb)(nil),
		(*SSMsg_NotifyServiceMessage)(nil),
		(*SSMsg_ServicePingReq)(nil),
		(*SSMsg_ServicePingRsp)(nil),
	}
}

func init() {
	proto.RegisterEnum("smproto.SSCmd", SSCmd_name, SSCmd_value)
	proto.RegisterEnum("smproto.SSError", SSError_name, SSError_value)
	proto.RegisterType((*ReqRegisterApp)(nil), "smproto.ReqRegisterApp")
	proto.RegisterType((*RspRegisterApp)(nil), "smproto.RspRegisterApp")
	proto.RegisterType((*ServiceInfo)(nil), "smproto.ServiceInfo")
	proto.RegisterType((*ReqRegisterService)(nil), "smproto.ReqRegisterService")
	proto.RegisterType((*RspRegisterService)(nil), "smproto.RspRegisterService")
	proto.RegisterType((*ReqUnRegisterService)(nil), "smproto.ReqUnRegisterService")
	proto.RegisterType((*NotifyServiceOnline)(nil), "smproto.NotifyServiceOnline")
	proto.RegisterType((*NotifyServiceMessage)(nil), "smproto.NotifyServiceMessage")
	proto.RegisterType((*NotifyNameServerHeartBeat)(nil), "smproto.NotifyNameServerHeartBeat")
	proto.RegisterType((*ReqServicePing)(nil), "smproto.ReqServicePing")
	proto.RegisterType((*RspServicePing)(nil), "smproto.RspServicePing")
	proto.RegisterType((*SSMsg)(nil), "smproto.SSMsg")
}

func init() { proto.RegisterFile("ssproto.proto", fileDescriptor_38a3ff22a5566682) }

var fileDescriptor_38a3ff22a5566682 = []byte{
	// 898 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x54, 0xcf, 0x73, 0xdb, 0x44,
	0x14, 0xb6, 0xec, 0x26, 0x6e, 0x9e, 0x5b, 0x47, 0xb3, 0x55, 0x82, 0xda, 0xb4, 0x4c, 0xd0, 0x30,
	0x43, 0xa7, 0x87, 0x1c, 0xca, 0x00, 0xc3, 0x0d, 0xc5, 0x56, 0x63, 0x0d, 0xb1, 0x6c, 0x76, 0x9d,
	0xcc, 0x70, 0xd2, 0x28, 0xd6, 0xda, 0x11, 0x13, 0xcb, 0xb2, 0x56, 0x65, 0x86, 0x5b, 0xef, 0xdc,
	0x81, 0x33, 0xff, 0x0a, 0x17, 0xe0, 0x04, 0xfc, 0x15, 0xfc, 0x3e, 0x73, 0x83, 0xd9, 0xd5, 0xea,
	0xa7, 0x45, 0x0b, 0x39, 0x64, 0x76, 0xdf, 0xbe, 0xf7, 0xbd, 0xef, 0xfb, 0xf4, 0x9e, 0xe1, 0x2e,
	0x63, 0x51, 0xbc, 0x4e, 0xd6, 0x27, 0xe2, 0x3f, 0xea, 0xb2, 0x95, 0x38, 0x18, 0xcf, 0xa0, 0x8f,
	0xe9, 0x06, 0xd3, 0x65, 0xc0, 0x12, 0x1a, 0x9b, 0x51, 0x84, 0x5e, 0x07, 0x20, 0x34, 0xfe, 0x94,
	0xc6, 0xa6, 0xef, 0xc7, 0xba, 0x72, 0xac, 0x3c, 0xde, 0xc3, 0xa5, 0x08, 0xd2, 0x60, 0xc7, 0x8c,
	0x22, 0x7b, 0xa8, 0xb7, 0xc5, 0x53, 0x7a, 0x31, 0x1e, 0x43, 0x1f, 0xb3, 0xa8, 0x8c, 0x73, 0x08,
	0xbb, 0x98, 0xb2, 0xe7, 0x37, 0x89, 0x48, 0xdc, 0xc1, 0xf2, 0x66, 0xac, 0xa0, 0xc7, 0xd1, 0x82,
	0x39, 0xb5, 0xc3, 0xc5, 0x1a, 0x1d, 0xe7, 0x57, 0xc7, 0x5b, 0x51, 0xd9, 0xaf, 0x1c, 0x42, 0x0f,
	0x61, 0x2f, 0x2b, 0xf0, 0x05, 0xd6, 0x2d, 0x5c, 0x04, 0x38, 0x5d, 0x4e, 0x6b, 0xe4, 0x85, 0xfe,
	0x0d, 0xd5, 0x3b, 0xe2, 0xb9, 0x14, 0x31, 0xc6, 0x80, 0x4a, 0x02, 0x65, 0x1d, 0x7a, 0x0f, 0xee,
	0xb0, 0xf4, 0xe8, 0x06, 0xe1, 0x62, 0x2d, 0xda, 0xf6, 0x9e, 0x6a, 0x27, 0xd2, 0x96, 0x93, 0x12,
	0x43, 0xdc, 0x63, 0xc5, 0xc5, 0x38, 0x07, 0x54, 0xd2, 0x99, 0xc1, 0x55, 0x49, 0x28, 0x75, 0x12,
	0xff, 0xea, 0xc5, 0xbb, 0xa0, 0x61, 0xba, 0xb9, 0x08, 0xff, 0x27, 0x9e, 0xf1, 0xb9, 0x02, 0xf7,
	0x9c, 0x75, 0x12, 0x2c, 0x3e, 0x93, 0x15, 0x93, 0xf0, 0x26, 0x08, 0xe9, 0x2b, 0xbf, 0x5d, 0x5d,
	0x76, 0xfb, 0x3f, 0xca, 0x46, 0x47, 0xb0, 0x17, 0x30, 0x77, 0x2d, 0xba, 0x08, 0x93, 0x6f, 0xe3,
	0xdb, 0x01, 0x4b, 0xbb, 0x1a, 0x2f, 0x14, 0xd0, 0x2a, 0x6c, 0xc6, 0x94, 0x31, 0x6f, 0x49, 0xd1,
	0x3b, 0xd0, 0x63, 0xf1, 0xdc, 0x95, 0x40, 0x2f, 0x35, 0x19, 0x58, 0x3c, 0xcf, 0xd4, 0x3f, 0x02,
	0xf0, 0x59, 0xe2, 0x5e, 0xa7, 0xea, 0xe5, 0x17, 0xf7, 0x59, 0x22, 0xcd, 0x44, 0x70, 0xcb, 0xf7,
	0x12, 0x4f, 0xd0, 0xb8, 0x83, 0xc5, 0xd9, 0x78, 0x1f, 0xee, 0xa7, 0x0c, 0xf8, 0xc4, 0xa4, 0x82,
	0x47, 0xd4, 0x8b, 0x93, 0x53, 0xea, 0x25, 0x62, 0x80, 0xe2, 0x79, 0xc5, 0xcc, 0x22, 0x60, 0x2c,
	0xc4, 0x06, 0xc8, 0xde, 0xd3, 0x20, 0x5c, 0x22, 0x15, 0x3a, 0x84, 0x6e, 0x64, 0x26, 0x3f, 0x72,
	0x84, 0x61, 0xd6, 0x3f, 0x23, 0x94, 0x07, 0xd0, 0x9b, 0x70, 0x97, 0xa4, 0xec, 0xa5, 0xf1, 0x1d,
	0x61, 0x7c, 0x35, 0x68, 0x7c, 0x20, 0x36, 0xe4, 0x95, 0x7d, 0x0a, 0xa6, 0xed, 0x3a, 0xd3, 0x6f,
	0x76, 0x61, 0x87, 0x90, 0x31, 0x5b, 0xa2, 0x63, 0xe8, 0xcc, 0x57, 0xbe, 0xa8, 0xec, 0x3f, 0xed,
	0x17, 0x86, 0x92, 0xc1, 0xca, 0xc7, 0xfc, 0x09, 0x0d, 0x40, 0x8d, 0xe5, 0x50, 0xb9, 0x5e, 0x14,
	0xb9, 0x31, 0xdd, 0xe8, 0xbe, 0xf0, 0xff, 0xb5, 0x3c, 0xbd, 0xba, 0xf8, 0xa3, 0x16, 0xee, 0xc7,
	0xc5, 0x15, 0xd3, 0xcd, 0x36, 0x08, 0x8b, 0x74, 0x5a, 0x07, 0xa9, 0x6c, 0x7d, 0x1d, 0x84, 0x45,
	0x68, 0x02, 0x5a, 0x0e, 0x92, 0x0d, 0x1f, 0x67, 0xb3, 0x10, 0x40, 0x47, 0x4d, 0x6c, 0xa4, 0x49,
	0xa3, 0x16, 0x46, 0x71, 0x35, 0xc4, 0x59, 0x35, 0x02, 0xb2, 0x48, 0x5f, 0xd6, 0x01, 0xb7, 0xf6,
	0xb4, 0x09, 0x90, 0x45, 0x08, 0xc3, 0x41, 0x28, 0x86, 0x27, 0x87, 0x93, 0x83, 0x7e, 0x2d, 0x10,
	0x1f, 0xe6, 0x88, 0x0d, 0x2b, 0x37, 0x6a, 0xe1, 0x7b, 0x61, 0xc3, 0x26, 0x5e, 0xc0, 0xe1, 0xf3,
	0xb0, 0x51, 0x77, 0x20, 0x40, 0x1f, 0x95, 0x75, 0x6f, 0xfd, 0x00, 0x8c, 0x5a, 0x58, 0x2b, 0xca,
	0x4b, 0xda, 0x2f, 0x41, 0x93, 0x54, 0x43, 0x6f, 0x45, 0x99, 0x98, 0x2e, 0xf7, 0xfa, 0x4a, 0xff,
	0x44, 0x80, 0x1a, 0x35, 0xa6, 0x0d, 0xcb, 0xc0, 0x2d, 0x08, 0xf3, 0xc7, 0x14, 0x60, 0x74, 0x85,
	0x2e, 0xe1, 0xb0, 0x66, 0xc1, 0x2a, 0xdd, 0x61, 0xfd, 0x5b, 0xa5, 0xc6, 0xb7, 0x69, 0xd3, 0x39,
	0xdf, 0xb0, 0xe9, 0x17, 0x60, 0x08, 0x6a, 0x06, 0x18, 0x05, 0xe1, 0x52, 0x18, 0xf0, 0x9d, 0xb2,
	0x3d, 0x87, 0xa5, 0xb5, 0xe0, 0x23, 0xc4, 0x8a, 0x2b, 0x57, 0xbd, 0x85, 0xc2, 0x22, 0xfd, 0x7b,
	0x65, 0x7b, 0x10, 0x5f, 0x82, 0xc2, 0xa2, 0xd3, 0x1d, 0xe8, 0x8c, 0xd9, 0xf2, 0xc9, 0xd7, 0x6d,
	0xbe, 0x45, 0x83, 0x95, 0x8f, 0xf6, 0xa1, 0x67, 0x3b, 0x97, 0xe6, 0xb9, 0x3d, 0x74, 0x07, 0xe3,
	0xa1, 0xda, 0x42, 0x07, 0xa0, 0x62, 0xeb, 0x23, 0x17, 0x5b, 0x67, 0x36, 0x99, 0x59, 0xd8, 0x35,
	0xa7, 0x53, 0xf5, 0xe7, 0xae, 0x08, 0x93, 0x69, 0x35, 0xfc, 0x4b, 0x17, 0xdd, 0x07, 0xad, 0x92,
	0x4d, 0x2c, 0x7c, 0x69, 0x0f, 0x2c, 0xf5, 0xd7, 0xf4, 0xa9, 0x5c, 0x91, 0x3d, 0xfd, 0xd6, 0x45,
	0x0f, 0xe0, 0xc0, 0x99, 0xcc, 0xec, 0x67, 0x1f, 0x67, 0x41, 0x77, 0xe2, 0x9c, 0xdb, 0x8e, 0xa5,
	0xfe, 0xde, 0x45, 0x47, 0x70, 0xc8, 0x11, 0x2f, 0x9c, 0xad, 0xc2, 0x3f, 0xba, 0xe8, 0x18, 0x8e,
	0x64, 0xa1, 0x63, 0x8e, 0x2d, 0xfe, 0x60, 0x61, 0x77, 0x64, 0x99, 0x78, 0x76, 0x6a, 0x99, 0x33,
	0xf5, 0x4f, 0x51, 0x5e, 0x83, 0x1e, 0x5b, 0x84, 0x98, 0x67, 0x96, 0xfa, 0xc3, 0x7e, 0xa6, 0x6d,
	0x6a, 0x3b, 0x67, 0x39, 0xea, 0x8f, 0xfb, 0x99, 0xb6, 0x4a, 0xf8, 0xa7, 0xfd, 0x27, 0x2f, 0x14,
	0xe8, 0x12, 0x62, 0xc5, 0xf1, 0x3a, 0x46, 0xbb, 0xd0, 0x9e, 0x7c, 0xa8, 0xb6, 0xd0, 0x1b, 0xa0,
	0x59, 0x58, 0xa8, 0x77, 0x9d, 0xc9, 0x2c, 0x17, 0xa7, 0x7e, 0xf5, 0xd7, 0xdf, 0xe9, 0x9f, 0x82,
	0xde, 0x82, 0x07, 0x3c, 0x25, 0x6b, 0x5f, 0x38, 0x76, 0x66, 0xda, 0x8e, 0xfa, 0x65, 0x91, 0x58,
	0xc2, 0x22, 0x16, 0x21, 0xf6, 0xc4, 0x71, 0xcf, 0x27, 0x64, 0xa6, 0x7e, 0x91, 0xa7, 0x5c, 0xed,
	0x8a, 0xcf, 0xfa, 0xf6, 0x3f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x8b, 0x77, 0xe8, 0xe2, 0xba, 0x08,
	0x00, 0x00,
}
