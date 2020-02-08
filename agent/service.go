package skymesh

import (
	smsync "github.com/xingshuo/skymesh/common/sync"
	"github.com/xingshuo/skymesh/log"
	"time"
)

type remoteService struct { //其他skymeshServer服务实例的简要信息
	serverAddr     string
	serviceAddr    *Addr
	lastPingSendMS int64   //毫秒
	rttMetrics     []int64 //近n次往返延迟采样
	curPingSeq     uint64  //当前Ping包Seq编号
	lastAckPingSeq uint64  //上一次收到回复的Ping包编号
}

func (rs *remoteService) onSendPing() {
	sendTime := int64(time.Now().UnixNano()/1e6)
	if rs.lastAckPingSeq == rs.curPingSeq { //已经收到回包
		rs.curPingSeq++
		rs.lastPingSendMS = sendTime
	} else {
		rs.rttMetrics = append(rs.rttMetrics, sendTime - rs.lastPingSendMS)
		if len(rs.rttMetrics) > AVERAGE_RTT_SAMPLING_NUM {
			rs.rttMetrics = append(rs.rttMetrics[:0], rs.rttMetrics[1:]...)
		}
		rs.lastAckPingSeq = rs.curPingSeq
		rs.curPingSeq++
		rs.lastPingSendMS = sendTime
	}
}

func (rs *remoteService) PingAck(ackSeq uint64) {
	if ackSeq != rs.curPingSeq { //之前ping包的回复,已过期
		return
	}
	if rs.lastAckPingSeq == ackSeq { //重复包
		return
	}
	recvTime := int64(time.Now().UnixNano()/1e6)
	rs.rttMetrics = append(rs.rttMetrics, recvTime - rs.lastPingSendMS)
	if len(rs.rttMetrics) > AVERAGE_RTT_SAMPLING_NUM {
		rs.rttMetrics = append(rs.rttMetrics[:0], rs.rttMetrics[1:]...)
	}
	rs.lastAckPingSeq = ackSeq
}

func (rs *remoteService) CalAverageRTT() int64 { //毫秒
	if len(rs.rttMetrics) == 0 {
		return  0
	}
	var sum int64
	for _,rtt := range rs.rttMetrics {
		sum = sum + rtt
	}
	return sum/int64(len(rs.rttMetrics))
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
	msgQueue chan Message
}

func (s *skymeshService) SendByRouter(serviceName string, msg []byte) error {
	return s.server.sendByRouter(s.addr, serviceName, msg)
}

func (s *skymeshService) Send(dstHandle uint64, msg []byte) error {
	return s.server.send(s.addr, dstHandle, msg)
}

func (s *skymeshService) GetLocalAddr() *Addr {
	return s.addr
}

func (s *skymeshService) PushMessage(msg Message) bool {
	select {
	case s.msgQueue <- msg:
		return true
	default:
		log.Errorf("push msg to %d failed.", msg.GetDstHandle())
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
			switch msg.GetMessageType() {
			case DATA_MESSAE:
				dataMsg := msg.(*DataMessage)
				s.service.OnMessage(dataMsg.srcAddr, dataMsg.data)
			case PING_MESSAGE:
				pingMsg := msg.(*PingMessage)
				s.server.sidecar.sendPingAck(pingMsg.dstHandle, pingMsg.seq, pingMsg.srcServerAddr)
			}
		case <-s.quit.Done():
			for len(s.msgQueue) > 0 {
				msg := <-s.msgQueue
				switch msg.GetMessageType() {
				case DATA_MESSAE:
					dataMsg := msg.(*DataMessage)
					s.service.OnMessage(dataMsg.srcAddr, dataMsg.data)
				case PING_MESSAGE:
					pingMsg := msg.(*PingMessage)
					s.server.sidecar.sendPingAck(pingMsg.dstHandle, pingMsg.seq, pingMsg.srcServerAddr)
				}
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
