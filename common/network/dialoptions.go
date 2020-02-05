//Author: lakefu
//Date:   2020.1.27
//Function: Provide dial Optional Config Parameters
package gonet

type dialOptions struct {
	maxRetryTimes int //connect失败,最大重试次数
	retryInterval int //connect重试间隔
	dialTimeout   int //connect的超时时长,秒级
}

type DialOption interface {
	apply(*dialOptions)
}

type funcDialOption struct {
	f func(*dialOptions)
}

func (fdo *funcDialOption) apply(do *dialOptions) {
	fdo.f(do)
}

func newFuncDialOption(f func(*dialOptions)) *funcDialOption {
	return &funcDialOption{
		f: f,
	}
}

func WithMaxRetryTimes(n int) DialOption {
	return newFuncDialOption(func(do *dialOptions) {
		do.maxRetryTimes = n
	})
}

func WithRetryInterval(n int) DialOption {
	return newFuncDialOption(func(do *dialOptions) {
		do.retryInterval = n
	})
}

func WithDialTimeout(n int) DialOption {
	return newFuncDialOption(func(do *dialOptions) {
		do.dialTimeout = n
	})
}

func defaultDialOptions() dialOptions {
	return dialOptions{
		maxRetryTimes: 10,
		retryInterval: 3,
		dialTimeout:   5,
	}
}
