package redis

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/redis/go-redis/v9"
	"strconv"
	"strings"
)

// Connection: redis: `client list` command parser
type Connection struct {
	Id    string `json:"id"`
	Addr  string `json:"addr,omitempty"`
	Name  string `json:"name,omitempty"`
	Flags string `json:"flags"`
	Cmd   string `json:"cmd,omitempty"`
}

func NewConnections(client *redis.Client) ([]*Connection, error) {
	clientList, err := ParseClientList(client)
	if err != nil {
		return nil, err
	}
	var connections []*Connection
	for _, sessionInfo := range clientList {
		var conn Connection
		if err := mapstructure.Decode(sessionInfo, &conn); err != nil {
			return nil, fmt.Errorf("failed to decode client list to Connection: %v", err)
		}
		connections = append(connections, &conn)
	}
	return connections, nil
}

func (c *Connection) GetHostPort() (string, int, error) {
	hostPort := strings.Split(c.Addr, ":")
	if len(hostPort) != 2 {
		return "", 0, fmt.Errorf("invalid addr format: %s", c.Addr)
	}
	host := hostPort[0]
	port, err := strconv.Atoi(hostPort[1])
	if err != nil {
		return "", 0, fmt.Errorf("invalid port format: %s", hostPort[1])
	}
	return host, port, nil
}
