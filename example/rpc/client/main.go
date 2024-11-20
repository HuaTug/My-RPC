package main

import (
	"context"
	"fmt"
	"time"

	"HuaTug.com/auth"
	"HuaTug.com/client"
	"HuaTug.com/plugin/etcd"
	"HuaTug.com/plugin/jaeger"
	"HuaTug.com/testdata"
)

func main() {

	tracer, err := jaeger.Init("localhost:6831")
	if err != nil {
		panic(err)
	}

	opts := []client.Option{
		//client.WithTarget("localhost:8000"),
		client.WithNetwork("tcp"),
		client.WithTimeout(2000 * time.Millisecond),
		client.WithSerializationType("msgpack"),
		client.WithPerRPCAuth(auth.NewOAuth2ByToken("token")),
		client.WithSelectorName(etcd.Name),
		client.WithInterceptor(jaeger.OpenTracingClientInterceptor(tracer, "/test.Greeter/Calculate")),
	}

	c := client.DefaultClient
	// req := &testdata.CalculateRequest{
	// 	Operation: "subtract",
	// 	Num1:      1,
	// 	Num2:      2,
	// }
	req := &testdata.CalculateRequest{
		Operation: "multiply",
		Num1:      3.5,
		Num2:      6,
	}
	rsp := &testdata.CalculateReply{}
	etcd.Init("localhost:2379")
	err = c.Call(context.Background(), "/test.Greeter/Calculate", req, rsp, opts...)
	fmt.Println(rsp.Result, err)
}
