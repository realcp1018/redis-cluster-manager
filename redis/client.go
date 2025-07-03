package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

func newRedisClient(hostPort string, password string) (*redis.Client, error) {
	opt := redis.Options{
		Addr:         hostPort,
		Password:     password,
		PoolSize:     3,
		MinIdleConns: 3,
		ReadTimeout:  3 * time.Second,
	}
	client := redis.NewClient(&opt)
	pingResult, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}
	if pingResult != "PONG" {
		return nil, fmt.Errorf("create redis client failed with a non-PONG response")
	}
	return client, nil
}
