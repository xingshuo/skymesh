package skymesh

import (
	smsync "github.com/xingshuo/skymesh/common/sync"
	"github.com/xingshuo/skymesh/log"
)

type remoteService struct { //其他skymeshServer服务实例的简要信息
	serverAddr  string
	serviceAddr *Addr
}

type Transport interface {
	SendByRouter(serviceName string, msg []byte) error //发送无状态服务消息,通过无状态路由策略选择发送目标服务
	Send(dstHandle uint64, msg []byte) error           //向指定服务发送消息
	GetLocalAddr() *Addr
}

type Service interface {
	OnRegister(trans Transport, result int32)
	OnUnRegister()
	OnMessage(rmtAddr *Addr, msg []byte)
}

type skymeshService struct {
	server   *skymeshServer
	addr     *Addr
	service  Service
	quit     *smsync.Event
	done     *smsync.Event
	msgQueue chan *Message
}

func (s *skymeshService) SendByRouter(serviceName string, msg []byte) error {
	return s.server.SendByRouter(s.addr, serviceName, msg)
}

func (s *skymeshService) Send(dstHandle uint64, msg []byte) error {
	return s.server.Send(s.addr, dstHandle, msg)
}

func (s *skymeshService) GetLocalAddr() *Addr {
	return s.addr
}

func (s *skymeshService) PushMessage(msg *Message) bool {
	select {
	case s.msgQueue <- msg:
		return true
	default:
		log.Errorf("push msg to %s failed.", msg.srcAddr)
		return false
	}
}

func (s *skymeshService) GetMessageSize() int {
	return len(s.msgQueue)
}

func (s *skymeshService) Serve() {
	for {
		select {
		case msg := <-s.msgQueue:
			s.service.OnMessage(msg.srcAddr, msg.data)
		case <-s.quit.Done():
			for len(s.msgQueue) > 0 {
				msg := <-s.msgQueue
				s.service.OnMessage(msg.srcAddr, msg.data)
			}
			s.done.Fire()
			return
		}
	}
}

func (s *skymeshService) Stop() {
	if s.quit.Fire() {
		<-s.done.Done()
		s.service.OnUnRegister()
	}
}
