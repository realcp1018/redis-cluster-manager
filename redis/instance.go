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

func NewInstance(hostPort, password string) (*Instance, error) {
	client, err := newRedisClient(hostPort, password)
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
// we need to convert some of the values thus we cannot use json.Unmarshal here
func (i *Instance) init() error {
	info, err := i.Client.Info(context.Background(), "all").Result()
	if err != nil {
		return err
	}
	infoMap := ParseInfoAll(info)
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
	if i.Slots == nil {
		return 0
	}
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

type SlotRange struct {
	Start     int // start of the slot range
	End       int // end of the slot range
	SlotCount int
}

// newSlotRanges creates a slice of SlotRange from a string like "1 2-100 101-200"
func newSlotRanges(slotStr string) []*SlotRange {
	var slotRanges []*SlotRange
	slots := strings.Split(slotStr, " ")
	if len(slots) == 0 {
		return nil
	}
	for _, slot := range slots {
		if strings.Contains(slot, "-") {
			// it's a range
			parts := strings.Split(slot, "-")
			start, _ := strconv.Atoi(parts[0])
			end, _ := strconv.Atoi(parts[1])
			slotRanges = append(slotRanges, &SlotRange{
				Start:     start,
				End:       end,
				SlotCount: end - start + 1,
			})
		} else {
			// it's a single slot
			slotNum, _ := strconv.Atoi(slot)
			slotRanges = append(slotRanges, &SlotRange{
				Start:     slotNum,
				End:       slotNum,
				SlotCount: 1,
			})
		}
	}
	return slotRanges
}

func (s *SlotRange) ContainsSlot(slot int) bool {
	return slot >= s.Start && slot <= s.End
}

func (s *SlotRange) String() string {
	return fmt.Sprintf("[%d-%d]", s.Start, s.End)
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
