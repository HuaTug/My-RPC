package json

import "encoding/json"

type SerializerJson struct {
}

func (s SerializerJson) Code() byte {
	return 1
}

func (s SerializerJson) Encode(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

func (s SerializerJson) Decode(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
