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
			log.Infof("register app %s result:%d", ndr.server.appID, ssmsg.GetRegisterAppRsp().Result)
			return
		}
		if ssmsg.Cmd == smproto.SSCmd_RSP_REGISTER_SERVICE {
			log.Info("recv register service rsp.\n")
			rsp := ssmsg.GetRegisterServiceRsp()
			msg := &RegServiceEvent{
				dstHandle: rsp.AddrHandle,
				result:    rsp.Result,
			}
			select {
			case ndr.server.eventQueue <- msg:
			default:
				log.Error("deliver register event msg block.\n")
			}
			return
		}
		if ssmsg.Cmd == smproto.SSCmd_NOTIFY_SERVICE_ONLINE {
			log.Info("recv online msg.\n")
			rsp := ssmsg.GetNotifyServiceOnline()
			msg := &OnlineEvent{
				serviceAddr: &Addr{
					ServiceName: rsp.ServiceInfo.ServiceName,
					ServiceId:   rsp.ServiceInfo.ServiceId,
					AddrHandle:  rsp.ServiceInfo.AddrHandle,
				},
				serverAddr: rsp.ServerAddr,
				isOnline:   rsp.IsOnline,
			}
			deadline := 2
			ticker := time.NewTicker(time.Duration(deadline)*time.Second)
			select {
			case ndr.server.eventQueue <- msg:
				ticker.Stop()
				log.Info("deliver online event msg succeed.\n")
			case <-ticker.C:
				log.Errorf("deliver online event msg block %ds.\n", deadline)
			}
			return
		}
		if ssmsg.Cmd == smproto.SSCmd_NOTIFY_SERVICE_SYNCATTR {
			log.Info("recv sync attr msg.\n")
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
			log.Info("recv election result msg.\n")
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
