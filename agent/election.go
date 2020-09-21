package skymesh

import (
	"errors"
	"sync"
)

type skymeshElection struct {
	svcName    string //watch or run for
	rwmu       sync.RWMutex
	server     *skymeshServer
	candidates map[uint64]AppElectionCandidate //key is service handle
	leader     *Addr
	watchers   map[uint64]AppElectionWatcher //key is service handle
}

func (se *skymeshElection) Init() {

}

func (se *skymeshElection) onRegisterApp(leader *Addr) {
	se.rwmu.Lock()
	se.leader = leader
	cand := se.candidates[leader.AddrHandle]
	se.rwmu.Unlock()

	//通知竞选人成功当选Leader
	if cand != nil {
		ldrSvc := se.server.getService(leader.AddrHandle)
		cand.OnRegisterLeader(ldrSvc, KElectionResultOK)
	}

	var watchers []AppElectionWatcher
	se.rwmu.RLock()
	for _,w := range se.watchers {
		if w != nil {
			watchers = append(watchers, w)
		}
	}
	se.rwmu.RUnlock()

	//通知选举结果监听者新Leader当选
	for _,w := range watchers {
		w.OnLeaderChange(leader, KGotElectionLeader)
	}
}

func (se *skymeshElection) onRunForElectionResultNotify(candidate *Addr, result int32) {
	if result == KElectionResultOK {
		newLeader := candidate
		var watchers []AppElectionWatcher
		se.rwmu.Lock()
		for _,w := range se.watchers {
			if w != nil {
				watchers = append(watchers, w)
			}
		}
		oldLeader := se.leader
		se.leader = newLeader
		newCan := se.candidates[newLeader.AddrHandle]
		se.rwmu.Unlock()

		if oldLeader != nil {
			if oldLeader.AddrHandle == newLeader.AddrHandle { //重复通知了??
				return
			}
			//注意: 目前只有旧leader主动放弃才会切换新leader,这里直接退出候选人列表.
			//     但以后可能会有仲裁节点主动更换leader的情况...
			se.rwmu.Lock()
			oldCan := se.candidates[oldLeader.AddrHandle]
			delete(se.candidates, oldLeader.AddrHandle)
			se.rwmu.Unlock()
			//通知竞选人退出成功
			if oldCan != nil {
				oldCan.OnUnRegisterLeader()
			}
			//通知选举结果监听者旧Leader退出
			for _,w := range watchers {
				w.OnLeaderChange(oldLeader, KLostElectionLeader)
			}
		}
		//通知竞选人成功当选Leader
		if newCan != nil {
			ldrSvc := se.server.getService(newLeader.AddrHandle)
			newCan.OnRegisterLeader(ldrSvc, KElectionResultOK)
		}
		//通知选举结果监听者新Leader当选
		for _,w := range watchers {
			w.OnLeaderChange(newLeader, KGotElectionLeader)
		}
	} else {
		//通知竞选人竞选失败
		se.rwmu.RLock()
		can := se.candidates[candidate.AddrHandle]
		se.rwmu.RUnlock()
		if can != nil {
			svc := se.server.getService(candidate.AddrHandle)
			can.OnRegisterLeader(svc, result)
		}
	}
}

func (se *skymeshElection) onGiveUpElectionResultNotify(candidate *Addr, result int32) {
	if result == KElectionResultQuitCandidate {
		se.rwmu.Lock()
		delete(se.candidates, candidate.AddrHandle)
		se.rwmu.Unlock()
		return
	} else if result == KElectionResultOK {
		se.rwmu.Lock()
		can := se.candidates[candidate.AddrHandle]
		delete(se.candidates, candidate.AddrHandle)
		oldLeader := se.leader
		if oldLeader != nil && oldLeader.AddrHandle == candidate.AddrHandle { //正常情况
			se.leader = nil
			var watchers []AppElectionWatcher
			for _,w := range se.watchers {
				watchers = append(watchers, w)
			}
			se.rwmu.Unlock()
			//通知竞选人退出成功
			if can != nil {
				can.OnUnRegisterLeader()
			}
			//通知选举结果监听者旧Leader退出
			for _,w := range watchers {
				w.OnLeaderChange(oldLeader, KLostElectionLeader)
			}
		} else { //异常情况
			se.rwmu.Unlock()
		}
	}
}

func (se *skymeshElection) unRegisterService(srcAddr *Addr) {
	//giveUpElection由NameServer接收SSCmd_REQ_UNREGISTER_SERVICE消息,再反向通知处理
	se.unWatchElection(srcAddr)
}

func (se *skymeshElection) runForElection(srcAddr *Addr, candidate AppElectionCandidate) error {
	se.rwmu.Lock()
	if se.candidates[srcAddr.AddrHandle] != nil {
		se.rwmu.Unlock()
		return errors.New("service run for election again")
	}
	se.candidates[srcAddr.AddrHandle] = candidate
	se.rwmu.Unlock()
	return nil
}

func (se *skymeshElection) giveUpElection(srcAddr *Addr) error {
	se.rwmu.RLock()
	can := se.candidates[srcAddr.AddrHandle]
	se.rwmu.RUnlock()
	if can == nil {
		return errors.New("service not run for election")
	}
	return nil
}

func (se *skymeshElection) watchElection(srcAddr *Addr, watcher AppElectionWatcher) error {
	se.rwmu.Lock()
	if se.watchers[srcAddr.AddrHandle] != nil {
		se.rwmu.Unlock()
		return errors.New("service watch election again")
	}
	se.watchers[srcAddr.AddrHandle] = watcher
	leader := se.leader
	se.rwmu.Unlock()
	if leader != nil {
		watcher.OnLeaderChange(leader, KGotElectionLeader)
	}
	return nil
}

func (se *skymeshElection) unWatchElection(srcAddr *Addr) error {
	se.rwmu.Lock()
	if se.watchers[srcAddr.AddrHandle] == nil {
		se.rwmu.Unlock()
		return errors.New("service not watch election")
	}
	delete(se.watchers, srcAddr.AddrHandle)
	se.rwmu.Unlock()
	return nil
}

func (se *skymeshElection) getElectionLeader() *Addr {
	se.rwmu.RLock()
	leader := se.leader
	se.rwmu.RUnlock()
	return leader
}

func (se *skymeshElection) isElectionLeader(srcAddr *Addr) bool {
	se.rwmu.RLock()
	leader := se.leader
	se.rwmu.RUnlock()
	if leader == nil {
		return false
	}
	return leader.AddrHandle == srcAddr.AddrHandle
}

func (se *skymeshElection) Release() {
	se.candidates = make(map[uint64]AppElectionCandidate)
	se.watchers = make(map[uint64]AppElectionWatcher)
}