package main

import (
	"flag"
	"fmt"
	skymesh "github.com/xingshuo/skymesh/agent"
	"github.com/xingshuo/skymesh/log"
	"syscall"
)

var (
	conf         string
	appID        = "testApp"
	svcUrl       = fmt.Sprintf("%s.weixin1.Client/101", appID)
	watchSvcName = fmt.Sprintf("%s.weixin1.Server", appID)
)

type Client struct {

}

func (c *Client) OnRegister(trans skymesh.MeshService, result int32) {}

func (c *Client) OnUnRegister() {}

func (c *Client) OnMessage(rmtAddr *skymesh.Addr, msg []byte) {}

type Listener struct {
	service skymesh.MeshService
}

func (l *Listener) OnRegisterLeader(svc skymesh.MeshService, result int32) {
	log.Infof("client run for leader result:%d\n", result)
}

func (l *Listener) OnUnRegisterLeader() {
	log.Infof("client giveup leader succeed\n")
}

func (l *Listener) OnLeaderChange(leader *skymesh.Addr, event skymesh.LeaderChangeEvent) {
	log.Infof("client [OnLeaderChange]: server %d %s\n", leader.ServiceId, event)
	if event == skymesh.KGotElectionLeader {
		l.service.SendBySvcNameAndInstID(watchSvcName, leader.ServiceId, []byte("GiveUp"))
	}
}


func main() {
	flag.StringVar(&conf, "conf", "config.json", "greeter client config")
	flag.Parse()
	s, err := skymesh.NewServer(conf, appID)
	if err != nil {
		log.Errorf("new server err:%v.\n", err)
		return
	}
	svc, err := s.Register(svcUrl, &Client{})
	if err != nil {
		log.Errorf("register %s err:%v\n", svcUrl, err)
		return
	}
	svc.SetElectionListener(&Listener{svc})
	svc.WatchElection(watchSvcName)
	skymesh.WaitSignalToStop(s, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	log.Info("server quit.\n")
}
