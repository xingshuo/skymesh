package skymesh

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"sync"

	smsync "github.com/xingshuo/skymesh/common/sync"
	"github.com/xingshuo/skymesh/log"
)

type Server interface {
	Register(serviceUrl string, service Service) error //注册服务
	UnRegister(serviceUrl string) error  //注销服务
	GetNameResolver(serviceName string) NameResolver //返回serviceName的名字解析器
	Send(srcServiceUrl string, dstHandle uint64, b []byte) error //定向发送, 适用有状态服务
	SendByRouter(srcServiceUrl string, dstServiceName string, b []byte) error //根据ServiceName的所有链路质量,选择最佳发送,适用无状态
	Serve() error  //阻塞循环
	GracefulStop() //优雅退出
}

func NewServer(conf string, appID string) (Server, error) { //每个进程每个appid只启动一个实例
	s := &skymeshServer{}
	err := s.Init(conf, appID)
	if err != nil {
		return nil, err
	}
	return s, err
}

type skymeshServer struct {
	appID string
	cfg   Config

	mu         sync.Mutex
	quit       *smsync.Event
	done       *smsync.Event
	errQueue   chan error
	recvQueue  chan Message
	eventQueue chan interface{}
	serviceWG  sync.WaitGroup

	urlServices       map[string]*skymeshService            //key is Service Url
	nameGroupServices map[string]map[uint64]*skymeshService //first floor key is ServiceName(不含实例id)
	handleServices    map[uint64]*skymeshService
	resolvers         map[string]*skymeshResolver //skymesh的名字解析列表
	sidecar           *skymeshSidecar
}

