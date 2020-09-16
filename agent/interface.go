package skymesh

//用户注册自定义服务需要实现的接口
type AppService interface {
	OnRegister(trans MeshService, result int32)
	OnUnRegister()
	OnMessage(rmtAddr *Addr, msg []byte)
}

//Service Mesh逻辑层服务节点上层会话(对应某个Cluster节点下属某一个Service)
type MeshService interface {
	SendBySvcNameAndInstID(serviceName string, instID uint64, msg []byte) error //instID非0时,定向发送. instID为0时,选择ServiceName最佳质量链路发送
	SendByHandle(dstHandle uint64, msg []byte) error             //向指定服务发送消息
	SendBySvcUrl(dstServiceUrl string, msg []byte) error         //根据服务url,定向发送
	BroadcastBySvcName(dstServiceName string, msg []byte) error  //根据ServiceName 广播给所有对应的服务
	GetLocalAddr() *Addr
	SetAttribute(attrs ServiceAttr) error             //设置服务实例属性/标签, 所有Watch该服务名的NameWatcher都会收到通知
	GetAttribute() ServiceAttr                        //获取服务实例属性/标签
}

//Service Mesh通信层Cluster节点上层会话(管理一个Sidecar和下属所有Service)
type MeshServer interface {
	Register(serviceUrl string, service AppService) (MeshService, error)            //注册服务
	UnRegister(serviceUrl string) error                                             //注销服务
	GetNameRouter(serviceName string) NameRouter                                    //返回serviceName的服务路由
	Serve() error                                                                   //阻塞循环
	GracefulStop()                                                                  //优雅退出
}

type NameWatcher interface {
	OnInstOnline(addr *Addr)
	OnInstOffline(addr *Addr)
	OnInstSyncAttr(addr *Addr, attrs ServiceAttr)
}

type NameRouter interface {
	Watch(watcher NameWatcher)
	UnWatch(watcher NameWatcher)
	GetInstsAddr() map[uint64]*Addr
	GetInstsAttr() map[uint64]ServiceAttr
}