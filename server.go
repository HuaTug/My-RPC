package rpcdemo

import (
	"context"
	"fmt"
	logs "log"
	"os"
	"os/signal"
	"reflect"
	"syscall"

	"github.com/HuaTug/My-RPC/interceptor"
	"github.com/HuaTug/My-RPC/log"
	"github.com/HuaTug/My-RPC/plugin"
	"github.com/HuaTug/My-RPC/plugin/jaeger"
)

type Server struct {
	opts    *ServerOptions
	service Service
	plugins []plugin.Plugin
	closing bool
}

func NewServer(opts ...ServerOption) *Server {
	s := &Server{
		opts: &ServerOptions{},
	}
	for _, o := range opts {
		o(s.opts)
	}

	s.service = NewService(s.opts)

	// add plugin
	for pluginName, plugin := range plugin.PluginMap {
		if !containPlugin(pluginName, s.opts.pluginNames) {
			continue
		}
		s.plugins = append(s.plugins, plugin)
	}

	return s
}

func NewService(opts *ServerOptions) Service {
	return &service{
		opts: opts,
	}
}

func containPlugin(pluginName string, plugins []string) bool {
	for _, plugin := range plugins {
		if plugin == pluginName {
			return true
		}
	}
	return false
}

type emptyInterface interface{}

/*
假设在未来，我们决定为服务添加特定的接口要求，例如所有服务都必须实现一个 DoSomething 方法。

	type ServiceInterface interface {
	    DoSomething(ctx context.Context, req *RequestType) (*ResponseType, error)
	}

此时，我们可以修改 ServiceDesc 中的 HandlerType 字段，使其不再是空接口，而是我们的 ServiceInterface：

	type ServiceDesc struct {
	    ServiceName string
	    HandlerType ServiceInterface // 修改为具体的接口
	    Svr         interface{}
	}

*/
// 服务端注册服务

func (s *Server) RegisterService(serviceName string, svr interface{}) error {
	svrType := reflect.TypeOf(svr)
	srvValue := reflect.ValueOf(svr)

	logs.Println("svrType is: ", svrType)
	logs.Println("svrValue is: ", srvValue)

	sd := &ServiceDesc{
		ServiceName: serviceName,
		HandlerType: (*emptyInterface)(nil),
		Svr:         svr,
	}

	methods, err := getServiceMethods(svrType, srvValue)
	if err != nil {
		return err
	}
	//logs.Println("methods is: ", methods[1].MethodName)
	sd.Methods = methods

	logs.Printf("register service: %s", serviceName)
	s.Register(sd, svr)

	return nil
}

func getServiceMethods(serviceType reflect.Type, servieValue reflect.Value) ([]*MethodDesc, error) {

	var methods []*MethodDesc

	for i := 0; i < serviceType.NumMethod(); i++ {
		method := serviceType.Method(i)

		if err := checkMethod(method.Type); err != nil {
			return nil, err
		}

		//这个函数被封装后，只有等到客户端去调用，才会执行
		methodHandler := func(ctx context.Context, svr interface{}, dec func(interface{}) error, ceps []interceptor.ServerInterceptor) (interface{}, error) {
			reqType := method.Type.In(2)

			req := reflect.New(reqType.Elem()).Interface()
			if err := dec(req); err != nil {
				return nil, err
			}

			if len(ceps) == 0 {
				//通过method.Func.Call完成了对方法的调用,其中Call的参数列表按照方法的参数列表顺序，以及类型填写
				values := method.Func.Call([]reflect.Value{servieValue, reflect.ValueOf(ctx), reflect.ValueOf(req)})
				return values[0].Interface(), nil
			}

			// 执行拦截器
			handler := func(ctx context.Context, reqbody interface{}) (interface{}, error) {
				values := method.Func.Call([]reflect.Value{servieValue, reflect.ValueOf(ctx), reflect.ValueOf(req)})

				return values[0].Interface(), nil
			}
			return interceptor.ServerIntercept(ctx, req, ceps, handler)
		}

		methods = append(methods, &MethodDesc{
			MethodName: method.Name,
			// ToDo: 精彩 通过映射map存储handler处理方法
			Handler: methodHandler,
		})
	}

	return methods, nil
}

