package skymesh

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	smsync "github.com/xingshuo/skymesh/common/sync"
	"github.com/xingshuo/skymesh/log"
)


func NewServer(conf string, appID string, doServe bool) (MeshServer, error) { //每个进程每个appid只启动一个实例
	s := &skymeshServer{}
	err := s.Init(conf, appID, doServe)
	if err != nil {
		return nil, err
	}
	return s, err
}

//GraceServer可以是skymeshServer, grpcServer...
type GraceServer interface {
	GracefulStop()
}

//接收指定信号，优雅退出接口
func WaitSignalToStop(s GraceServer, sigs ...os.Signal) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, sigs...)
	sig := <-c
	log.Infof("MeshServer(%v) exit with signal(%d)\n", syscall.Getpid(), sig)
	s.GracefulStop()
}

type skymeshServer struct {
	appID string
	cfg   Config

	mu         sync.Mutex //Todo: 替换成sync.RWMutex
	quit       *smsync.Event
	done       *smsync.Event
	errQueue   chan error
	recvQueue  chan Message
	eventQueue chan interface{}
	serviceWG  sync.WaitGroup
	state 	   int32

	urlServices        map[string]*skymeshService            //key is AppService Url
	nameGroupServices  map[string]map[uint64]*skymeshService //first floor key is ServiceName(不含实例id)
	handleServices     map[uint64]*skymeshService
	nameRouters        map[string]*skymeshNameRouter //skymesh的名字解析列表

	electionCandidates map[string]map[uint64]*skymeshService //{svcName: {handle: service}}
	electionLeaders    map[string]*Addr //svcName: *Addr
	electionWatchers   map[string]map[uint64]*skymeshService //{watchSvcName: {handle: service}}
	sidecar            *skymeshSidecar
}

func (s *skymeshServer) Init(conf string, appID string, doServe bool) error {
	//加载配置(必须放第一步)
	err := s.loadConfig(conf)
	if err != nil {
		return err
	}
	//数据结构初始化
	s.appID = appID
	s.state = kSkymeshServerIdle
	s.quit = smsync.NewEvent("skymesh.skymeshServer.quit")
	s.done = smsync.NewEvent("skymesh.skymeshServer.done")
	s.errQueue = make(chan error, 1)
	s.recvQueue = make(chan Message, s.cfg.RecvQueueSize)
	s.eventQueue = make(chan interface{}, s.cfg.EventQueueSize)
	s.urlServices = make(map[string]*skymeshService)
	s.nameGroupServices = make(map[string]map[uint64]*skymeshService)
	s.handleServices = make(map[uint64]*skymeshService)
	s.nameRouters = make(map[string]*skymeshNameRouter)
	s.electionCandidates = make(map[string]map[uint64]*skymeshService)
	s.electionWatchers = make(map[string]map[uint64]*skymeshService)
	s.electionLeaders = make(map[string]*Addr)
	//sidecar 初始化
	s.sidecar = &skymeshSidecar{server: s}
	err = s.sidecar.Init()
	if err != nil {
		return err
	}

	if doServe {
		//这里是否需要阻塞确认Running状态再返回?
		errNotify := make(chan error, 1)
		go func() {
			err := s.Serve()
			if err != nil {
				log.Errorf("run serve error:%v", err)
				errNotify <- err
			}
		}()
		select {
		case <-time.After(2 * time.Second):
			if s.isInState(kSkymeshServerRunning) {
				return nil
			} else {
				return errors.New(fmt.Sprintf("do serve timeout:%v", s.state))
			}
		case err := <- errNotify:
			return err
		}
	}
	return nil
}

func (s *skymeshServer) loadConfig(conf string) error {
	data, err := ioutil.ReadFile(conf)
	if err != nil {
		log.Errorf("load config %s failed:%v\n", conf, err)
		return err
	}
	err = json.Unmarshal(data, &s.cfg)
	if err != nil {
		log.Errorf("load config %s failed:%v.\n", conf, err)
		return err
	}
	err = s.cfg.CheckConfig()
	if err != nil {
		log.Errorf("check config err:%v.\n", err)
		return err
	}
	return nil
}

