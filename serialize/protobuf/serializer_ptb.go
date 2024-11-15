package protobuf

import (
	"errors"

	"google.golang.org/protobuf/proto"
)

type SerializeProto struct {
}

func (s SerializeProto) Code() byte {
	return 2
}

func (s SerializeProto) Encode(val interface{}) ([]byte, error) {
	msg, ok := val.(proto.Message)
	if !ok {
		return nil, errors.New("not a proto.Message")
	}
	return proto.Marshal(msg)
}

func (s SerializeProto) Decode(data []byte, val interface{}) error {
	msg, ok := val.(proto.Message)
	if !ok {
		return errors.New("not a proto.Message")
	}
	return proto.Unmarshal(data, msg)
}
