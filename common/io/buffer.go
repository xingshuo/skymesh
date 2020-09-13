package smio

import (
	"errors"
	"sync"
	"time"

	"github.com/xingshuo/skymesh/log"
)

var (
	ErrCancel                     = errors.New("remote is closed")
	ErrTimeout                    = errors.New("io wait timeout")
	IoLimitDuration time.Duration = 1<<63 - 1
)

type itemNode struct {
	it   interface{}
	next *itemNode
}

type itemList struct {
	head *itemNode
	tail *itemNode
}

func (il *itemList) enqueue(i interface{}) {
	n := &itemNode{it: i}
	if il.tail == nil {
		il.head, il.tail = n, n
		return
	}
	il.tail.next = n
	il.tail = n
}

func (il *itemList) peek() interface{} {
	return il.head.it
}

//nolint
func (il *itemList) dequeue() interface{} {
	if il.head == nil {
		return nil
	}
	i := il.head.it
	il.head = il.head.next
	if il.head == nil {
		il.tail = nil
	}
	return i
}

//nolint
func (il *itemList) dequeueAll() *itemNode {
	h := il.head
	il.head, il.tail = nil, nil
	return h
}

func (il *itemList) isEmpty() bool {
	return il.head == nil
}

type readStream struct {
	buf    []byte
	offset int
}

type IoBuffer struct {
	ch              chan struct{}
	mu              sync.Mutex
	consumerWaiting bool
	list            *itemList
	done            <-chan struct{}
	timeout         *time.Timer
	err             error
}

func NewIoBuffer(done <-chan struct{}) *IoBuffer {
	return &IoBuffer{
		ch:      make(chan struct{}, 1),
		list:    &itemList{},
		done:    done,
		timeout: time.NewTimer(IoLimitDuration),
	}
}

func (c *IoBuffer) Reset(done <-chan struct{}) {
	c.done = done
}

func (c *IoBuffer) Clean() {
	c.consumerWaiting = false
	c.list.dequeueAll()
	c.done = nil
	c.timeout.Stop()
	c.err = nil
}

func (c *IoBuffer) Put(b []byte) (bool, error) {
	var wakeUp bool
	c.mu.Lock()
	if c.err != nil {
		c.mu.Unlock()
		return false, c.err
	}
	if c.consumerWaiting { //ch buffer len 1
		wakeUp = true
		c.consumerWaiting = false
	}

	//注意：需要考虑此处是否需要copy
	s := &readStream{
		offset: 0,
		buf:    b, // make([]byte, len(b)),
	}
	// copy(s.buf[0:], b[0:])
	c.list.enqueue(s)

	c.mu.Unlock()
	if wakeUp {
		select {
		case c.ch <- struct{}{}:
		default:
			log.Info("wakeup iobuffer read failed.\n")
		}
	}
	return true, nil
}

// 从IoBuffer中取出第一段数据
func (c *IoBuffer) Pick() ([]byte, error) {
	for {
	checklist:
		c.mu.Lock()
		c.consumerWaiting = false
		if !c.list.isEmpty() {
			s := c.list.peek().(*readStream)
			b := s.buf[s.offset:]
			c.list.dequeue()
			c.mu.Unlock()
			return b, nil
		}
		if c.err != nil {
			c.mu.Unlock()
			return nil, c.err
		}
		c.consumerWaiting = true
		c.mu.Unlock()
		select {
		case <-c.ch:
		case <-c.timeout.C: //once timer
			c.mu.Lock()
			c.consumerWaiting = false
			c.mu.Unlock()
			return nil, ErrTimeout
		case <-c.done:
			if c.Finish() {
				goto checklist
			}
			return nil, c.err
		}
	}
}

// 从IoBuffer中读取指定slice长度的数据
func (c *IoBuffer) Get(b []byte) (n int, err error) {
	for {
	checklist:
		c.mu.Lock()
		c.consumerWaiting = false
		if !c.list.isEmpty() {
			n := 0
			for {
				s := c.list.peek().(*readStream)
				inc := copy(b[n:], s.buf[s.offset:])
				n += inc
				s.offset += inc
				if s.offset >= len(s.buf) {
					c.list.dequeue()
					if c.list.isEmpty() {
						break
					}
				}
				if n >= len(b) {
					break
				}
			}
			c.mu.Unlock()
			return n, nil
		}
		if c.err != nil {
			c.mu.Unlock()
			return 0, c.err
		}
		c.consumerWaiting = true
		c.mu.Unlock()

		select {
		case <-c.ch:
		case <-c.timeout.C: //once timer
			c.mu.Lock()
			c.consumerWaiting = false
			c.mu.Unlock()
			return 0, ErrTimeout
		case <-c.done:
			if c.Finish() {
				goto checklist
			}
			return 0, c.err
		}
	}
}

func (c *IoBuffer) Finish() bool {
	c.mu.Lock()
	if c.err != nil {
		c.mu.Unlock()
		return false
	}
	c.err = ErrCancel
	c.mu.Unlock()
	return true
}

func (c *IoBuffer) SetTimeout(t time.Time) {
	if t.IsZero() {
		c.timeout.Stop()
	} else {
		d := time.Until(t)
		c.timeout.Reset(d)
	}
}
