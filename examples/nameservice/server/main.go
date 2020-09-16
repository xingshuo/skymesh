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
}

func (s *Server) OnRegister(trans skymesh.MeshService, result int32) {
	log.Info("greeter server register ok.\n")
}

func (s *Server) OnUnRegister() {

}

func (s *Server) OnMessage(rmtAddr *skymesh.Addr, msg []byte) {
	log.Infof("recv client msg: %s from %s.\n", string(msg),rmtAddr)
}

func main() {
	flag.StringVar(&conf,"conf", "config.json", "server config")
	flag.Uint64Var(&svcInstId, "instid", svcInstId, "name service inst id")
	flag.Parse()
	s,err := skymesh.NewServer(conf, appID, false)
	if err != nil {
		log.Errorf("new server err:%v.\n", err)
		return
	}
	go skymesh.WaitSignalToStop(s, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	svc := &Server{}
	svcUrl := fmt.Sprintf("%s.weixin1.Server/%d", appID, svcInstId)
	_,err = s.Register(svcUrl, svc)
	if err != nil {
		log.Errorf("register %s err:%v\n", svcUrl,err)
		return
	}
	log.Info("ready to serve.\n")
	if err = s.Serve(); err != nil {
		log.Errorf("serve err:%v.\n", err)
	}
	log.Info("server quit.\n")
}
