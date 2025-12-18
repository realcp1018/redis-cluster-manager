package redis

import (
	"fmt"
	"redis-cluster-manager/vars"
	"testing"
)

func TestCmd(t *testing.T) {
	vars.Password = "redis"
	client, err := newRedisClient("127.0.0.1:6379")
	if err != nil {
		fmt.Println(err)
		return
	}
	clusterInfo, err := ParseClusterInfo(client)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(clusterInfo["cluster_state"])
}
