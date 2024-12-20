package client

import (
	"context"
	"fmt"
	"log"

	"github.com/HuaTug/My-RPC/codec"
	"github.com/HuaTug/My-RPC/codes"
	connpool "github.com/HuaTug/My-RPC/pool"
	"github.com/HuaTug/My-RPC/interceptor"
	"github.com/HuaTug/My-RPC/metadata"
	"github.com/HuaTug/My-RPC/protocol"
	"github.com/HuaTug/My-RPC/selector"
	"github.com/HuaTug/My-RPC/stream"
	"github.com/HuaTug/My-RPC/transport"
	"github.com/HuaTug/My-RPC/utils"

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

	// servicePath example : /test.Greeter
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
	// 这是使用stream流进行操作，基于Http协议
	clientStream.WithServiceName(serviceName)
	clientStream.WithMethod(method)

	// execute the interceptor first
	//log.Println("invoke interceptor...", c.opts.interceptors)
	return interceptor.ClientIntercept(newCtx, req, rsp, c.opts.interceptors, c.invoke)
}

func (c *defaultClient) invoke(ctx context.Context, req, rsp interface{}) error {
	//此时的serialization是msgpack（序列化协议由Client客户端传入的决定）
	serialization := codec.GetSerialization(c.opts.serializationType)
	payload, err := serialization.Marshal(req)
	if err != nil {
		return codes.NewFrameworkError(codes.ClientMsgErrorCode, "request marshal failed ...")
	}
	//log.Println("request payload : ", payload) 用于编码使用
	//进行编码到网络层，然后网络在传输到目标主机
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
	// 先进行了序列化，然后再对其进行编码操作（序列化是将数据转化为二进制流，而编码则是将二进制流转化为特定的传输格式）

	// 接着便是进行传输层，即Client端的传输
	clientTransport := c.NewClientTransport()
	clientTransportOpts := []transport.ClientTransportOption{
		transport.WithServiceName(c.opts.serviceName),
		transport.WithClientTarget(c.opts.target),
		transport.WithClientNetwork(c.opts.network),
		transport.WithClientPool(connpool.GetPool("default")),
		transport.WithSelector(selector.GetSelector(c.opts.selectorName)),
		transport.WithTimeout(c.opts.timeout),
	}

	// clientTransport实现了Send方法
	frame, err := clientTransport.Send(ctx, reqbody, clientTransportOpts...)
	if err != nil {
		return err
	}

	rspbuf, err := clientCodec.Decode(frame)
	if err != nil {
		return err
	}

	// parse protocol header
	response := &protocol.Response{}
	if err = proto.Unmarshal(rspbuf, response); err != nil {
		return err
	}

	if response.RetCode != 0 {
		return codes.New(response.RetCode, response.RetMsg)
	}

	// return serialization.Unmarshal(response.Payload, rsp)
	// 在完成了解码后，最后一步便是反序列化，将二进制数据转化为对象
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
	//这个md是用来设置client的上下文ctx，用于传输数据（例如认证信息）
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
