package rpcdemo

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"reflect"

	"HuaTug.com/message"
	"HuaTug.com/serialize"
	"HuaTug.com/serialize/json"
)

type Server struct {
	services    map[string]reflectionStub
	serializers []serialize.Serializer
}

type reflectionStub struct {
	value       reflect.Value
	serializers []serialize.Serializer
}

func NewServer() *Server {
	res := &Server{
		services:    map[string]reflectionStub{},
		serializers: make([]serialize.Serializer, 256),
	}

	res.RegisterSerializer(json.SerializerJson{}) // 默认支持json序列化

	return res
}

func (s *Server) MustRegister(service Service) {
	if err := s.RegisterService(service); err != nil {
		panic(err)
	}
}

func (s *Server) RegisterService(service Service) error {
	s.services[service.Name()] = reflectionStub{
		value:       reflect.ValueOf(service),
		serializers: s.serializers,
	}
	return nil
}
func (s *Server) RegisterSerializer(serializer serialize.Serializer) {
	s.serializers[serializer.Code()] = serializer
}

func (s *Server) Start(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			//log.Println("accpect connection failed,err: ", err)
			continue
		}
		//log.Println("receiver connection: ", conn.RemoteAddr(), "->", conn.LocalAddr())
		go func() {
			if err := s.HandleConn(conn); err != nil {
				//log.Println("handle connection failed,err: ", err)
				conn.Close()
				return
			}
		}()
	}
}

// RPC服务端来处理请求
func (s *Server) HandleConn(conn net.Conn) error {
	for {
		// 读请求
		// 执行
		// 写回响应

		reqMsg, err := ReadMsg(conn)
		if err != nil {
			return err
		}

		req := message.DecodeReq(reqMsg)
		//log.Println(req)
		resp := &message.Response{
			MessageId:  req.MessageId,
			Version:    req.Version,
			Compressor: req.Compressor,
			Serializer: req.Serializer,
		}
		// ping探活请求处理
		if req.Ping == PingPong {
			resp.Pong = PingPong
			resp.CalcHeadLength()
			_, err = conn.Write(message.EncodeResp(resp))
			if err != nil {
				return err
			}

			continue
		}
		log.Print("Name: ", req.ServiceName)
		// 找到本地对应的服务
		service, ok := s.services[req.ServiceName]
		if !ok {
			resp.Error = []byte("找不到对应服务")
			resp.CalcHeadLength()
			_, err = conn.Write(message.EncodeResp(resp))
			if err != nil {
				return err
			}

			continue
		}
		ctx := context.Background()
		//log.Print("执行本地服务:", req)
		data, err := service.invoke(ctx, req)
		if err != nil {
			resp.Error = []byte(err.Error())
			resp.CalcHeadLength()
			_, err = conn.Write(message.EncodeResp(resp))
			if err != nil {
				return err
			}

			continue
		}

		resp.Data = data
		resp.BodyLength = uint32(len(data))
		resp.CalcHeadLength()
		bitFlow := message.EncodeResp(resp)
		_, err = conn.Write(bitFlow)
		if err != nil {
			return err
		}

	}
}

func (s *reflectionStub) invoke(ctx context.Context, req *message.Request) ([]byte, error) {
	methodName := req.MethodName
	data := req.Data

	serializer := s.serializers[req.Serializer]
	if serializer == nil {
		return nil, errors.New("不支持的序列化协议")
	}

	// 具体来说，s.value.MethodByName(methodName) 会根据提供的 methodName 字符串查找并返回对应的方法。
	method := s.value.MethodByName(methodName)
	if !method.IsValid() {
		return nil, errors.New(fmt.Sprintf("%s%s", "服务下不存在该方法，方法名为：", methodName))
	}
	// method.Type() 是一个反射相关的方法，它用于获取一个 reflect.Value 类型的值所表示的方法的类型信息
	inType := method.Type().In(1)                  // 获取索引为1的参数的类型
	in := reflect.New(inType.Elem())               // 创建一个新的值，返回一个指向该类型(inType.Elem)的指针
	err := serializer.Decode(data, in.Interface()) //根据in的类型，将data反序列化为in
	if err != nil {
		return nil, err
	}
	res := method.Call([]reflect.Value{reflect.ValueOf(ctx), in})
	if len(res) > 1 && !res[1].IsZero() {
		return nil, res[1].Interface().(error)
	}

	return serializer.Encode(res[0].Interface())
}
