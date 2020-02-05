package skymesh

import (
	"fmt"
	"sync"

	"github.com/golang/protobuf/proto"
	gonet "github.com/xingshuo/skymesh/common/network"
	"github.com/xingshuo/skymesh/log"
	smpack "github.com/xingshuo/skymesh/proto"
	smproto "github.com/xingshuo/skymesh/proto/generate"
)

//负责其他进程agent的连接池管理
type AgentDialerMgr struct {
	mu      sync.Mutex
	dialers map[string]*gonet.Dialer //ServerUrl(ip:port) : dialer
}

func (ds *AgentDialerMgr) Init() {
	ds.dialers = make(map[string]*gonet.Dialer)
}

func (ds *AgentDialerMgr) GetDialer(serverUrl string) (*gonet.Dialer, error) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	d := ds.dialers[serverUrl]
	if d != nil {
		return d, nil
	}
	newReceiver := func() gonet.Receiver {
		return &AgentDialerReceiver{}
	}
	d, err := gonet.NewDialer(serverUrl, newReceiver)
	if err != nil {
		return nil, err
	}
	err = d.Start()
	if err != nil {
		return nil, err
	}
	ds.dialers[serverUrl] = d
	return d, nil
}

func (ds *AgentDialerMgr) Release() {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	for _, dialer := range ds.dialers {
		dialer.Shutdown()
	}
	ds.dialers = make(map[string]*gonet.Dialer)
}

//负责mesh远端服务信息采集、管理以及通信等相关功能
type skymeshSidecar struct {
	mu                  sync.Mutex
	server              *skymeshServer
	nameserverDialer    *gonet.Dialer
	meshserverListener  *gonet.Listener
	agentDialers        *AgentDialerMgr
	remoteServices      map[uint64]*remoteService
	remoteGroupServices map[string]map[uint64]*remoteService //ServiceName(不含实例id):{ handle: remoteService}
}

func (sc *skymeshSidecar) Init() error {
	sc.remoteServices = make(map[uint64]*remoteService)
	sc.remoteGroupServices = make(map[string]map[uint64]*remoteService)
	sc.agentDialers = new(AgentDialerMgr)
	sc.agentDialers.Init()

	//连接名字服务
	newDialerReceiver := func() gonet.Receiver {
		return &NSDialerReceiver{sc.server}
	}
	d, err := gonet.NewDialer(sc.server.cfg.NameserverAddress, newDialerReceiver)
	if err != nil {
		log.Errorf("new dialer failed:%v", err)
		return err
	}
	err = d.Start()
	if err != nil {
		log.Errorf("dialer err:%v", err)
		return err
	}
	sc.nameserverDialer = d

	//注册app信息到NameServer
	go func() {
		msg := &smproto.SSMsg{
			Cmd: smproto.SSCmd_REQ_REGISTER_APP,
			Msg: &smproto.SSMsg_RegisterAppReq{
				RegisterAppReq: &smproto.ReqRegisterApp{
					ServerAddr: sc.server.cfg.MeshserverAddress,
					AppID:      sc.server.appID,
				},
			},
		}
		b, err := proto.Marshal(msg)
		if err != nil {
			log.Errorf("pb marshal err:%v.\n", err)
			sc.server.errQueue <- err
			return
		}
		d.Send(b)
	}()

	//监听接收服务消息端口
	newListenerReceiver := func() gonet.Receiver {
		return &lisConnReceiver{sc.server}
	}
	l, err := gonet.NewListener(sc.server.cfg.MeshserverAddress, newListenerReceiver)
	if err != nil {
		log.Errorf("new listener failed:%v", err)
		return err
	}
	sc.meshserverListener = l
	go func() {
		err := l.Serve()
		if err != nil {
			log.Errorf("listener serve err:%v", err)
			sc.server.errQueue <- err
		} else {
			sc.server.GracefulStop()
		}
		log.Info("listener quit.")
	}()
	return nil
}