func (s *skymeshServer) Init(conf string, appID string) error {
	//加载配置(必须放第一步)
	err := s.loadConfig(conf)
	if err != nil {
		return err
	}
	//数据结构初始化
	s.appID = appID
	s.quit = smsync.NewEvent("skymesh.skymeshServer.quit")
	s.done = smsync.NewEvent("skymesh.skymeshServer.done")
	s.errQueue = make(chan error, 1)
	s.recvQueue = make(chan Message, s.cfg.RecvQueueSize)
	s.eventQueue = make(chan interface{}, s.cfg.EventQueueSize)
	s.urlServices = make(map[string]*skymeshService)
	s.nameGroupServices = make(map[string]map[uint64]*skymeshService)
	s.handleServices = make(map[uint64]*skymeshService)
	s.resolvers = make(map[string]*skymeshResolver)
	//sidecar 初始化
	s.sidecar = &skymeshSidecar{server: s}
	err = s.sidecar.Init()
	if err != nil {
		return err
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

//serviceUrl format : game_id.env_name.svc_name/inst_id
func (s *skymeshServer) Register(serviceUrl string, service Service) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	log.Infof("register service: %s\n", serviceUrl)
	if _, ok := s.urlServices[serviceUrl]; ok {
		return errors.New("skymesh register service repeated")
	}
	addr, err := SkymeshUrl2Addr(serviceUrl, false)
	if err != nil {
		return errors.New("service url format err, should be skymesh://service_name/service_id")
	}
	addr.AddrHandle = MakeServiceHandle(s.cfg.MeshserverAddress, addr.ServiceName, addr.ServiceId)
	log.Infof("%s bind handle %d.\n", serviceUrl, addr.AddrHandle)
	err = s.sidecar.RegisterServiceToNameServer(addr, true)
	if err != nil {
		return err
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
	return nil
}

//serviceUrl format : game_id.env_name.svc_name/inst_id
func (s *skymeshServer) UnRegister(serviceUrl string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if svc := s.urlServices[serviceUrl]; svc != nil {
		s.sidecar.RegisterServiceToNameServer(svc.GetLocalAddr(), false)
		svc.Stop()
		delete(s.urlServices, serviceUrl)
		delete(s.handleServices, svc.addr.AddrHandle)
		delete(s.nameGroupServices[svc.addr.ServiceName], svc.addr.AddrHandle)
	}
	return nil
}

func (s *skymeshServer) Serve() error {
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
				resolver := s.resolvers[rsname]
				s.mu.Unlock()
				if resolver != nil {
					resolver.notifyInstsChange(e.isOnline, rh, e.serviceAddr.ServiceId)
				}
			case *RegServiceEvent:
				e := event.(*RegServiceEvent)
				dh := e.dstHandle
				result := e.result
				s.mu.Lock()
				svc := s.handleServices[dh]
				s.mu.Unlock()
				if svc != nil {
					svc.service.OnRegister(svc, result)
				} else {
					log.Errorf("event not exist handle %d.\n", dh)
				}
			}
		case msg := <-s.recvQueue: //这里要优先处理消息??
			dh := msg.GetDstHandle()
			s.mu.Lock()
			svc := s.handleServices[dh]
			if svc != nil {
				svc.PushMessage(msg)
			}
			s.mu.Unlock()

		case <-s.quit.Done():
			log.Info("start server quit.\n")
			s.sidecar.Release()
			log.Info("start handle remain msg.\n")
			for len(s.recvQueue) > 0 {
				msg := <-s.recvQueue
				dh := msg.GetDstHandle()
				s.mu.Lock()
				svc := s.handleServices[dh]
				if svc != nil {
					svc.PushMessage(msg)
				}
				s.mu.Unlock()
			}
			log.Info("handle remain msg done.\n")
			s.mu.Lock()
			for _, svc := range s.handleServices {
				s.sidecar.RegisterServiceToNameServer(svc.GetLocalAddr(), false)
				svc.Stop()
			}
			s.urlServices = make(map[string]*skymeshService)
			s.handleServices = make(map[uint64]*skymeshService)
			s.nameGroupServices = make(map[string]map[uint64]*skymeshService)
			s.resolvers = make(map[string]*skymeshResolver)
			s.mu.Unlock()
			log.Info("stop all services done.\n")
			s.serviceWG.Wait()
			log.Warning("skymesh server quit!.\n")
			s.done.Fire()
			return nil

		case err := <-s.errQueue:
			log.Errorf("server err:%v\n", err)
			s.mu.Lock()
			for _, svc := range s.handleServices {
				svc.Stop()
			}
			s.mu.Unlock()
			log.Warning("skymesh server quit.")
			s.done.Fire()
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

func (s *skymeshServer) SendByRouter(srcServiceUrl string, dstServiceName string, b []byte) error {
	s.mu.Lock()
	srcSvc := s.urlServices[srcServiceUrl]
	s.mu.Unlock()
	if srcSvc == nil {
		return fmt.Errorf("SendByRouter not find service %s", srcServiceUrl)
	}
	return s.sendByRouter(srcSvc.addr, dstServiceName, b)
}

func (s *skymeshServer) Send(srcServiceUrl string, dstHandle uint64, b []byte) error {
	s.mu.Lock()
	srcSvc := s.urlServices[srcServiceUrl]
	s.mu.Unlock()
	if srcSvc == nil {
		return fmt.Errorf("Send not find service %s", srcServiceUrl)
	}
	return s.send(srcSvc.addr, dstHandle, b)
}

func (s *skymeshServer) sendByRouter(srcAddr *Addr, serviceName string, b []byte) error { //针对无状态服务
	rmAddr := s.GetBestQualityService(serviceName)
	if rmAddr == nil {
		return fmt.Errorf("not find service %s router", serviceName)
	}
	return s.send(srcAddr, rmAddr.AddrHandle, b)
}

func (s *skymeshServer) send(srcAddr *Addr, dstHandle uint64, b []byte) error {
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

func (s *skymeshServer) GetNameResolver(serviceName string) NameResolver {
	s.mu.Lock()
	ns := s.resolvers[serviceName]
	if ns == nil {
		ns = &skymeshResolver{
			svcName:   serviceName,
			instAddrs: make(map[uint64]*Addr),
			watchers:  make(map[NameWatcher]bool),
		}
		s.resolvers[serviceName] = ns
	}
	s.mu.Unlock()
	handles, insts := s.getServiceInsts(serviceName)
	for idx, instID := range insts {
		ns.instAddrs[instID] = &Addr{ServiceName: serviceName, ServiceId: instID, AddrHandle: handles[idx]}
	}
	return ns
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
