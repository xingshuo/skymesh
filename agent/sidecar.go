package skymesh

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

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
	remoteUrlServices   map[string]*remoteService
	remoteGroupServices map[string]map[uint64]*remoteService //ServiceName:{ handle: remoteService}
	healthTicker        *time.Ticker
	keepaliveTicker     *time.Ticker
}

func (sc *skymeshSidecar) Init() error {
	sc.remoteServices = make(map[uint64]*remoteService)
	sc.remoteUrlServices = make(map[string]*remoteService)
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

	//注册app信息到NameServer(要保证先于Register AppService)
	msg := &smproto.SSMsg{
		Cmd: smproto.SSCmd_REQ_REGISTER_APP,
		Msg: &smproto.SSMsg_RegisterAppReq{
			RegisterAppReq: &smproto.ReqRegisterApp{
				ServerAddr: sc.server.cfg.MeshserverAddress,
				AppID:      sc.server.appID,
			},
		},
	}
	b, err := smpack.PackSSMsg(msg)
	if err != nil {
		log.Errorf("pb marshal err:%v.\n", err)
		return err
	}
	d.Send(b)

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

	sc.keepaliveTicker = time.NewTicker(time.Duration(sc.server.cfg.ServiceKeepAliveInterval) * time.Second)
	go func() {
		for range sc.keepaliveTicker.C {
			sc.keepalive()
		}
	}()

	sc.healthTicker = time.NewTicker(time.Duration(sc.server.cfg.ServicePingInterval) * time.Second)
	go func() {
		for range sc.healthTicker.C {
			sc.healthCheck()
		}
	}()
	return nil
}

func (sc *skymeshSidecar) keepalive() {
	for handle, svc := range sc.server.getAllServices() {
		msg := &smproto.SSMsg{
			Cmd: smproto.SSCmd_NOTIFY_NAMESERVER_HEARTBEAT,
			Msg: &smproto.SSMsg_NotifyNameserverHb{
				NotifyNameserverHb: &smproto.NotifyNameServerHeartBeat{
					SrcHandle: handle,
				},
			},
		}
		data, err := smpack.PackSSMsg(msg)
		if err != nil {
			log.Errorf("keepalive %s pb marshal err:%v.\n", svc.addr, err)
			continue
		}
		sc.nameserverDialer.Send(data)
	}
}

func (sc *skymeshSidecar) healthCheck() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	for _, rmSvc := range sc.remoteServices {
		d, err := sc.agentDialers.GetDialer(rmSvc.serverAddr)
		if err != nil {
			log.Errorf("health check not find %s to %s dialer err:%v\n", rmSvc.serviceAddr, rmSvc.serverAddr, err)
			continue
		}
		rmSvc.onSendPing()
		msg := &smproto.SSMsg{
			Cmd: smproto.SSCmd_REQ_PING_SERVICE,
			Msg: &smproto.SSMsg_ServicePingReq{
				ServicePingReq: &smproto.ReqServicePing{
					DstHandle:     rmSvc.serviceAddr.AddrHandle,
					Seq:           rmSvc.curPingSeq,
					SrcServerAddr: sc.server.cfg.MeshserverAddress,
				},
			},
		}
		data, err := smpack.PackSSMsg(msg)
		if err != nil {
			log.Errorf("health check %s pb marshal err:%v.\n", rmSvc.serviceAddr, err)
			continue
		}
		d.Send(data)
	}
}

func (sc *skymeshSidecar) sendPingAck(srcHandle uint64, ackSeq uint64, dstServerAddr string) {
	d, err := sc.agentDialers.GetDialer(dstServerAddr)
	if err != nil {
		log.Errorf("sendPingAck not find %s dialer %d %d\n", dstServerAddr, srcHandle, ackSeq)
		return
	}
	msg := &smproto.SSMsg{
		Cmd: smproto.SSCmd_RSP_PING_SERVICE,
		Msg: &smproto.SSMsg_ServicePingRsp{
			ServicePingRsp: &smproto.RspServicePing{
				Seq:       ackSeq,
				SrcHandle: srcHandle,
			},
		},
	}
	data, err := smpack.PackSSMsg(msg)
	if err != nil {
		log.Errorf("sendPingAck %s %d %d pb marshal err:%v.\n", dstServerAddr, srcHandle, ackSeq, err)
		return
	}
	d.Send(data)
}

func (sc *skymeshSidecar) recvPingAck(srcHandle uint64, ackSeq uint64) {
	sc.mu.Lock()
	rmSvc := sc.remoteServices[srcHandle]
	sc.mu.Unlock()
	if rmSvc == nil {
		log.Errorf("recvPingAck not find service %d %d", srcHandle, ackSeq)
		return
	}
	rmSvc.PingAck(ackSeq)
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

		url := SkymeshAddr2Url(e.serviceAddr, false)
		sc.remoteUrlServices[url] = rmsvc

		rsvcs := sc.remoteGroupServices[rsname]
		if rsvcs == nil {
			rsvcs = make(map[uint64]*remoteService)
		}
		rsvcs[rh] = rmsvc
		sc.remoteGroupServices[rsname] = rsvcs
	} else {
		rh := e.serviceAddr.AddrHandle
		url := SkymeshAddr2Url(e.serviceAddr, false)
		rsname := e.serviceAddr.ServiceName
		rmsvc := sc.remoteServices[rh]
		if rmsvc == nil {
			log.Errorf("recv service %s repeat offline msg.\n", e.serviceAddr)
		}
		delete(sc.remoteServices, rh)
		delete(sc.remoteUrlServices, url)
		delete(sc.remoteGroupServices[rsname], rh)
	}
}

