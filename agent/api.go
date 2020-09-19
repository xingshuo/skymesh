package skymesh

import (
	"github.com/xingshuo/skymesh/log"
	"os"
	"os/signal"
	"syscall"
)

//注意:
//  1.每个进程每个appid只启动一个实例
//  2.同步阻塞
func NewServer(conf string, appID string) (MeshServer, error) {
	s := &skymeshServer{}
	err := s.Init(conf, appID, true)
	if err != nil {
		return nil, err
	}
	return s, err
}

//GraceServer可以是skymeshServer, grpcServer...
type GraceServer interface {
	GracefulStop()
}

//接收指定信号，优雅退出接口
func WaitSignalToStop(s GraceServer, sigs ...os.Signal) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, sigs...)
	sig := <-c
	log.Infof("MeshServer(%v) exit with signal(%d)\n", syscall.Getpid(), sig)
	s.GracefulStop()
}