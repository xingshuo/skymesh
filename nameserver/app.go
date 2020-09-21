package nameserver

import (
	"container/list"
	skymesh "github.com/xingshuo/skymesh/agent"
	"github.com/xingshuo/skymesh/log"
	smpack "github.com/xingshuo/skymesh/proto"
	smproto "github.com/xingshuo/skymesh/proto/generate"
	"time"
)

type AppInfo struct { //对应一个游戏或应用
	server   *Server
	appID    string
	services map[uint64]*list.Element
	svclruList *list.List
	electionCandidates map[string]map[uint64]*list.Element //{svcName: {handle: svcListItem}}
	electionLeaders    map[string]*list.Element //{svcName: svcListItem}
}

func (a *AppInfo) Init(appid string) {
	a.appID = appid
	a.services = make(map[uint64]*list.Element)
	a.electionCandidates = make(map[string]map[uint64]*list.Element)
	a.electionLeaders = make(map[string]*list.Element)
	a.svclruList = list.New() //lru链表优化服务活跃检查时间复杂度
}

func (a *AppInfo) AddItem(si *ServiceInfo) {
	a.services[si.serviceAddr.AddrHandle] = a.svclruList.PushBack(si)
}

func (a *AppInfo) RemoveItem(handle uint64) {
	siItem := a.services[handle]
	if siItem != nil {
		delete(a.services, handle)
		a.svclruList.Remove(siItem)
	}
}

func (a *AppInfo) Broadcast(excludeSI *ServiceInfo, b []byte) {
	notifys := make(map[string]bool)
	for _, siItem := range a.services {
		service := siItem.Value.(*ServiceInfo)
		if excludeSI != nil && service.serverAddr == excludeSI.serverAddr { //只通知其他进程App
			continue
		}
		notifys[service.serverAddr] = true
	}
	for saddr := range notifys {
		lr := a.server.GetSession(saddr, a.appID)
		if lr != nil {
			lr.Send(b)
		}
	}
}

func (a *AppInfo) BroadcastSyncServiceAttr(si *ServiceInfo, attrs []byte) {
	msg := &smproto.SSMsg{
		Cmd: smproto.SSCmd_NOTIFY_SERVICE_SYNCATTR,
		Msg: &smproto.SSMsg_NotifyServiceAttr{
			NotifyServiceAttr: &smproto.NotifyServiceSyncAttr{
				ServiceInfo: &smproto.ServiceInfo{
					ServiceName: si.serviceAddr.ServiceName,
					ServiceId:   si.serviceAddr.ServiceId,
					AddrHandle:  si.serviceAddr.AddrHandle,
				},
				Data: attrs,
			},
		},
	}
	b, err := smpack.PackSSMsg(msg)
	if err != nil {
		log.Errorf("pb marshal err:%v.\n", err)
		return
	}
	a.Broadcast(si, b)
}

func (a *AppInfo) BroadcastOnlineToOthers(si *ServiceInfo, is_online bool) {
	msg := &smproto.SSMsg{
		Cmd: smproto.SSCmd_NOTIFY_SERVICE_ONLINE,
		Msg: &smproto.SSMsg_NotifyServiceOnline{
			NotifyServiceOnline: &smproto.NotifyServiceOnline{
				ServerAddr: si.serverAddr,
				ServiceInfo: &smproto.ServiceInfo{
					ServiceName: si.serviceAddr.ServiceName,
					ServiceId:   si.serviceAddr.ServiceId,
					AddrHandle:  si.serviceAddr.AddrHandle,
				},
				IsOnline: is_online,
			},
		},
	}
	b, err := smpack.PackSSMsg(msg)
	if err != nil {
		log.Errorf("pb marshal err:%v.\n", err)
		return
	}
	a.Broadcast(si, b)
}

