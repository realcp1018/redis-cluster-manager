package redis

import (
	"context"
	"fmt"
	"testing"
)

func TestRedis(t *testing.T) {
	client, err := newRedisClient("127.0.0.1:6379")
	if err != nil {
		t.Skipf("skip live Redis test: %v", err)
	}
	infoAllResult, err := ParseInfo(client, "all")
	if err != nil {
		t.Fatalf("ParseInfo() error = %v", err)
	}
	fmt.Printf("Redis Info: %v\n", infoAllResult)

	slotsInfo, err := client.ClusterSlots(context.Background()).Result()
	fmt.Printf("Redis Cluster Slots: %v\n", slotsInfo)
}
