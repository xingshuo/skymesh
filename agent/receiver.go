package skymesh

import (
	"encoding/json"
	"github.com/golang/protobuf/proto"
	gonet "github.com/xingshuo/skymesh/common/network"
	"github.com/xingshuo/skymesh/log"
	smpack "github.com/xingshuo/skymesh/proto"
	smproto "github.com/xingshuo/skymesh/proto/generate"
	"time"
)

type lisConnReceiver struct {
	server *skymeshServer
}

func (lr *lisConnReceiver) OnConnected(s gonet.Sender) error {
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
		if ssmsg.Cmd == smproto.SSCmd_NOTIFY_SERVICE_MESSAGE {
			notify := ssmsg.GetNotifyServiceMessage()
			msg := &DataMessage{
				dstHandle: notify.DstHandle,
				srcAddr: &Addr{
					ServiceName: notify.SrcService.ServiceName,
					ServiceId:   notify.SrcService.ServiceId,
					AddrHandle:  notify.SrcService.AddrHandle,
				},
				data: notify.Data,
			}
			lr.server.recvQueue <- msg
			return
		}
		if ssmsg.Cmd == smproto.SSCmd_REQ_PING_SERVICE {
			req := ssmsg.GetServicePingReq()
			msg := &PingMessage{
				srcServerAddr: req.SrcServerAddr,
				seq:           req.Seq,
				dstHandle:     req.DstHandle,
			}
			lr.server.recvQueue <- msg
			return
		}
		if ssmsg.Cmd == smproto.SSCmd_RSP_PING_SERVICE {
			rsp := ssmsg.GetServicePingRsp()
			lr.server.sidecar.recvPingAck(rsp.SrcHandle, rsp.Seq)
			return
		}
	}
	return
}

func (lr *lisConnReceiver) OnClosed(s gonet.Sender) error {
	return nil
}

type NSDialerReceiver struct {
	server *skymeshServer
}

func (ndr *NSDialerReceiver) OnConnected(s gonet.Sender) error {
	msg := &smproto.SSMsg {
		Cmd: smproto.SSCmd_NOTIFY_NAMESERVER_AGENT_INFO,
		Msg: &smproto.SSMsg_NotifyNameserverAgentInfo {
			NotifyNameserverAgentInfo: &smproto.NotifyNameServerAgentInfo{
				ServerAddr: ndr.server.cfg.MeshserverAddress,
				AppID:      ndr.server.appID,
			},
		},
	}
	b, err := smpack.PackSSMsg(msg)
	if err != nil {
		log.Errorf("pb marshal err:%v.\n", err)
		return err
	}
	s.Send(b)
	return nil
}

