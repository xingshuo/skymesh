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
)

type Server struct {
	service skymesh.MeshService
}

func (s *Server) OnRegister(svc skymesh.MeshService, result int32) {
	log.Info("greeter server register ok.\n")
	s.service = svc
}

func (s *Server) OnUnRegister() {

}

func (s *Server) OnMessage(rmtAddr *skymesh.Addr, msg []byte) {
	smsg := string(msg)
	log.Infof("recv client msg: %s.\n", string(smsg))
	attrs := s.service.GetAttribute()
	if attrs[smsg] == nil {
		attrs[smsg] = 1
	} else {
		attrs[smsg] = attrs[smsg].(int) + 1
	}
	s.service.SetAttribute(attrs)
}

func main() {
	flag.StringVar(&conf,"conf", "config.json", "server config")
	flag.Uint64Var(&svcInstId, "instid", svcInstId, "name service inst id")
	flag.Parse()
	s,err := skymesh.NewServer(conf, appID)
	if err != nil {
		log.Errorf("new server err:%v.\n", err)
		return
	}
	svc := &Server{}
	svcUrl := fmt.Sprintf("%s.weixin1.Server/%d", appID, svcInstId)
	_,err = s.Register(svcUrl, svc)
	if err != nil {
		log.Errorf("register %s err:%v\n", svcUrl,err)
		return
	}
	skymesh.WaitSignalToStop(s, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	log.Info("server quit.\n")
}
