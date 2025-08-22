package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"math"
	"net/netip"
	"strconv"
	"strings"
)

type Instance struct {
	Addr           string
	Client         *redis.Client
	NodeID         string       // node ID if this is a cluster instance
	Role           string       // master or slave
	SlaveInit      bool         // master_sync_in_progress of a slave
	Master         string       // master addr if this is a slave, will be "" if this is a master
	MaxMemory      float64      // maxmemory in GB
	UsedMemory     float64      // used memory in GB
	MaxClients     int          // maximum number of clients allowed to connect to this instance
	ClientsCount   int          // number of clients connected to this instance
	ClusterEnabled bool         // true if this instance is part of a Redis Cluster
	Slots          []*SlotRange // list of SlotRange assigned to this instance
	KeysCount      string       // number of keys in this instance
	Version        string       // redis version
}

func NewInstance(hostPort string) (*Instance, error) {
	client, err := newRedisClient(hostPort)
	if err != nil {
		return nil, err
	}
	instance := &Instance{
		Addr:   hostPort,
		Client: client,
	}
	if err := instance.init(); err != nil {
		return nil, fmt.Errorf("failed to init instance %s: %v", hostPort, err)
	}
	return instance, nil
}

// init initializes the Redis instance by fetching its basic info
func (i *Instance) init() error {
	infoAllOutput, err := i.Client.Info(context.Background(), "all").Result()
	if err != nil {
		return err
	}
	infoMap := ParseInfo(infoAllOutput)
	i.Role = infoMap["role"]
	if i.Role == "slave" {
		i.Master = infoMap["master_host"] + ":" + infoMap["master_port"]
		if infoMap["master_sync_in_progress"] == "1" {
			i.SlaveInit = true
		}
	}

	maxMemoryBytes, _ := strconv.ParseFloat(infoMap["maxmemory"], 64)
	usedMemoryBytes, _ := strconv.ParseFloat(infoMap["used_memory"], 64)
	i.MaxMemory = math.Round(maxMemoryBytes/1024/1024/1024*100) / 100 // convert to GB and round to 2 digits
	i.UsedMemory = math.Round(usedMemoryBytes/1024/1024/1024*100) / 100

	i.ClientsCount, _ = strconv.Atoi(infoMap["connected_clients"])
	i.MaxClients, _ = strconv.Atoi(infoMap["maxclients"]) // if no maxclients it returns (0,error), we ignore this error

	clusterEnabled := infoMap["cluster_enabled"]
	if clusterEnabled == "1" {
		i.ClusterEnabled = true
	} else {
		i.ClusterEnabled = false
	}

	i.Version = infoMap["redis_version"]
	db0Info, exist := infoMap["db0"]
	if exist {
		i.KeysCount = strings.Split(strings.Split(db0Info, ",")[0], "=")[1]
	} else {
		i.KeysCount = "NaN"
	}
	return nil
}

// GetMasterSlaveMembers return members of a master-slave mechanism
// the first element of returned slice is the master addr
func (i *Instance) GetMasterSlaveMembers() ([]string, error) {
	if i.Role == "slave" {
		master, err := NewInstance(i.Master)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to master %s: %v", i.Master, err)
		}
		defer master.Close()
		if master.Role != "master" {
			return nil, fmt.Errorf("cascading replication is not supported")
		}
		return master.GetMasterSlaveMembers()
	}
	members := []string{i.Addr}
	infoReplOutput, err := i.Client.Info(context.Background(), "replication").Result()
	if err != nil {
		return nil, err
	}
	replInfo := ParseInfo(infoReplOutput)
	for k, v := range replInfo {
		if strings.HasPrefix(k, "slave") {
			slaveInfo := strings.Split(v, ",")
			slaveIpInfo, slavePortInfo := slaveInfo[0], slaveInfo[1]
			slaveIp := strings.Split(slaveIpInfo, "=")[1]
			slavePort := strings.Split(slavePortInfo, "=")[1]
			members = append(members, fmt.Sprintf("%s:%s", slaveIp, slavePort))
		}
	}
	return members, nil
}

// UpdateNodeIdAndSlots updates the NodeID and Slots of the instance using ParseClusterNodes output
func (i *Instance) UpdateNodeIdAndSlots(clusterNodesInfo [][]string) {
	for _, nodeInfo := range clusterNodesInfo {
		nodeID := nodeInfo[0]
		addr := nodeInfo[1]
		slotsStr := nodeInfo[2]
		if i.Addr == addr {
			i.NodeID = nodeID
			if i.Role == "master" {
				i.Slots = newSlotRanges(slotsStr)
			}
		}
	}
}

func (i *Instance) GetSlotCount() int {
	slotCount := 0
	for _, slotRange := range i.Slots {
		slotCount += slotRange.SlotCount
	}
	return slotCount

}

// StringSlots returns a string representation of the slots assigned to this instance: "[1] [2-100] [101-200]"
func (i *Instance) StringSlots() string {
	if i.Slots == nil {
		return "[]"
	}
	slotStr := ""
	for _, slotRange := range i.Slots {
		slotStr += slotRange.String() + ""
	}
	return strings.TrimRight(slotStr, " ")
}

func (i *Instance) Close() {
	i.Client.Close()
}

// Sorts Instances by Addr
type InstancesAscByAddr []*Instance

func (e InstancesAscByAddr) Len() int { return len(e) }

func (e InstancesAscByAddr) Less(i, j int) bool {
	addrPortI, _ := netip.ParseAddrPort(e[i].Addr)
	addrPortJ, _ := netip.ParseAddrPort(e[j].Addr)
	return addrPortI.Compare(addrPortJ) == -1
}

func (e InstancesAscByAddr) Swap(i, j int) { e[i], e[j] = e[j], e[i] }
