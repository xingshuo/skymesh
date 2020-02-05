package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	skymesh "github.com/xingshuo/skymesh/agent"
	"github.com/xingshuo/skymesh/log"
)

var (
	conf string
)

type greeterClient struct {
	trans skymesh.Transport
}

func (c *greeterClient) OnRegister(trans skymesh.Transport, result int32) {
	log.Info("greeter client register ok.\n")
	c.trans = trans
}

func (c *greeterClient) OnUnRegister() {

}

func (c *greeterClient) OnMessage(rmtAddr *skymesh.Addr, msg []byte) {

}

func (c *greeterClient) SendMessage(dstHandle uint64, msg []byte) error {
	return c.trans.Send(dstHandle, msg)
}

type greeterServerWatcher struct {
	client *greeterClient
}

func (w *greeterServerWatcher) OnInstOnline(addr *skymesh.Addr) {
	log.Info("service %s online.",addr)
	err := w.client.SendMessage(addr.AddrHandle, []byte("hello world"))
	if err != nil {
		log.Errorf("greeter client send msg err:%v.\n", err)
	}
}
func (w *greeterServerWatcher) OnInstOffline(addr *skymesh.Addr) {
	log.Info("service %s offline.",addr)
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
	flag.StringVar(&conf,"conf", "config.json", "nameserver config")
	flag.Parse()
	s,err := skymesh.NewServer(conf, "testApp")
	if err != nil {
		log.Errorf("new server err:%v.\n", err)
		return
	}
	go handleSignal(s)
	svcName := "testApp.weixin1.greeterClient/101"
	c := &greeterClient{}
	err = s.Register(svcName, c)
	if err != nil {
		log.Errorf("register %s err:%v\n",svcName,err)
		return
	}
	ns := s.GetNameResolver("testApp.weixin1.greeterServer")
	ns.Watch(&greeterServerWatcher{c})
	log.Info("ready to serve.\n")
	if err = s.Serve(); err != nil {
		log.Errorf("serve err:%v.\n", err)
	}
	log.Info("server quit.\n")
}
