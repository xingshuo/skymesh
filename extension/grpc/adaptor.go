package skymesh_grpc //nolint

import (
	"net"
	"sync"

	skymesh "github.com/xingshuo/skymesh/agent"
)

var (
	adaptor = &grpcAdaptor{
		dia : make(map[string]*skymeshDialer),
		proto : &VirConnProtoDefault{},
		servers : make(map[string]skymesh.MeshServer),
	}
)

type grpcAdaptor struct {
	mu  sync.Mutex
	dia map[string]*skymeshDialer //as client, key:localServiceUrl ,usually only one dialer
	proto       VirConnProto
	servers     map[string]skymesh.MeshServer // {appID:MeshServer}
}

func (ad *grpcAdaptor) register(conf string, appID string) error {
	ad.mu.Lock()
	s := ad.servers[appID]
	ad.mu.Unlock()
	if s != nil {
		return ErrAppReRegister
	}
	s,err := skymesh.NewServer(conf, appID, true)
	if err != nil {
		return err
	}
	ad.mu.Lock()
	ad.servers[appID] = s
	ad.mu.Unlock()
	return nil
}

func (ad *grpcAdaptor) release() {
	for _,s := range ad.servers {
		s.GracefulStop()
	}
	ad.mu.Lock()
	ad.dia = make(map[string]*skymeshDialer) //只创建,不销毁,交给Server统一释放
	ad.servers = make(map[string]skymesh.MeshServer)
	ad.mu.Unlock()
}

func (ad *grpcAdaptor) setProto(proto VirConnProto) {
	ad.proto = proto
}

func (ad *grpcAdaptor) newDialer(svcName string, target string) (*skymeshDialer, error) {
	ad.mu.Lock()
	d := ad.dia[svcName]
	ad.mu.Unlock()
	if d != nil {
		_,err := d.getNameResolver(target)
		return d, err
	}
	appID, _, _, _, err := skymesh.ParseSkymeshUrl(svcName)
	if err != nil {
		return nil, err
	}
	ad.mu.Lock()
	s := ad.servers[appID]
	ad.mu.Unlock()
	if s == nil {
		return nil, ErrAppNoRegister
	}
	d, err = newSkymeshDialer(svcName, ad.proto, s)
	if err != nil {
		return nil, err
	}
	_,err = d.getNameResolver(target)
	ad.dia[svcName] = d
	return d, err
}

func (ad *grpcAdaptor) getDialer(svcURL string) *skymeshDialer {
	ad.mu.Lock()
	defer ad.mu.Unlock()
	if d, ok := ad.dia[svcURL]; ok {
		return d
	}
	return nil
}

func (ad *grpcAdaptor) newListener(svcName string) (net.Listener, error) {
	appID, _, _, _, err := skymesh.ParseSkymeshUrl(svcName)
	if err != nil {
		return nil, err
	}
	ad.mu.Lock()
	s := ad.servers[appID]
	ad.mu.Unlock()
	if s == nil {
		return nil, ErrAppNoRegister
	}
	return newSkymeshListener(svcName, ad.proto, s)
}
