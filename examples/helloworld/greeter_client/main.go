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
	trans skymesh.MeshService
}

func (c *greeterClient) OnRegister(trans skymesh.MeshService, result int32) {
	log.Infof("greeter client register status %d.\n", result)
	c.trans = trans
}

func (c *greeterClient) OnUnRegister() {}

func (c *greeterClient) OnMessage(rmtAddr *skymesh.Addr, msg []byte) {
	log.Infof("recv server reply %s from %s.\n", string(msg), rmtAddr)
}

func main() {
	flag.StringVar(&conf, "conf", "config.json", "greeter client config")
	flag.Parse()
	s, err := skymesh.NewServer(conf, appID)
	if err != nil {
		log.Errorf("new server err:%v.\n", err)
		return
	}
	c := &greeterClient{}
	_, err = s.Register(svcUrl, c)
	if err != nil {
		log.Errorf("register %s err:%v\n", svcUrl, err)
		return
	}
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		err := c.trans.SendBySvcUrl(dstUrl, []byte(greetMessage))
		if err != nil {
			log.Errorf("send message err:%v\n", err)
		}
	}
	skymesh.WaitSignalToStop(s, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	ticker.Stop()
	log.Info("server quit.\n")
}