func (sc *skymeshSidecar) notifyServiceSyncAttr(e *SyncAttrEvent) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	rh := e.serviceAddr.AddrHandle
	rmsvc := sc.remoteServices[rh]
	if rmsvc != nil {
		rmsvc.SetAttribute(e.attributes)
	}
}

func (sc *skymeshSidecar) GetBestQualityService(serviceName string) *remoteService {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	var bestSvc *remoteService
	var bestRTT int64
	for _, rmsvc := range sc.remoteGroupServices[serviceName] {
		if bestSvc == nil {
			bestSvc = rmsvc
			bestRTT = rmsvc.CalAverageRTT()
			continue
		}
		rtt := rmsvc.CalAverageRTT()
		if rtt < bestRTT {
			bestSvc = rmsvc
			bestRTT = rtt
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

	b, err := smpack.PackSSMsg(msg)
	if err != nil {
		log.Errorf("pb marshal err:%v.\n", err)
		return err
	}
	sc.nameserverDialer.Send(b)
	return nil
}

func (sc *skymeshSidecar) SyncServiceAttrToNameServer(serviceAddr *Addr, attrs ServiceAttr) error {
	data,err := json.Marshal(attrs)
	if err != nil {
		return err
	}
	msg := &smproto.SSMsg{
		Cmd: smproto.SSCmd_NOTIFY_NAMESERVER_SYNCATTR,
		Msg: &smproto.SSMsg_NotifyNameserverAttr{
			NotifyNameserverAttr: &smproto.NotifyNameServerSyncAttr{
				SrcHandle: serviceAddr.AddrHandle,
				Data: data,
			},
		},
	}

	b, err := smpack.PackSSMsg(msg)
	if err != nil {
		log.Errorf("pb marshal err:%v.\n", err)
		return err
	}
	sc.nameserverDialer.Send(b)
	return nil
}

func (sc *skymeshSidecar) SyncNameServerElection(serviceAddr *Addr, event int32) error {
	msg := &smproto.SSMsg{
		Cmd: smproto.SSCmd_NOTIFY_NAMESERVER_ELECTION,
		Msg: &smproto.SSMsg_NotifyNameserverElection{
			NotifyNameserverElection: &smproto.NotifyNameServerElection{
				SrcHandle:            serviceAddr.AddrHandle,
				Event:                event,
			},
		},
	}

	b, err := smpack.PackSSMsg(msg)
	if err != nil {
		log.Errorf("pb marshal err:%v.\n", err)
		return err
	}
	sc.nameserverDialer.Send(b)
	return nil
}

func (sc *skymeshSidecar) SendAllRemote(srcAddr *Addr, serviceName string, b []byte) {
	var rhs []uint64
	sc.mu.Lock()
	rmSvc := sc.remoteGroupServices[serviceName]
	for rh := range rmSvc {
		rhs = append(rhs, rh)
	}
	sc.mu.Unlock()
	for _, rh := range rhs {
		sc.SendRemote(srcAddr, rh, b)
	}
}

func (sc *skymeshSidecar) SendRemoteBySvcUrl(srcAddr *Addr, dstSvcUrl string, b []byte) error {
	sc.mu.Lock()
	rmsvc := sc.remoteUrlServices[dstSvcUrl]
	sc.mu.Unlock()
	if rmsvc == nil {
		return fmt.Errorf("not find dst service url %s", dstSvcUrl)
	}
	return sc.SendRemote(srcAddr, rmsvc.serviceAddr.AddrHandle, b)
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
		log.Errorf("not find %s agent dialer %s.\n", rmsvc.serviceAddr, rmsvc.serverAddr)
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
	data, err := smpack.PackSSMsg(msg)
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

func (sc *skymeshSidecar) getRemoteService(handle uint64) *remoteService {
	sc.mu.Lock()
	rmsvc := sc.remoteServices[handle]
	sc.mu.Unlock()
	return rmsvc
}

func (sc *skymeshSidecar) Release() {
	sc.mu.Lock()
	sc.remoteServices = make(map[uint64]*remoteService)
	sc.remoteUrlServices = make(map[string]*remoteService)
	sc.remoteGroupServices = make(map[string]map[uint64]*remoteService)
	sc.mu.Unlock()
	sc.keepaliveTicker.Stop()
	sc.healthTicker.Stop()
	sc.nameserverDialer.Shutdown()
	sc.meshserverListener.GracefulStop()
	sc.agentDialers.Release()
}
