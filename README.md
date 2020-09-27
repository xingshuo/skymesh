Skymesh
========
    基于golang编写的service mesh理念实践的一种实现方式.
    目前支持主要功能:
        *名字服务能力
        *支持自定义协议rpc通信的基础服务(inner_service)
        *grpc支持(skymesh名字服务打通 && 继承其原生能力)
        *路由选择方案支持(轮询、取模hash、最佳质量)
        *主备节点选举
        *服务实例设置属性(对应路由节点获取属性变更通知)/获取属性
    待补充的功能:
        *metric模块
        *trace模块
        *log模块扩充
        *有状态路由选择方案支持(一致性hash...)
    待优化的功能:
        *性能优化(sync.Mutex -> sync.RWMutex, map -> sync.Map...)
        *skymeshServer锁粒度检查 && 死锁check
        *节点上下线流程完善(防止数据比上线事件先到之类)
        *节点Event和Data使用统一recv chan
        *nameserver改成无状态集群 + db

Summary
-------
    1. ServiceUrl格式: [skymesh://]appid.env_name.service_name/instance_id
        appid为应用名称, env_name 为二级范围(如游戏的大区)
        service_name为真正的服务名, instance_id为该名称服务的具体实例id.
        每个ServiceUrl 全网格唯一确定一个服务实例
    2. 服务间通信协议格式: 4字节包头长度 + pb压缩内容

PrepareEnv
-------
    golang 1.13及以上

Platform
-------
    Linux/Windows

Architecture
-------
![flowchart](https://github.com/xingshuo/skymesh/blob/master/flowchart.png)

Api && Interface
-----
    //--------Api使用示例开始--------
    //每个进程每个类型app只启动一个实例
    //s是skymeshServer外部能力的抽象接口MeshServer的实例
    s, err := skymesh.NewServer(conf, appID)
    
    //svc是skymeshService外部能力的抽象接口MeshService的实例
    svc, err = s.Register(svcUrl, AppServiceImpl)
    svc.SendBySvcUrl(dstUrl, []byte("hello"))
    
    skymesh.WaitSignalToStop(s, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
    //--------Api使用示例结束---------
    
    Api定义详见:  agent/api.go
    Interface定义详见:  agent/interface.go
     
Tips
-----
    1.为了消除不同版本protoc生成的pb.go不同引入的问题, 所有proto生成的pb.go统一以生成好的源码方式提供.
    2.框架内部通信协议描述文件为proto\ssproto.proto, 协议对应pb.go文件生成方式(以windows为例):
        protoc --proto_path=.\proto\ --go_out=.\proto\generate .\proto\*.proto
        
Examples
-----
    Windows:
      Build: .\build.bat
      Dir:   .\examples\
      Run <helloworld> demo:
          1.start nameserver
            Cmd: .\nameserver.bat
          2.start server
            Cmd: .\helloworld_server.bat
          3.start client
            Cmd: .\helloworld_client.bat
            
      Run <features\name_router> demo:
          1.start nameserver
            Cmd: .\nameserver.bat
          2.start server 1
            Cmd: .\features_name_router_server.bat 1
          3.start server 2
            Cmd: .\features_name_router_server.bat 2
          4.start client
            Cmd: .\features_name_router_client.bat
            
      Run <inner_service> demo: 
          1.start nameserver
            Cmd: .\nameserver.bat
          2.start server
            Cmd: .\inner_service_server.bat
          3.start client
            Cmd: .\inner_service_client.bat
      
      Run <grpc\helloworld> demo:
          1.start nameserver
            Cmd: .\nameserver.bat
          2.start server
            Cmd: .\grpc_helloworld_server.bat
          3.start client
            Cmd: .\grpc_helloworld_client.bat
      
      Run <features\election> demo:
          1.start nameserver
            Cmd: .\nameserver.bat
          2.start server 1
            Cmd: .\features_election_server.bat 1
          3.start server 2
            Cmd: .\features_election_server.bat 2
          4.start client
            Cmd: .\features_election_client.bat
         
    Linux:
      运行`sh generate_linux_shell.sh`生成对应的build.sh和examples/目录下的*.sh, 其余同windows