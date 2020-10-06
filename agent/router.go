package skymesh

import (
	"sort"
	"sync"
)

type ConsistentHashInfo struct {
	key    uint64
	instID uint64
}

type skymeshNameRouter struct {
	server    *skymeshServer
	mu        sync.Mutex
	svcName   string
	instAddrs map[uint64]*Addr          //后续考虑改成sync.Map
	watchers  map[AppRouterWatcher]bool //后续考虑改成sync.Map
	instAttrs map[uint64]ServiceAttr
	instOpts  map[uint64]ServiceOptions
	instIDs   []uint64
	ConsHashList []ConsistentHashInfo
	loopSeq   int
	state     NameRouterState
}

func (sr *skymeshNameRouter) Watch(w AppRouterWatcher) {
	sr.mu.Lock()
	sr.watchers[w] = true
	sr.mu.Unlock()
	for _,addr := range sr.instAddrs {
		w.OnInstOnline(addr)
	}
}

func (sr *skymeshNameRouter) UnWatch(w AppRouterWatcher) {
	sr.mu.Lock()
	delete(sr.watchers, w)
	sr.mu.Unlock()
}

func (sr *skymeshNameRouter) AddInstsAddr(instID uint64, instAddr *Addr) {
	sr.mu.Lock()
	v := sr.instAddrs[instID]
	if v == nil {
		sr.instAddrs[instID] = instAddr
		sr.instIDs = append(sr.instIDs, instID)
	}
	sr.mu.Unlock()
}

func (sr *skymeshNameRouter) AddInstsOpts(instID uint64, opts ServiceOptions) {
	sr.mu.Lock()
	sr.instOpts[instID] = opts
	sr.refreshConsHashList()
	sr.mu.Unlock()
}

func (sr *skymeshNameRouter) GetInstsAddr() map[uint64]*Addr {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	copy := make(map[uint64]*Addr)
	for inst,addr := range sr.instAddrs {
		copy[inst] = addr
	}
	return copy
}

func (sr *skymeshNameRouter) refreshConsHashList() {
	var list []ConsistentHashInfo
	for instID,opts := range sr.instOpts {
		list = append(list, ConsistentHashInfo{opts.ConsistentHashKey, instID})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].key < list[j].key
	})
	sr.ConsHashList = list
}

func (sr *skymeshNameRouter) notifyInstsChange(isOnline bool, handle uint64, instID uint64, opts ServiceOptions) {
	if isOnline {
		sr.mu.Lock()
		instAddr := sr.instAddrs[instID]
		if instAddr != nil {
			sr.mu.Unlock()
			return
		}
		instAddr = &Addr{ServiceName: sr.svcName, ServiceId: instID, AddrHandle: handle}
		sr.instAddrs[instID] = instAddr
		sr.instIDs = append(sr.instIDs, instID)
		sr.instOpts[instID] = opts
		sr.refreshConsHashList()
		sr.state = NameRouterReady
		var wcopy []AppRouterWatcher
		for w := range sr.watchers {
			wcopy = append(wcopy, w)
		}
		sr.mu.Unlock()
		for _,w := range wcopy {
			w.OnInstOnline(instAddr)
		}
	} else {
		sr.mu.Lock()
		instAddr := sr.instAddrs[instID]
		if instAddr == nil {
			sr.mu.Unlock()
			return
		}
		delete(sr.instAddrs, instID)
		delete(sr.instAttrs, instID)
		delete(sr.instOpts, instID)
		sr.refreshConsHashList()
		for idx,v := range sr.instIDs {
			if v == instID {
				sr.instIDs = append(sr.instIDs[:idx], sr.instIDs[idx+1:]...)
				break
			}
		}
		var wcopy []AppRouterWatcher
		for w := range sr.watchers {
			wcopy = append(wcopy, w)
		}
		sr.mu.Unlock()
		for _,w := range wcopy {
			w.OnInstOffline(instAddr)
		}
	}
}

func (sr *skymeshNameRouter) notifyInstsAttrs(instID uint64, attrs ServiceAttr) {
	sr.mu.Lock()
	instAddr := sr.instAddrs[instID]
	sr.mu.Unlock()
	if instAddr == nil {
		return
	}
	sr.mu.Lock()
	sr.instAttrs[instID] = attrs
	var wcopy []AppRouterWatcher
	for w := range sr.watchers {
		wcopy = append(wcopy, w)
	}
	sr.mu.Unlock()
	for _,w := range wcopy {
		w.OnInstSyncAttr(instAddr, attrs)
	}
}

func (sr *skymeshNameRouter) AddInstsAttr(instID uint64, attrs ServiceAttr) {
	sr.mu.Lock()
	sr.instAttrs[instID] = attrs
	sr.mu.Unlock()
}

func (sr *skymeshNameRouter) GetInstsAttr() map[uint64]ServiceAttr {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	copy := make(map[uint64]ServiceAttr)
	for inst,addr := range sr.instAttrs {
		copy[inst] = addr
	}
	return copy
}

func (sr *skymeshNameRouter) setState(s NameRouterState) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.state = s
}

func (sr *skymeshNameRouter) SelectRouterByConsistentHash(key uint64) uint64 {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	if len(sr.instIDs) == 0 || sr.state != NameRouterReady || len(sr.ConsHashList) == 0 {
		return INVALID_ROUTER_ID
	}
	shareIdx := 0 //超过最大ConsistentHashkey,也分配到索引为0的实例
	for idx,info := range sr.ConsHashList {
		if key < info.key {
			shareIdx = idx
			break
		}
	}
	return sr.ConsHashList[shareIdx].instID
}

func (sr *skymeshNameRouter) SelectRouterByModHash(key uint64) uint64 {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	if len(sr.instIDs) == 0 || sr.state != NameRouterReady {
		return INVALID_ROUTER_ID
	}
	return sr.instIDs[key % uint64(len(sr.instIDs))]
}

func (sr *skymeshNameRouter) SelectRouterByLoop() uint64 {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	if len(sr.instIDs) == 0 || sr.state != NameRouterReady {
		return INVALID_ROUTER_ID
	}
	seq := sr.loopSeq
	sr.loopSeq++
	return sr.instIDs[seq % len(sr.instIDs)]
}

func (sr *skymeshNameRouter) SelectRouterByQuality() uint64 {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	if len(sr.instIDs) == 0 || sr.state != NameRouterReady {
		return INVALID_ROUTER_ID
	}
	dstAddr := sr.server.GetBestQualityService(sr.svcName)
	if dstAddr == nil { //出现Routers和Service不一致情况
		return sr.instIDs[0]
	}
	for _,id := range sr.instIDs {
		if id == dstAddr.ServiceId {
			return id
		}
	}
	return sr.instIDs[0] //出现Routers和Service不一致情况
}