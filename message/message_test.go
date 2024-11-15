package message

import (
	"bytes"
	"testing"
)

func TestEncodeDecodeRequest(t *testing.T) {
	req := &Request{
		HeadLength:  0,
		BodyLength:  12,
		MessageId:   12345,
		Version:     1,
		Compressor:  2,
		Serializer:  3,
		Ping:        0,
		ServiceName: "MyService",
		MethodName:  "MyMethod",
		Meta:        map[string]string{"key1": "value1", "key2": "value2"},
		Data:        []byte("Hello, World"),
	}

	// 计算头部长度
	req.CalcHeadLength()

	// 编码请求
	encoded := EncodeReq(req)

	// 解码请求
	decoded := DecodeReq(encoded)

	// 比较原始请求和解码后的请求
	if decoded.HeadLength != req.HeadLength {
		t.Errorf("HeadLength mismatch: got %d, want %d", decoded.HeadLength, req.HeadLength)
	}
	if decoded.BodyLength != req.BodyLength {
		t.Errorf("BodyLength mismatch: got %d, want %d", decoded.BodyLength, req.BodyLength)
	}
	if decoded.MessageId != req.MessageId {
		t.Errorf("MessageId mismatch: got %d, want %d", decoded.MessageId, req.MessageId)
	}
	if decoded.Version != req.Version {
		t.Errorf("Version mismatch: got %d, want %d", decoded.Version, req.Version)
	}
	if decoded.Compressor != req.Compressor {
		t.Errorf("Compressor mismatch: got %d, want %d", decoded.Compressor, req.Compressor)
	}
	if decoded.Serializer != req.Serializer {
		t.Errorf("Serializer mismatch: got %d, want %d", decoded.Serializer, req.Serializer)
	}
	if decoded.Ping != req.Ping {
		t.Errorf("Ping mismatch: got %d, want %d", decoded.Ping, req.Ping)
	}
	if decoded.ServiceName != req.ServiceName {
		t.Errorf("ServiceName mismatch: got %s, want %s", decoded.ServiceName, req.ServiceName)
	}
	if decoded.MethodName != req.MethodName {
		t.Errorf("MethodName mismatch: got %s, want %s", decoded.MethodName, req.MethodName)
	}
	t.Logf("Original Request: %+v", req)
	t.Logf("Encoded Data: %v", encoded)
	t.Logf("Decoded Request: %+v", decoded)
	if len(decoded.Meta) != len(req.Meta) {
		t.Errorf("Meta length mismatch: got %d, want %d", len(decoded.Meta), len(req.Meta))
	} else {
		for k, v := range req.Meta {
			if decoded.Meta[k] != v {
				t.Errorf("Meta value mismatch for key %s: got %s, want %s", k, decoded.Meta[k], v)
			}
		}
	}
	if !bytes.Equal(decoded.Data, req.Data) {
		t.Errorf("Data mismatch: got %v, want %v", decoded.Data, req.Data)
	}
}

func TestEncodeDecodeResponse(t *testing.T) {
	resp := &Response{
		HeadLength: 0,
		BodyLength: 12,
		MessageId:  12345,
		Version:    1,
		Compressor: 2,
		Serializer: 3,
		Pong:       0,
		Error:      []byte("Some error"),
		Data:       []byte("Hello, World!"),
	}

	// 计算头部长度
	resp.CalcHeadLength()

	// 编码响应
	encoded := EncodeResp(resp)

	// 解码响应
	decoded := DecodeResp(encoded)

	// 比较原始响应和解码后的响应
	if decoded.HeadLength != resp.HeadLength {
		t.Errorf("HeadLength mismatch: got %d, want %d", decoded.HeadLength, resp.HeadLength)
	}
	if decoded.BodyLength != resp.BodyLength {
		t.Errorf("BodyLength mismatch: got %d, want %d", decoded.BodyLength, resp.BodyLength)
	}
	if decoded.MessageId != resp.MessageId {
		t.Errorf("MessageId mismatch: got %d, want %d", decoded.MessageId, resp.MessageId)
	}
	if decoded.Version != resp.Version {
		t.Errorf("Version mismatch: got %d, want %d", decoded.Version, resp.Version)
	}
	if decoded.Compressor != resp.Compressor {
		t.Errorf("Compressor mismatch: got %d, want %d", decoded.Compressor, resp.Compressor)
	}
	if decoded.Serializer != resp.Serializer {
		t.Errorf("Serializer mismatch: got %d, want %d", decoded.Serializer, resp.Serializer)
	}
	if decoded.Pong != resp.Pong {
		t.Errorf("Pong mismatch: got %d, want %d", decoded.Pong, resp.Pong)
	}
	if !bytes.Equal(decoded.Error, resp.Error) {
		t.Errorf("Error mismatch: got %v, want %v", decoded.Error, resp.Error)
	}
	if !bytes.Equal(decoded.Data, resp.Data) {
		t.Errorf("Data mismatch: got %v, want %v", decoded.Data, resp.Data)
	}
}
