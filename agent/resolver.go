package skymesh

import "sync"

type skymeshResolver struct {
	mu sync.Mutex
	svcName   string
	instAddrs map[uint64]*Addr //后续考虑改成sync.Map
	watchers  map[NameWatcher]bool //后续考虑改成sync.Map
	instAttrs map[uint64]ServiceAttr
}

func (sr *skymeshResolver) Watch(w NameWatcher) {
	sr.mu.Lock()
	sr.watchers[w] = true
	sr.mu.Unlock()
}

func (sr *skymeshResolver) UnWatch(w NameWatcher) {
	sr.mu.Lock()
	delete(sr.watchers, w)
	sr.mu.Unlock()
}

func (sr *skymeshResolver) AddInstsAddr(instID uint64, instAddr *Addr) {
	sr.mu.Lock()
	sr.instAddrs[instID] = instAddr
	sr.mu.Unlock()
}

func (sr *skymeshResolver) GetInstsAddr() map[uint64]*Addr {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	copy := make(map[uint64]*Addr)
	for inst,addr := range sr.instAddrs {
		copy[inst] = addr
	}
	return copy
}

func (sr *skymeshResolver) notifyInstsChange(isOnline bool, handle uint64, instID uint64) {
	if isOnline {
		sr.mu.Lock()
		instAddr := sr.instAddrs[instID]
		if instAddr != nil {
			sr.mu.Unlock()
			return
		}
		instAddr = &Addr{ServiceName: sr.svcName, ServiceId: instID, AddrHandle: handle}
		sr.instAddrs[instID] = instAddr
		var wcopy []NameWatcher
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
		var wcopy []NameWatcher
		for w := range sr.watchers {
			wcopy = append(wcopy, w)
		}
		sr.mu.Unlock()
		for _,w := range wcopy {
			w.OnInstOffline(instAddr)
		}
	}
}

func (sr *skymeshResolver) notifyInstsAttrs(instID uint64, attrs ServiceAttr) {
	sr.mu.Lock()
	instAddr := sr.instAddrs[instID]
	sr.mu.Unlock()
	if instAddr == nil {
		return
	}
	sr.mu.Lock()
	sr.instAttrs[instID] = attrs
	var wcopy []NameWatcher
	for w := range sr.watchers {
		wcopy = append(wcopy, w)
	}
	sr.mu.Unlock()
	for _,w := range wcopy {
		w.OnInstSyncAttr(instAddr, attrs)
	}
}

func (sr *skymeshResolver) AddInstsAttr(instID uint64, attrs ServiceAttr) {
	sr.mu.Lock()
	sr.instAttrs[instID] = attrs
	sr.mu.Unlock()
}

func (sr *skymeshResolver) GetInstsAttr() map[uint64]ServiceAttr {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	copy := make(map[uint64]ServiceAttr)
	for inst,addr := range sr.instAttrs {
		copy[inst] = addr
	}
	return copy
}
