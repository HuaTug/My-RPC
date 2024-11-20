package main

import (
	"context"
	"errors"
	"net/http"
	"time"

	rpcdemo "HuaTug.com"
	"HuaTug.com/auth"
	"HuaTug.com/metadata"
	"HuaTug.com/plugin/etcd"
	"HuaTug.com/plugin/jaeger"
	"HuaTug.com/testdata"
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
		rpcdemo.WithSelectorSvrAddr("localhost:2379"),
		rpcdemo.WithPlugin(etcd.Name),
		rpcdemo.WithTracingSvrAddr("localhost:6831"),
		rpcdemo.WithTracingSpanName("test.Greeter"),
		rpcdemo.WithPlugin(jaeger.Name),
	}
	s := rpcdemo.NewServer(opts...)
	if err := s.RegisterService("/test.Greeter", new(testdata.CalculatorService)); err != nil {
		panic(err)
	}
	s.Serve()
}

func pprof() {
	go func() {
		http.ListenAndServe("localhost:6060", http.DefaultServeMux)
	}()
}
