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
	svcUrl       = fmt.Sprintf("%s.weixin1.SMgreeterServer/101", appID)
	replyMessage = "Nice to meet you "
	method       = "SayHello"
)

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

func greetInterceptor(ctx context.Context, b []byte, handler inner_service.HandlerFunc) (interface{}, error) {
	startTime := time.Now()
	rsp,err := handler(ctx, b)
	log.Infof("handle client greet cost time:%vms\n", float64(time.Since(startTime)) / 1e6)
	return rsp,err
}

func StartServer(s skymesh.MeshServer) {
	service,err := inner_service.RegisterService(s, svcUrl)
	if err != nil {
		log.Errorf("register SMService err:%v.\n", err)
		return
	}
	service.ApplyServerInterceptors(greetInterceptor)
	service.RegisterFunc(method, onSayHello)
}

func main() {
	flag.StringVar(&conf,"conf", "config.json", "greeter server config")
	flag.Parse()
	s,err := skymesh.NewServer(conf, appID, false)
	if err != nil {
		log.Errorf("new server err:%v.\n", err)
		return
	}
	go skymesh.WaitSignalToStop(s, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go StartServer(s)

	log.Info("ready to serve.\n")
	if err = s.Serve(); err != nil {
		log.Errorf("serve err:%v.\n", err)
	}
	log.Info("server quit.\n")
}