//serviceUrl format : [skymesh://]game_id.env_name.svc_name/inst_id
func (s *skymeshServer) Register(serviceUrl string, service AppService) (MeshService, error) {
	serviceUrl = StripSkymeshUrlPrefix(serviceUrl)
	s.mu.Lock()
	defer s.mu.Unlock()
	log.Infof("register service: %s\n", serviceUrl)
	if svc, ok := s.urlServices[serviceUrl]; ok {
		return svc, errors.New("skymesh register service repeated")
	}
	addr, err := SkymeshUrl2Addr(serviceUrl)
	if err != nil {
		return nil, errors.New("service url format err, should be skymesh://service_name/service_id")
	}
	addr.AddrHandle = MakeServiceHandle(s.cfg.MeshserverAddress, addr.ServiceName, addr.ServiceId)
	log.Infof("%s bind handle %d.\n", serviceUrl, addr.AddrHandle)
	err = s.sidecar.RegisterServiceToNameServer(addr, true)
	if err != nil {
		return nil, err
	}

	svc := &skymeshService{
		server:   s,
		addr:     addr,
		service:  service,
		quit:     smsync.NewEvent("skymesh.skymeshService.quit"),
		done:     smsync.NewEvent("skymesh.skymeshService.done"),
		msgQueue: make(chan Message, s.cfg.ServiceQueueSize),
	}
	s.urlServices[serviceUrl] = svc
	s.handleServices[addr.AddrHandle] = svc
	if s.nameGroupServices[addr.ServiceName] == nil {
		s.nameGroupServices[addr.ServiceName] = make(map[uint64]*skymeshService)
	}
	s.nameGroupServices[addr.ServiceName][addr.AddrHandle] = svc

	s.serviceWG.Add(1)
	go func() {
		svc.Serve()
		log.Infof("service %s quit.\n", svc.addr)
		s.serviceWG.Done()
	}()
	return svc, nil
}

//serviceUrl format : [skymesh://]game_id.env_name.svc_name/inst_id
func (s *skymeshServer) UnRegister(serviceUrl string) error {
	serviceUrl = StripSkymeshUrlPrefix(serviceUrl)
	s.mu.Lock()
	svc := s.urlServices[serviceUrl]
	s.mu.Unlock()
	if svc != nil {
		s.sidecar.RegisterServiceToNameServer(svc.GetLocalAddr(), false)
		svc.Stop()
		s.mu.Lock()
		delete(s.urlServices, serviceUrl)
		delete(s.handleServices, svc.addr.AddrHandle)
		delete(s.nameGroupServices[svc.addr.ServiceName], svc.addr.AddrHandle)
		delete(s.electionCandidates[svc.addr.ServiceName], svc.addr.AddrHandle)
		for _,v := range s.electionWatchers {
			delete(v, svc.addr.AddrHandle)
		}
		leader := s.electionLeaders[svc.addr.ServiceName]
		if leader != nil && leader.AddrHandle == svc.addr.AddrHandle {
			delete(s.electionLeaders, svc.addr.ServiceName)
		}
		for _,wSvc := range s.electionWatchers[svc.addr.ServiceName] {
			elis := wSvc.GetElectionListener()
			if elis != nil {
				elis.OnLeaderChange(leader, KLostElectionLeader)
			}
		}
		s.mu.Unlock()
	}
	return nil
}

func (s *skymeshServer) setState(st int32) bool {
	oldSt := atomic.LoadInt32(&s.state)
	if oldSt != st {
		return atomic.CompareAndSwapInt32(&s.state, oldSt, st)
	}
	return false
}

func (s *skymeshServer) isInState(st int32) bool {
	return atomic.LoadInt32(&s.state) == st
}

func (s *skymeshServer) Serve() error {
	if !s.setState(kSkymeshServerRunning) {
		return errors.New("server already on serving")
	}
	for {
		select {
		case event := <-s.eventQueue:
			switch event.(type) {
			case *OnlineEvent:
				e := event.(*OnlineEvent)
				s.sidecar.notifyServiceOnline(e)
				rh := e.serviceAddr.AddrHandle
				rsname := e.serviceAddr.ServiceName
				s.mu.Lock()
				router := s.nameRouters[rsname]
				s.mu.Unlock()
				if router != nil {
					router.notifyInstsChange(e.isOnline, rh, e.serviceAddr.ServiceId)
				}
			case *RegServiceEvent:
				e := event.(*RegServiceEvent)
				dh := e.dstHandle
				result := e.result
				s.mu.Lock()
				svc := s.handleServices[dh]
				s.mu.Unlock()
				if svc != nil {
					svc.OnRegister(result)
				} else {
					log.Errorf("event not exist handle %d.\n", dh)
				}
			case *SyncAttrEvent:
				e := event.(*SyncAttrEvent)
				s.sidecar.notifyServiceSyncAttr(e)
				rsname := e.serviceAddr.ServiceName
				s.mu.Lock()
				router := s.nameRouters[rsname]
				s.mu.Unlock()
				if router != nil {
					router.notifyInstsAttrs(e.serviceAddr.ServiceId, e.attributes)
				}
			case *ElectionEvent:
				e := event.(*ElectionEvent)
				if e.event == KElectionRunForLeader {
					s.onRunForElectionResultNotify(e.candidate, e.result)
				} else if e.event == KElectionGiveUpLeader {
					s.onGiveUpElectionResultNotify(e.candidate, e.result)
				}
			}
		case msg := <-s.recvQueue: //这里要优先处理消息??
			dh := msg.GetDstHandle()
			s.mu.Lock()
			svc := s.handleServices[dh]
			s.mu.Unlock()
			if svc != nil {
				svc.PushMessage(msg)
			}

		case <-s.quit.Done():
			s.Release()
			return nil

		case err := <-s.errQueue:
			log.Errorf("server err:%v\n", err)
			s.Release()
			return err
		}
	}
	return nil
}

