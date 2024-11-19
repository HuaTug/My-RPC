package balance

type LoadBalancer interface {
	Select(servers []string) string
}


