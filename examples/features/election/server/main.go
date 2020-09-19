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
	svcInstId    uint64 = 101
	greetMessage = "Nice to meet you too."
)

type Server struct {
	service skymesh.MeshService
}

func (s *Server) OnRegister(svc skymesh.MeshService, result int32) {
	log.Info("server register ok.\n")
	s.service = svc
}

func (s *Server) OnUnRegister() {}

func (s *Server) OnMessage(rmtAddr *skymesh.Addr, msg []byte) {
	if string(msg) == "GiveUp" {
		s.service.GiveUpElection()
		log.Infof("server %d give up leader.\n", s.service.GetLocalAddr().ServiceId)
	}
}

type Listener struct {
	service skymesh.MeshService
}

func (l *Listener) OnRegisterLeader(svc skymesh.MeshService, result int32) {
	log.Infof("server %d run for leader result:%d\n", svc.GetLocalAddr().ServiceId, result)
	l.service = svc
}

func (l *Listener) OnUnRegisterLeader() {
	log.Infof("server %d give up leader succeed\n", l.service.GetLocalAddr().ServiceId)
}

func (l *Listener) OnLeaderChange(leader *skymesh.Addr, event skymesh.LeaderChangeEvent) {
	log.Infof("server %d recv [OnLeaderChange] msg\n", l.service.GetLocalAddr().ServiceId)
}

func main() {
	flag.StringVar(&conf,"conf", "config.json", "greeter server config")
	flag.Uint64Var(&svcInstId, "instid", svcInstId, "name service inst id")
	flag.Parse()
	s,err := skymesh.NewServer(conf, appID)
	if err != nil {
		log.Errorf("new server err:%v.\n", err)
		return
	}
	svcUrl := fmt.Sprintf("%s.weixin1.Server/%d", appID, svcInstId)
	service,err := s.Register(svcUrl, &Server{})
	if err != nil {
		log.Errorf("register %s err:%v\n", svcUrl,err)
		return
	}
	service.SetElectionListener(&Listener{})
	service.RunForElection()
	log.Infof("server %d run for leader.\n", service.GetLocalAddr().ServiceId)
	skymesh.WaitSignalToStop(s, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	log.Info("server quit.\n")
}
