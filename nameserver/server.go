package nameserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	skymesh "github.com/xingshuo/skymesh/agent"
	gonet "github.com/xingshuo/skymesh/common/network"
	smsync "github.com/xingshuo/skymesh/common/sync"
	"github.com/xingshuo/skymesh/log"
	smpack "github.com/xingshuo/skymesh/proto"
	smproto "github.com/xingshuo/skymesh/proto/generate"
)

type AppSession struct {
	serverAddress string //ip:port
	appid         string
}

type SessionMgr struct {
	mu       sync.Mutex
	sessions map[AppSession]*lisConnReceiver
}

func (sm *SessionMgr) Init() {
	sm.sessions = make(map[AppSession]*lisConnReceiver)
}

func (sm *SessionMgr) AddSession(serverAddress, appid string, lr *lisConnReceiver) {
	sm.mu.Lock()
	sm.sessions[AppSession{serverAddress, appid}] = lr
	sm.mu.Unlock()
}

func (sm *SessionMgr) GetSession(serverAddress, appid string) *lisConnReceiver {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.sessions[AppSession{serverAddress, appid}]
}

func (sm *SessionMgr) RemoveSession(serverAddress, appid string) {
	sm.mu.Lock()
	delete(sm.sessions, AppSession{serverAddress: serverAddress, appid: appid})
	sm.mu.Unlock()
}

func (sm *SessionMgr) Release() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sessions = make(map[AppSession]*lisConnReceiver)
}

type ServiceInfo struct {
	serverAddr  string
	appID       string
	serviceAddr *skymesh.Addr
	sessMgr     *SessionMgr
	lastRecvHbTime int64 //秒级
}

func (si *ServiceInfo) NotifyApp(b []byte) {
	lr := si.sessMgr.GetSession(si.serverAddr, si.appID)
	if lr == nil {
		log.Errorf("lost session(%s,%s) when broadcast.\n", si.serverAddr, si.appID)
		return
	}
	lr.Send(b)
}

type Server struct {
	cfg            Config
	sess_mgr       *SessionMgr
	msg_queue      chan interface{}
	err_queue      chan error
	handleServices map[uint64]*ServiceInfo
	apps           map[string]*AppInfo
	quit           *smsync.Event
	ticker         *time.Ticker
}

func (s *Server) Init(conf string) error {
	err := s.loadConfig(conf)
	if err != nil {
		return err
	}
	s.sess_mgr = &SessionMgr{}
	s.sess_mgr.Init()
	s.msg_queue = make(chan interface{}, s.cfg.MsgQueueSize)
	s.err_queue = make(chan error, 1)
	s.handleServices = make(map[uint64]*ServiceInfo)
	s.apps = make(map[string]*AppInfo)
	s.quit = smsync.NewEvent("nameserver.MeshServer.quit")

	newReceiver := func() gonet.Receiver {
		return &lisConnReceiver{server: s}
	}
	l, err := gonet.NewListener(s.cfg.ServerAddress, newReceiver)
	if err != nil {
		log.Errorf("new listener failed:%v", err)
		return err
	}
	go func() {
		err := l.Serve()
		if err != nil {
			log.Errorf("listener serve err:%v", err)
		}
		s.err_queue <- err
		log.Info("listener quit.")
	}()
	s.ticker = time.NewTicker(time.Duration(s.cfg.ServerTickInterval) * time.Second)
	go func() {
		for range s.ticker.C {
			s.msg_queue <- &TickMsg{}
		}
	}()
	return nil
}

func (s *Server) Serve() error {
	for {
		select {
		case msg := <-s.msg_queue:
			err := s.onMessage(msg)
			if err != nil {
				log.Errorf("on server message err:%v.", err)
			}
		case err := <-s.err_queue:
			log.Errorf("server error:%v.\n", err)
			return err
		case <-s.quit.Done():
			for len(s.msg_queue) > 0 {
				msg := <-s.msg_queue
				s.onMessage(msg)
			}
			return nil
		}
	}
}

func (s *Server) loadConfig(conf string) error {
	data, err := ioutil.ReadFile(conf)
	if err != nil {
		log.Errorf("load config %s failed:%v", conf, err)
		return err
	}
	err = json.Unmarshal(data, &s.cfg)
	if err != nil {
		log.Errorf("load config %s failed:%v.", conf, err)
		return err
	}
	return nil
}

func (s *Server) onMessage(msg interface{}) error {
	switch msg.(type) {
	case *RegAppMsg:
		appID := msg.(*RegAppMsg).appid
		serverAddr := msg.(*RegAppMsg).serverAddr
		return s.RegisterApp(serverAddr, appID)
	case *RegServiceMsg:
		appID := msg.(*RegServiceMsg).appid
		svrAddr := msg.(*RegServiceMsg).serverAddr
		svcAddr := msg.(*RegServiceMsg).serviceAddr
		return s.RegisterService(appID, svrAddr, svcAddr)
	case *UnRegServiceMsg:
		h := msg.(*UnRegServiceMsg).addrHandle
		return s.UnRegisterService(h)
	case *ServiceHeartbeat:
		h := msg.(*ServiceHeartbeat).addrHandle
		return s.OnServiceHeartbeat(h)
	case *TickMsg:
		s.OnTick()
		return nil
	case *ServiceSyncAttr:
		h := msg.(*ServiceSyncAttr).addrHandle
		attrs := msg.(*ServiceSyncAttr).attrs
		return s.OnServiceSyncAttr(h, attrs)
	default:
		return fmt.Errorf("unknow msg type")
	}
}

