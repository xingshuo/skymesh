Skymesh
========
    基于golang编写的service mesh理念实践的一种实现方式, 目前主要提供了名字服务能力, 其他功能正在陆续补充

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
    protoc 3.x
    protoc-gen-go

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
    /* type Server interface {
            Register(serviceName string, service Service) error //注册服务
            UnRegister(serviceName string) error  //注销服务
            GetNameResolver(serviceName string) NameResolver //返回serviceName的名字解析器
            Send(srcServiceUrl string, dstHandle uint64, b []byte) error //定向发送, 适用有状态服务
            SendByRouter(srcServiceUrl string, dstServiceName string, b []byte) error //根据ServiceName的所有链路质量,选择最佳发送,适用无状态
            Serve() error  //阻塞循环
            GracefulStop() //优雅退出
       }*/
     //详见examples
     
Test
-----
    Windows:
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
    
    Linux:
      待补充