package rpcdemo

import (
	"encoding/binary"
	"errors"
	"net"
)

const lenBytes = 8

func ReadMsg(conn net.Conn) ([]byte, error) {
	// 读取消息长度
	msgLenBytes := make([]byte, lenBytes)
	length, err := conn.Read(msgLenBytes)
	if err != nil {
		return nil, err
	}

	if length != lenBytes {
		conn.Close()
		return nil, errors.New("read length data failed")
	}

	headLength := binary.BigEndian.Uint32(msgLenBytes[:4])
	bodyLength := binary.BigEndian.Uint32(msgLenBytes[4:lenBytes])
	bs := make([]byte, headLength+bodyLength)
	n, err := conn.Read(bs[lenBytes:])
	if n != int(headLength+bodyLength-lenBytes) {
		conn.Close()
		return nil, errors.New("tcp not read enough data")
	}
	copy(bs, msgLenBytes)

	return bs, err
}