//这里可以一条协议通知某cluster上所有服务给自己,后面优化
func (a *AppInfo) NotifyOthersOnlineToSelf(serverAddr string, lr *lisConnReceiver) {
	for _, siItem := range a.services {
		service := siItem.Value.(*ServiceInfo)
		if service.serverAddr == serverAddr { //只通知自己其他进程service上线消息
			continue
		}
		msg := &smproto.SSMsg{
			Cmd: smproto.SSCmd_NOTIFY_SERVICE_ONLINE,
			Msg: &smproto.SSMsg_NotifyServiceOnline{
				NotifyServiceOnline: &smproto.NotifyServiceOnline{
					ServerAddr: service.serverAddr,
					ServiceInfo: &smproto.ServiceInfo{
						ServiceName: service.serviceAddr.ServiceName,
						ServiceId:   service.serviceAddr.ServiceId,
						AddrHandle:  service.serviceAddr.AddrHandle,
					},
					IsOnline: true,
				},
			},
		}
		b, err := smpack.PackSSMsg(msg)
		if err != nil {
			log.Errorf("pack %s online pb marshal err:%v.\n", service.serviceAddr, err)
			continue
		}
		lr.Send(b)
	}
}

func (a *AppInfo) CheckServiceAlive() {
	now := time.Now().Unix()
	var expired [CheckAliveMaxNumPerTickPerApp]uint64
	var num int
	for siItem := a.svclruList.Front(); siItem != nil && num < CheckAliveMaxNumPerTickPerApp; siItem = siItem.Next() {
		si := siItem.Value.(*ServiceInfo)
		if si.lastRecvHbTime + a.server.cfg.ServiceKeepaliveTime <= now {
			expired[num] = si.serviceAddr.AddrHandle
			num++
		} else {
			break
		}
	}

	for i := 0; i < num; i++ {
		h := expired[i]
		a.server.UnRegisterService(h, true)
	}
}

func (a *AppInfo) OnServiceHeartbeat(handle uint64) {
	siItem := a.services[handle]
	if siItem != nil {
		si := siItem.Value.(*ServiceInfo)
		si.lastRecvHbTime = time.Now().Unix()
		a.svclruList.MoveToBack(siItem)
	}
}

func (a *AppInfo) GetLeaders() []*ServiceInfo {
	var leaders []*ServiceInfo
	for _,siItem := range a.electionLeaders {
		leaders = append(leaders, siItem.Value.(*ServiceInfo))
	}
	return leaders
}

func (a *AppInfo) OnServiceRunForElection(handle uint64) {
	siItem := a.services[handle]
	si := siItem.Value.(*ServiceInfo)
	svcName := si.serviceAddr.ServiceName
	v := a.electionCandidates[svcName]
	if v == nil {
		v = make(map[uint64]*list.Element)
		a.electionCandidates[svcName] = v
	}
	var result int32
	if v[handle] != nil { //已经是候选人了
		result = skymesh.KElectionResultRunForFail
		if a.electionLeaders[svcName] != nil {
			leader := a.electionLeaders[svcName].Value.(*ServiceInfo)
			if leader.serviceAddr.AddrHandle == handle { //已经是Leader了
				result = skymesh.KElectionResultAlreadyLeader
			}
		}
	} else {
		v[handle] = siItem
		if a.electionLeaders[svcName] != nil {
			result = skymesh.KElectionResultRunForFail
		} else {
			result = skymesh.KElectionResultOK
			a.electionLeaders[svcName] = siItem
		}
	}

	msg := &smproto.SSMsg{
		Cmd: smproto.SSCmd_NOTIFY_SERVICE_ELECTION_RESULT,
		Msg: &smproto.SSMsg_NotifyServiceElectionResult{
			NotifyServiceElectionResult: &smproto.NotifyServiceElectionResult{
				Candidate:            &smproto.ServiceInfo{
					ServiceName: 		si.serviceAddr.ServiceName,
					ServiceId:   		si.serviceAddr.ServiceId,
					AddrHandle:  		si.serviceAddr.AddrHandle,
				},
				Event:                  skymesh.KElectionRunForLeader,
				Result:                 result,
			},
		},
	}
	b, err := smpack.PackSSMsg(msg)
	if err != nil {
		log.Errorf("pb marshal err:%v.\n", err)
		return
	}
	if result == skymesh.KElectionResultOK { //成功,通知所有节点更新
		a.Broadcast(nil, b)
	} else { //失败,只通知发起者
		lr := a.server.GetSession(si.serverAddr, a.appID)
		if lr != nil {
			lr.Send(b)
		}
	}
}

