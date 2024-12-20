package etcd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/HuaTug/My-RPC/plugin"
	"github.com/HuaTug/My-RPC/selector"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Etcd implements the server discovery specification
type Etcd struct {
	opts         *plugin.Options
	client       *clientv3.Client
	balancerName string // load balancing mode, including random, polling, weighted polling, consistent hash, etc
}

const Name = "etcd"

// plugin和jaeger都是有一个init函数，自己执行，注册插件
func init() {
	plugin.Register(Name, EtcdSvr)
	selector.RegisterSelector(Name, EtcdSvr)
}

// 通过Etcd实现了Plugin中的ResolvePlugin接口的方法，进而将EtcdSvr和ResolvePlugin联系了起来
// global etcd objects for framework
var EtcdSvr = &Etcd{
	opts: &plugin.Options{},
}

// InitConfig initializes the etcd client configuration and creates the client connection
func (e *Etcd) InitConfig() error {
	// etcd client配置
	config := clientv3.Config{
		Endpoints: []string{e.opts.SelectorSvrAddr},
	}
	client, err := clientv3.New(config)
	if err != nil {
		return err
	}
	e.client = client
	return nil
}

// Resolve resolves service names to a list of nodes from etcd
func (e *Etcd) Resolve(serviceName string) ([]*selector.Node, error) {
	ctx := context.Background()
	resp, err := e.client.Get(ctx, serviceName, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, fmt.Errorf("no services find in path : %s", serviceName)
	}
	var nodes []*selector.Node
	for _, kv := range resp.Kvs {
		nodes = append(nodes, &selector.Node{
			Key:   string(kv.Key),
			Value: kv.Value,
		})
	}
	return nodes, nil
}

// Select selects a service node based on the load balancing strategy
func (e *Etcd) Select(serviceName string) (string, error) {
	nodes, err := e.Resolve(serviceName)
	if nodes == nil || len(nodes) == 0 || err != nil {
		return "", err
	}
	balancer := selector.GetBalancer(e.balancerName)
	node := balancer.Balance(serviceName, nodes)
	if node == nil {
		return "", fmt.Errorf("no services find in %s", serviceName)
	}
	return parseAddrFromNode(node)
}

func parseAddrFromNode(node *selector.Node) (string, error) {
	if node.Key == "" {
		return "", fmt.Errorf("addr is empty")
	}
	strs := strings.Split(node.Key, "/")
	return strs[len(strs)-1], nil
}

// Init initializes the etcd configuration and registers services
func (e *Etcd) Init(opts ...plugin.Option) error {
	for _, o := range opts {
		o(e.opts)
	}
	log.Println(e.opts.Services)
	if len(e.opts.Services) == 0 || e.opts.SvrAddr == "" || e.opts.SelectorSvrAddr == "" {
		return fmt.Errorf("etcd init error, len(services) : %d, svrAddr : %s, selectorSvrAddr : %s",
			len(e.opts.Services), e.opts.SvrAddr, e.opts.SelectorSvrAddr)
	}
	if err := e.InitConfig(); err != nil {
		return err
	}
	for _, serviceName := range e.opts.Services {
		nodeName := fmt.Sprintf("%s/%s", serviceName, e.opts.SvrAddr)
		ctx := context.Background()
		// 注册服务
		_, err := e.client.Put(ctx, nodeName, e.opts.SvrAddr)
		if err != nil {
			return err
		}
	}
	return nil
}

// Init implements the initialization of the etcd configuration when the framework is loaded
func Init(etcdSvrAddr string, opts ...plugin.Option) error {
	for _, o := range opts {
		o(EtcdSvr.opts)
	}
	EtcdSvr.opts.SelectorSvrAddr = etcdSvrAddr
	return EtcdSvr.InitConfig()
}