func (s *skymeshServer) GracefulStop() {
	if s.quit.Fire() {
		log.Warning("skymesh quit fire.\n")
		<-s.done.Done()
		log.Warning("skymesh graceful stop.\n")
	}
}

func (s *skymeshServer) Release() {
	if !s.isInState(kSkymeshServerRunning) {
		log.Error("try to release no running server.\n")
		return
	}
	log.Info("release start.\n")
	s.setState(kSkymeshServerStop)
	for len(s.recvQueue) > 0 {
		msg := <-s.recvQueue
		dh := msg.GetDstHandle()
		s.mu.Lock()
		svc := s.handleServices[dh]
		s.mu.Unlock()
		if svc != nil {
			svc.PushMessage(msg)
		}
	}
	log.Info("handle remain msg done.\n")
	var hSvcs []*skymeshService
	s.mu.Lock()
	for _, svc := range s.handleServices {
		hSvcs = append(hSvcs, svc)
	}
	s.mu.Unlock()
	for _,svc := range hSvcs {
		s.sidecar.RegisterServiceToNameServer(svc.GetLocalAddr(), false)
		svc.Stop()
	}
	log.Info("stop all services.\n")
	s.sidecar.Release()
	s.mu.Lock()
	s.urlServices = make(map[string]*skymeshService)
	s.handleServices = make(map[uint64]*skymeshService)
	s.nameGroupServices = make(map[string]map[uint64]*skymeshService)
	s.nameRouters = make(map[string]*skymeshNameRouter)
	s.electionCandidates = make(map[string]map[uint64]*skymeshService)
	s.electionWatchers = make(map[string]map[uint64]*skymeshService)
	s.electionLeaders = make(map[string]*Addr)
	s.mu.Unlock()
	s.serviceWG.Wait()
	log.Warning("all services stop done!.\n")
	s.done.Fire()
}

func (s *skymeshServer) GetBestQualityService(svcName string) *Addr { //优先本地,再远程
	var best *skymeshService
	for _, svc := range s.nameGroupServices[svcName] {
		if best == nil {
			best = svc
		}
		if svc.GetMessageSize() < best.GetMessageSize() {
			best = svc
		}
	}
	if best != nil {
		return best.GetLocalAddr()
	}
	rmSvc := s.sidecar.GetBestQualityService(svcName)
	if rmSvc != nil {
		return rmSvc.serviceAddr
	}
	return nil
}

func (s *skymeshServer) broadcastBySvcName(srcAddr *Addr, dstServiceName string, b []byte) error {
	//本地广播
	s.mu.Lock()
	for lh,dstSvc := range s.nameGroupServices[dstServiceName] {
		msg := &DataMessage{
			srcAddr:   srcAddr,
			dstHandle: lh,
			data:      b,
		}
		dstSvc.PushMessage(msg)
	}
	s.mu.Unlock()
	//远程广播
	s.sidecar.SendAllRemote(srcAddr, dstServiceName, b)
	return nil
}

func (s *skymeshServer) sendBySvcUrl(srcAddr *Addr, dstServiceUrl string, b []byte) error {
	s.mu.Lock()
	dstSvc := s.urlServices[dstServiceUrl]
	s.mu.Unlock()
	if dstSvc != nil {
		msg := &DataMessage{
			srcAddr:   srcAddr,
			dstHandle: dstSvc.addr.AddrHandle,
			data:      b,
		}
		dstSvc.PushMessage(msg)
		return nil
	}
	return s.sidecar.SendRemoteBySvcUrl(srcAddr, dstServiceUrl, b)
}