func (ndr *NSDialerReceiver) OnMessage(s gonet.Sender, b []byte) (skipLen int, err error) {
	skipLen, data := smpack.Unpack(b)
	if skipLen > 0 {
		var ssmsg smproto.SSMsg
		err = proto.Unmarshal(data, &ssmsg)
		if err != nil {
			log.Errorf("pb unmarshal err:%v.\n", err)
			return
		}
		if ssmsg.Cmd == smproto.SSCmd_RSP_REGISTER_APP {
			rsp := ssmsg.GetRegisterAppRsp()
			log.Debugf("register app %s result:%d", ndr.server.appID, rsp.Result)
			leaders := make([]*Addr, len(rsp.Leaders))
			for i,pbAddr := range rsp.Leaders {
				leaders[i] = &Addr {
					ServiceName: pbAddr.ServiceName,
					ServiceId: pbAddr.ServiceId,
					AddrHandle: pbAddr.AddrHandle,
				}
			}
			msg := &RegAppEvent{
				leaders: leaders,
				result: rsp.Result,
			}
			select {
			case ndr.server.eventQueue <- msg:
			default:
				log.Error("deliver register app event msg block.\n")
			}
			return
		}
		if ssmsg.Cmd == smproto.SSCmd_RSP_REGISTER_SERVICE {
			log.Debug("recv register service rsp.\n")
			rsp := ssmsg.GetRegisterServiceRsp()
			msg := &RegServiceEvent{
				dstHandle: rsp.AddrHandle,
				result:    rsp.Result,
			}
			select {
			case ndr.server.eventQueue <- msg:
			default:
				log.Error("deliver register service event msg block.\n")
			}
			return
		}
		if ssmsg.Cmd == smproto.SSCmd_NOTIFY_SERVICE_ONLINE {
			log.Debug("recv online msg.\n")
			rsp := ssmsg.GetNotifyServiceOnline()
			msg := &OnlineEvent{
				serviceAddr: &Addr {
					ServiceName: rsp.ServiceInfo.ServiceName,
					ServiceId:   rsp.ServiceInfo.ServiceId,
					AddrHandle:  rsp.ServiceInfo.AddrHandle,
				},
				serverAddr:  rsp.ServerAddr,
				isOnline:    rsp.IsOnline,
				serviceOpts: ServiceOptions {
					ConsistentHashKey: rsp.Options.ConsistentHashKey,
				},
			}
			deadline := 2
			select {
			case ndr.server.eventQueue <- msg:
				log.Debug("deliver online event msg succeed.\n")
			case <-time.After(time.Duration(deadline) * time.Second):
				log.Errorf("deliver online event msg block %ds.\n", deadline)
			}
			return
		}
		if ssmsg.Cmd == smproto.SSCmd_NOTIFY_SERVICE_SYNCATTR {
			log.Debug("recv sync attr msg.\n")
			rsp := ssmsg.GetNotifyServiceAttr()
			var attrs ServiceAttr
			err = json.Unmarshal(rsp.Data, &attrs)
			if err != nil {
				log.Errorf("Unmarshal sync attr err:%v.\n", err)
				return
			}
			msg := &SyncAttrEvent{
				serviceAddr: &Addr{
					ServiceName: rsp.ServiceInfo.ServiceName,
					ServiceId:   rsp.ServiceInfo.ServiceId,
					AddrHandle:  rsp.ServiceInfo.AddrHandle,
				},
				attributes:  attrs,
			}
			select {
			case ndr.server.eventQueue <- msg:
			default:
				log.Error("deliver sync attr event msg block.\n")
			}
			return
		}
		if ssmsg.Cmd == smproto.SSCmd_NOTIFY_SERVICE_ELECTION_RESULT {
			log.Debug("recv election result msg.\n")
			rsp := ssmsg.GetNotifyServiceElectionResult()
			msg := &ElectionEvent{
				candidate: &Addr{
					ServiceName: rsp.Candidate.ServiceName,
					ServiceId:   rsp.Candidate.ServiceId,
					AddrHandle:  rsp.Candidate.AddrHandle,
				},
				event:  rsp.Event,
				result: rsp.Result,
			}
			select {
			case ndr.server.eventQueue <- msg:
			default:
				log.Error("deliver election event msg block.\n")
			}
			return
		}
		if ssmsg.Cmd == smproto.SSCmd_NOTIFY_SERVICE_KICK_OFF {
			log.Debug("recv kick off msg.\n")
			rsp := ssmsg.GetNotifyServiceKickoff()
			msg := &KickOffEvent {
				dstHandle: rsp.AddrHandle,
			}
			select {
			case ndr.server.eventQueue <- msg:
			default:
				log.Error("deliver election event msg block.\n")
			}
			return
		}
		if ssmsg.Cmd == smproto.SSCmd_REQ_AGENT_ROUTER_UPDATE {
			log.Debug("recv router update msg.\n")
			rsp := ssmsg.GetAgentRouterUpdateReq()
			msg := &RouterUpdateEvent {
				serviceName: rsp.ServiceName,
				cmd: NameRouterCmd(rsp.Cmd),
			}
			select {
			case ndr.server.eventQueue <- msg:
			default:
				log.Error("deliver router update event msg block.\n")
			}
			return
		}
	}
	return
}

func (ndr *NSDialerReceiver) OnClosed(s gonet.Sender) error {
	return nil
}

type AgentDialerReceiver struct {
}

func (adr *AgentDialerReceiver) OnConnected(s gonet.Sender) error {
	return nil
}

func (adr *AgentDialerReceiver) OnMessage(s gonet.Sender, b []byte) (n int, err error) {
	log.Error("agent dialer recv unexpect stream %v.\n", b)
	return 0, nil
}

func (adr *AgentDialerReceiver) OnClosed(s gonet.Sender) error {
	return nil
}
