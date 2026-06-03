package cluster

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"redis-cluster-manager/perf"
	r "redis-cluster-manager/redis"
	"redis-cluster-manager/vars"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

var (
	showSlots bool // whether to show slots info or not
)

var StatusCmd = &cobra.Command{
	Use:     "status",
	Short:   "Show cluster status",
	Long:    `Show cluster status`,
	Args:    cobra.ExactArgs(1),
	Example: fmt.Sprintf("%s cluster status <seed-node> -a \"password\"", vars.AppName),
	RunE: func(cmd *cobra.Command, args []string) error {
		vars.HostPort = args[0]
		f := perf.StartCpuProfile()
		defer perf.StopCpuProfile(f)
		err := printClusterStatus(vars.HostPort)
		if err != nil {
			return err
		}
		perf.MemProfile()
		return nil
	},
}

func InitStatus() {
	StatusCmd.Flags().BoolVarP(&showSlots, "show-slots", "s", false, "Show slots info or not, default false")
}

// printClusterStatus
// if cluster is a sharding cluster, it shows sharding cluster status
// if cluster is a master-slave/sentinel cluster, it calls printMasterSlaveStatus
func printClusterStatus(hostPort string) error {
	// we call the provided node as `the seed node`(vars.HostPort above)
	seedNode, err := r.NewInstance(hostPort)
	if err != nil {
		return err
	}
	defer seedNode.Close()
	if !seedNode.ClusterEnabled {
		return printMasterSlaveStatus(seedNode)
	}
	clusterInfo, err := r.ParseClusterInfo(seedNode.Client)
	if err != nil {
		return err
	}
	clusterState := clusterInfo["cluster_state"]
	if clusterState == "fail" {
		return fmt.Errorf("seed node cluster mode ON, but it's cluster state is fail, might be a orphaned node")
	}
	// get cluster instances by running `cluster nodes` on seed node
	clusterNodesInfo, err := r.ParseClusterNodes(seedNode.Client)
	if err != nil {
		return err
	}
	// get cluster instances simultaneously
	var (
		clusterInstances []*r.Instance // all instances that can be connected
		displayedSlaves  []*r.Instance // slaves that have alive master
		OrphanedSlaves   []*r.Instance // slaves that have no alive master, clusterInstances - displayedSlaves
		mu               sync.Mutex
		wg               sync.WaitGroup
		warnings         sync.Map     // instances can not be created err messages
		errInstanceCount atomic.Int32 // instances can not be created counter
		slotsCount       int          // slots count of all masters
	)
	for _, nodeInfo := range clusterNodesInfo {
		nodeId := nodeInfo[0]
		addr := nodeInfo[1]
		wg.Add(1)
		go func(addr, nodeId string) {
			defer wg.Done()
			if i, err := r.NewInstance(addr); err != nil {
				warnings.Store(fmt.Sprintf("%s,%s", addr, nodeId), err)
				errInstanceCount.Add(1)
				return
			} else {
				i.UpdateNodeClusterInfo(clusterNodesInfo)
				mu.Lock()
				clusterInstances = append(clusterInstances, i)
				mu.Unlock()
			}
		}(addr, nodeId)
	}
	wg.Wait()
	// Print Cluster Basic Info
	fmt.Println(strings.Repeat("=", 155))
	fmt.Printf("%-16s:\t%s\n", "Cluster Version", seedNode.Version)
	fmt.Println(strings.Repeat("=", 155))
	// Print Node Banner
	color.Cyan("%-45s%-24s%-16s%-16s%-16s%-16s%-12s%s\n", "NodeID", "Address", "Role", "Memory(GB)",
		"KeysCount", "Clients", "Slots", "SlotRanges")
	fmt.Printf("%-45s%-24s%-16s%-16s%-16s%-16s%-12s%s\n", "------", "-------", "----", "----------",
		"---------", "-------", "-----", "----------")
	// get all masters
	var clusterMasters []*r.Instance
	for _, i := range clusterInstances {
		if i.Role == "master" {
			clusterMasters = append(clusterMasters, i)
			for _, slot := range i.Slots {
				slotsCount += slot.SlotCount
			}
		}
	}
	sortedMasters := r.InstancesAscByAddr(clusterMasters)
	sort.Sort(sortedMasters)
	for _, m := range sortedMasters {
		// print master info
		fmt.Print(color.RedString("%-45s", m.NodeID))
		fmt.Print(color.RedString("%-24s", m.Addr))
		fmt.Printf("%-16s", formatRole(m, false))
		fmt.Printf("%-16s", formatMemory(m))
		fmt.Printf("%-16s", formatKeysCount(m))
		fmt.Printf("%-16s", formatClients(m))
		fmt.Printf("%-12d", m.GetSlotCount())
		if showSlots {
			fmt.Printf("%s\n", m.StringSlots())
		} else {
			fmt.Print("...\n")
		}
		// print slaves info: get slaves and sort them then print them
		var slaves []*r.Instance
		for _, i := range clusterInstances {
			if i.Master == m.Addr {
				slaves = append(slaves, i)
				displayedSlaves = append(displayedSlaves, i)
			}
		}
		sortedSlaves := r.InstancesAscByAddr(slaves)
		sort.Sort(sortedSlaves)
		for _, s := range sortedSlaves {
			fmt.Printf("%-45s", s.NodeID)
			fmt.Printf("%-24s", s.Addr)
			fmt.Printf("%-16s", formatRole(s, true))
			fmt.Printf("%-16s", formatMemory(s))
			fmt.Printf("%-16s", formatKeysCount(s))
			fmt.Printf("%-16s", formatClients(s))
			// skip slot info for slave
			fmt.Printf("%-12s", "")
			fmt.Printf("%s\n", "")
		}
	}
	// when a master's connection count reaches maxclients, no new connections are allowed,
	// that means it is alive, but we can not create an instance for it, so it seems it's slaves are orphaned.
	// we need to display these orphaned slaves in the end
	for _, i := range clusterInstances {
		if i.Role == "slave" {
			slaveDisplayed := false
			for _, ds := range displayedSlaves {
				if i.Addr == ds.Addr {
					slaveDisplayed = true
					break
				}
			}
			if !slaveDisplayed {
				OrphanedSlaves = append(OrphanedSlaves, i)
			}
		}
	}
	if len(OrphanedSlaves) > 0 {
		color.Red("Orphaned Slaves (whose masters can not connected, see Warnings below):")
		for _, os := range OrphanedSlaves {
			fmt.Printf("%-45s", os.NodeID)
			fmt.Printf("%-24s", os.Addr)
			fmt.Printf("%-16s", formatRole(os, true))
			fmt.Printf("%-16s", formatMemory(os))
			fmt.Printf("%-16s", formatKeysCount(os))
			fmt.Printf("%-16s", formatClients(os))
			// skip slot info for slave
			fmt.Printf("%-12s", "")
			fmt.Printf("%s\n", "")
		}
	}
	for _, i := range clusterInstances {
		i.Close()
	}
	color.Cyan("Total up masters in cluster: %d\n", len(clusterMasters))
	color.Cyan("Total up members in cluster: %d\n", len(clusterInstances))
	if errInstanceCount.Load() != 0 {
		color.Cyan("Warnings:")
		warnings.Range(func(nodeInfo, err interface{}) bool {
			n := strings.Split(nodeInfo.(string), ",")
			color.Red("failed to create instance for node [addr=%s] [node_id=%s], error: %v\n", n[0], n[1], err)
			return true
		})
		color.Cyan("Error nodes in cluster: %d\n", errInstanceCount.Load())
	}
	if slotsCount != 16384 {
		color.Red("Master slot count is not 16384(%d). Some slots missing or migrating. Please check your cluster status.", slotsCount)
	}
	return nil
}

