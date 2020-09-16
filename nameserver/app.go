package nameserver

import (
	"container/list"
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
}

func (a *AppInfo) Init(appid string) {
	a.appID = appid
	a.services = make(map[uint64]*list.Element)
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

func (a *AppInfo) BroadcastToOthers(si *ServiceInfo, b []byte) {
	notifys := make(map[string]bool)
	for _, siItem := range a.services {
		service := siItem.Value.(*ServiceInfo)
		if si != nil && service.serverAddr == si.serverAddr { //只通知其他进程App
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
	a.BroadcastToOthers(si, b)
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
	a.BroadcastToOthers(si, b)
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
		a.server.UnRegisterService(h)
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

func (a *AppInfo) Release() {
	a.services = make(map[uint64]*list.Element)
	a.svclruList = list.New()
	a.server = nil
}
