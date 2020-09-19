package skymesh

import (
	"fmt"
)

const (
	AVERAGE_RTT_SAMPLING_NUM = 5 //计算平均延迟的采样次数
	INVALID_ROUTER_ID = 0 //无效的ServiceId
	INVALID_HANDLE = 0 //无效的ServiceHandle
)

const (
	kSkymeshServerIdle int32 = iota
	kSkymeshServerRunning
	kSkymeshServerStop
)

// if req is KElectionRunForLeader {
// 		rsp is {KElectionRunForLeader: KElectionResultOK}
//         or {KElectionRunForLeader: KElectionResultRunForFail}
//         or {KElectionRunForLeader: KElectionResultAlreadyLeader}
// }
// if req is KElectionGiveUpLeader {
// 		rsp is {KElectionGiveUpLeader: KElectionResultOK} //放弃leader成功,但未选举出新leader
//     	   or {KElectionRunForLeader: KElectionResultOK} //放弃leader成功,并成功选举出新leader,直接通知KElectionRunForLeader,Agent自行通知KLostElectionLeader和KGotElectionLeader
// }
const (
	//发起事件
	KElectionRunForLeader = 1001
	KElectionGiveUpLeader = 1002
	//结果通知事件
	KElectionResultOK            = 0
	KElectionResultRunForFail    = -1001
	KElectionResultAlreadyLeader = -1002
)

type LeaderChangeEvent int

const (
	KGotElectionLeader LeaderChangeEvent = iota
	KLostElectionLeader
)

func (e LeaderChangeEvent) String() string {
	switch e {
	case KGotElectionLeader:
		return "GotLeader"
	case KLostElectionLeader:
		return "LostLeader"
	default:
		return "Invalid-Event"
	}
}

type ServiceAttr map[string]interface{}

type Config struct {
	NameserverAddress string //名字服务ip:port
	MeshserverAddress string //本进程/容器 mesh地址 ip:port
	RecvQueueSize     int    //全局消息队容量
	ServiceQueueSize  int    //单个服务消息队列容量
	EventQueueSize    int    //全局事件队列消息上限
	ServicePingInterval int  //远端服务健康监测周期(秒级)
	ServiceKeepAliveInterval int //agent服务与nameserver保活的心跳间隔(秒级)
}

func (c *Config) CheckConfig() error {
	if len(c.NameserverAddress) < 3 { //最小 :80这种??
		return fmt.Errorf("err nameserver address: %s", c.NameserverAddress)
	}
	if len(c.MeshserverAddress) < 3 {
		return fmt.Errorf("err meshserver address: %s", c.MeshserverAddress)
	}
	if c.RecvQueueSize <= 0 {
		c.RecvQueueSize = 1000
	}
	if c.ServiceQueueSize <= 0 {
		c.ServiceQueueSize = 5000
	}
	if c.EventQueueSize <= 0 {
		c.EventQueueSize = 2000
	}
	if c.ServicePingInterval <= 0 {
		c.ServicePingInterval = 5
	}
	if c.ServiceKeepAliveInterval <= 0 {
		c.ServiceKeepAliveInterval = 5
	}
	return nil
}

type Addr struct {
	ServiceName string //appid.env_name.service_name
	ServiceId   uint64 //instance id
	AddrHandle  uint64 //唯一标识
}

func (addr *Addr) String() string {
	return fmt.Sprintf("Addr - ServiceName[%s] ServiceId[%v] Handle[%v]", addr.ServiceName, addr.ServiceId, addr.AddrHandle)
}

func (addr *Addr) Network() string {
	return skymeshSchema
}

func (addr *Addr) ServiceUrl() string {
	return fmt.Sprintf("%s/%d", addr.ServiceName, addr.ServiceId)
}

func (addr *Addr) GetDetail() (ServiceName string, ServiceId uint64, AddrHandle uint64) {
	return addr.ServiceName, addr.ServiceId, addr.AddrHandle
}

const (
	DATA_MESSAE = 1
	PING_MESSAGE = 2
	REGISTER_MESSAGE = 3
)

type Message interface {
	GetMessageType() int
	GetDstHandle() uint64 //投递给哪个服务
	String() string
}

type DataMessage struct {
	srcAddr   *Addr //消息源服务地址(服务间消息)
	dstHandle uint64
	data      []byte
}

func (m *DataMessage) GetMessageType() int {
	return DATA_MESSAE
}

func (m *DataMessage) GetDstHandle() uint64 {
	return m.dstHandle
}

func (m *DataMessage) String() string {
	return fmt.Sprintf("Data Message(%d) from %s to %d ", len(m.data), m.srcAddr, m.dstHandle)
}

type PingMessage struct {
	srcServerAddr string
	seq uint64
	dstHandle uint64
}

func (m *PingMessage) GetMessageType() int {
	return PING_MESSAGE
}

func (m *PingMessage) GetDstHandle() uint64 {
	return m.dstHandle
}

func (m *PingMessage) String() string {
	return fmt.Sprintf("Ping Message(%d) from %s to %d ", m.seq, m.srcServerAddr, m.dstHandle)
}

type RegisterMessage struct {
	result int32
	svcHandle uint64
}

func (m *RegisterMessage) GetMessageType() int {
	return REGISTER_MESSAGE
}

func (m *RegisterMessage) GetDstHandle() uint64 {
	return m.svcHandle
}

func (m *RegisterMessage) String() string {
	return fmt.Sprintf("Register Message on %d result: %d", m.svcHandle, m.result)
}

type OnlineEvent struct {
	serviceAddr *Addr
	serverAddr  string
	isOnline    bool
}

type RegAppEvent struct {
	leaders   []*Addr
	result    int32
}

type RegServiceEvent struct {
	dstHandle uint64
	result    int32
}

type SyncAttrEvent struct {
	serviceAddr *Addr
	attributes  ServiceAttr
}

type ElectionEvent struct {
	candidate *Addr
	event     int32
	result    int32
}

type KickOffEvent struct {
	dstHandle uint64
}
