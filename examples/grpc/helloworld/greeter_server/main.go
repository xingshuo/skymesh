package main

import (
	"context"
	"flag"
	"fmt"
	skymesh "github.com/xingshuo/skymesh/agent"
	skymesh_grpc "github.com/xingshuo/skymesh/extension/grpc"
	"log"
	"sync"
	"google.golang.org/grpc"
	pb "github.com/xingshuo/skymesh/examples/grpc/helloworld"
	"syscall"
)

var (
	conf         string
	appID        = "testGrpc"
	serverUrl    = fmt.Sprintf("skymesh://%s.qq1.greeterServer/", appID)
	addrs = []string{serverUrl + "50051", serverUrl + "50052"}
)

type server struct{
	addr string
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	fmt.Printf("\033[0;31m%s receive request: %v\033[0m\n", s.addr, in.Name)
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

func startServer(addr string) {
	lis, err := skymesh_grpc.Listen(addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{addr: addr})
	log.Printf("serving on %s\n", addr)
	go skymesh.WaitSignalToStop(s, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func main() {
	flag.StringVar(&conf, "conf", "config.json", "greeter client config")
	flag.Parse()
	err := skymesh_grpc.RegisterApp(conf, appID)
	if err != nil {
		return
	}
	defer skymesh_grpc.Release()
	var wg sync.WaitGroup
	for _, addr := range addrs {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			startServer(addr)
		}(addr)
	}
	wg.Wait()
}