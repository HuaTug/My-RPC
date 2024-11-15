package rpcdemo

import (
	"errors"
	"net"

	"HuaTug.com/message"
)

const PingPong = 1

func Ping(i interface{}) error {
	conn, ok := i.(net.Conn)
	if !ok {
		return errors.New("req must implements net.Conn")
	}

	//conn.SetDeadline(time.Now().Add(2 * time.Second))
	req := &message.Request{
		Ping: PingPong,
	}
	req.CalcHeadLength()
	data := message.EncodeReq(req)
	wLen, err := conn.Write(data)
	if err != nil {
		return err
	}

	if wLen != len(data) {
		return errors.New("ping: failed to write all data")
	}

	respMsg, err := ReadMsg(conn)
	if err != nil {
		return err
	}

	res := message.DecodeResp(respMsg)
	if res.Pong != PingPong {
		return errors.New("ping failed")
	}

	return nil
}
