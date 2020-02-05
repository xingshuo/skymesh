package smpack

import (
	"encoding/binary"
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
