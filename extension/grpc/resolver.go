package skymesh_grpc //nolint

import (
	"fmt"
	"log"
	"time"

	"github.com/xingshuo/skymesh/agent"
	"google.golang.org/grpc/resolver"
)


var (
	skymeshCltName = "__grpc_client"
)

type grpcSkymeshResolverBuilder struct{}

func (*grpcSkymeshResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn,
	_ resolver.BuildOptions) (resolver.Resolver, error) {
	r := &grpcSkymeshResolver{
		target: resolver.Target{
			Scheme:    target.Scheme,
			Authority: target.Authority,
			Endpoint:  target.Endpoint,
		},
		cc: cc,
	}
	go r.start()
	return r, nil
}

func (*grpcSkymeshResolverBuilder) Scheme() string { return skymesh.GetSkymeshSchema() }

type grpcSkymeshResolver struct {
	target resolver.Target // 其中的Endpoint务必为完整的3段式服务名
	cc     resolver.ClientConn
	nw     *dialNameWatcher //这里需要具体的子类实例
}

func genSkymeshInstID() uint64 {
	return uint64(time.Now().UnixNano())
}

func (r *grpcSkymeshResolver) start() {
	gameID, envName, _,_ , err := skymesh.ParseSkymeshUrl(r.target.Endpoint)
	if err != nil {
		log.Fatalf("skymesh new router monitor failed:%v", err)
	}
	svcName := fmt.Sprintf("%s.%s.%s/%v", gameID, envName, skymeshCltName, genSkymeshInstID())
	d, err := adaptor.newDialer(svcName, r.target.Endpoint)
	if err != nil {
		log.Fatalf("skymesh new router monitor failed:%v", err)
	}
	nw := d.newNameWatcher(r.target.Endpoint)
	if nw == nil {
		log.Fatalf("skymesh new router monitor failed")
	}

	r.nw = nw
	for {
		rsAddrs := make([]resolver.Address, 0, 32)
		for _,instAddr := range nw.nameResolver.GetInstsAddr() {
			da := dialAddr{Laddr: svcName, Raddr: *instAddr}
			addr, err := dumpDialAddr(da)
			if err != nil {
				panic(fmt.Sprintf("dump dial addr failed: %v\n", err))
			}
			rsAddrs = append(rsAddrs, resolver.Address{Addr: addr})
		}
		r.cc.UpdateState(resolver.State{Addresses: rsAddrs})
		// wait instance notify
		err := nw.wait(-1)
		if err == ErrRouterMonitorClosed {
			return
		}
	}
}

func (*grpcSkymeshResolver) ResolveNow(_ resolver.ResolveNowOptions) {}

func (r *grpcSkymeshResolver) Close() {
	if r.nw != nil {
		r.nw.close()
	}
}

//nolint
func init() {
	resolver.Register(&grpcSkymeshResolverBuilder{})
}
