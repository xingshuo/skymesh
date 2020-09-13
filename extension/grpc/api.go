package skymesh_grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/xingshuo/skymesh/log"
)

//一个app注册一次
func RegisterApp(conf string, appID string) error {
	err := adaptor.register(conf, appID)
	if err != nil {
		log.Errorf("app:%s skymesh grpc init failed:%v\n", appID, err)
	}
	return err
}

//所有app只释放一次
func Release() {
	adaptor.release()
}

//grpc schem解析target时格式为：schem://authority/service_name
//authority后续考虑在skymesh做权限认证，即为当前传''
//因此传入到grpc.Dial的target格式务必为：skymesh:///service_name，注意"skymesh:"后有3个'/'
func Dial(_ context.Context, addr string) (net.Conn, error) {
	da, err := loadDialAddr(addr)
	if err != nil {
		return nil, err
	}
	d := adaptor.getDialer(da.Laddr)
	if d != nil {
		return d.dial(&da.Raddr)
	}
	return nil, fmt.Errorf("get no dialer")
}

// address格式: [skymesh://]app_id.env_name.svc_name/inst_id
func Listen(address string) (net.Listener, error) {
	return adaptor.newListener(address)
}