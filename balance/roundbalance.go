package balance

import (
	"log"
	"sync"
)

// RoundRobinBalancer 轮询负载均衡器结构体，实现了LoadBalancer接口
type RoundRobinBalancer struct {
	mu         sync.Mutex
	instances  []string // 存储可用服务实例地址的列表
	currentIdx int      // 当前选择的实例索引
}

// AddInstance 添加服务实例到负载均衡器的实例列表中
func (r *RoundRobinBalancer) AddInstance(instance string) {
	r.mu.Lock()
	r.instances = append(r.instances, instance)
	r.mu.Unlock()
}

// Select 按照轮询策略从实例列表中选择服务实例，实现了LoadBalancer接口的Select方法
// 如果没有可用实例，返回错误
func (r *RoundRobinBalancer) Select(servers []string) string {
	r.mu.Lock()
	if len(r.instances) == 0 {
		r.mu.Unlock()
		return ""
	}
	instance := r.instances[r.currentIdx]
	log.Println("Index:", r.currentIdx)
	r.currentIdx = (r.currentIdx + 1) % len(r.instances)
	r.mu.Unlock()
	return instance
}
