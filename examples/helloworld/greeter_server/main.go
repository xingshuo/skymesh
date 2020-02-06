package main

import (
	"flag"
	"fmt"
	skymesh "github.com/xingshuo/skymesh/agent"
	"github.com/xingshuo/skymesh/log"
	"os"
	"os/signal"
	"syscall"
)

var (
	conf         string
	svcUrl       = "testApp.weixin1.greeterServer/101"
	greetMessage = "Nice to meet you too."
)

type greeterServer struct {
	server skymesh.Server
}

func (s *greeterServer) OnRegister(trans skymesh.Transport, result int32) {
	log.Info("greeter server register ok.\n")
}

func (s *greeterServer) OnUnRegister() {

}

func (s *greeterServer) OnMessage(rmtAddr *skymesh.Addr, msg []byte) {
	log.Infof("recv client msg %s from %s.\n", string(msg),rmtAddr)
	s.server.Send(svcUrl, rmtAddr.AddrHandle, []byte(greetMessage))
}

func handleSignal(s skymesh.Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c)
	for sig := range c {
		fmt.Printf("recv sig %d\n", sig)
		if sig == syscall.SIGINT || sig == syscall.SIGTERM || sig == syscall.SIGQUIT {
			s.GracefulStop()
		}
	}
}

func main() {
	flag.StringVar(&conf,"conf", "config.json", "greeter server config")
	flag.Parse()
	s,err := skymesh.NewServer(conf, "testApp")
	if err != nil {
		log.Errorf("new server err:%v.\n", err)
		return
	}
	go handleSignal(s)
	gs := &greeterServer{server:s}
	err = s.Register(svcUrl, gs)
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
