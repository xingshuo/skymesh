package inner_service

import (
	"sync"
	"sync/atomic"
	"time"
)

type RpcFramework struct {
	pending map[uint64]chan interface{}
	mu      sync.Mutex
	seq     uint64
}

func (fw *RpcFramework) Init() {
	fw.pending = make(map[uint64]chan interface{})
}

func (fw *RpcFramework) Exit() {
	fw.mu.Lock()
	fw.pending = make(map[uint64]chan interface{})
	fw.mu.Unlock()
}

func (fw *RpcFramework) setSeqID(id uint64) {
	atomic.StoreUint64(&fw.seq, id)
}

func (fw *RpcFramework) genSessID() uint64 {
	return atomic.AddUint64(&fw.seq, 1)
}

func (fw *RpcFramework) genSessIdWithExclude(exclusions []uint64) uint64 {
retry:
	seq := atomic.AddUint64(&fw.seq, 1)
	for _,exc := range exclusions {
		if seq == exc {
			goto retry
		}
	}
	return seq
}

func (fw *RpcFramework) WakeUp(session uint64, args ...interface{}) error {
	fw.mu.Lock()
	done, ok := fw.pending[session]
	if !ok {
		fw.mu.Unlock()
		return RPC_SESSION_NOEXIST_ERR
	}
	delete(fw.pending, session)
	fw.mu.Unlock()
	select {
	case done <- args:
		return nil
	default:
		return RPC_WAKEUP_ERR
	}
}

//call in independent goroutine if async needed
func (fw *RpcFramework) Wait(session uint64, timeout_ms int) (error, []interface{}) {
	fw.mu.Lock()
	_, ok := fw.pending[session]
	if ok {
		fw.mu.Unlock()
		return RPC_SESSION_REPEAT_ERR, nil
	}
	done := make(chan interface{}, 1)
	fw.pending[session] = done
	fw.mu.Unlock()
	if timeout_ms > 0 {
		tempDelay := time.Duration(timeout_ms) * time.Millisecond
		timer := time.NewTimer(tempDelay)
		select {
		case <-timer.C:
			fw.mu.Lock()
			delete(fw.pending, session)
			fw.mu.Unlock()
			return RPC_TIMEOUT_ERR, nil
		case args := <-done:
			timer.Stop()
			return nil, args.([]interface{})
		}
	} else {
		select {
		case args := <-done:
			return nil, args.([]interface{})
		}
	}
}