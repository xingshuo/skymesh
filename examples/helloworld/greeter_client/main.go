package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
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
	server skymesh.Server
}

func (c *greeterClient) OnRegister(trans skymesh.Transport, result int32) {
	log.Infof("greeter client register status %d.\n", result)
}

func (c *greeterClient) OnUnRegister() {

}

func (c *greeterClient) OnMessage(rmtAddr *skymesh.Addr, msg []byte) {
	log.Infof("recv server reply %s from %s.\n", string(msg), rmtAddr)
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
	flag.StringVar(&conf, "conf", "config.json", "greeter client config")
	flag.Parse()
	s, err := skymesh.NewServer(conf, appID)
	if err != nil {
		log.Errorf("new server err:%v.\n", err)
		return
	}
	go handleSignal(s)
	c := &greeterClient{server: s}
	err = s.Register(svcUrl, c)
	if err != nil {
		log.Errorf("register %s err:%v\n", svcUrl, err)
		return
	}
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for range ticker.C {
			err := s.SendBySvcUrl(svcUrl, dstUrl, []byte(greetMessage))
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
