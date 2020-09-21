//Author: lakefu
//Date:   2020.1.27
//Function: tcp Dial流程封装, 并提供断线重连的能力
package gonet

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/xingshuo/skymesh/common/sync"
)

type Dialer struct {
	opts        dialOptions
	conn        *Conn
	address     string
	newReceiver func() Receiver
	wg          sync.WaitGroup
	quit        *smsync.Event
}

func (d *Dialer) dial() error {
	var (
		err  error
		rawConn net.Conn
	)
	if d.opts.dialTimeout > 0 {
		rawConn, err = net.DialTimeout("tcp", d.address, time.Duration(d.opts.dialTimeout)*time.Second)
	} else {
		rawConn, err = net.Dial("tcp", d.address)
	}
	if err != nil {
		return err
	}
	d.conn = new(Conn)
	err = d.conn.Init(rawConn, d.newReceiver())
	if err != nil {
		return err
	}
	d.wg.Add(2)
	go func() {
		err := d.conn.loopRead()
		if err != nil {
			d.conn.Close()
			log.Printf("loop read err:%v", err)
		}
		d.wg.Done()
	}()
	go func() {
		err := d.conn.loopWrite()
		if err != nil {
			d.conn.Close()
			log.Printf("loop write err:%v", err)
		}
		d.wg.Done()
	}()
	return nil
}

func (d *Dialer) reConnect() error {
	for {
		select {
		case <-d.conn.Done():
			log.Printf("connection close")
		case <-d.quit.Done():
			log.Printf("quit dial")
			return nil
		}

		log.Printf("wait to reconnect..")
		d.wg.Wait()
		log.Printf("start reconnect")

		//---下面是重连逻辑---
		retryTimes := 0
	retry:
		for {
			select {
			case <-d.quit.Done():
				log.Printf("quit dial.")
				return nil
			default:
				retryTimes++
				if err := d.dial(); err != nil {
					log.Printf("reconnect %dth failed %v", retryTimes, err)
					if retryTimes >= d.opts.maxRetryTimes { //达到重连次数上限,退出
						return fmt.Errorf("reconnect %d times failed", retryTimes)
					}
					time.Sleep(time.Duration(d.opts.retryInterval) * time.Second)
					continue
				}
				log.Printf("reconnect %dth succeed!", retryTimes)
				break retry
			}
		}
	}
}

//外部调用接口
func (d *Dialer) Start() error {
	if err := d.dial(); err != nil {
		return err
	}
	go d.reConnect()
	return nil
}

func (d *Dialer) Shutdown() error {
	err := fmt.Errorf("repeat shutdown")
	if d.quit.Fire() {
		d.conn.Close()
		err = nil
	}
	return err
}

func (d *Dialer) Send(b []byte) {
	d.conn.Send(b)
}
