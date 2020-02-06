package smpack

import (
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"github.com/xingshuo/skymesh/log"
	smproto "github.com/xingshuo/skymesh/proto/generate"
)

//协议格式: 4字节包头长度 + 内容
const PkgHeadLen = 4

func Pack(b []byte) []byte {
	bodyLen := len(b)
	data := make([]byte, bodyLen+PkgHeadLen)
	binary.BigEndian.PutUint32(data, uint32(bodyLen))
	copy(data[PkgHeadLen:], b)
	return data
}

func PackSSMsg(msg *smproto.SSMsg) ([]byte, error) { //推荐使用
	b, err := proto.Marshal(msg)
	if err != nil {
		log.Errorf("pb marshal err:%v.\n", err)
		return nil, err
	}
	data := Pack(b)
	return data, nil
}

func Unpack(b []byte) (int, []byte) { //返回(消耗字节数,实际内容)
	if len(b) < PkgHeadLen { //不够包头长度
		return 0, nil
	}
	bodyLen := int(binary.BigEndian.Uint32(b))
	if len(b) < PkgHeadLen+bodyLen { //不够body长度
		return 0, nil
	}
	msgLen := PkgHeadLen + bodyLen
	return msgLen, b[PkgHeadLen:msgLen]
}
