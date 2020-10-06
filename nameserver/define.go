package nameserver

import skymesh "github.com/xingshuo/skymesh/agent"

const (
	CheckAliveMaxNumPerTickPerApp = 200
	ConfirmExpiredSecond = 5
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
	serviceOpts *skymesh.ServiceOptions
}

type UnRegServiceMsg struct {
	addrHandle uint64
}

type TickMsg struct {

}

type ServiceHeartbeat struct {
	addrHandle uint64
}

type ServiceSyncAttr struct {
	addrHandle uint64
	attrs      []byte
}

type ServiceElection struct {
	addrHandle uint64
	event      int32
}

type RegisterNameRouter struct {
	serverAddr   string
	appid        string
	watchSvcName string
}