package main

import (
	"flag"
	"syscall"

	"github.com/xingshuo/skymesh/log"
	"github.com/xingshuo/skymesh/nameserver"
	skymesh "github.com/xingshuo/skymesh/agent"
)

var (
	conf string
)

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
	go skymesh.WaitSignalToStop(s, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	log.Info("ready to serve.\n")
	if err = s.Serve(); err != nil {
		log.Errorf("server serve err:%v\n", err)
	}
	log.Info("server quit.\n")
}
