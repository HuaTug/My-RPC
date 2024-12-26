package main

import (
	"context"
	"errors"
	"net/http"
	"time"

	rpcdemo "github.com/HuaTug/My-RPC"
	"github.com/HuaTug/My-RPC/auth"
	"github.com/HuaTug/My-RPC/metadata"
	"github.com/HuaTug/My-RPC/plugin/consul"
	"github.com/HuaTug/My-RPC/plugin/jaeger"
	"github.com/HuaTug/My-RPC/testdata"
)

func main() {
	af := func(ctx context.Context) (context.Context, error) {
		md := metadata.ServerMetadata(ctx)

		if len(md) == 0 {
			return ctx, errors.New("token nil")
		}
		v := md["authorization"]
		if string(v) != "Bearer token" {
			return ctx, errors.New("token invalid")
		}
		return ctx, nil
	}

	pprof()

	opts := []rpcdemo.ServerOption{
		rpcdemo.WithAddress("localhost:8000"),
		rpcdemo.WithNetwork("tcp"),
		rpcdemo.WithSerializationType("msgpack"),
		rpcdemo.WithTimeout(time.Millisecond * 2000),
		rpcdemo.WithInterceptor(auth.BuildAuthInterceptor(af)),
		rpcdemo.WithSelectorSvrAddr("localhost:8500"),
		rpcdemo.WithPlugin(consul.Name, jaeger.Name),
		rpcdemo.WithTracingSvrAddr("localhost:6831"),
		rpcdemo.WithTracingSpanName("test.Greeter"),
	}
	s := rpcdemo.NewServer(opts...)
	if err := s.RegisterService("test.Greeter", new(testdata.CalculatorService)); err != nil {
		panic(err)
	}
	s.Serve()
}

func pprof() {
	go func() {
		http.ListenAndServe("localhost:6060", http.DefaultServeMux)
	}()
}
