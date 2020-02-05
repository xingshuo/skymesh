//Author: lakefu
//Date:   2020.1.27
//Function: 对外提供接口
package gonet

import (
	"sync"

	smsync "github.com/xingshuo/skymesh/common/sync"
)

func NewDialer(address string, newReceiver func() Receiver, opts ...DialOption) (*Dialer, error) {
	d := &Dialer{
		opts:        defaultDialOptions(),
		address:     address,
		quit:        smsync.NewEvent("gonet.Dialer.quit"),
		newReceiver: newReceiver,
	}
	//处理参数
	for _, opt := range opts {
		opt.apply(&d.opts)
	}
	return d, nil
}

func NewListener(address string, newReceiver func() Receiver) (*Listener, error) {
	l := &Listener{
		conns:       make(map[*Conn]bool),
		address:     address,
		quit:        smsync.NewEvent("gonet.Listener.quit"),
		done:        smsync.NewEvent("gonet.Listener.done"),
		newReceiver: newReceiver,
	}
	l.cv = sync.NewCond(&l.mu)
	return l, nil
}
