package rpcdemo

import (
	"context"

	"HuaTug.com/message"
)

type Proxy interface {
	//按照协议调用服务 call(ctx,req) (resp,err)
	Invoke(ctx context.Context, req *message.Request) (*message.Response, error)
}