func (s *Server) GetSession(serverAddr, appID string) *lisConnReceiver {
	lr := s.sess_mgr.GetSession(serverAddr, appID)
	if lr == nil {
		log.Errorf("get (%s, %s)session failed.\n", serverAddr, appID)
	}
	return lr
}

func (s *Server) RegisterApp(serverAddr, appID string) error {
	lr := s.sess_mgr.GetSession(serverAddr, appID)
	if lr == nil {
		return fmt.Errorf("session (%s,%s) not find", serverAddr, appID)
	}
	msg := &smproto.SSMsg{
		Cmd: smproto.SSCmd_RSP_REGISTER_APP,
		Msg: &smproto.SSMsg_RegisterAppRsp{
			RegisterAppRsp: &smproto.RspRegisterApp{
				Result: int32(smproto.SSError_OK),
			},
		},
	}
	b, err := smpack.PackSSMsg(msg)
	if err != nil {
		log.Errorf("pb marshal err:%v.\n", err)
		return err
	}
	lr.Send(b)

	app := s.apps[appID]
	if app != nil {
		log.Infof("re-register app %s.\n", appID)
	} else {
		app = &AppInfo{server:s}
		app.Init(appID)
		s.apps[appID] = app
	}
	app.NotifyOthersOnlineToSelf(serverAddr, lr)
	return nil
}

func (s *Server) RegisterService(appID string, serverAddr string, serviceAddr *skymesh.Addr) error {
	si := s.handleServices[serviceAddr.AddrHandle]
	if si != nil {
		return fmt.Errorf("re-register service %s", serviceAddr)
	}
	app := s.apps[appID]
	if app == nil {
		return fmt.Errorf("register not exist appID %s.", appID)
	}
	si = &ServiceInfo{
		serverAddr:  serverAddr,
		appID:       appID,
		serviceAddr: serviceAddr,
		sessMgr:     s.sess_mgr,
		lastRecvHbTime: time.Now().Unix(),
	}
	s.handleServices[serviceAddr.AddrHandle] = si
	app.AddItem(si)
	//通知Sidecar
	msg := &smproto.SSMsg{
		Cmd: smproto.SSCmd_RSP_REGISTER_SERVICE,
		Msg: &smproto.SSMsg_RegisterServiceRsp{
			RegisterServiceRsp: &smproto.RspRegisterService{
				AddrHandle: serviceAddr.AddrHandle,
				Result:     int32(smproto.SSError_OK),
			},
		},
	}
	b, err := smpack.PackSSMsg(msg)
	if err != nil {
		log.Errorf("pb marshal err:%v.\n", err)
	} else {
		log.Infof("notify service %s register rsp\n", serviceAddr)
		si.NotifyApp(b)
	}
	app.BroadcastOnlineToOthers(si, true)
	return nil
}

func (s *Server) UnRegisterService(addrHandle uint64) error {
	log.Infof("un register service %d.\n", addrHandle)
	si := s.handleServices[addrHandle]
	if si == nil {
		return fmt.Errorf("unregister not exist handle: %d", addrHandle)
	}
	app := s.apps[si.appID]
	if app == nil {
		return fmt.Errorf("unregister not exist appID %s.", si.appID)
	}
	delete(s.handleServices, addrHandle)
	app.RemoveItem(si.serviceAddr.AddrHandle)
	app.BroadcastOnlineToOthers(si, false)
	return nil
}

func (s *Server) OnServiceHeartbeat(addrHandle uint64) error {
	si := s.handleServices[addrHandle]
	if si == nil {
		return fmt.Errorf("service heartbeat not exist handle: %d", addrHandle)
	}
	app := s.apps[si.appID]
	if app == nil {
		return fmt.Errorf("service heartbeat not exist appID %s.", si.appID)
	}
	app.OnServiceHeartbeat(addrHandle)
	return nil
}

func (s *Server) OnServiceSyncAttr(addrHandle uint64, attrs []byte) error {
	si := s.handleServices[addrHandle]
	if si == nil {
		return fmt.Errorf("service sync attr not exist handle: %d", addrHandle)
	}
	app := s.apps[si.appID]
	if app == nil {
		return fmt.Errorf("service sync attr not exist appID %s.", si.appID)
	}
	app.BroadcastSyncServiceAttr(si, attrs)
	return nil
}

func (s *Server) OnTick() {
	for _,app := range s.apps {
		app.CheckServiceAlive()
	}
}

func (s *Server) Stop() {
	s.ticker.Stop()
	s.quit.Fire()
	s.sess_mgr.Release()
	s.sess_mgr = nil
	for _,app := range s.apps {
		app.Release()
	}
	s.apps = make(map[string]*AppInfo)
	s.handleServices = make(map[uint64]*ServiceInfo)
}
