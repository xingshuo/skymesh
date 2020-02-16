package inner_service

import (
	"context"
	"fmt"
	skymesh "github.com/xingshuo/skymesh/agent"
	"github.com/xingshuo/skymesh/log"
	"time"
)


type HandlerFunc = func(context.Context, []byte) (interface{}, error)

type SMInterceptor = func(context.Context, []byte, HandlerFunc) (interface{}, error)

type SMService interface {
	// 注册处理方法接口，as server
	RegisterFunc(methodName string, handler HandlerFunc) error
	// 发送响应，用于Handler中发送响应
	SendResponse(ctx context.Context, rmtAddr *skymesh.Addr, rspSeq uint64, rsp []byte) error
	// 添加服务端拦截器
	ApplyServerInterceptors(interceptors ...SMInterceptor)
	// 单向请求，不需要响应
	NotifyService(ctx context.Context, targetSvcUrl string, method string, req []byte) error
	// 同步请求，as client
	CallService(ctx context.Context, targetSvcUrl string, method string, req []byte, timeout_ms int) (rsp []byte, err error)
	// 添加客户端拦截器
	ApplyClientInterceptors(interceptors ...SMInterceptor)
}


// 注册SM服务
func RegisterService(server skymesh.Server, serviceUrl string) (SMService, error) {
	s := &SMServiceImpl{
		serviceUrl:   serviceUrl,
		server:       server,
		rpcFramework: new(RpcFramework),
		regResult:    make(chan error),
		methods:      make(map[string]methodHandler),
	}
	err := s.init()
	return s, err
}

// 注销SM服务
func UnRegisterService(service SMService) error {
	serviceImpl, ok := service.(*SMServiceImpl)
	if !ok {
		return fmt.Errorf("unexpect service type")
	}
	return serviceImpl.close()
}


type methodHandler interface {
	onMethod(context.Context, []byte) (interface{}, error)
}

type chainMethodHandler struct {
	handler      HandlerFunc
	interceptors []SMInterceptor
}

func chainHandler(ctx context.Context, msg []byte, interceptors []SMInterceptor, handler HandlerFunc) (interface{}, error) {
	chainer := func(currentInter SMInterceptor, currentHandler HandlerFunc) HandlerFunc {
		return func(currentCtx context.Context, currentMsg []byte) (interface{}, error) {
			return currentInter(currentCtx, currentMsg, currentHandler)
		}
	}

	n := len(interceptors)
	chainedHandler := handler
	for i := n - 1; i >= 0; i-- {
		chainedHandler = chainer(interceptors[i], chainedHandler)
	}

	return chainedHandler(ctx, msg)
}

func (h *chainMethodHandler) onMethod(ctx context.Context, msg []byte) (interface{}, error) {
	return chainHandler(ctx, msg, h.interceptors, h.handler)
}

type SMServiceImpl struct {
	serviceUrl         string
	server             skymesh.Server
	regResult          chan error		// 注册完成时的通知,将skymeshService的异步注册转为同步
	seq                uint64
	methods            map[string]methodHandler
	serverInterceptors []SMInterceptor
	clientInterceptors []SMInterceptor
	rpcFramework       *RpcFramework
	packBuffer         [PACK_BUFFER_LEN]byte //公用的数据打包buffer
}

func (s *SMServiceImpl) init() error {
	s.rpcFramework.Init()
	err := s.server.Register(s.serviceUrl, s)
	if err != nil {
		return err
	}
	// 同步等待注册完成的skymesh通知
	err = <-s.regResult
	return err
}

// 注销并清理资源
func (s *SMServiceImpl) close() error {
	return s.server.UnRegister(s.serviceUrl)
}

//Server Method
func (s *SMServiceImpl) RegisterFunc(methodName string, handler HandlerFunc) error {
	s.methods[methodName] = &chainMethodHandler{handler: handler, interceptors: s.serverInterceptors}
	return nil
}

func (s *SMServiceImpl) SendResponse(ctx context.Context, rmtAddr *skymesh.Addr, rspSeq uint64, rsp []byte) error {
	targetSvcUrl := skymesh.SkymeshAddr2Url(rmtAddr, false)
	return s.rawSend(ctx, targetSvcUrl, kMsgTypeRsp, rspSeq, "", rsp)
}

