package skymesh_grpc //nolint

import (
	"encoding/binary"
	"unsafe"
)

type VConnCmd uint32

const (
	KVConnCmdData    VConnCmd = 0x1
	KVConnCmdConnect VConnCmd = 0x2
	KVConnCmdClose   VConnCmd = 0x4
)

// skymesh协议之上虚拟链路协议
type VirConnProto interface {
	PackPacket(cmd VConnCmd, connID uint64, msg []byte, ext interface{}) (packets []byte, err error)
	UnpackPacket(packet []byte) (cmd VConnCmd, connID uint64, msg []byte, ext interface{}, err error)
}

// 虚拟链路协议头
type VirConnHead struct {
	vcmd   VConnCmd
	seq    uint32 //包连续性校验
	connID uint64
}

func (vch *VirConnHead) PackHead() []byte {
	b := make([]byte, vch.Size())
	var pos uintptr
	binary.LittleEndian.PutUint32(b[pos:], uint32(vch.vcmd))
	pos += unsafe.Sizeof(vch.vcmd)

	binary.LittleEndian.PutUint32(b[pos:], vch.seq)
	pos += unsafe.Sizeof(vch.seq)

	binary.LittleEndian.PutUint64(b[pos:], vch.connID)
	pos += unsafe.Sizeof(vch.connID)
	return b[0:pos]
}

func (vch *VirConnHead) UnpackHead(msg []byte) (uintptr, error) {
	var pos uintptr
	vch.vcmd = *(*VConnCmd)(unsafe.Pointer(&msg[pos]))
	pos += unsafe.Sizeof(vch.vcmd)

	vch.seq = *(*uint32)(unsafe.Pointer(&msg[pos]))
	pos += unsafe.Sizeof(vch.seq)

	vch.connID = *(*uint64)(unsafe.Pointer(&msg[pos]))
	pos += unsafe.Sizeof(vch.connID)
	return pos, nil
}

func (vch *VirConnHead) Size() uintptr {
	return unsafe.Sizeof(*vch)
}

// 虚拟链路协议实现
type VirConnProtoDefault struct {
}

func (vc *VirConnProtoDefault) PackPacket(cmd VConnCmd, connID uint64, msg []byte, ext interface{}) (
	packets []byte, err error) {
	var vch VirConnHead
	seq, ok := ext.(uint32)
	if ok {
		vch.seq = seq
	}
	vch.vcmd = cmd
	vch.connID = connID
	//后续优化
	packets = append(packets, vch.PackHead()...)
	packets = append(packets, msg...)
	return
}

func (vc *VirConnProtoDefault) UnpackPacket(packet []byte) (
	cmd VConnCmd, connID uint64, msg []byte, ext interface{}, err error) {
	var (
		vch       VirConnHead
		unpackLen uintptr
	)
	unpackLen, err = vch.UnpackHead(packet)
	if err != nil {
		return
	}
	cmd = vch.vcmd
	connID = vch.connID
	msg = packet[unpackLen:]
	ext = vch.seq
	return
}
