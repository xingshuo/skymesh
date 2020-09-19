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
	svcUrl       = fmt.Sprintf("%s.weixin1.Client/101", appID)
	watchUrl     = fmt.Sprintf("%s.weixin1.Server", appID)
	greetMessage = "Hello."
)

type Client struct {
	transport skymesh.MeshService
}

func (c *Client) OnRegister(trans skymesh.MeshService, result int32) {
	log.Infof("greeter client register status %d.\n", result)
	c.transport = trans
}

func (c *Client) OnUnRegister() {}

func (c *Client) OnMessage(rmtAddr *skymesh.Addr, msg []byte) {
	log.Infof("recv server reply %s from %s.\n", string(msg), rmtAddr)
}

type serverWatcher struct {
	transport skymesh.MeshService
}

func (w *serverWatcher) OnInstOnline(addr *skymesh.Addr) {
	log.Infof("service %s inst online.", addr)
	err := w.transport.SendByHandle(addr.AddrHandle, []byte(greetMessage))
	if err != nil {
		log.Errorf("client send greet msg err:%v.\n", err)
	} else {
		log.Info("client send greet msg ok.\n")
	}
}
func (w *serverWatcher) OnInstOffline(addr *skymesh.Addr) {
	log.Infof("service %s inst offline.", addr)
}

func (w *serverWatcher) OnInstSyncAttr(addr *skymesh.Addr, attrs skymesh.ServiceAttr) {
	log.Infof("service: %d sync attr: %v.\n", addr.ServiceId, attrs)
}

func main() {
	flag.StringVar(&conf, "conf", "config.json", "greeter client config")
	flag.Parse()
	s, err := skymesh.NewServer(conf, appID)
	if err != nil {
		log.Errorf("new server err:%v.\n", err)
		return
	}
	c := &Client{}
	_,err = s.Register(svcUrl, c)
	if err != nil {
		log.Errorf("register %s err:%v\n", svcUrl, err)
		return
	}
	ns := s.GetNameRouter(watchUrl)
	ns.Watch(&serverWatcher{c.transport})

	ticker := time.NewTicker(5 * time.Second)
	loopSeq := 1
	for range ticker.C {
		router_id := ns.SelectRouterByLoop()
		c.transport.SendBySvcNameAndInstID(watchUrl, router_id, []byte("RouterByLoop"))
		router_id = ns.SelectRouterByModHash(uint64(loopSeq))
		c.transport.SendBySvcNameAndInstID(watchUrl, router_id, []byte("RouterByModHash"))
		loopSeq++
	}
	skymesh.WaitSignalToStop(s, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	ticker.Stop()
	log.Info("server quit.\n")
}