// 服务端能力实现
func (s *SMServiceImpl) ApplyServerInterceptors(interceptors ...SMInterceptor) {
	for _, interceptor := range interceptors {
		s.serverInterceptors = append(s.serverInterceptors, interceptor)
	}
}

func (s *SMServiceImpl) rawSend(ctx context.Context, targetSvcUrl string, msgType uint8, seq uint64, method string, msgBody []byte) error {
	if len(method) > RPC_METHOD_MAX_LEN {
		return RPC_METHOD_LEN_OVER_ERR
	}
	head := &SMRpcHead{}
	head.Init()
	if len(msgBody) + int(head.Size()) > PACK_BUFFER_LEN {
		return PACK_BUFFER_SHORT_ERR
	}
	head.msgType = msgType
	head.methodLen = uint8(len(method))
	copy(head.method[:], []byte(method))
	head.timeUS = uint64(time.Now().UnixNano()/1000)
	head.seq = seq
	err,packLen := head.Pack(s.packBuffer[:])
	if err != nil {
		return err
	}
	n := copy(s.packBuffer[packLen:], msgBody)
	packLen = packLen + uintptr(n)
	return s.server.SendBySvcUrl(s.serviceUrl, targetSvcUrl, s.packBuffer[:packLen])
}

// 客户端能力实现
func (s *SMServiceImpl) NotifyService(ctx context.Context, targetSvcUrl string, method string, req []byte) error {
	return s.rawSend(ctx, targetSvcUrl, kMsgTypeReq,0, method, req)
}

// 客户端能力实现
func (s *SMServiceImpl) CallService(ctx context.Context, targetSvcUrl string, method string, req []byte, timeout_ms int) ([]byte, error) {
	call := func(currCtx context.Context, currMsg []byte) (interface{}, error) {
		reqSession := s.rpcFramework.genSessIdWithExclude([]uint64{0})
		err := s.rawSend(currCtx, targetSvcUrl, kMsgTypeReq, reqSession, method, currMsg)
		if err != nil {
			return nil, err
		}
		err,rets := s.rpcFramework.Wait(reqSession, timeout_ms)
		if err != nil {
			return nil, err
		}
		if len(rets) != 1 {
			return nil, RPC_RETVALUE_NUM_ERR
		}
		_,ok := rets[0].([]byte)
		if !ok {
			return nil, RPC_RETVALUE_TYPE_ERR
		}
		return rets[0], nil
	}

	rsp,err := chainHandler(ctx, req, s.clientInterceptors, call)
	return rsp.([]byte), err
}

// 客户端能力实现
func (s *SMServiceImpl) ApplyClientInterceptors(interceptors ...SMInterceptor) {
	for _, interceptor := range interceptors {
		s.clientInterceptors = append(s.clientInterceptors, interceptor)
	}
}

// impl of skymesh Service
func (s *SMServiceImpl) OnRegister(trans skymesh.Transport, result int32) {
	if result == 0 {
		s.regResult <- nil
	} else {
		s.regResult <- fmt.Errorf("register fail(%d)", result)
	}
}

// impl of skymesh Service
func (s *SMServiceImpl) OnUnRegister() {
}

// impl of skymesh Service
func (s *SMServiceImpl) OnMessage(rmtAddr *skymesh.Addr, data []byte) {
	head := &SMRpcHead{}
	head.Init()
	err,unpackLen := head.Unpack(data)
	if err != nil {
		log.Errorf("unpack SMService failed:%v\n", err)
		return
	}

	if head.msgType == kMsgTypeRsp {
		err := s.rpcFramework.WakeUp(head.seq, data[unpackLen:])
		if err != nil {
			log.Errorf("wake up session %d failed:%v\n", head.seq, err)
		}
		return
	}

	if head.msgType == kMsgTypeReq {
		method := string(head.method[:head.methodLen])
		handler := s.methods[method]
		if handler == nil {
			log.Errorf("can not find method(%s)\n", method)
			return
		}
		rsp, err := handler.onMethod(context.Background(), data[unpackLen:])
		if err != nil { //Todo: 错误是否要做返回处理???
			log.Errorf("handle method(%s) failed:%v\n", method, err)
			return
		}
		if head.seq != 0 { //need response
			s.SendResponse(context.Background(), rmtAddr, head.seq, rsp.([]byte))
		}
	}
}