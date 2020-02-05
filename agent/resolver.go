package skymesh

type NameWatcher interface {
	OnInstOnline(addr *Addr)
	OnInstOffline(addr *Addr)
}

type NameResolver interface {
	Watch(watcher NameWatcher)
	UnWatch(watcher NameWatcher)
	GetInstsAddr() map[uint64]*Addr
}

type skymeshResolver struct {
	svcName   string
	instAddrs map[uint64]*Addr
	watchers  map[NameWatcher]bool
}

func (sr *skymeshResolver) Watch(w NameWatcher) {
	sr.watchers[w] = true
}

func (sr *skymeshResolver) UnWatch(w NameWatcher) {
	delete(sr.watchers, w)
}

func (sr *skymeshResolver) GetInstsAddr() map[uint64]*Addr {
	return sr.instAddrs
}

func (sr *skymeshResolver) notifyInstsChange(isOnline bool, handle uint64, instID uint64) {
	if isOnline {
		_, ok := sr.instAddrs[instID]
		if ok {
			return
		}
		instAddr := &Addr{ServiceName: sr.svcName, ServiceId: instID, AddrHandle: handle}
		sr.instAddrs[instID] = instAddr
		for w := range sr.watchers {
			w.OnInstOnline(instAddr)
		}
	} else {
		instAddr, ok := sr.instAddrs[instID]
		if !ok {
			return
		}
		for w := range sr.watchers {
			w.OnInstOffline(instAddr)
		}
		delete(sr.instAddrs, instID)
	}
}
