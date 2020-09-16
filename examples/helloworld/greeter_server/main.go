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
	svcUrl       = fmt.Sprintf("%s.weixin1.greeterServer/101", appID)
	greetMessage = "Nice to meet you too."
)

type greeterServer struct {
	transport skymesh.MeshService
}

func (s *greeterServer) OnRegister(trans skymesh.MeshService, result int32) {
	log.Info("greeter server register ok.\n")
}

func (s *greeterServer) OnUnRegister() {

}

func (s *greeterServer) OnMessage(rmtAddr *skymesh.Addr, msg []byte) {
	log.Infof("recv client msg %s from %s.\n", string(msg),rmtAddr)
	if s.transport != nil {
		s.transport.SendByHandle(rmtAddr.AddrHandle, []byte(greetMessage))
	}
}

func main() {
	flag.StringVar(&conf,"conf", "config.json", "greeter server config")
	flag.Parse()
	s,err := skymesh.NewServer(conf, appID, false)
	if err != nil {
		log.Errorf("new server err:%v.\n", err)
		return
	}
	go skymesh.WaitSignalToStop(s, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	gs := &greeterServer{}
	meshSvc,err := s.Register(svcUrl, gs)
	if err != nil {
		log.Errorf("register %s err:%v\n", svcUrl,err)
		return
	}
	gs.transport = meshSvc
	log.Info("ready to serve.\n")
	if err = s.Serve(); err != nil {
		log.Errorf("serve err:%v.\n", err)
	}
	log.Info("server quit.\n")
}
