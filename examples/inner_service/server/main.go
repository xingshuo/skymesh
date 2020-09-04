package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	skymesh "github.com/xingshuo/skymesh/agent"
	"github.com/xingshuo/skymesh/log"
	"github.com/golang/protobuf/proto"
	"github.com/xingshuo/skymesh/log"
	pb "github.com/xingshuo/skymesh/examples/inner_service"
	inner_service "github.com/xingshuo/skymesh/extension/inner_service"
)

var (
	conf         string
	appID        = "testApp"
	svcUrl       = fmt.Sprintf("%s.weixin1.SMgreeterServer/101", appID)
	replyMessage = "Nice to meet you "
	method       = "SayHello"
)


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

func onSayHello(ctx context.Context, msg []byte) (interface{}, error) {
	req := new(pb.HelloRequest)
	err := proto.Unmarshal(msg, req)
	if err != nil {
		return nil, fmt.Errorf("invail request")
	}
	log.Infof("recv greet msg from %s\n", req.Name)
	rsp := &pb.HelloReply{Message: replyMessage + req.Name}
	return proto.Marshal(rsp)
}

func StartServer(s skymesh.Server) {
	service,err := inner_service.RegisterService(s, svcUrl)
	if err != nil {
		log.Errorf("register SMService err:%v.\n", err)
		return
	}
	service.RegisterFunc(method, onSayHello)
}

func main() {
	flag.StringVar(&conf,"conf", "config.json", "greeter server config")
	flag.Parse()
	s,err := skymesh.NewServer(conf, appID)
	if err != nil {
		log.Errorf("new server err:%v.\n", err)
		return
	}
	go handleSignal(s)
	go StartServer(s)

	log.Info("ready to serve.\n")
	if err = s.Serve(); err != nil {
		log.Errorf("serve err:%v.\n", err)
	}
	log.Info("server quit.\n")
}