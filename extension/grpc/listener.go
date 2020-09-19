package skymesh_grpc //nolint

import (
	"fmt"
	"github.com/xingshuo/skymesh/agent"
	"github.com/xingshuo/skymesh/log"
	"net"
)

type skymeshListener struct {
	virConnProto VirConnProto      // virtual conn proto

	connMgr   *ConnMgr        // accept conn manager
	connected chan *SkymeshConn // sync channel for accept conn
	quit      chan struct{}

	trans     skymesh.MeshService
	server    skymesh.MeshServer
	resolvers map[string]skymesh.NameRouter
}

func newSkymeshListener(serviceName string, proto VirConnProto, s skymesh.MeshServer) (*skymeshListener, error) {
	l := &skymeshListener{
		virConnProto: proto,
		connected:    make(chan *SkymeshConn, 1024),
		quit:         make(chan struct{}),
		resolvers:    make(map[string]skymesh.NameRouter),
	}
	_, err := s.Register(serviceName, l)
	if err != nil {
		return nil, err
	}
	l.connMgr = NewConnMgr()
	l.server = s
	return l, nil
}

// skymesh AppService interface
func (l *skymeshListener) OnRegister(trans skymesh.MeshService, result int32) {
	if result == 0 {
		l.trans = trans
	}
}

// skymesh AppService interface
func (l *skymeshListener) OnUnRegister() {
	l.connMgr.Close()
}

func (l *skymeshListener) resetConn(rmtAddr *skymesh.Addr, connID uint64, err error) {
	log.Errorf("Reset remote[%v] conn[%v] reason[%v]\n", rmtAddr.String(), connID, err)
	if packets, err := l.virConnProto.PackPacket(KVConnCmdClose, connID, nil, nil); err == nil {
		_ = l.trans.SendByHandle(rmtAddr.AddrHandle, packets)
	}
}

// skymesh AppService interface
func (l *skymeshListener) OnMessage(rmtAddr *skymesh.Addr, packet []byte) {
	cmd, connID, msg, ext, err := l.virConnProto.UnpackPacket(packet)
	if err != nil {
		log.Errorf("Unpack skymesh message fail: %v\n", err)
		return
	}
	if (cmd & KVConnCmdConnect) != 0 {
		// 尝试将主动链接的服务加入到skymesh的名字监听中，保证可以收到实例的下线通知来处理链接的销毁
		_, ok := l.resolvers[rmtAddr.ServiceName]
		if !ok {
			ns := l.server.GetNameRouter(rmtAddr.ServiceName)
			l.resolvers[rmtAddr.ServiceName] = ns
			ns.Watch(l)
		}
		conn, err := NewSkymeshConn(connID, rmtAddr, l.trans, l.virConnProto)
		if err != nil {
			l.resetConn(rmtAddr, connID, err)
			return
		}
		err = l.connMgr.AddConn(rmtAddr.AddrHandle, connID, conn)
		if err != nil {
			l.resetConn(rmtAddr, connID, err)
			return
		}
		// notify conn to http server accept channel
		select {
		case l.connected <- conn:
		default:
			log.Infof("send connected conn %v failed.\n", conn)
		}
	}
	if (cmd & KVConnCmdData) != 0 {
		conn := l.connMgr.GetConn(rmtAddr.AddrHandle, connID, true)
		if conn == nil {
			l.resetConn(rmtAddr, connID, fmt.Errorf("conn not exist"))
			return
		}
		err = conn.OnRecv(msg, ext)
		if err != nil {
			l.resetConn(rmtAddr, connID, err)
		}
	}
	if (cmd & KVConnCmdClose) != 0 {
		if l.connMgr.DelConn(rmtAddr.AddrHandle, connID) {
			log.Debugf("remote[%v] close conn[%v]\n", rmtAddr.AddrHandle, connID)
		}
	}
}

// skymesh AppRouterWatcher interface
func (l *skymeshListener) OnInstOnline(_ *skymesh.Addr) {
}

// skymesh AppRouterWatcher interface
func (l *skymeshListener) OnInstOffline(addr *skymesh.Addr) {
	l.connMgr.DelConns(addr.AddrHandle)
}

// skymesh AppRouterWatcher interface
func (l *skymeshListener) OnInstSyncAttr(addr *skymesh.Addr, attrs skymesh.ServiceAttr) {

}

// net Listener interface
func (l *skymeshListener) Accept() (net.Conn, error) {
	for {
		select {
		case c := <-l.connected:
			return c, nil
		case <-l.quit:
			return nil, fmt.Errorf("stop")
		}
	}
}

// net Listener interface
func (l *skymeshListener) Close() error {
	//暂时不做引用计数处理,资源由framework统一释放
	close(l.quit)
	return nil
}

// net Listener interface
func (l *skymeshListener) Addr() net.Addr {
	if l.trans != nil {
		return l.trans.GetLocalAddr()
	}
	return nil
}
