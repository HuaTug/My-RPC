package main

import (
	rpcdemo "HuaTug.com"
	"HuaTug.com/serialize/json"
	"HuaTug.com/serialize/protobuf"
)

func main() {
	srv := rpcdemo.NewServer()

	// 注册服务
	srv.MustRegister(&UserService{})
	srv.MustRegister(&UserParentService{})

	// 注册server支持的序列化协议
	srv.RegisterSerializer(json.SerializerJson{})
	srv.RegisterSerializer(protobuf.SerializeProto{})

	if err := srv.Start("localhost:8080"); err != nil {
		panic(err)
	}
}