func (sc *skymeshSidecar) notifyServiceOnline(e *OnlineEvent) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if e.isOnline {
		rh := e.serviceAddr.AddrHandle
		rsname := e.serviceAddr.ServiceName
		rmsvc := sc.remoteServices[rh]
		if rmsvc != nil {
			log.Errorf("recv service %s repeat online msg.\n", e.serviceAddr)
			return
		}
		rmsvc = &remoteService{
			serverAddr:  e.serverAddr,
			serviceAddr: e.serviceAddr,
		}
		sc.remoteServices[rh] = rmsvc

		rsvcs := sc.remoteGroupServices[rsname]
		if rsvcs == nil {
			rsvcs = make(map[uint64]*remoteService)
		}
		rsvcs[rh] = rmsvc
		sc.remoteGroupServices[rsname] = rsvcs
	} else {
		rh := e.serviceAddr.AddrHandle
		rsname := e.serviceAddr.ServiceName
		sc.mu.Lock()
		rmsvc := sc.remoteServices[rh]
		if rmsvc == nil {
			log.Errorf("recv service %s repeat offline msg.\n", e.serviceAddr)
		}
		delete(sc.remoteServices, rh)
		delete(sc.remoteGroupServices[rsname], rh)
	}
}

func (sc *skymeshSidecar) GetBestQualityService(serviceName string) *remoteService { //Todo: 改成通过心跳包探测不同链路延迟,取最佳
	sc.mu.Lock()
	defer sc.mu.Unlock()
	var bestSvc *remoteService
	for rh, rmsvc := range sc.remoteGroupServices[serviceName] {
		if bestSvc == nil {
			bestSvc = rmsvc
		}
		if bestSvc.serviceAddr.AddrHandle < rh {
			bestSvc = rmsvc
		}
	}
	return bestSvc
}

func (sc *skymeshSidecar) RegisterServiceToNameServer(serviceAddr *Addr, isRegister bool) error {
	var msg *smproto.SSMsg
	if isRegister {
		msg = &smproto.SSMsg{
			Cmd: smproto.SSCmd_REQ_REGISTER_SERVICE,
			Msg: &smproto.SSMsg_RegisterServiceReq{
				RegisterServiceReq: &smproto.ReqRegisterService{
					ServiceInfo: &smproto.ServiceInfo{
						ServiceName: serviceAddr.ServiceName,
						ServiceId:   serviceAddr.ServiceId,
						AddrHandle:  serviceAddr.AddrHandle,
					},
				},
			},
		}
	} else { //UnRegister
		msg = &smproto.SSMsg{
			Cmd: smproto.SSCmd_REQ_UNREGISTER_SERVICE,
			Msg: &smproto.SSMsg_UnregisterServiceReq{
				UnregisterServiceReq: &smproto.ReqUnRegisterService{
					AddrHandle: serviceAddr.AddrHandle,
				},
			},
		}
	}

	b, err := proto.Marshal(msg)
	if err != nil {
		log.Errorf("pb marshal err:%v.\n", err)
		return nil
	}
	data := smpack.Pack(b)
	sc.nameserverDialer.Send(data)
	return nil
}

func (sc *skymeshSidecar) SendRemote(srcAddr *Addr, dstHandle uint64, b []byte) error {
	sc.mu.Lock()
	rmsvc := sc.remoteServices[dstHandle]
	sc.mu.Unlock()
	if rmsvc == nil {
		log.Errorf("not find dst service handle %d.\n", dstHandle)
		return fmt.Errorf("not find dst service handle %d", dstHandle)
	}
	d, err := sc.agentDialers.GetDialer(rmsvc.serverAddr)
	if err != nil {
		log.Errorf("not find %s agent dialer %s.", rmsvc.serviceAddr, rmsvc.serverAddr)
		return err
	}
	msg := &smproto.SSMsg{
		Cmd: smproto.SSCmd_NOTIFY_SERVICE_MESSAGE,
		Msg: &smproto.SSMsg_NotifyServiceMessage{
			NotifyServiceMessage: &smproto.NotifyServiceMessage{
				SrcService: &smproto.ServiceInfo{
					ServiceName: srcAddr.ServiceName,
					ServiceId:   srcAddr.ServiceId,
					AddrHandle:  srcAddr.AddrHandle,
				},
				DstHandle: dstHandle,
				Data:      b,
			},
		},
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		log.Errorf("pb marshal err:%v.\n", err)
		return err
	}
	d.Send(data)
	return nil
}

func (sc *skymeshSidecar) getRemoteServiceInsts(svcName string) (handles []uint64, insts []uint64) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	for rh, svc := range sc.remoteGroupServices[svcName] {
		handles = append(handles, rh)
		insts = append(insts, svc.serviceAddr.ServiceId)
	}
	return
}

func (sc *skymeshSidecar) Release() {
	sc.nameserverDialer.Shutdown()
	sc.meshserverListener.GracefulStop()
	sc.agentDialers.Release()
}
