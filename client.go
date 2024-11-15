package rpcdemo

import (
	"context"
	"errors"
	"fmt"
	"net"
	"reflect"
	"sync/atomic"
	"time"

	"HuaTug.com/message"
	"HuaTug.com/serialize"
	"github.com/silenceper/pool"
)

var messageId uint32 = 0

type Service interface {
	Name() string
}

type Client struct {
	coonPool   pool.Pool
	serializer serialize.Serializer
}

func NewClient(addr string, serializer serialize.Serializer) (*Client, error) {
	p, err := pool.NewChannelPool(&pool.Config{
		InitialCap: 10,
		MaxCap:     100,
		MaxIdle:    30,
		Factory: func() (interface{}, error) {
			return net.Dial("tcp", addr)
		}, // 这个pool缓存的内容是连接
		Close: func(i interface{}) error {
			return i.(net.Conn).Close()
		},
		Ping:        Ping, // 根据连接池源码的结构设计出来的函数，用来检测连接是否可用
		IdleTimeout: time.Minute,
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		coonPool:   p,
		serializer: serializer,
	}, nil

}

func (c *Client) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	obj, err := c.coonPool.Get()
	defer func() {
		c.coonPool.Put(obj)
	}()
	if err != nil {
		return nil, err
	}

	conn := obj.(net.Conn)
	data := message.EncodeReq(req)
	wLen, err := conn.Write(data)
	if err != nil {
		return nil, err
	}

	if wLen != len(data) { // 几乎遇不到，如果遇到了，有没有更好的处理方法？
		return nil, errors.New("rpc: 未写入全部数据")
	}

	respMsg, err := ReadMsg(conn)
	if err != nil {
		return nil, err
	}

	return message.DecodeResp(respMsg), nil
}

func (c *Client) Close() {
	c.coonPool.Release()
}

func (c *Client) InitStub(service Service) error {
	// 校验，确保它是一个指向结构体的指针
	if reflect.TypeOf(service).Kind() != reflect.PtrTo(reflect.TypeOf(reflect.Struct)).Kind() {
		return errors.New("service必须是指向结构体的指针")
	}

	val := reflect.ValueOf(service).Elem() //val是指向结构体的指针指向的值，也就是实际的值
	typ := reflect.TypeOf(service).Elem()
	//log.Print("初始化服务: ", val.NumField(), ",", typ)
	numField := val.NumField()
	for i := 0; i < numField; i++ {
		fieldValue := val.Field(i)
		fieldType := typ.Field(i)
		//log.Print("初始化服务2: ", fieldValue, ",", fieldType)
		if !fieldValue.CanSet() {
			continue
		}
		if fieldType.Type.Kind() != reflect.Func {
			continue
		}

		// 替换实现
		fn := reflect.MakeFunc(fieldType.Type, func(args []reflect.Value) (results []reflect.Value) {
			ctx, ok := args[0].Interface().(context.Context)
			if !ok {
				panic("The first request param must be context.Context!")
			}
			arg := args[1].Interface()

			outType := fieldType.Type.Out(0)
			//log.Print("序列化参数: ", outType)
			bs, err := c.serializer.Encode(arg)
			if err != nil {
				//reflect.Zero() 用于创建一个零值
				results = append(results, reflect.Zero(outType))
				results = append(results, reflect.ValueOf(err))
				return
			}
			//log.Print("调用服务: ", service.Name(), ".", fieldType.Name, "(", arg, ")")
			msgId := atomic.AddUint32(&messageId, 1)
			req := &message.Request{
				Compressor:  0, // 可扩展
				Serializer:  c.serializer.Code(),
				MessageId:   msgId, // 可扩展
				Version:     0,     // 可扩展
				BodyLength:  uint32(len(bs)),
				ServiceName: service.Name(),
				MethodName:  fieldType.Name,
				Data:        bs,
			}
			req.CalcHeadLength()
			res, err := c.Invoke(ctx, req)
			if err != nil {
				results = append(results, reflect.Zero(outType))
				results = append(results, reflect.ValueOf(err))
				return
			}

			first := reflect.New(outType.Elem()).Interface()
			if len(res.Data) > 0 {
				err = c.serializer.Decode(res.Data, first)
				if err != nil {
					results = append(results, reflect.Zero(outType))
					results = append(results, reflect.ValueOf(fmt.Sprintf("%s%v", "decode response body failed, err: ", err)))
					return
				}
			}

			results = append(results, reflect.ValueOf(first))
			// 第二个返回值，是error
			if len(res.Error) > 0 {
				results = append(results, reflect.ValueOf(errors.New(string(res.Error))))
			} else {
				results = append(results, reflect.Zero(reflect.TypeOf(new(error)).Elem()))
			}
			return
		})
		fieldValue.Set(fn)
	}
	return nil
}
