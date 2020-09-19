package skymesh_grpc //nolint

import (
	"fmt"
	"sync"
	"time"

	"github.com/xingshuo/skymesh/log"
	"github.com/xingshuo/skymesh/agent"
)

type skymeshDialer struct { //only consider service grpc client
	mu           sync.Mutex
	virConnProto VirConnProto      // virtual conn proto
	connMgr      *ConnMgr
	trans        skymesh.MeshService
	server       skymesh.MeshServer
	resolvers    map[string]skymesh.NameRouter
}

func newSkymeshDialer(serviceName string, proto VirConnProto, s skymesh.MeshServer) (*skymeshDialer, error) {
	d := &skymeshDialer{}
	d.server = s
	d.virConnProto = proto
	d.connMgr = NewConnMgr()
	d.resolvers = make(map[string]skymesh.NameRouter)
	_, err := s.Register(serviceName, d)
	if err != nil {
		return nil, err
	}
	return d, nil
}

// skymesh AppService interface
func (d *skymeshDialer) OnRegister(trans skymesh.MeshService, result int32) {
	d.trans = trans
}

// skymesh AppService interface
func (d *skymeshDialer) OnUnRegister() {
	d.connMgr.Close()
}

func (d *skymeshDialer) resetConn(rmtAddr *skymesh.Addr, connID uint64, err error) {
	log.Errorf("Reset remote[%v] conn[%v] reason[%v]\n", rmtAddr.String(), connID, err)
	if packets, err := d.virConnProto.PackPacket(KVConnCmdClose, connID, nil, nil); err == nil {
		_ = d.trans.SendByHandle(rmtAddr.AddrHandle, packets)
	}
}

// skymesh AppService interface
func (d *skymeshDialer) OnMessage(rmtAddr *skymesh.Addr, packet []byte) {
	cmd, connID, msg, ext, err := d.virConnProto.UnpackPacket(packet)
	if err != nil {
		log.Errorf("Unpack skymesh message fail: %v\n", err)
		return
	}

	if (cmd & KVConnCmdData) != 0 {
		conn := d.connMgr.GetConn(rmtAddr.AddrHandle, connID, true)
		if conn == nil {
			d.resetConn(rmtAddr, connID, fmt.Errorf("conn not exits"))
			return
		}
		err = conn.OnRecv(msg, ext)
		if err != nil {
			d.resetConn(rmtAddr, connID, err)
		}
	}
	if (cmd & KVConnCmdClose) != 0 {
		if d.connMgr.DelConn(rmtAddr.AddrHandle, connID) {
			log.Debugf("remote[%v] close conn[%v]\n", rmtAddr.String(), connID)
		}
	}
}

func (d *skymeshDialer) dial(rmtAddr *skymesh.Addr) (*SkymeshConn, error) {
	connID := d.connMgr.GenerateConnID()
	conn, err := NewSkymeshConn(connID, rmtAddr, d.trans, d.virConnProto)
	if err != nil {
		return nil, err
	}
	err = d.connMgr.AddConn(rmtAddr.AddrHandle, connID, conn)
	if err != nil {
		return nil, err
	}
	err = conn.Connect()
	if err != nil {
		return nil, err
	}
	return conn, nil
}

//nolint
func (d *skymeshDialer) getNameResolver(target string) (skymesh.NameRouter, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	nr, ok := d.resolvers[target]
	if ok {
		return nr, nil
	}
	nr = d.server.GetNameRouter(target)
	d.resolvers[target] = nr
	return nr, nil
}

func (d *skymeshDialer) newNameWatcher(target string) *dialNameWatcher {
	nr, ok := d.resolvers[target]
	if !ok {
		return nil
	}
	nw := new(dialNameWatcher)
	nw.init(nr, target)
	nr.Watch(nw)
	return nw
}


type dialNameWatcher struct { //one name resolver instance correspond one
	nameResolver skymesh.NameRouter
	target       string //watch service name
	waiting      chan struct{}
	quit         chan struct{}
}

func (nw *dialNameWatcher) init(nr skymesh.NameRouter, target string) {
	nw.target = target
	nw.waiting = make(chan struct{}, 1)
	nw.quit = make(chan struct{})
	nw.nameResolver = nr
}

func (nw *dialNameWatcher) OnInstOnline(addr *skymesh.Addr) {
	log.Infof("AppService(%s) inst(%v) online\n", nw.target, addr.ServiceId)
	select {
	case nw.waiting <- struct{}{}:
	default:
		log.Infof("notify inst %v online failed\n", addr)
	}
}

func (nw *dialNameWatcher) OnInstOffline(addr *skymesh.Addr) {
	log.Infof("AppService(%s) inst(%v) offline\n", nw.target, addr.ServiceId)
	select {
	case nw.waiting <- struct{}{}:
	default:
		log.Infof("notify inst %v offline failed\n", addr)
	}
}

func (nw *dialNameWatcher) OnInstSyncAttr(addr *skymesh.Addr, attrs skymesh.ServiceAttr) {

}

func (nw *dialNameWatcher) wait(waitMs int) error {
	var deadline time.Duration
	if waitMs > 0 {
		deadline = time.Duration(time.Now().UnixNano()) + time.Duration(waitMs)*time.Millisecond
	}
	for {
		if waitMs == 0 {
			return ErrWaitTimeout
		}
		if waitMs > 0 {
			tempDelay := deadline - time.Duration(time.Now().UnixNano())
			if tempDelay <= 0 {
				return ErrWaitTimeout
			}
			timer := time.NewTimer(tempDelay)
			select {
			case <-timer.C:
				return ErrWaitTimeout
			case <-nw.waiting:
				timer.Stop()
				return nil
			case <-nw.quit:
				timer.Stop()
				return ErrRouterMonitorClosed
			}
		}
		select {
		case <-nw.waiting:
			return nil
		case <-nw.quit:
			return ErrRouterMonitorClosed
		}
	}
}

func (nw *dialNameWatcher) close() {
	nr := nw.nameResolver
	if nr != nil {
		nr.UnWatch(nw)
		nw.nameResolver = nil
		close(nw.quit)
	}
}
