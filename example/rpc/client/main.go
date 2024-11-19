package main

import (
	"context"
	"fmt"
	"time"

	"HuaTug.com/client"
	"HuaTug.com/testdata"
)

func main() {
	opts := []client.Option{
		client.WithTarget("localhost:8000"),
		client.WithNetwork("tcp"),
		client.WithTimeout(2000 * time.Millisecond),
		client.WithSerializationType("msgpack"),
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
	err := c.Call(context.Background(), "/test.Greeter/Calculate", req, rsp, opts...)
	fmt.Print(rsp.Result, err)

}
