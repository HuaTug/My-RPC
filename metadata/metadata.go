package metadata

import "context"

type clientMD struct{}
type serverMD struct{}

type clientMetadata map[string][]byte
type serverMetadata map[string][]byte

// ClientMetadata creates a new context with key-value pairs attached.
func ClientMetadata(ctx context.Context) clientMetadata {
	if md, ok := ctx.Value(clientMD{}).(clientMetadata); ok {
		return md
	}
	md := make(map[string][]byte)
	WithClientMetadata(ctx, md)
	return md
}

// WithClientMetadata creates a new context with the specified metadata
func WithClientMetadata(ctx context.Context, metadata map[string][]byte) context.Context {
	return context.WithValue(ctx, clientMD{}, clientMetadata(metadata))
}

// ServerMetadata creates a new context with key-value pairs attached.
func ServerMetadata(ctx context.Context) serverMetadata {
	if md, ok := ctx.Value(serverMD{}).(serverMetadata); ok {
		return md
	}
	md := make(map[string][]byte)
	WithServerMetadata(ctx, md)
	return md
}

// WithServerMetadata creates a new context with the specified metadata
func WithServerMetadata(ctx context.Context, metadata map[string][]byte) context.Context {
	return context.WithValue(ctx, serverMD{}, serverMetadata(metadata))
}


/*
这段代码定义在 metadata 包中，主要围绕着处理在 context.Context 类型上下文中添加、获取和管理元数据（metadata，以键值对的形式存在，键为字符串，值为字节切片）的功能，分别针对客户端和服务器端的场景进行了相应操作的封装，方便在基于 context 的应用中传递和使用额外的相关信息。
*/