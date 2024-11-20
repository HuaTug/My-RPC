package client

import (
	"context"
	"fmt"
	"log"

	"HuaTug.com/codec"
	"HuaTug.com/codes"
	"HuaTug.com/interceptor"
	"HuaTug.com/metadata"
	connpool "HuaTug.com/pool"
	"HuaTug.com/protocol"
	"HuaTug.com/selector"
	"HuaTug.com/stream"
	"HuaTug.com/transport"
	"HuaTug.com/utils"
	"github.com/golang/protobuf/proto"
)

// global client interface
type Client interface {
	Invoke(ctx context.Context, req, rsp interface{}, path string, opts ...Option) error
}

// use a global client
var DefaultClient = New()

var New = func() *defaultClient {
	return &defaultClient{
		opts: &Options{
			protocol: "proto",
		},
	}
}

type defaultClient struct {
	opts *Options
}

// call by reflect
func (c *defaultClient) Call(ctx context.Context, servicePath string, req interface{}, rsp interface{},
	opts ...Option) error {

	// reflection calls need to be serialized using msgpack
	callOpts := make([]Option, 0, len(opts)+1)
	callOpts = append(callOpts, opts...)
	callOpts = append(callOpts, WithSerializationType(codec.MsgPack))

	// servicePath example : /helloworld.Greeter/SayHello
	err := c.Invoke(ctx, req, rsp, servicePath, callOpts...)
	if err != nil {
		return err
	}

	return nil
}

func (c *defaultClient) Invoke(ctx context.Context, req, rsp interface{}, path string, opts ...Option) error {

	for _, o := range opts {
		o(c.opts)
	}

	if c.opts.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.opts.timeout)
		defer cancel()
	}

	// set serviceName, method
	newCtx, clientStream := stream.NewClientStream(ctx)

	serviceName, method, err := utils.ParseServicePath(path)
	if err != nil {
		return err
	}

	c.opts.serviceName = serviceName
	c.opts.method = method

	// TODO : delete or not
	clientStream.WithServiceName(serviceName)
	clientStream.WithMethod(method)

	// execute the interceptor first
	//log.Println("invoke interceptor...", c.opts.interceptors)
	return interceptor.ClientIntercept(newCtx, req, rsp, c.opts.interceptors, c.invoke)
}

func (c *defaultClient) invoke(ctx context.Context, req, rsp interface{}) error {

	serialization := codec.GetSerialization(c.opts.serializationType)
	payload, err := serialization.Marshal(req)
	if err != nil {
		return codes.NewFrameworkError(codes.ClientMsgErrorCode, "request marshal failed ...")
	}
	//log.Println("request payload : ", payload)
	clientCodec := codec.GetCodec(c.opts.protocol)

	// assemble header
	request := addReqHeader(ctx, c, payload)
	reqbuf, err := proto.Marshal(request)
	if err != nil {
		return err
	}

	reqbody, err := clientCodec.Encode(reqbuf)
	if err != nil {
		return err
	}

	clientTransport := c.NewClientTransport()
	clientTransportOpts := []transport.ClientTransportOption{
		transport.WithServiceName(c.opts.serviceName),
		transport.WithClientTarget(c.opts.target),
		transport.WithClientNetwork(c.opts.network),
		transport.WithClientPool(connpool.GetPool("default")),
		transport.WithSelector(selector.GetSelector(c.opts.selectorName)),
		transport.WithTimeout(c.opts.timeout),
	}
	frame, err := clientTransport.Send(ctx, reqbody, clientTransportOpts...)
	if err != nil {
		return err
	}

	rspbuf, err := clientCodec.Decode(frame)
	if err != nil {
		return err
	}
	// parse protocol header 
	// 客户端和服务端之间的通信使用proto协议（这是第一层）
	response := &protocol.Response{}
	if err = proto.Unmarshal(rspbuf, response); err != nil {
		return err
	}

	if response.RetCode != 0 {
		return codes.New(response.RetCode, response.RetMsg)
	}

	// return serialization.Unmarshal(response.Payload, rsp)
	// 客户端和服务端之间第二次通信使用msgpack协议（这是第二层）,此时再解析消息的时候需要再次对其进行反序列化
	err = serialization.Unmarshal(response.Payload, rsp)
	if err != nil {
		return err
	}
	log.Println("response payload : ", rsp)
	return nil
}

func (c *defaultClient) NewClientTransport() transport.ClientTransport {
	return transport.GetClientTransport(c.opts.protocol)
}

func addReqHeader(ctx context.Context, client *defaultClient, payload []byte) *protocol.Request {
	clientStream := stream.GetClientStream(ctx)
	//log.Println("clientStream : ", clientStream)
	servicePath := fmt.Sprintf("/%s/%s", clientStream.ServiceName, clientStream.Method)
	md := metadata.ClientMetadata(ctx)

	// fill the authentication information
	for _, pra := range client.opts.perRPCAuth {
		authMd, _ := pra.GetMetadata(ctx)
		log.Println(authMd)
		for k, v := range authMd {
			md[k] = []byte(v)
		}
	}

	request := &protocol.Request{
		ServicePath: servicePath,
		Payload:     payload,
		Metadata:    md,
	}

	return request
}
