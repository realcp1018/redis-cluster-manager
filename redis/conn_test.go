package redis

import (
	"fmt"
	"redis-cluster-manager/vars"
	"testing"
)

func TestNewConnection(t *testing.T) {
	vars.Password = "redis"
	client, err := newRedisClient("127.0.0.1:6379")
	if err != nil {
		t.Skipf("skip live Redis test: %v", err)
	}
	connections, err := NewConnections(client)
	if err != nil {
		t.Fatalf("NewConnections() error = %v", err)
	}
	for _, conn := range connections {
		fmt.Println(conn.Id, conn.Addr, conn.Flags, conn.Cmd)
	}
}
