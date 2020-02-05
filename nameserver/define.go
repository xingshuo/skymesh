package nameserver

import skymesh "github.com/xingshuo/skymesh/agent"

type Config struct {
	ServerAddress string
	MsgQueueSize  int
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
