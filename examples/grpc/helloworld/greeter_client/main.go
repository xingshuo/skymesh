package main

import (
	"context"
	"flag"
	"fmt"
	skymesh_grpc "github.com/xingshuo/skymesh/extension/grpc"
	"github.com/xingshuo/skymesh/log"
	pb "github.com/xingshuo/skymesh/examples/grpc/helloworld"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"time"
)

var (
	conf         string
	appID        = "testGrpc"
	serverUrl    = fmt.Sprintf("skymesh:///%s.qq1.greeterServer", appID)
)

func main() {
	flag.StringVar(&conf, "conf", "config.json", "greeter client config")
	flag.Parse()
	err := skymesh_grpc.RegisterApp(conf, appID)
	if err != nil {
		return
	}
	defer skymesh_grpc.Release()

	conn, err := grpc.Dial(
		serverUrl,
		grpc.WithInsecure(),
		grpc.WithContextDialer(skymesh_grpc.Dial),
		grpc.WithBalancerName("round_robin"),
	)
	if err != nil {
		log.Errorf("connect url(%s) error(%v)\n", serverUrl, err)
		return
	}
	defer conn.Close()
	log.Infof("connect url(%s) succ\n", serverUrl)
	// Contact the server and print out its response.
	c := pb.NewGreeterClient(conn)
	name := "world"
	for i:=0; i<2; i++ {
		ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
		rmtPeer := &peer.Peer{}
		r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name}, grpc.Peer(rmtPeer))
		if err != nil {
			fmt.Printf("\033[0;31mgrpc call fail(%v)\033[0m\n", err)
		} else {
			fmt.Printf("\033[0;31mGreeting: %s\033[0m\n", r.Message)
			fmt.Printf("\033[0;31mremote service: %v\033[0m\n", rmtPeer.Addr)
		}
	}

}