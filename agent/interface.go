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
type AppElectionListener interface {
	OnRegisterLeader(trans MeshService, result int32)     //RunForElection Callback
	OnUnRegisterLeader()                                  //GiveUpElection Callback When GiveUp Leader Succeed
	OnLeaderChange(leader *Addr, event LeaderChangeEvent) //WatchElection Callback
}


//Service Mesh通信层Cluster节点上层会话(管理一个Sidecar和下属所有Service)
type MeshServer interface {
	Register(serviceUrl string, service AppService) (MeshService, error)            //注册服务
	UnRegister(serviceUrl string) error                                             //注销服务
	GetNameRouter(serviceName string) NameRouter                                    //返回ServiceName的服务路由
	GetElectionLeader(serviceName string) *Addr										//获取ServiceName选举的leader地址
	Serve() error                                                                   //阻塞循环
	GracefulStop()                                                                  //优雅退出
}

//Service Mesh逻辑层服务节点上层会话(对应某个Cluster节点下属某一个Service)
type MeshService interface {
	SendBySvcNameAndInstID(serviceName string, instID uint64, msg []byte) error //instID有效时,定向发送. instID为INVALID_ROUTER_ID时,选择ServiceName最佳质量链路发送
	SendByHandle(dstHandle uint64, msg []byte) error             //向指定服务发送消息
	SendBySvcUrl(dstServiceUrl string, msg []byte) error         //根据服务url,定向发送
	BroadcastBySvcName(dstServiceName string, msg []byte) error  //根据ServiceName 广播给所有对应的服务
	GetLocalAddr() *Addr
	SetAttribute(attrs ServiceAttr) error             //设置服务实例属性/标签, 所有Watch该服务名的NameWatcher都会收到通知
	GetAttribute() ServiceAttr                        //获取服务实例属性/标签
	RunForElection() error							  //当前服务实例参与ServiceName Leader选举
	GiveUpElection() error							  //当前服务实例退出ServiceName Leader选举
	WatchElection(watchSvcName string) error		  //当前服务实例关注ServiceName Leader选举结果
	UnWatchElection(watchSvcName string) error		  //当前服务实例取消关注ServiceName Leader选举结果
	IsLeader() bool 								  //当前服务实例是否是ServiceName选举的Leader
	SetElectionListener(listener AppElectionListener) //设置选举事件监控器,目前限定只能设置一个
}

type NameRouter interface {
	Watch(watcher AppRouterWatcher)
	UnWatch(watcher AppRouterWatcher)
	GetInstsAddr() map[uint64]*Addr
	GetInstsAttr() map[uint64]ServiceAttr
	SelectRouterByModHash(key uint64) uint64 //返回ServiceId
	SelectRouterByLoop() uint64 //返回ServiceId
	SelectRouterByQuality() uint64 //返回ServiceId
}