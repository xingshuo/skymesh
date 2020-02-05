//Author: lakefu
//Date:   2020.1.27
//Function: 确保事件只执行一次的原语封装,并提供通过channel是否读阻塞和bool类型两种判定事件是否执行的接口
package smsync

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type Event struct {
	fired       int32
	tag         string
	notifyFired chan struct{}
	fireOnce    sync.Once
}

func (e *Event) Fire() bool {
	ret := false
	e.fireOnce.Do(func() {
		atomic.StoreInt32(&e.fired, 1)
		close(e.notifyFired)
		ret = true
	})
	return ret
}

func (e *Event) Done() <-chan struct{} {
	return e.notifyFired
}

func (e *Event) HasFired() bool {
	return atomic.LoadInt32(&e.fired) == 1
}

func (e *Event) String() string {
	return fmt.Sprintf("[syncEvent]:%s", e.tag)
}

func NewEvent(tag string) *Event {
	return &Event{notifyFired: make(chan struct{}), tag: tag}
}