func (a *AppInfo) OnServiceGiveupElection(handle uint64) {
	siItem := a.services[handle]
	si := siItem.Value.(*ServiceInfo)
	svcName := si.serviceAddr.ServiceName
	if a.electionCandidates[svcName] == nil || a.electionCandidates[svcName][handle] == nil {
		return
	}
	delete(a.electionCandidates[svcName], handle)
	leaderChanged := false
	leaderItem := a.electionLeaders[svcName]
	if leaderItem == nil { //有候选人,没leader??
		log.Errorf("no leader finded when %s give up election\n", si.serviceAddr)
	} else {
		leader := leaderItem.Value.(*ServiceInfo)
		if leader.serviceAddr.AddrHandle == handle {
			leaderChanged = true
		}
	}
	//未发生leader变更, 仅通知发起者退出候选人队列
	if !leaderChanged {
		msg := &smproto.SSMsg {
			Cmd: smproto.SSCmd_NOTIFY_SERVICE_ELECTION_RESULT,
			Msg: &smproto.SSMsg_NotifyServiceElectionResult {
				NotifyServiceElectionResult: &smproto.NotifyServiceElectionResult{
					Candidate:            &smproto.ServiceInfo{
						ServiceName: 		si.serviceAddr.ServiceName,
						ServiceId:   		si.serviceAddr.ServiceId,
						AddrHandle:  		si.serviceAddr.AddrHandle,
					},
					Event:                  skymesh.KElectionGiveUpLeader,
					Result:                 skymesh.KElectionResultQuitCandidate,
				},
			},
		}
		b, err := smpack.PackSSMsg(msg)
		if err != nil {
			log.Errorf("pb marshal err:%v.\n", err)
			return
		}
		lr := a.server.GetSession(si.serverAddr, a.appID)
		if lr != nil {
			lr.Send(b)
		}
		return
	}

	delete(a.electionLeaders, svcName)
	//重新选举
	var newLeader *ServiceInfo = nil
	for _,item := range a.electionCandidates[svcName] {
		svcInfo := item.Value.(*ServiceInfo)
		if newLeader == nil {
			newLeader = svcInfo
		} else if svcInfo.lastRecvHbTime > newLeader.lastRecvHbTime {
			newLeader = svcInfo
		}
	}

	var msg *smproto.SSMsg
	if newLeader != nil { //重新选举成功
		a.electionLeaders[svcName] = a.services[newLeader.serviceAddr.AddrHandle]
		//只广播新Leader竞选成功事件
		msg = &smproto.SSMsg{
			Cmd: smproto.SSCmd_NOTIFY_SERVICE_ELECTION_RESULT,
			Msg: &smproto.SSMsg_NotifyServiceElectionResult{
				NotifyServiceElectionResult: &smproto.NotifyServiceElectionResult{
					Candidate:            &smproto.ServiceInfo{
						ServiceName: 		newLeader.serviceAddr.ServiceName,
						ServiceId:   		newLeader.serviceAddr.ServiceId,
						AddrHandle:  		newLeader.serviceAddr.AddrHandle,
					},
					Event:                  skymesh.KElectionRunForLeader,
					Result:                 skymesh.KElectionResultOK,
				},
			},
		}
	} else { //重新选举失败
		//广播旧Leader弃选成功事件
		msg = &smproto.SSMsg{
			Cmd: smproto.SSCmd_NOTIFY_SERVICE_ELECTION_RESULT,
			Msg: &smproto.SSMsg_NotifyServiceElectionResult{
				NotifyServiceElectionResult: &smproto.NotifyServiceElectionResult{
					Candidate:            &smproto.ServiceInfo{
						ServiceName: 		si.serviceAddr.ServiceName,
						ServiceId:   		si.serviceAddr.ServiceId,
						AddrHandle:  		si.serviceAddr.AddrHandle,
					},
					Event:                  skymesh.KElectionGiveUpLeader,
					Result:                 skymesh.KElectionResultOK,
				},
			},
		}
	}

	b, err := smpack.PackSSMsg(msg)
	if err != nil {
		log.Errorf("pb marshal err:%v.\n", err)
		return
	}
	a.Broadcast(nil, b)
}

func (a *AppInfo) Release() {
	a.services = make(map[uint64]*list.Element)
	a.electionCandidates = make(map[string]map[uint64]*list.Element)
	a.electionLeaders = make(map[string]*list.Element)
	a.svclruList = list.New()
	a.server = nil
}
