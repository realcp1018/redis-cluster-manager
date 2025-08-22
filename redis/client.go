package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"redis-cluster-manager/vars"
	"strings"
)

func newRedisClient(hostPort string) (*redis.Client, error) {
	opt := redis.Options{
		Addr:         hostPort,
		Password:     vars.Password,
		PoolSize:     3,
		MinIdleConns: 3,
		DialTimeout:  vars.Timeout,
		ReadTimeout:  vars.Timeout,
		WriteTimeout: vars.Timeout,
	}
	client := redis.NewClient(&opt)
	pingResult, err := client.Ping(context.Background()).Result()
	if err != nil {
		if strings.Contains(err.Error(), "LOADING Redis is loading the dataset in memory") {
			return client, nil
		}
		return nil, fmt.Errorf("create redis client failed with error: %v", err)
	}
	if pingResult != "PONG" {
		return nil, fmt.Errorf("create redis client failed with a non-PONG response: [%s]", pingResult)
	}
	return client, nil
}
