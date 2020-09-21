package skymesh

//用户注册自定义服务需要实现的接口
type AppService interface {
	OnRegister(trans MeshService, result int32)
	OnUnRegister()
	OnMessage(rmtAddr *Addr, msg []byte)
}

//用户注册名字路由状态变化监控器
type AppRouterWatcher interface {
	OnInstOnline(addr *Addr)
	OnInstOffline(addr *Addr)
	OnInstSyncAttr(addr *Addr, attrs ServiceAttr)
}

//用户注册选举事件监控器
type AppElectionCandidate interface {
	OnRegisterLeader(trans MeshService, result int32)     //RunForElection Callback (Only Candidates && Leader)
	OnUnRegisterLeader()                                  //GiveUpElection Callback When GiveUp Leader Succeed (Only Leader)
}

//用户注册选举事件监控器
type AppElectionWatcher interface {
	OnLeaderChange(leader *Addr, event LeaderChangeEvent) //WatchElection Callback (Only Watchers)
}

//============================以上为用户自定义接口, 以下为框架提供接口==========================

//Service Mesh通信层Cluster节点上层会话(管理一个Sidecar和下属所有Service)
type MeshServer interface {
	//注册服务, 同步阻塞
	Register(serviceUrl string, service AppService) (MeshService, error)

	//注销服务, 非阻塞
	UnRegister(serviceUrl string) error

	//返回ServiceName的服务路由
	GetNameRouter(serviceName string) NameRouter

	//获取ServiceName选举的leader地址
	GetElectionLeader(serviceName string) *Addr

	//优雅退出
	GracefulStop()
}

//Service Mesh逻辑层服务节点上层会话(对应某个Cluster节点下属某一个Service)
type MeshService interface {
	//instID有效时,定向发送. instID为INVALID_ROUTER_ID时,选择ServiceName最佳质量链路发送
	SendBySvcNameAndInstID(serviceName string, instID uint64, msg []byte) error

	//向指定服务发送消息
	SendByHandle(dstHandle uint64, msg []byte) error

	//根据服务url,定向发送
	SendBySvcUrl(dstServiceUrl string, msg []byte) error

	//根据ServiceName 广播给所有对应的服务
	BroadcastBySvcName(dstServiceName string, msg []byte) error

	//获取当前服务地址
	GetLocalAddr() *Addr

	//设置服务实例属性/标签, 所有Watch该服务名的NameWatcher都会收到通知
	SetAttribute(attrs ServiceAttr) error

	//获取服务实例属性/标签
	GetAttribute() ServiceAttr

	//当前服务实例参与ServiceName Leader选举
	RunForElection(candidate AppElectionCandidate) error

	//当前服务实例退出ServiceName Leader选举
	GiveUpElection() error

	//当前服务实例关注ServiceName Leader选举结果
	WatchElection(watchName string, watcher AppElectionWatcher) error

	//当前服务实例取消关注ServiceName Leader选举结果
	UnWatchElection(watchName string) error

	//当前服务实例是否是ServiceName选举的Leader
	IsLeader() bool
}

//对应ServiceName的名字路由管理器
type NameRouter interface {
	//用户注册名字路由状态变化监控器
	Watch(watcher AppRouterWatcher)

	//用户注销名字路由状态变化监控器
	UnWatch(watcher AppRouterWatcher)

	//获取ServiceName所有服务实例地址, map key为ServiceId
	GetInstsAddr() map[uint64]*Addr

	//获取ServiceName所有服务实例属性, map key为ServiceId
	GetInstsAttr() map[uint64]ServiceAttr

	//返回ServiceId
	SelectRouterByModHash(key uint64) uint64

	//返回ServiceId
	SelectRouterByLoop() uint64

	//返回ServiceId
	SelectRouterByQuality() uint64
}