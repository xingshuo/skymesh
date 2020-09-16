package main

import (
	"flag"
	"fmt"
	"syscall"
	"time"

	skymesh "github.com/xingshuo/skymesh/agent"
	"github.com/xingshuo/skymesh/log"
)

var (
	conf         string
	appID        = "testApp"
	svcUrl       = fmt.Sprintf("%s.weixin1.greeterClient/101", appID)
	dstUrl       = fmt.Sprintf("%s.weixin1.greeterServer/101", appID)
	greetMessage = "Nice to meet you."
)

type greeterClient struct {
}

func (c *greeterClient) OnRegister(trans skymesh.MeshService, result int32) {
	log.Infof("greeter client register status %d.\n", result)
}

func (c *greeterClient) OnUnRegister() {

}

func (c *greeterClient) OnMessage(rmtAddr *skymesh.Addr, msg []byte) {
	log.Infof("recv server reply %s from %s.\n", string(msg), rmtAddr)
}

func main() {
	flag.StringVar(&conf, "conf", "config.json", "greeter client config")
	flag.Parse()
	s, err := skymesh.NewServer(conf, appID, false)
	if err != nil {
		log.Errorf("new server err:%v.\n", err)
		return
	}
	go skymesh.WaitSignalToStop(s, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	c := &greeterClient{}
	meshSvc, err := s.Register(svcUrl, c)
	if err != nil {
		log.Errorf("register %s err:%v\n", svcUrl, err)
		return
	}
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for range ticker.C {
			err := meshSvc.SendBySvcUrl(dstUrl, []byte(greetMessage))
			if err != nil {
				log.Errorf("send message err:%v\n", err)
			}
		}
	}()
	log.Info("ready to serve.\n")
	if err = s.Serve(); err != nil {
		log.Errorf("serve err:%v.\n", err)
	}
	ticker.Stop()
	log.Info("server quit.\n")
}
