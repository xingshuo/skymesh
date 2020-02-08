package nameserver

import skymesh "github.com/xingshuo/skymesh/agent"

const (
	CheckAliveMaxNumPerTickPerApp = 200
)

type Config struct {
	ServerAddress string
	MsgQueueSize  int
	ServiceKeepaliveTime int64 //服务没收到心跳时,保活上限(秒级)
	ServerTickInterval int //服务自检查间隔(秒级)
}

type RegAppMsg struct {
	serverAddr string
	appid      string
}

type RegServiceMsg struct {
	serverAddr  string
	appid       string
	serviceAddr *skymesh.Addr
}

type UnRegServiceMsg struct {
	addrHandle uint64
}

type TickMsg struct {

}

type ServiceHeartbeat struct {
	addrHandle uint64
}
