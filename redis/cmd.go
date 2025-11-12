package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strings"
)

// ParseInfo parses the Redis `info [section]` command output
func ParseInfo(client *redis.Client, section string) (map[string]string, error) {
	result := make(map[string]string)
	cmdOutput, err := client.Info(context.Background(), section).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to run info all: %v", err)
	}
	lines := strings.Split(cmdOutput, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "#") || len(strings.TrimSpace(line)) == 0 {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		result[parts[0]] = strings.TrimSpace(parts[1])
	}
	return result, nil
}

// ParseClusterNodes parses the Redis `cluster nodes` command output and returns a slice of slice(nodeID, addr, slots)
func ParseClusterNodes(client *redis.Client) ([][]string, error) {
	var nodes [][]string
	cmdOutput, err := client.ClusterNodes(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster nodes: %v", err)
	}
	lines := strings.Split(cmdOutput, "\n")
	for _, line := range lines {
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}
		parts := strings.Split(line, " ")
		nodeID := parts[0]
		addr := strings.Split(parts[1], "@")[0]
		slots := strings.Join(parts[8:], " ")
		nodes = append(nodes, []string{nodeID, addr, slots})
	}
	return nodes, nil
}

// ParseClientList parses the Redis `client list` command output and returns []map[string]string
func ParseClientList(client *redis.Client) ([]map[string]string, error) {
	var results []map[string]string
	cmdOutput, err := client.ClientList(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get client list: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(cmdOutput), "\n")
	for _, line := range lines {
		fields := strings.Split(line, " ")
		m := make(map[string]string)
		for _, field := range fields {
			parts := strings.SplitN(field, "=", 2)
			if len(parts) == 2 {
				m[parts[0]] = parts[1]
			}
		}
		results = append(results, m)
	}
	return results, nil
}

// ParseConfigGet: config get goes well even parameter is not exist, so no error check here
func ParseConfigGet(client *redis.Client, parameter string) string {
	cmdOutput, err := client.ConfigGet(context.Background(), parameter).Result()
	if err != nil || len(cmdOutput) == 0 {
		return ""
	}
	return cmdOutput[parameter]
}
