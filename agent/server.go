package skymesh

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"sync"
	"sync/atomic"
	"time"

	smsync "github.com/xingshuo/skymesh/common/sync"
	"github.com/xingshuo/skymesh/log"
)


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
	nameElections      map[string]*skymeshElection

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
	s.nameElections = make(map[string]*skymeshElection)

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
	svc := s.urlServices[serviceUrl]
	s.mu.Unlock()
	log.Infof("register service: %s\n", serviceUrl)
	if svc != nil {
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

	svc = &skymeshService {
		server:   s,
		addr:     addr,
		service:  service,
		quit:     smsync.NewEvent("skymesh.skymeshService.quit"),
		done:     smsync.NewEvent("skymesh.skymeshService.done"),
		msgQueue: make(chan Message, s.cfg.ServiceQueueSize),
		regNotify:make(chan error, 1),
	}
	s.mu.Lock()
	s.urlServices[serviceUrl] = svc
	s.handleServices[addr.AddrHandle] = svc
	if s.nameGroupServices[addr.ServiceName] == nil {
		s.nameGroupServices[addr.ServiceName] = make(map[uint64]*skymeshService)
	}
	s.nameGroupServices[addr.ServiceName][addr.AddrHandle] = svc
	s.mu.Unlock()

	s.serviceWG.Add(1)
	go func() {
		svc.Serve()
		log.Infof("service %s quit.\n", svc.addr)
		s.serviceWG.Done()
	}()
	// 同步等待注册结果通知
	select {
	case <-time.After(3 * time.Second):
		s.UnRegister(serviceUrl)
		return nil, fmt.Errorf("register service timeout")
	case err := <- svc.regNotify:
		if err != nil {
			s.UnRegister(serviceUrl)
			return nil, err
		}
	}
	return svc, nil
}

//serviceUrl format : [skymesh://]game_id.env_name.svc_name/inst_id
func (s *skymeshServer) UnRegister(serviceUrl string) error {
	serviceUrl = StripSkymeshUrlPrefix(serviceUrl)
	s.mu.Lock()
	svc := s.urlServices[serviceUrl]
	s.mu.Unlock()
	if svc != nil {
		s.sidecar.RegisterServiceToNameServer(svc.addr, false)
		elec := s.getNameElection(svc.addr.ServiceName)
		elec.unRegisterService(svc.addr)
		svc.Stop()
		s.mu.Lock()
		delete(s.urlServices, serviceUrl)
		delete(s.handleServices, svc.addr.AddrHandle)
		delete(s.nameGroupServices[svc.addr.ServiceName], svc.addr.AddrHandle)
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
			case *RegAppEvent:
				e := event.(*RegAppEvent)
				s.onRegisterApp(e)
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
				elec := s.getNameElection(e.candidate.ServiceName)
				if e.event == KElectionRunForLeader {
					elec.onRunForElectionResultNotify(e.candidate, e.result)
				} else if e.event == KElectionGiveUpLeader {
					elec.onGiveUpElectionResultNotify(e.candidate, e.result)
				}
			case *KickOffEvent:
				e := event.(*KickOffEvent)
				dh := e.dstHandle
				s.mu.Lock()
				svc := s.handleServices[dh]
				s.mu.Unlock()
				if svc != nil {
					s.UnRegister(svc.addr.ServiceUrl())
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
	log.Debug("release start.\n")
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
	log.Debug("handle remain msg done.\n")
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
	log.Debug("stop all services.\n")
	s.sidecar.Release()
	s.mu.Lock()
	s.urlServices = make(map[string]*skymeshService)
	s.handleServices = make(map[uint64]*skymeshService)
	s.nameGroupServices = make(map[string]map[uint64]*skymeshService)
	s.nameRouters = make(map[string]*skymeshNameRouter) //没有逐个release资源??
	s.nameElections = make(map[string]*skymeshElection) //没有逐个release资源??
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
	group := s.nameGroupServices[dstServiceName]
	s.mu.Unlock()
	for lh,dstSvc := range group {
		msg := &DataMessage{
			srcAddr:   srcAddr,
			dstHandle: lh,
			data:      b,
		}
		dstSvc.PushMessage(msg)
	}
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

func (s *skymeshServer) getService(handle uint64) *skymeshService {
	s.mu.Lock()
	svc := s.handleServices[handle]
	s.mu.Unlock()
	return svc
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

func (s *skymeshServer) onRegisterApp(e *RegAppEvent) {
	for _,addr := range e.leaders {
		elec := s.getNameElection(addr.ServiceName)
		elec.onRegisterApp(addr)
	}
}

func (s *skymeshServer) getNameElection(svcName string) *skymeshElection {
	s.mu.Lock()
	elec := s.nameElections[svcName]
	if elec == nil {
		elec = &skymeshElection{
			svcName:    svcName,
			server:     s,
			candidates: make(map[uint64]AppElectionCandidate),
			watchers:   make(map[uint64]AppElectionWatcher),
		}
		s.nameElections[svcName] = elec
	}
	s.mu.Unlock()
	return elec
}

func (s *skymeshServer) runForElection(srcAddr *Addr, candidate AppElectionCandidate) error {
	if candidate == nil {
		return errors.New("empty app election candidate")
	}
	elec := s.getNameElection(srcAddr.ServiceName)
	err := elec.runForElection(srcAddr, candidate)
	if err != nil {
		return err
	}
	return s.sidecar.SyncNameServerElection(srcAddr, KElectionRunForLeader)
}

func (s *skymeshServer) giveUpElection(srcAddr *Addr) error {
	elec := s.getNameElection(srcAddr.ServiceName)
	err := elec.giveUpElection(srcAddr)
	if err != nil {
		return err
	}
	return s.sidecar.SyncNameServerElection(srcAddr, KElectionGiveUpLeader)
}

func (s *skymeshServer) watchElection(srcAddr *Addr, watchName string, watcher AppElectionWatcher) error {
	if watcher == nil {
		return errors.New("empty app election watcher")
	}
	elec := s.getNameElection(watchName)
	return elec.watchElection(srcAddr, watcher)
}

func (s *skymeshServer) unWatchElection(srcAddr *Addr, watchName string) error {
	elec := s.getNameElection(watchName)
	return elec.unWatchElection(srcAddr)
}

func (s *skymeshServer) GetElectionLeader(svcName string) *Addr {
	elec := s.getNameElection(svcName)
	return elec.getElectionLeader()
}

func (s *skymeshServer) isElectionLeader(srcAddr *Addr) bool {
	elec := s.getNameElection(srcAddr.ServiceName)
	return elec.isElectionLeader(srcAddr)
}