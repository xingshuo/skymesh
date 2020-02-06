package skymesh

import "fmt"

type Config struct {
	NameserverAddress string //名字服务ip:port
	MeshserverAddress string //本进程/容器 mesh地址 ip:port
	RecvQueueSize     int    //全局消息队容量
	ServiceQueueSize  int    //单个服务消息队列容量
	EventQueueSize    int    //全局事件队列消息上限
}

type Addr struct {
	ServiceName string //appid.env_name.service_name
	ServiceId   uint64 //instance id
	AddrHandle  uint64 //唯一标识
}

func (addr *Addr) String() string {
	return fmt.Sprintf("Addr - ServiceName[%s] ServiceId[%v]", addr.ServiceName, addr.ServiceId)
}

type Message struct {
	srcAddr   *Addr //消息源服务地址(服务间消息)
	dstHandle uint64
	data      []byte
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
