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
	svcUrl       = fmt.Sprintf("%s.weixin1.Client/101", appID)
	watchUrl     = fmt.Sprintf("%s.weixin1.Server", appID)
	greetMessage = "Hello."
)

type Client struct {
	server skymesh.Server
}

func (c *Client) OnRegister(trans skymesh.Transport, result int32) {
	log.Infof("greeter client register status %d.\n", result)
}

func (c *Client) OnUnRegister() {

}

func (c *Client) OnMessage(rmtAddr *skymesh.Addr, msg []byte) {
	log.Infof("recv server reply %s from %s.\n", string(msg), rmtAddr)
}

type serverWatcher struct {
	server skymesh.Server
}

func (w *serverWatcher) OnInstOnline(addr *skymesh.Addr) {
	log.Infof("service %s inst online.", addr)
	err := w.server.Send(svcUrl, addr.AddrHandle, []byte(greetMessage))
	if err != nil {
		log.Errorf("client send greet msg err:%v.\n", err)
	} else {
		log.Info("client send greet msg ok.\n")
	}
}
func (w *serverWatcher) OnInstOffline(addr *skymesh.Addr) {
	log.Infof("service %s inst offline.", addr)
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
	c := &Client{server: s}
	err = s.Register(svcUrl, c)
	if err != nil {
		log.Errorf("register %s err:%v\n", svcUrl, err)
		return
	}
	ns := c.server.GetNameResolver(watchUrl)
	ns.Watch(&serverWatcher{s})
	//向已经上线的Server端服务发送greetMessage
	for _, addr := range ns.GetInstsAddr() {
		s.Send(svcUrl, addr.AddrHandle, []byte(greetMessage))
	}

	ticker := time.NewTicker(5 * time.Second)
	go func() {
		cnt := 1
		for range ticker.C {
			if cnt % 4 == 0 {
				s.BroadcastBySvcName(svcUrl, watchUrl, []byte("broadcast by name"))
			} else {
				s.SendBySvcName(svcUrl, watchUrl, []byte("notify one by name"))
			}
			cnt++
		}
	}()
	log.Info("ready to serve.\n")
	if err = s.Serve(); err != nil {
		log.Errorf("serve err:%v.\n", err)
	}
	ticker.Stop()
	log.Info("server quit.\n")
}
