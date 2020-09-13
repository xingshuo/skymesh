package skymesh_grpc //nolint

import (
	"fmt"
	"net"
	"time"

	"github.com/xingshuo/skymesh/log"
	"github.com/xingshuo/skymesh/agent"
	"github.com/xingshuo/skymesh/common/io"
)

// impl of net.Conn
type SkymeshConn struct {
	connID   uint64
	trans    skymesh.Transport
	rmtAddr  skymesh.Addr
	virProto VirConnProto

	done    chan struct{}
	cbuf    *smio.IoBuffer
	sendSeq uint32
	recvSeq uint32

	connMgr    *ConnMgr
	lastUpTime int64 // 上次激活时间
}

func (c *SkymeshConn) Connect() error {
	return c.Send(KVConnCmdConnect, nil)
}

func (c *SkymeshConn) OnRecv(msg []byte, ext interface{}) error {
	c.lastUpTime = time.Now().Unix()
	seq, ok := ext.(uint32)
	if !ok || c.recvSeq != seq {
		_ = c.Send(KVConnCmdClose, nil)
		return fmt.Errorf("invaild recv seq")
	}
	c.recvSeq++
	// put message to buffer, wait http transport read
	_, err := c.cbuf.Put(msg)
	log.Debugf("conn(%v) recv (%d)bytes\n", c, len(msg))
	return err
}

func (c *SkymeshConn) Send(vConnCmd VConnCmd, msg []byte) error {
	packets, err := c.virProto.PackPacket(vConnCmd, c.connID, msg, c.sendSeq)
	if err != nil {
		return err
	}
	err = c.trans.Send(c.rmtAddr.AddrHandle, packets)
	log.Debugf("conn(%v) send cmd(%v) seq(%v) (%d)bytes err(%v)\n", c, vConnCmd, c.sendSeq, len(msg), err)
	if err != nil {
		return err
	}
	// only data frame incr seq
	if (vConnCmd & KVConnCmdData) != 0 {
		c.sendSeq++
	}
	return nil
}

func (c *SkymeshConn) SetConnMngr(mgr *ConnMgr) {
	c.connMgr = mgr
}

func (c *SkymeshConn) OnClose() error {
	// 借助conn manager保证只关闭一次
	close(c.done)
	return nil
}

func (c *SkymeshConn) CheckExpired(now int64) bool {
	return c.lastUpTime + SessionExpiredSecond < now
}

// interface net.Conn
func (c *SkymeshConn) Read(b []byte) (int, error) {
	return c.cbuf.Get(b)
}

// interface net.Conn
func (c *SkymeshConn) Write(b []byte) (int, error) {
	err := c.Send(KVConnCmdData, b)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}

// interface net.Conn
func (c *SkymeshConn) Close() error {
	_ = c.Send(KVConnCmdClose, nil)
	// 从管理器中删除本链接
	if c.connMgr.DelConn(c.rmtAddr.AddrHandle, c.connID) {
		return nil
	}
	return fmt.Errorf("conn has closed")
}

// interface net.Conn
func (c *SkymeshConn) LocalAddr() net.Addr {
	return c.trans.GetLocalAddr()
}

// interface net.Conn
func (c *SkymeshConn) RemoteAddr() net.Addr {
	return &c.rmtAddr
}

// interface net.Conn
func (c *SkymeshConn) SetDeadline(t time.Time) error {
	_ = c.SetReadDeadline(t)
	return nil
}

// interface net.Conn
func (c *SkymeshConn) SetReadDeadline(t time.Time) error {
	c.cbuf.SetTimeout(t)
	return nil
}

// interface net.Conn
func (c *SkymeshConn) SetWriteDeadline(_ time.Time) error {
	return nil
}

func (c *SkymeshConn) String() string {
	return fmt.Sprintf("[%s->%s (%d,%d)]", c.trans.GetLocalAddr().String(),
		c.rmtAddr.String(), c.rmtAddr.AddrHandle, c.connID)
}


func NewSkymeshConn(connID uint64, rmtAddr *skymesh.Addr,
	trans skymesh.Transport, virProto VirConnProto) (*SkymeshConn, error) {
	c := &SkymeshConn{
		connID:     connID,
		trans:      trans,
		rmtAddr:    *rmtAddr,
		virProto:   virProto,
		done:       make(chan struct{}),
		cbuf:       nil,
		sendSeq:    0,
		recvSeq:    0,
		lastUpTime: time.Now().Unix(),
	}
	c.cbuf = smio.NewIoBuffer(c.done)
	return c, nil
}
