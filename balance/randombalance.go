package balance

import (
	"math/rand"
	"time"
)

type RandomBalance struct{}

func (r *RandomBalance) Select(servers []string) string {
	if len(servers) == 0 {
		return ""
	}
	rand.Seed(time.Now().UnixNano())
	return servers[rand.Intn(len(servers))]
}
