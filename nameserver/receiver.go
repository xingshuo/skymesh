package nameserver

import (
	"github.com/golang/protobuf/proto"
	skymesh "github.com/xingshuo/skymesh/agent"
	gonet "github.com/xingshuo/skymesh/common/network"
	"github.com/xingshuo/skymesh/log"
	smpack "github.com/xingshuo/skymesh/proto"
	smproto "github.com/xingshuo/skymesh/proto/generate"
	"sync"
)

type lisConnReceiver struct {
	serverAddress string //ip:port
	appid         string
	server        *Server
	sender        gonet.Sender
	mu            sync.Mutex
}

func (lr *lisConnReceiver) OnConnected(s gonet.Sender) error {
	lr.mu.Lock()
	lr.sender = s
	lr.mu.Unlock()
	return nil
}

func (lr *lisConnReceiver) OnMessage(s gonet.Sender, b []byte) (skipLen int, err error) {
	skipLen, data := smpack.Unpack(b)
	if skipLen > 0 {
		var ssmsg smproto.SSMsg
		err = proto.Unmarshal(data, &ssmsg)
		if err != nil {
			log.Errorf("pb unmarshal err:%v.\n", err)
			return
		}
		if ssmsg.Cmd == smproto.SSCmd_NOTIFY_NAMESERVER_AGENT_INFO {
			lr.OnNotifiedAgentInfo(&ssmsg)
			return
		}
		if ssmsg.Cmd == smproto.SSCmd_REQ_REGISTER_APP {
			lr.OnRegisterApp(&ssmsg)
			return
		}
		if ssmsg.Cmd == smproto.SSCmd_REQ_REGISTER_SERVICE {
			lr.OnRegisterService(&ssmsg)
			return
		}
		if ssmsg.Cmd == smproto.SSCmd_REQ_UNREGISTER_SERVICE {
			lr.OnUnRegisterService(&ssmsg)
			return
		}
		if ssmsg.Cmd == smproto.SSCmd_NOTIFY_NAMESERVER_HEARTBEAT {
			lr.OnNotifiedServiceHeartbeat(&ssmsg)
			return
		}
		if ssmsg.Cmd == smproto.SSCmd_NOTIFY_NAMESERVER_SYNCATTR {
			lr.OnServiceSyncAttr(&ssmsg)
			return
		}
		if ssmsg.Cmd == smproto.SSCmd_NOTIFY_NAMESERVER_ELECTION {
			lr.OnServiceElection(&ssmsg)
			return
		}
		if ssmsg.Cmd == smproto.SSCmd_NOTIFY_NAMESERVER_REGISTER_ROUTER {
			lr.OnRegisterRouter(&ssmsg)
			return
		}
	}
	return
}

func (lr *lisConnReceiver) OnClosed(s gonet.Sender) error {
	lr.server.sess_mgr.RemoveSession(lr.serverAddress, lr.appid)
	return nil
}

func (lr *lisConnReceiver) Send(b []byte) {
	lr.mu.Lock()
	s := lr.sender
	lr.mu.Unlock()
	if s != nil {
		s.Send(b)
	} else {
		log.Error("lr sender not init.")
	}
}

func (lr *lisConnReceiver) OnNotifiedAgentInfo(ssmsg *smproto.SSMsg) {
	log.Debug("on notified agent info\n")
	req := ssmsg.GetNotifyNameserverAgentInfo()
	lr.appid = req.AppID
	lr.serverAddress = req.ServerAddr
	lr.server.sess_mgr.AddSession(lr.serverAddress, lr.appid, lr)
}

func (lr *lisConnReceiver) OnRegisterApp(ssmsg *smproto.SSMsg) {
	log.Debug("on register app\n")
	req := ssmsg.GetRegisterAppReq()
	lr.appid = req.AppID
	lr.serverAddress = req.ServerAddr
	lr.server.sess_mgr.AddSession(lr.serverAddress, lr.appid, lr)
	msg := &RegAppMsg {
		appid:      lr.appid,
		serverAddr: lr.serverAddress,
	}
	select {
	case lr.server.msg_queue <- msg:
	default:
		log.Error("deliver register app msg block.\n")
	}
}

func (lr *lisConnReceiver) OnRegisterService(ssmsg *smproto.SSMsg) {
	log.Debug("on register service\n")
	info := ssmsg.GetRegisterServiceReq().GetServiceInfo()
	opts := ssmsg.GetRegisterServiceReq().GetOptions()
	msg := &RegServiceMsg{
		appid:      lr.appid,
		serverAddr: lr.serverAddress,
		serviceAddr: &skymesh.Addr{
			ServiceName: info.ServiceName,
			ServiceId:   info.ServiceId,
			AddrHandle:  info.AddrHandle,
		},
		serviceOpts: &skymesh.ServiceOptions{opts.ConsistentHashKey},
	}
	log.Debugf("register service %s\n", msg.serviceAddr)
	select {
	case lr.server.msg_queue <- msg:
	default:
		log.Error("deliver register service msg block.\n")
	}
}

func (lr *lisConnReceiver) OnUnRegisterService(ssmsg *smproto.SSMsg) {
	log.Debug("on un-register service\n")
	req := ssmsg.GetUnregisterServiceReq()
	msg := &UnRegServiceMsg{
		addrHandle: req.AddrHandle,
	}
	select {
	case lr.server.msg_queue <- msg:
	default:
		log.Error("deliver unregister service msg block.\n")
	}
}

func (lr *lisConnReceiver) OnNotifiedServiceHeartbeat(ssmsg *smproto.SSMsg) {
	notify := ssmsg.GetNotifyNameserverHb()
	msg := &ServiceHeartbeat{
		addrHandle:notify.SrcHandle,
	}
	select {
	case lr.server.msg_queue <- msg:
	default:
		log.Error("deliver service heartbeat msg block.\n")
	}
}

func (lr *lisConnReceiver) OnServiceSyncAttr(ssmsg *smproto.SSMsg) {
	notify := ssmsg.GetNotifyNameserverAttr()
	msg := &ServiceSyncAttr {
		addrHandle: notify.SrcHandle,
		attrs: notify.Data,
	}
	select {
	case lr.server.msg_queue <- msg:
	default:
		log.Error("deliver service sync attr msg block.\n")
	}
}

func (lr *lisConnReceiver) OnServiceElection(ssmsg *smproto.SSMsg) {
	notify := ssmsg.GetNotifyNameserverElection()
	msg := &ServiceElection {
		addrHandle: notify.SrcHandle,
		event: notify.Event,
	}
	select {
	case lr.server.msg_queue <- msg:
	default:
		log.Error("deliver service election msg block.\n")
	}
}

func (lr *lisConnReceiver) OnRegisterRouter(ssmsg *smproto.SSMsg) {
	notify := ssmsg.GetNotifyNameserverRegisterRouter()
	msg := &RegisterNameRouter {
		serverAddr: lr.serverAddress,
		appid: lr.appid,
		watchSvcName: notify.ServiceName,
	}
	select {
	case lr.server.msg_queue <- msg:
	default:
		log.Error("deliver register router msg block.\n")
	}
}