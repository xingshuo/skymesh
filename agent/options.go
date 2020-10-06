package skymesh

type ServiceOptions struct {
	ConsistentHashKey uint64
}

type RegisterOption interface {
	apply(*ServiceOptions)
}

type funcRegisterOption struct {
	f func(*ServiceOptions)
}

func (fro *funcRegisterOption) apply(ro *ServiceOptions) {
	fro.f(ro)
}

func newFuncRegisterOption(f func(*ServiceOptions)) *funcRegisterOption {
	return &funcRegisterOption{
		f: f,
	}
}

func WithConsistentHashKey(key uint64) RegisterOption {
	return newFuncRegisterOption(func(ro *ServiceOptions) {
		ro.ConsistentHashKey = key
	})
}

func defaultRegisterOptions() ServiceOptions {
	return ServiceOptions{
		ConsistentHashKey: INVALID_CONSISTENT_HASH_KEY,
	}
}