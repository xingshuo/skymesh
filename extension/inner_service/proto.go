package inner_service

import (
	"encoding/binary"
	"reflect"
	"unsafe"
)

const (
	RPC_METHOD_MAX_LEN = 32
	PACK_BUFFER_LEN = 4096
)

const (
	kMsgTypeReq = 1
	kMsgTypeRsp = 2
)

type SMRpcHead struct {
	msgType   uint8
	methodLen uint8
	method [RPC_METHOD_MAX_LEN]byte
	seq    uint64 //0 is notify, otherwise rpc
	timeUS uint64
	version uint8
}

func (h *SMRpcHead) Init() { //as C++ constructor
}

func (h *SMRpcHead) Pack(b []byte) (error, uintptr) {
	if len(b) < int(h.Size()) { //检查buffer够不够上限
		return PACK_BUFFER_SHORT_ERR, 0
	}

	var pos uintptr
	copy(b[pos:], []byte{h.msgType})
	pos = pos + unsafe.Sizeof(h.msgType)

	copy(b[pos:], []byte{h.methodLen})
	pos = pos + unsafe.Sizeof(h.methodLen)

	copy(b[pos:], h.method[:h.methodLen])
	pos = pos + uintptr(h.methodLen)

	binary.LittleEndian.PutUint64(b[pos:], h.seq)
	pos = pos + unsafe.Sizeof(h.seq)

	binary.LittleEndian.PutUint64(b[pos:], h.timeUS)
	pos = pos + unsafe.Sizeof(h.timeUS)

	copy(b[pos:], []byte{h.version})
	pos = pos + unsafe.Sizeof(h.version)
	return nil, pos
}

func (h *SMRpcHead) Unpack(b []byte) (error, uintptr) {
	var (
		pos uintptr
		nextPos uintptr
		size uintptr
	)

	size = unsafe.Sizeof(h.msgType)
	nextPos = pos + size
	if len(b) < int(nextPos) {
		return UNPACK_BUFFER_SHORT_ERR, pos
	}
	h.msgType = uint8(b[pos])
	pos = nextPos

	size = unsafe.Sizeof(h.methodLen)
	nextPos = pos + size
	if len(b) < int(nextPos) {
		return UNPACK_BUFFER_SHORT_ERR, pos
	}
	h.methodLen = uint8(b[pos])
	pos = nextPos

	size = uintptr(h.methodLen)
	nextPos = pos + size
	if len(b) < int(nextPos) {
		return UNPACK_BUFFER_SHORT_ERR, pos
	}
	copy(h.method[:], b[pos:nextPos])
	pos = nextPos

	size = unsafe.Sizeof(h.seq)
	nextPos = pos + size
	if len(b) < int(nextPos) {
		return UNPACK_BUFFER_SHORT_ERR, pos
	}
	h.seq = binary.LittleEndian.Uint64(b[pos:])
	pos = nextPos

	size = unsafe.Sizeof(h.timeUS)
	nextPos = pos + size
	if len(b) < int(nextPos) {
		return UNPACK_BUFFER_SHORT_ERR, pos
	}
	h.timeUS = binary.LittleEndian.Uint64(b[pos:])
	pos = nextPos

	size = unsafe.Sizeof(h.version)
	nextPos = pos + size
	if len(b) < int(nextPos) {
		return UNPACK_BUFFER_SHORT_ERR, pos
	}
	h.version = uint8(b[pos])
	pos = nextPos

	return nil, pos
}

func (h *SMRpcHead) Size() uintptr {
	var size uintptr
	hType := reflect.TypeOf(h)
	for i := 0; i < hType.Elem().NumField(); i++ {
		field := hType.Elem().Field(i)
		size = size + field.Type.Size()
	}
	return size
}