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
	svcUrl = "testApp.weixin1.greeterClient/101"
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
	log.Infof("recv server reply %s from %s.\n", string(msg),rmtAddr)
}

type greeterServerWatcher struct {
	server skymesh.Server
}

func (w *greeterServerWatcher) OnInstOnline(addr *skymesh.Addr) {
	log.Infof("service %s inst online.",addr)
	err := w.server.Send(svcUrl, addr.AddrHandle, []byte(greetMessage))
	if err != nil {
		log.Errorf("greeter client send msg err:%v.\n", err)
	} else {
		log.Info("send greet msg ok.\n")
	}
}
func (w *greeterServerWatcher) OnInstOffline(addr *skymesh.Addr) {
	log.Info("service %s inst offline.",addr)
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
	flag.StringVar(&conf,"conf", "config.json", "greeter client config")
	flag.Parse()
	s,err := skymesh.NewServer(conf, "testApp")
	if err != nil {
		log.Errorf("new server err:%v.\n", err)
		return
	}
	go handleSignal(s)
	c := &greeterClient{server:s}
	err = s.Register(svcUrl, c)
	if err != nil {
		log.Errorf("register %s err:%v\n",svcUrl,err)
		return
	}
	ns := c.server.GetNameResolver("testApp.weixin1.greeterServer")
	ns.Watch(&greeterServerWatcher{s})
	//向已经上线的Server端服务发送greetMessage
	for _,addr := range ns.GetInstsAddr() {
		s.Send(svcUrl, addr.AddrHandle, []byte(greetMessage))
	}
	log.Info("ready to serve.\n")
	if err = s.Serve(); err != nil {
		log.Errorf("serve err:%v.\n", err)
	}
	log.Info("server quit.\n")
}