func (s *skymeshServer) sendByRouter(srcAddr *Addr, serviceName string, b []byte) error { //针对无状态服务
	dstAddr := s.GetBestQualityService(serviceName)
	if dstAddr == nil {
		return fmt.Errorf("not find service %s router", serviceName)
	}
	return s.sendByHandle(srcAddr, dstAddr.AddrHandle, b)
}

func (s *skymeshServer) sendByHandle(srcAddr *Addr, dstHandle uint64, b []byte) error {
	s.mu.Lock()
	dstSvc := s.handleServices[dstHandle]
	s.mu.Unlock()
	if dstSvc != nil { //local service message
		msg := &DataMessage{
			srcAddr:   srcAddr,
			dstHandle: dstHandle,
			data:      b,
		}
		dstSvc.PushMessage(msg)
		return nil
	}
	return s.sidecar.SendRemote(srcAddr, dstHandle, b)
}

func (s *skymeshServer) GetNameRouter(serviceName string) NameRouter {
	s.mu.Lock()
	nr := s.nameRouters[serviceName]
	if nr == nil {
		nr = &skymeshNameRouter{
			server:    s,
			svcName:   serviceName,
			instAddrs: make(map[uint64]*Addr),
			watchers:  make(map[AppRouterWatcher]bool),
			instAttrs: make(map[uint64]ServiceAttr),
		}
		s.nameRouters[serviceName] = nr
	}
	s.mu.Unlock()
	handles, insts := s.getServiceInsts(serviceName)
	for idx, instID := range insts {
		nr.AddInstsAddr(instID, &Addr{ServiceName: serviceName, ServiceId: instID, AddrHandle: handles[idx]})
		h := handles[idx]
		s.mu.Lock()
		svc := s.handleServices[h]
		s.mu.Unlock()
		if svc != nil { //local service
			nr.AddInstsAttr(instID, svc.GetAttribute())
		} else {
			rmtSvc := s.sidecar.getRemoteService(h)
			if rmtSvc != nil {
				nr.AddInstsAttr(instID, rmtSvc.GetAttribute())
			}
		}

	}
	return nr
}

func (s *skymeshServer) getServiceInsts(serviceName string) (handles []uint64, insts []uint64) {
	//先收集本地service实例信息
	s.mu.Lock()
	for lh, svc := range s.nameGroupServices[serviceName] {
		handles = append(handles, lh)
		insts = append(insts, svc.addr.ServiceId)
	}
	s.mu.Unlock()
	//再收集其他进程服务service实例信息
	rhs, rinsts := s.sidecar.getRemoteServiceInsts(serviceName)
	//合并
	handles = append(handles, rhs...)
	insts = append(insts, rinsts...)
	return
}

func (s *skymeshServer) getAllServices() map[uint64]*skymeshService {
	s.mu.Lock()
	defer s.mu.Unlock()
	services := make(map[uint64]*skymeshService)
	for handle,svc := range s.handleServices {
		services[handle] = svc
	}
	return services
}

func (s *skymeshServer) setAttribute(srcAddr *Addr, attrs ServiceAttr) error {
	s.mu.Lock()
	nr := s.nameRouters[srcAddr.ServiceName]
	s.mu.Unlock()
	if nr != nil {
		nr.notifyInstsAttrs(srcAddr.ServiceId, attrs)
	}
	return s.sidecar.SyncServiceAttrToNameServer(srcAddr, attrs)
}

func (s *skymeshServer) onRunForElectionResultNotify(candidate *Addr, result int32) {
	if result == KElectionResultOK {
		newLeader := candidate
		electionName := newLeader.ServiceName
		oldLeader := s.electionLeaders[electionName]
		s.electionLeaders[electionName] = newLeader
		if oldLeader != nil {
			if oldLeader.AddrHandle == newLeader.AddrHandle { //重复通知了??
				return
			}
			//通知竞选人退出成功
			oldSvc := s.handleServices[oldLeader.AddrHandle]
			if oldSvc != nil {
				oldLis := oldSvc.GetElectionListener()
				if oldLis != nil {
					oldLis.OnUnRegisterLeader()
				}
			}
			//通知选举结果监听者旧Leader退出
			for _,svc := range s.electionWatchers[electionName] {
				elis := svc.GetElectionListener()
				if elis != nil {
					elis.OnLeaderChange(oldLeader, KLostElectionLeader)
				}
			}
		}
		//通知竞选人成功当选Leader
		newSvc := s.handleServices[newLeader.AddrHandle]
		if newSvc != nil {
			newLis := newSvc.GetElectionListener()
			if newLis != nil {
				newLis.OnRegisterLeader(newSvc, KElectionResultOK)
			}
		}
		//通知选举结果监听者新Leader当选
		for _,svc := range s.electionWatchers[electionName] {
			elis := svc.GetElectionListener()
			if elis != nil {
				elis.OnLeaderChange(newLeader, KGotElectionLeader)
			}
		}
	} else {
		//通知竞选人竞选失败
		canSvc := s.handleServices[candidate.AddrHandle]
		if canSvc != nil {
			canLis := canSvc.GetElectionListener()
			if canLis != nil {
				canLis.OnRegisterLeader(canSvc, result)
			}
		}
	}
}

