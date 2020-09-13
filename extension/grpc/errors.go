package skymesh_grpc

import (
	"fmt"
)

// TimeoutError is returned for an expired deadline.
type TimeoutError struct{}

// Implement the net.Error interface.
func (e *TimeoutError) Error() string   { return "skymesh i/o timeout" }
func (e *TimeoutError) Timeout() bool   { return true }
func (e *TimeoutError) Temporary() bool { return true }

var ErrWaitTimeout = fmt.Errorf("skymesh wait timeout")

var ErrRouterMonitorClosed = fmt.Errorf("skymesh router monitor closed")

var ErrAppReRegister = fmt.Errorf("skymesh app re-register")

var ErrAppNoRegister = fmt.Errorf("skymesh app no-register")