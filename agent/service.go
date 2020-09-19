package skymesh

import (
	"fmt"
	smsync "github.com/xingshuo/skymesh/common/sync"
	"github.com/xingshuo/skymesh/log"
	"sync"
	"time"
)

type remoteService struct { //其他skymeshServer服务实例的简要信息
	serverAddr     string
	serviceAddr    *Addr
	lastPingSendMS int64   //毫秒
	rttMetrics     []int64 //近n次往返延迟采样
	curPingSeq     uint64  //当前Ping包Seq编号
	lastAckPingSeq uint64  //上一次收到回复的Ping包编号
	mu             sync.RWMutex
	Attributes     ServiceAttr //远端服务属性同步缓存
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

//这里可能需要DeepCopy
func (rs *remoteService) SetAttribute(attrs ServiceAttr) {
	copy := make(ServiceAttr)
	for k,v := range attrs {
		copy[k] = v
	}
	rs.mu.Lock()
	rs.Attributes = copy
	rs.mu.Unlock()
}

//这里可能需要DeepCopy
func (rs *remoteService) GetAttribute() ServiceAttr {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	copy := make(ServiceAttr)
	for k,v := range rs.Attributes {
		copy[k] = v
	}
	return copy
}

type skymeshService struct {
	server       *skymeshServer
	addr         *Addr
	service      AppService
	quit         *smsync.Event
	done         *smsync.Event
	msgQueue     chan Message
	mu           sync.RWMutex
	attrs        ServiceAttr
	elecListener AppElectionListener
	regNotify    chan error
}

func (s *skymeshService) SendBySvcNameAndInstID(serviceName string, instID uint64, msg []byte) error {
	if instID == INVALID_ROUTER_ID { //选择最佳链路
		return s.server.sendByRouter(s.addr, serviceName, msg)
	}
	return s.server.sendBySvcUrl(s.addr, MakeSkymeshUrl(serviceName, instID), msg)
}

func (s *skymeshService) SendByHandle(dstHandle uint64, msg []byte) error {
	return s.server.sendByHandle(s.addr, dstHandle, msg)
}

func (s *skymeshService) SendBySvcUrl(dstServiceUrl string, msg []byte) error {
	return s.server.sendBySvcUrl(s.addr, dstServiceUrl, msg)
}

func (s *skymeshService) BroadcastBySvcName(dstServiceName string, msg []byte) error {
	return s.server.broadcastBySvcName(s.addr, dstServiceName, msg)
}

func (s *skymeshService) GetLocalAddr() *Addr {
	return s.addr
}

//这里可能需要DeepCopy
func (s *skymeshService) SetAttribute(attrs ServiceAttr) error {
	copy := make(ServiceAttr)
	for k,v := range attrs {
		copy[k] = v
	}
	s.mu.Lock()
	s.attrs = copy
	s.mu.Unlock()
	return s.server.setAttribute(s.addr, copy)
}

//这里可能需要DeepCopy
func (s *skymeshService) GetAttribute() ServiceAttr {
	s.mu.RLock()
	defer s.mu.RUnlock()
	copy := make(ServiceAttr)
	for k,v := range s.attrs {
		copy[k] = v
	}
	return copy
}

func (s *skymeshService) RunForElection() error {
	return s.server.runForElection(s.addr)
}

func (s *skymeshService) GiveUpElection() error {
	return s.server.giveUpElection(s.addr)
}

func (s *skymeshService) WatchElection(watchSvcName string) error {
	return s.server.watchElection(s.addr, watchSvcName)
}

func (s *skymeshService) UnWatchElection(watchSvcName string) error {
	return s.server.unWatchElection(s.addr, watchSvcName)
}

func (s *skymeshService) IsLeader() bool {
	return s.server.isElectionLeader(s.addr)
}

func (s *skymeshService) SetElectionListener(listener AppElectionListener) {
	s.mu.Lock()
	s.elecListener = listener
	s.mu.Unlock()
}

func (s *skymeshService) GetElectionListener() AppElectionListener {
	s.mu.Lock()
	lis := s.elecListener
	s.mu.Unlock()
	return lis
}

func (s *skymeshService) OnRegister(result int32) {
	msg := &RegisterMessage{
		result:result,
		svcHandle:s.addr.AddrHandle,
	}
	s.PushMessage(msg)
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
			case REGISTER_MESSAGE:
				regMsg := msg.(*RegisterMessage)
				s.service.OnRegister(s, regMsg.result)
				if regMsg.result == 0 {
					s.regNotify <- nil
				} else {
					s.regNotify <- fmt.Errorf("register service fail(%d)", regMsg.result)
				}
			}
		case <-s.quit.Done():
			log.Debugf("service %s recv quit (%d).\n", s.addr, len(s.msgQueue))
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
			log.Debug("handle service remain msg done.\n")
			s.done.Fire()
			return
		}
	}
}

func (s *skymeshService) Stop() {
	if s.quit.Fire() {
		log.Infof("service %s quit fired.\n", s.addr)
		<-s.done.Done()
		s.service.OnUnRegister()
	}
}
