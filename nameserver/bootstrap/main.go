package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/xingshuo/skymesh/log"
	"github.com/xingshuo/skymesh/nameserver"
)

var (
	conf string
)

func handleSignal(s *nameserver.Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c)
	for sig := range c {
		fmt.Printf("recv sig %d\n", sig)
		if sig == syscall.SIGINT || sig == syscall.SIGTERM || sig == syscall.SIGQUIT {
			s.Stop()
		}
	}
}

func main() {
	flag.StringVar(&conf,"conf", "config.json", "nameserver config")
	flag.Parse()
	log.Infof("server load conf:%s\n", conf)
	s := &nameserver.Server{}
	err := s.Init(conf)
	if err != nil {
		log.Errorf("server init err:%v\n", err)
		return
	}
	go handleSignal(s)
	log.Info("ready to serve.\n")
	if err = s.Serve(); err != nil {
		log.Errorf("server serve err:%v\n", err)
	}
	log.Info("server quit.\n")
}
