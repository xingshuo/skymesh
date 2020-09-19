Skymesh
========
    基于golang编写的service mesh理念实践的一种实现方式.
    目前支持主要功能:
        *名字服务能力
        *支持自定义协议rpc通信的基础服务(inner_service)
        *grpc支持(skymesh名字服务打通 && 继承其原生能力)
    待补充的功能:
        *metric模块
        *trace模块
        *log模块扩充
        *路由选择方案支持(轮询、一致性hash、用户自定义...)
        *主备节点选举
        *路由节点状态变更通知
    待优化的功能:
        *性能优化(sync.Mutex -> sync.RWMutex, map -> sync.Map...)
        *节点上下线流程完善(防止数据比上线事件先到之类)
        *节点Event和Data使用统一recv chan
        *gonet模块重构(copy on write)
        *nameserver改成无状态集群 + db

summary
-------
    1. ServiceUrl格式: [skymesh://]appid.env_name.service_name/instance_id
        appid为应用名称, env_name 为二级范围(如游戏的大区)
        service_name为真正的服务名, instance_id为该名称服务的具体实例id.
        每个ServiceUrl 全网格唯一确定一个服务实例
    2. 服务间通信协议格式: 4字节包头长度 + pb压缩内容

pre-env
-------
    golang 1.13及以上

platform
-----
    Linux/Windows

Architecture
-------
![flowchart](https://github.com/xingshuo/skymesh/blob/master/flowchart.png)

Api
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
    功能用例详见: examples/
     
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
          2.start client
            Cmd: .\features_name_router_client.bat
          3.start server 1
            Cmd: .\features_name_router_server.bat 1
          4.start server 2
            Cmd: .\features_name_router_server.bat 2
            
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
         
    Linux:
      Build: ./build.sh
      Dir:   ./examples/
      Generate Shell Script: sh generate_linux_shell.sh
      Run <helloworld> demo:
          1.start nameserver
            Cmd: sh nameserver.sh
          2.start server
            Cmd: sh helloworld_server.sh
          3.start client
            Cmd: sh helloworld_client.sh

      Run <features/name_router> demo:
          1.start nameserver
            Cmd: sh nameserver.sh
          2.start client
            Cmd: sh features_name_router_client.sh
          3.start server 1
            Cmd: sh features_name_router_server.sh 1
          4.start server 2
            Cmd: sh features_name_router_server.sh 2
      
      Run <inner_service> demo:
          1.start nameserver
            Cmd: sh nameserver.sh
          2.start server
            Cmd: sh inner_service_server.sh
          3.start client
            Cmd: sh inner_service_client.sh
      
      Run <grpc/helloworld> demo:
          1.start nameserver
            Cmd: sh nameserver.sh
          2.start server
            Cmd: sh grpc_helloworld_server.sh
          3.start client
            Cmd: sh grpc_helloworld_client.sh