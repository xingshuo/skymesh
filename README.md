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
    //core api:
    s, err := skymesh.NewServer(conf, appID)
    //每个进程每个类型app只启动一个实例
    //s是skymeshServer外部能力的抽象接口Server的实例
    //Server定义
    /*type Server interface {
       	Register(serviceUrl string, service AppService) (MeshService, error)            //注册服务
       	UnRegister(serviceUrl string) error                                             //注销服务
       	GetNameRouter(serviceName string) NameRouter                                    //返回serviceName的服务路由
       	GracefulStop()                                                                  //优雅退出
      }*/
     //详见examples
     
Examples
-----
    Windows:
      helloworld demo:
        Build:
          .\build.bat
        Run:
          1.start nameserver
            Dir: examples\helloworld\nameserver
            Cmd: .\nameserver.bat
          2.start greeter server
            Dir: examples\helloworld\greeter_server
            Cmd: .\greeter_server.bat
          3.start greeter client
            Dir: examples\helloworld\greeter_client
            Cmd: .\greeter_client.bat
    
      nameservice demo:
        Build:
          .\build.bat
        Run:
          1.start nameserver
            Dir: examples\nameservice\nameserver
            Cmd: .\nameserver.bat
          2.start client
            Dir: examples\nameservice\client
            Cmd: .\client.bat
          3.start server 1
            Dir: examples\nameservice\server
            Cmd: .\server.bat server1.json 101
          4.start server 2
            Dir: examples\nameservice\server
            Cmd: .\server.bat server2.json 102
            
      inner_service demo: 
        Build:
          .\build.bat
        Run:
          1.start nameserver
            Dir: examples\inner_service\nameserver
            Cmd: .\nameserver.bat
          2.start server
            Dir: examples\inner_service\server
            Cmd: .\server.bat
          3.start client
            Dir: examples\inner_service\client
            Cmd: .\client.bat
      
      grpc helloworld demo:
        Build:
          .\build.bat
        Run:
          1.start nameserver
            Dir: examples\grpc\helloworld\nameserver
            Cmd: .\nameserver.bat
          2.start server
            Dir: examples\grpc\helloworld\greeter_server
            Cmd: .\greeter_server.bat
          3.start client
            Dir: examples\grpc\helloworld\greeter_client
            Cmd: .\greeter_client.bat
         
    Linux:
      helloworld demo:
        Build:
          sh build.sh
        Run:
          1.start nameserver
            Dir: examples/helloworld/nameserver
            Cmd: sh nameserver.sh
          2.start greeter server
            Dir: examples/helloworld/greeter_server
            Cmd: sh greeter_server.sh
          3.start greeter client
            Dir: examples/helloworld/greeter_client
            Cmd: sh greeter_client.sh

      nameservice demo:
        Build:
          sh build.sh
        Run:
          1.start nameserver
            Dir: examples/nameservice/nameserver
            Cmd: sh nameserver.sh
          2.start client
            Dir: examples/nameservice/client
            Cmd: sh client.sh
          3.start server 1
            Dir: examples/nameservice/server
            Cmd: sh server.sh server1.json 101
          4.start server 2
            Dir: examples/nameservice/server
            Cmd: sh server.sh server2.json 102
      
      inner_service demo:
        Build:
          sh build.sh
        Run:
          1.start nameserver
            Dir: examples/inner_service/nameserver
            Cmd: sh nameserver.sh
          2.start server
            Dir: examples/inner_service/server
            Cmd: sh server.sh
          3.start client
            Dir: examples/inner_service/client
            Cmd: sh client.sh
      
      grpc helloworld demo:
        Build:
          sh build.sh
        Run:
          1.start nameserver
            Dir: examples/grpc/helloworld/nameserver
            Cmd: sh nameserver.sh
          2.start server
            Dir: examples/grpc/helloworld/greeter_server
            Cmd: sh greeter_server.sh
          3.start client
            Dir: examples/grpc/helloworld/greeter_client
            Cmd: sh greeter_client.sh