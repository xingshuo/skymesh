package main

import (
	"context"
	"flag"
	"fmt"
	"syscall"
	"time"

	"github.com/golang/protobuf/proto"
	skymesh "github.com/xingshuo/skymesh/agent"
	pb "github.com/xingshuo/skymesh/examples/inner_service"
	inner_service "github.com/xingshuo/skymesh/extension/inner_service"
	"github.com/xingshuo/skymesh/log"
)

var (
	conf         string
	appID        = "testApp"
	svcUrl       = fmt.Sprintf("%s.weixin1.SMgreeterClient/101", appID)
	dstUrl       = fmt.Sprintf("%s.weixin1.SMgreeterServer/101", appID)
	dstMethod    = "SayHello"
	NotifyName   = "Jerry"
	CallName     = "Bob"
)

func greetInterceptor(ctx context.Context, b []byte, handler inner_service.HandlerFunc) (interface{}, error) {
	startTime := time.Now()
	rsp,err := handler(ctx, b)
	log.Infof("greet rpc cost time:%vms\n", float64(time.Since(startTime)) / 1e6)
	return rsp,err
}

func greetServer(service inner_service.SMService) {
	reqMsg, err := proto.Marshal(&pb.HelloRequest{Name:NotifyName})
	if err != nil {
		log.Errorf("notify proto Marshal err:%v.\n", err)
	}

	err = service.NotifyService(context.Background(), dstUrl, dstMethod, reqMsg)
	if err != nil {
		log.Errorf("notify SMService err:%v.\n", err)
		return
	}

	reqMsg, err = proto.Marshal(&pb.HelloRequest{Name:CallName})
	if err != nil {
		log.Errorf("call proto Marshal err:%v.\n", err)
		return
	}

	rspMsg,err := service.CallService(context.Background(), dstUrl, dstMethod, reqMsg, 10000)
	if err != nil {
		log.Errorf("call SMService err:%v.\n", err)
		return
	}
	rsp := new(pb.HelloReply)
	err = proto.Unmarshal(rspMsg, rsp)
	if err != nil {
		log.Errorf("call proto Unmarshal err:%v.\n", err)
	}
	log.Infof("Greeting: %v\n", rsp.Message)
}

func main() {
	flag.StringVar(&conf, "conf", "config.json", "greeter client config")
	flag.Parse()
	s, err := skymesh.NewServer(conf, appID)
	if err != nil {
		log.Errorf("new server err:%v.\n", err)
		return
	}

	ticker := time.NewTicker(3 * time.Second)
	service,err := inner_service.RegisterService(s, svcUrl)
	if err != nil {
		log.Errorf("register SMService err:%v.\n", err)
		return
	}
	service.ApplyClientInterceptors(greetInterceptor)
	for range ticker.C {
		greetServer(service)
	}

	skymesh.WaitSignalToStop(s, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	ticker.Stop()
	log.Info("server quit.\n")
}