func checkMethod(method reflect.Type) error {

	// 要保证有两个自己给的参数，外加一个自己的参数 个数>=3
	if method.NumIn() < 3 {
		return fmt.Errorf("method %s invalid,the number of params < 2", method.Name())
	}

	if method.NumOut() != 2 {
		return fmt.Errorf("method %s invalid, the number of return values != 2", method.Name())
	}

	ctxType := method.In(1)
	var contextType = reflect.TypeOf((*context.Context)(nil)).Elem()
	if !ctxType.Implements(contextType) {
		return fmt.Errorf("method %s invalid, the first param is not context.Context", method.Name())
	}

	argType := method.In(2)
	if argType.Kind() != reflect.Ptr {
		return fmt.Errorf("method %s invalid, req type is not a pointer", method.Name())
	}

	replyType := method.Out(0)
	if replyType.Kind() != reflect.Ptr {
		return fmt.Errorf("method %s invalid, reply type is not a pointer", method.Name())
	}

	errType := method.Out(1)
	var errorType = reflect.TypeOf((*error)(nil)).Elem()
	if !errType.Implements(errorType) {
		return fmt.Errorf("method %s invalid, returns %s , not error", method.Name(), errType.Name())
	}

	return nil
}

func (s *Server) Register(sd *ServiceDesc, svr interface{}) {
	if sd == nil || svr == nil {
		return
	}

	ht := reflect.TypeOf(sd.HandlerType).Elem()
	st := reflect.TypeOf(svr)
	logs.Print("st is:", st)
	logs.Print("ht is:", ht)
	if !st.Implements(ht) {
		log.Fatalf("handlerType %v not match service : %v ", ht, st)
	}

	ser := &service{
		svr:         svr,
		serviceName: sd.ServiceName,
		handlers:    make(map[string]Handler),
	}

	for _, method := range sd.Methods {
		// logs.Println("method name: ",method.MethodName)
		// logs.Println("method handler: ",method.Handler)
		ser.handlers[method.MethodName] = method.Handler
	}

	s.service = ser
}

func (s *Server) Serve() {
	err := s.InitPlugins()
	if err != nil {
		panic(err)
	}

	s.service.Serve(s.opts)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGSEGV)
	<-ch

	s.Close()
}

type emptyServicee struct{}

func (s *Server) ServerHttp() {
	if err := s.RegisterService("/http", new(emptyServicee)); err != nil {
		panic(err)
	}

	s.Serve()
}

func (s *Server) Close() {
	s.closing = true
	s.service.Close()
}
func (s *Server) InitPlugins() error {
	for _, p := range s.plugins {

		switch val := p.(type) {
		case plugin.ResolverPlugin:
			var services []string
			// logs.Println("plugin is ",s.service.Name())
			// logs.Println("selector is ",s.opts.selectorSvrAddr)
			services = append(services, s.service.Name())
			pluginOpts := []plugin.Option{
				plugin.WithSelectorSvrAddr(s.opts.selectorSvrAddr),
				plugin.WithSvrAddr(s.opts.address),
				plugin.WithServices(services),
			}
			if err := val.Init(pluginOpts...); err != nil {
				log.Errorf("resolver init error: %v", err)
				return err
			}
		case plugin.TracingPlugin:
			pluginOpts := []plugin.Option{
				plugin.WithTracingSvrAddr(s.opts.tracingSvrAddr),
			}

			tracer, err := val.Init(pluginOpts...)
			if err != nil {
				log.Errorf("tracing init error: %v", err)
				return err
			}

			s.opts.interceptors = append(s.opts.interceptors, jaeger.OpenTracingServerInterceptor(tracer, s.opts.tracingSpanName))
		default:

		}
	}
	return nil
}
