package main

import (
	"time"

	rpcdemo "HuaTug.com"
	"HuaTug.com/testdata"
)

func main() {
	opts := []rpcdemo.ServerOption{
		rpcdemo.WithAddress("localhost:8000"),
		rpcdemo.WithNetwork("tcp"),
		rpcdemo.WithSerializationType("msgpack"),
		rpcdemo.WithTimeout(time.Millisecond * 2000),
	}
	s := rpcdemo.NewServer(opts ...)
	if err := s.RegisterService("/test.Greeter", new(testdata.CalculatorService)); err != nil {
		panic(err)
	}
	s.Serve()
}
