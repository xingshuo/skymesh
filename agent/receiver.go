package skymesh

import (
	"github.com/golang/protobuf/proto"
	gonet "github.com/xingshuo/skymesh/common/network"
	"github.com/xingshuo/skymesh/log"
	smpack "github.com/xingshuo/skymesh/proto"
	smproto "github.com/xingshuo/skymesh/proto/generate"
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
			msg := &Message{
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
			select {
			case ndr.server.eventQueue <- msg:
			default:
				log.Error("deliver online event msg block.\n")
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
	log.Error("agent dialer recv unexpect stream %v.", b)
	return 0, nil
}

func (adr *AgentDialerReceiver) OnClosed(s gonet.Sender) error {
	return nil
}