func formatRole(i *r.Instance, slavePrefix bool) string {
	role := i.Role
	if role == "" {
		role = "unknown"
	}
	if i.LoadingError || i.SlaveInit {
		role += "(init)"
	}
	if slavePrefix {
		return "-" + role
	}
	return role
}

func formatMemory(i *r.Instance) string {
	if i.LoadingError {
		return "-"
	}
	return fmt.Sprintf("%.2f/%.2f", i.UsedMemory, i.MaxMemory)
}

func formatKeysCount(i *r.Instance) string {
	if i.LoadingError || i.KeysCount == "" {
		return "-"
	}
	return i.KeysCount
}

func formatClients(i *r.Instance) string {
	if i.LoadingError {
		return "-"
	}
	return fmt.Sprintf("%d/%d", i.ClientsCount, i.MaxClients)
}

// printMasterSlaveStatus print status of a master-slave/sentinel cluster, called by printClusterStatus
func printMasterSlaveStatus(seedNode *r.Instance) error {
	members, err := seedNode.GetMasterSlaveMembers()
	if err != nil {
		return err
	}
	var (
		master         *r.Instance
		upSlaves       []*r.Instance
		mu             sync.Mutex
		wg             sync.WaitGroup
		warnings       sync.Map     // instances can not be created err messages
		errSlavesCount atomic.Int32 // slaves can not be created counter
	)
	// members can be retrieved means master is ok
	master, _ = r.NewInstance(members[0])
	defer master.Close()
	// get slave instances simultaneously
	for _, slave := range members[1:] {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			if i, err := r.NewInstance(addr); err != nil {
				warnings.Store(addr, err)
				errSlavesCount.Add(1)
				return
			} else {
				mu.Lock()
				upSlaves = append(upSlaves, i)
				mu.Unlock()
			}
		}(slave)
	}
	wg.Wait()
	// Print Cluster Basic Info
	fmt.Println(strings.Repeat("=", 79))
	fmt.Printf("%-16s:\t%s\n", "Cluster Version", seedNode.Version)
	fmt.Println(strings.Repeat("=", 79))
	// Print Node Banner
	color.Cyan("%-24s%-16s%-16s%-16s%s\n", "Address", "Role", "Memory(GB)", "KeysCount", "Clients")
	fmt.Printf("%-24s%-16s%-16s%-16s%s\n", "-------", "----", "----------", "---------", "-------")
	// print master info
	fmt.Print(color.RedString("%-24s", master.Addr))
	fmt.Printf("%-16s", formatRole(master, false))
	fmt.Printf("%-16s", formatMemory(master))
	fmt.Printf("%-16s", formatKeysCount(master))
	fmt.Printf("%s\n", formatClients(master))
	// print slaves info
	sortedUpSlaves := r.InstancesAscByAddr(upSlaves)
	sort.Sort(sortedUpSlaves)
	for _, s := range sortedUpSlaves {
		fmt.Printf("%-24s", s.Addr)
		fmt.Printf("%-16s", formatRole(s, true))
		fmt.Printf("%-16s", formatMemory(s))
		fmt.Printf("%-16s", formatKeysCount(s))
		fmt.Printf("%s\n", formatClients(s))
	}
	for _, i := range upSlaves {
		i.Close()
	}
	color.Cyan("Total up slaves in cluster: %d\n", len(upSlaves))
	if errSlavesCount.Load() != 0 {
		color.Cyan("Warnings:")
		warnings.Range(func(addr, err interface{}) bool {
			color.Red("failed to create instance for slave [addr=%v], error: %v\n", addr, err)
			return true
		})
		color.Cyan("Error slaves in cluster: %d\n", errSlavesCount.Load())
	}
	return nil
}
