package nameserver

import (
	"fmt"
	skymesh "github.com/xingshuo/skymesh/agent"
	"github.com/xingshuo/skymesh/log"
	smpack "github.com/xingshuo/skymesh/proto"
	smproto "github.com/xingshuo/skymesh/proto/generate"
	"time"
)

type NameRouterState int

const (
	Preparing NameRouterState = iota
	Ready
)

type NameRouter struct {
	svcName    string
	routers    map[AppSession]NameRouterState
	confirmDeadline int64  //Second
	onlineSvcHandle uint64 //预上线的服务实例handle
}

func (nr *NameRouter) AddItem(a *AppInfo ,serverAddr string) error {
	as := AppSession{serverAddr, a.appID}
	_,ok := nr.routers[as]
	if ok {
		return fmt.Errorf("add name router repeat")
	}
	if nr.onlineSvcHandle != skymesh.INVALID_SERVICE_HANDLE { //Todo: 通知NameRouter Preparing
		nr.routers[as] = Preparing
		lr := a.server.GetSession(serverAddr, a.appID)
		if lr == nil {
			return fmt.Errorf("%s not find session", as)
		}
		msg := &smproto.SSMsg {
			Cmd: smproto.SSCmd_REQ_AGENT_ROUTER_UPDATE,
			Msg: &smproto.SSMsg_AgentRouterUpdateReq {
				AgentRouterUpdateReq: &smproto.ReqAgentRouterUpdate {
					Cmd: int32(skymesh.KNotifyAgentPrepare),
					ServiceName: nr.svcName,
				},
			},
		}
		b, err := smpack.PackSSMsg(msg)
		if err != nil {
			log.Errorf("pb marshal err:%v.\n", err)
			return fmt.Errorf("pb marshal err")
		}
		lr.Send(b)
	} else {
		nr.routers[as] = Ready
	}
	return nil
}

func (nr *NameRouter) StartConfirm(a *AppInfo, si *ServiceInfo) error {
	if nr.onlineSvcHandle != skymesh.INVALID_SERVICE_HANDLE { //上一个实例还没完成上线流程
		err := fmt.Errorf("service register lineup")
		log.Errorf("Start2StageConfirm: %v\n", err)
		a.NotifyConfirmResult(si.serviceAddr.AddrHandle, smproto.SSError_ERR_SERVICE_REGISTER_LINEUP)
		return err
	}
	readyNum := 0
	for app,st := range nr.routers {
		if st != Ready { //理论上不应该!!
			log.Errorf("Start2StageConfirm [%s, %s] not in ready\n", app.appid, app.serverAddress)
			nr.routers[app] = Ready
		}
		readyNum++
	}

	if readyNum == 0 {
		a.NotifyConfirmResult(si.serviceAddr.AddrHandle, smproto.SSError_OK)
	} else {
		for app := range nr.routers {
			nr.routers[app] = Preparing
		}
		nr.confirmDeadline = time.Now().Unix() + ConfirmExpiredSecond
		nr.onlineSvcHandle = si.serviceAddr.AddrHandle
		nr.BroadcastCmd(a, skymesh.KNotifyAgentPrepare)
	}
	return nil
}

func (nr *NameRouter) CheckConfirm(a *AppInfo, serverAddr string, cmd skymesh.NameRouterCmd) {
	as := AppSession{serverAddr, a.appID}
	st,ok := nr.routers[as]
	if !ok {
		return
	}
	if st != Preparing {
		return
	}
	if cmd == skymesh.KReplyNSPrepareOK {
		nr.routers[as] = Ready
		allPrepared := true
		for _,st := range nr.routers {
			if st != Ready {
				allPrepared = false
				break
			}
		}
		if allPrepared {
			nr.CommitConfirm(a, skymesh.KNotifyAgentCommit, smproto.SSError_OK)
		}
	} else if cmd == skymesh.KReplyNSPrepareError {
		nr.routers[as] = Ready
		log.Infof("name router %s session %s prepare error\n", nr.svcName, as)
		nr.CommitConfirm(a, skymesh.KNotifyAgentPrepareAbort, smproto.SSError_ERR_ROUTER_PREPARE_ERR)
	}
}

func (nr *NameRouter) CheckExpired(a *AppInfo) {
	if nr.onlineSvcHandle == skymesh.INVALID_SERVICE_HANDLE {
		return
	}
	if nr.confirmDeadline > time.Now().Unix() {
		return
	}
	log.Infof("name router %s handle:%d online expired\n", nr.svcName, nr.onlineSvcHandle)
	for app,st := range nr.routers {
		if st != Ready {
			log.Infof("name router %s session %s prepare timeout\n", nr.svcName, app)
		}
	}
	nr.CommitConfirm(a, skymesh.KNotifyAgentPrepareAbort, smproto.SSError_ERR_ROUTER_PREPARE_TIMEOUT)
}

func (nr *NameRouter) CommitConfirm(a *AppInfo, cmd skymesh.NameRouterCmd, result smproto.SSError) {
	if nr.onlineSvcHandle == skymesh.INVALID_SERVICE_HANDLE {
		return
	}
	log.Infof("name router %s handle:%d CommitConfirm cmd:%v result:%v\n", nr.svcName, nr.onlineSvcHandle, cmd, result)
	//通知agent routers 第二段提交结果
	nr.BroadcastCmd(a, cmd)
	//通知上线实例结果 && 上线成功广播上线消息
	a.NotifyConfirmResult(nr.onlineSvcHandle, result)
	nr.confirmDeadline = 0
	nr.onlineSvcHandle = skymesh.INVALID_SERVICE_HANDLE
	for app,st := range nr.routers {
		if st != Ready {
			nr.routers[app] = Ready
		}
	}
}

func (nr *NameRouter) BroadcastCmd(a *AppInfo, cmd skymesh.NameRouterCmd) {
	if cmd == skymesh.KNotifyAgentCommit { //Commit由SSCmd_NOTIFY_SERVICE_ONLINE通知
		return
	}
	for app := range nr.routers {
		lr := a.server.GetSession(app.serverAddress, app.appid)
		if lr == nil {
			continue
		}
		msg := &smproto.SSMsg {
			Cmd: smproto.SSCmd_REQ_AGENT_ROUTER_UPDATE,
			Msg: &smproto.SSMsg_AgentRouterUpdateReq {
				AgentRouterUpdateReq: &smproto.ReqAgentRouterUpdate {
					Cmd: int32(cmd),
					ServiceName: nr.svcName,
				},
			},
		}
		b, err := smpack.PackSSMsg(msg)
		if err != nil {
			log.Errorf("pb marshal err:%v.\n", err)
			continue
		}
		lr.Send(b)
	}
}