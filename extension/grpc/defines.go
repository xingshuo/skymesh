package skymesh_grpc

import (
	"encoding/json"
	"fmt"
	skymesh "github.com/xingshuo/skymesh/agent"
)

const (
	SessionExpiredSecond = 60
)

type dialAddr struct {
	Laddr string
	Raddr skymesh.Addr
}

func dumpDialAddr(da dialAddr) (string, error) {
	adr, err := json.Marshal(da)
	if err != nil {
		return "", fmt.Errorf("json pack error %v", err)
	}
	return string(adr), nil
}

func loadDialAddr(adr string) (dialAddr, error) {
	var da dialAddr
	err := json.Unmarshal([]byte(adr), &da)
	if err != nil {
		return da, err
	}
	return da, nil
}