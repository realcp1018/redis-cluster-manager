package redis

import (
	"context"
	"fmt"
	"testing"
)

func Test_Redis(t *testing.T) {
	client, err := newRedisClient("127.0.0.1:6379", "redis")
	if err != nil {
		panic(err)
	}
	infoAllOutput, err := client.Info(context.Background(), "all").Result()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Redis Info: %v\n", ParseInfo(infoAllOutput))

	slotsInfo, err := client.ClusterSlots(context.Background()).Result()
	fmt.Printf("Redis Cluster Slots: %v\n", slotsInfo)
}
