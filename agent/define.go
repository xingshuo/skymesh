package skymesh

import "fmt"

const (
	AVERAGE_RTT_SAMPLING_NUM = 5 //计算平均延迟的采样次数
)

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
	return fmt.Sprintf("Addr - ServiceName[%s] ServiceId[%v]", addr.ServiceName, addr.ServiceId)
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

type RegServiceEvent struct {
	dstHandle uint64
	result    int32
}