func (s *skymeshServer) onGiveUpElectionResultNotify(candidate *Addr, result int32) {
	if result != KElectionResultOK { //目前GiveUp只有成功会通知
		return
	}
	oldLeader := candidate
	electionName := oldLeader.ServiceName
	delete(s.electionLeaders, electionName)
	//通知竞选人退出成功
	oldSvc := s.handleServices[oldLeader.AddrHandle]
	if oldSvc != nil {
		oldLis := oldSvc.GetElectionListener()
		if oldLis != nil {
			oldLis.OnUnRegisterLeader()
		}
	}
	//通知选举结果监听者旧Leader退出
	for _,svc := range s.electionWatchers[electionName] {
		elis := svc.GetElectionListener()
		if elis != nil {
			elis.OnLeaderChange(oldLeader, KLostElectionLeader)
		}
	}
}

func (s *skymeshServer) runForElection(srcAddr *Addr) error {
	svc := s.handleServices[srcAddr.AddrHandle]
	if svc == nil {
		return errors.New("No exist service run for election")
	}
	if s.electionCandidates[srcAddr.ServiceName] == nil {
		s.electionCandidates[srcAddr.ServiceName] = make(map[uint64]*skymeshService)
	}
	if s.electionCandidates[srcAddr.ServiceName][srcAddr.AddrHandle] != nil {
		return errors.New("service run for election again")
	}
	s.electionCandidates[srcAddr.ServiceName][srcAddr.AddrHandle] = svc
	return s.sidecar.SyncNameServerElection(srcAddr, KElectionRunForLeader)
}

func (s *skymeshServer) giveUpElection(srcAddr *Addr) error {
	svc := s.handleServices[srcAddr.AddrHandle]
	if svc == nil {
		return errors.New("No exist service give up election")
	}
	if s.electionCandidates[srcAddr.ServiceName] == nil {
		return errors.New("service not run for election")
	}
	if s.electionCandidates[srcAddr.ServiceName][srcAddr.AddrHandle] == nil {
		return errors.New("service not run for election")
	}
	delete(s.electionCandidates[srcAddr.ServiceName], srcAddr.AddrHandle)
	return s.sidecar.SyncNameServerElection(srcAddr, KElectionGiveUpLeader)
}

func (s *skymeshServer) watchElection(srcAddr *Addr, watchSvcName string) error {
	svc := s.handleServices[srcAddr.AddrHandle]
	if svc == nil {
		return errors.New("No exist service watch election")
	}
	if s.electionWatchers[watchSvcName] == nil {
		s.electionWatchers[watchSvcName] = make(map[uint64]*skymeshService)
	}
	if s.electionWatchers[watchSvcName][srcAddr.AddrHandle] != nil {
		return errors.New("service watch election again")
	}
	s.electionWatchers[watchSvcName][srcAddr.AddrHandle] = svc
	leader := s.electionLeaders[watchSvcName]
	elis := svc.GetElectionListener()
	if leader != nil && elis != nil {
		elis.OnLeaderChange(leader, KGotElectionLeader)
	}
	return nil
}

func (s *skymeshServer) unWatchElection(srcAddr *Addr, watchSvcName string) error {
	svc := s.handleServices[srcAddr.AddrHandle]
	if svc == nil {
		return errors.New("No exist service unwatch election")
	}
	if s.electionWatchers[watchSvcName] == nil {
		return errors.New("service not watch election")
	}
	if s.electionWatchers[watchSvcName][srcAddr.AddrHandle] == nil {
		return errors.New("service not watch election")
	}
	delete(s.electionWatchers[watchSvcName], srcAddr.AddrHandle)
	return nil
}

func (s *skymeshServer) GetElectionLeader(svcName string) *Addr {
	return s.electionLeaders[svcName]
}

func (s *skymeshServer) isElectionLeader(srcAddr *Addr) bool {
	leader := s.electionLeaders[srcAddr.ServiceName]
	if leader == nil {
		return false
	}
	return leader.AddrHandle == srcAddr.AddrHandle
}