package redis

import (
	"strings"
)

// cmd.go: contains methods used to parse redis cmd string outputs

// ParseInfoAll parses the Redis `info all` command output
func ParseInfoAll(cmdOutput string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(cmdOutput, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "#") || len(strings.TrimSpace(line)) == 0 {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		result[parts[0]] = strings.TrimSpace(parts[1])
	}
	return result
}

// ParseClusterNodes parses the Redis `cluster nodes` command output and returns a slice of slice(nodeID, addr, slots)
func ParseClusterNodes(cmdOutput string) [][]string {
	var nodes [][]string
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
	return nodes
}
