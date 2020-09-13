package skymesh_grpc //nolint

import (
	"container/list"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xingshuo/skymesh/log"
)

const (
	CheckExpiredDuration          = time.Second
	MaxExpiredConnNumPerCheck int = 128
)

type ConnMgr struct {
	mu          sync.Mutex
	connLruList *list.List
	conns       map[uint64]map[uint64]*list.Element // map{handle, map{connid, list.Element{SkymeshConn}}}

	connSeq  uint32 // connID生成时的序号，16bit回绕，最多65536个conn
	connSeed uint32 // connID生成时的种子，默认为对象生成时的ms时间，填充为ConnID的低32bit

	ticker       *time.Ticker                          // 检查conn是否expired的定时器
	expitedConns [MaxExpiredConnNumPerCheck]*SkymeshConn // 检查过程中用于保存过期的SkymeshConn的临时数组
}

func NewConnMgr() *ConnMgr {
	cm := new(ConnMgr)
	cm.conns = make(map[uint64]map[uint64]*list.Element)
	cm.connLruList = list.New()
	cm.connSeed = uint32(time.Now().UnixNano() / 1000000)
	cm.ticker = time.NewTicker(CheckExpiredDuration)
	go func() {
		for range cm.ticker.C {
			cm.CheckExpired()
		}
	}()
	return cm
}

func (cm *ConnMgr) GenerateConnID() uint64 {
	// 以ConnMgr创建时的ms作为[0..31]bits；以seq按16bits回绕作为[32..47]bits
	seq := atomic.AddUint32(&cm.connSeq, 1)
	return (uint64(seq&0xffff) << 32) | uint64(cm.connSeed)
}

func (cm *ConnMgr) AddConn(handle uint64, connID uint64, conn *SkymeshConn) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if _, ok := cm.conns[handle]; !ok {
		cm.conns[handle] = make(map[uint64]*list.Element)
	}
	if _, ok := cm.conns[handle][connID]; ok {
		return fmt.Errorf("dump connID[%v]", connID)
	}
	cm.conns[handle][connID] = cm.connLruList.PushBack(conn)
	conn.SetConnMngr(cm)
	log.Infof("SkymeshConn[%v] add to manager\n", connID)
	return nil
}

func (cm *ConnMgr) GetConn(handle uint64, connID uint64, updateLru bool) *SkymeshConn {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	connItems, ok := cm.conns[handle]
	if !ok {
		return nil
	}
	connItem, ok := connItems[connID]
	if !ok {
		return nil
	}
	if updateLru {
		cm.connLruList.MoveToBack(connItem)
	}
	if c, ok := connItem.Value.(*SkymeshConn); ok {
		return c
	}
	return nil
}

func (cm *ConnMgr) DelConn(handle uint64, connID uint64) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	connItems, ok := cm.conns[handle]
	if !ok {
		return false
	}
	connItem, ok := connItems[connID]
	if !ok {
		return false
	}
	if c, ok := connItem.Value.(*SkymeshConn); ok {
		_ = c.OnClose()
	}
	cm.connLruList.Remove(connItem)
	delete(connItems, connID)
	log.Infof("SkymeshConn[%v] del from manager\n", connID)
	return true
}

func (cm *ConnMgr) DelConns(handle uint64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	connItems, ok := cm.conns[handle]
	if !ok {
		return
	}
	for _, connItem := range connItems {
		cm.connLruList.Remove(connItem)
	}
	delete(cm.conns, handle)
}

func (cm *ConnMgr) CheckExpired() {
	now := time.Now().Unix()
	expiredConnNum := 0
	cm.mu.Lock()
	for connItem := cm.connLruList.Front(); connItem != nil &&
		expiredConnNum < MaxExpiredConnNumPerCheck; connItem = connItem.Next() {
		conn, ok := connItem.Value.(*SkymeshConn)
		if !ok {
			// remove unexpected value, resume lru list
			cm.connLruList.Remove(connItem)
			continue
		}
		if !conn.CheckExpired(now) {
			break
		}
		cm.expitedConns[expiredConnNum] = conn
		expiredConnNum++
	}
	cm.mu.Unlock()
	// 在锁外执行Close，避免conn在关闭过程中使用manager的锁
	for idx := 0; idx < expiredConnNum; idx++ {
		_ = cm.expitedConns[idx].Close()
	}
}

func (cm *ConnMgr) Close() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.conns = make(map[uint64]map[uint64]*list.Element)
	cm.connLruList = list.New()
}
