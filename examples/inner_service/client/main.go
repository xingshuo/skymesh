package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"context"
	"time"

	skymesh "github.com/xingshuo/skymesh/agent"
	"github.com/xingshuo/skymesh/log"
	"github.com/golang/protobuf/proto"
	pb "github.com/xingshuo/skymesh/examples/inner_service"
	inner_service "github.com/xingshuo/skymesh/extension/inner_service"
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
	s, err := skymesh.NewServer(conf, appID, false)
	if err != nil {
		log.Errorf("new server err:%v.\n", err)
		return
	}
	go handleSignal(s)

	ticker := time.NewTicker(3 * time.Second)
	go func() {
		service,err := inner_service.RegisterService(s, svcUrl)
		if err != nil {
			log.Errorf("register SMService err:%v.\n", err)
			return
		}
		service.ApplyClientInterceptors(greetInterceptor)
		for range ticker.C {
			greetServer(service)
		}
	}()

	log.Info("ready to serve.\n")
	if err = s.Serve(); err != nil {
		log.Errorf("serve err:%v.\n", err)
	}
	ticker.Stop()
	log.Info("server quit.\n")
}