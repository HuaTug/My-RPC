package transport

import (
	"context"
	"log"

	"github.com/HuaTug/My-RPC/codes"
)

type clientTransport struct {
	opts *ClientTransportOptions
}

var clientTransportMap = make(map[string]ClientTransport)

func init() {
	clientTransportMap["default"] = DefaultClientTransport
}

// RegisterClientTransport supports business custom registered ClientTransport
func RegisterClientTransport(name string, clientTransport ClientTransport) {
	if clientTransportMap == nil {
		clientTransportMap = make(map[string]ClientTransport)
	}
	clientTransportMap[name] = clientTransport
}

// Get the ServerTransport
func GetClientTransport(transport string) ClientTransport {

	if v, ok := clientTransportMap[transport]; ok {
		return v
	}

	return DefaultClientTransport
}

// The default ClientTransport
var DefaultClientTransport = New()

// Use the singleton pattern to create a ClientTransport
var New = func() ClientTransport {
	return &clientTransport{
		opts: &ClientTransportOptions{},
	}
}

func (c *clientTransport) Send(ctx context.Context, req []byte, opts ...ClientTransportOption) ([]byte, error) {

	for _, o := range opts {
		o(c.opts)
	}

	if c.opts.Network == "tcp" {
		return c.SendTcpReq(ctx, req)
	}

	if c.opts.Network == "udp" {
		return c.SendUdpReq(ctx, req)
	}

	return nil, codes.NetworkNotSupportedError
}

func (c *clientTransport) SendTcpReq(ctx context.Context, req []byte) ([]byte, error) {

	// service discovery
	// 这里的c.opts.ServiceName表示为客户端的服务，即客户端可以发送想要调用的服务（服务名）
	log.Println("SendTcpReq service_name: ", c.opts.ServiceName)
	addr, err := c.opts.Selector.Select(c.opts.ServiceName)
	log.Println("Select the addr is :", addr)
	if err != nil {
		return nil, err
	}

	// defaultSelector returns "", use the target as address
	if addr == "" {
		addr = c.opts.Target
	}

	// 表示为从连接池中获取连接
	conn, err := c.opts.Pool.Get(ctx, c.opts.Network, addr)
	//	conn, err := net.DialTimeout("tcp", addr, c.opts.Timeout);
	if err != nil {
		return nil, err
	}

	defer conn.Close()
	// 模仿gRPC 发送数据(分块发送数据)
	sendNum := 0
	num := 0
	for sendNum < len(req) {
		num, err = conn.Write(req[sendNum:])
		if err != nil {
			return nil, err
		}
		sendNum += num

		if err = isDone(ctx); err != nil {
			return nil, err
		}
	}
	// parse frame
	wrapperConn := wrapConn(conn)
	// ReadFrame is for checking the frame header
	frame, err := wrapperConn.framer.ReadFrame(conn)
	if err != nil {
		return nil, err
	}
	return frame, err
}

// isDone 判断是否超时或者被异常中断
func isDone(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return nil
}
