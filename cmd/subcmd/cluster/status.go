package cluster

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
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
	Example: fmt.Sprintf("%s cluster status <seed-node> -a \"redis\"", vars.AppName),
	RunE: func(cmd *cobra.Command, args []string) error {
		vars.HostPort = args[0]
		err := printClusterStatus(vars.HostPort, vars.Password)
		if err != nil {
			return err
		}
		return nil
	},
}

func InitStatus() {
	StatusCmd.Flags().BoolVarP(&showSlots, "show-slots", "s", false, "Show slots info or not, default false")
}

func printClusterStatus(hostPort string, password string) error {
	// we call the provided node as `the seed node`(vars.HostPort&vars.Password above)
	seedNode, err := r.NewInstance(hostPort, password)
	if err != nil {
		return err
	}
	defer seedNode.Close()
	if !seedNode.ClusterEnabled {
		return fmt.Errorf("only redis sharding cluster supported")
	}
	// get cluster instances by running `cluster nodes` on seed node
	clusterNodesOutput, err := seedNode.Client.ClusterNodes(context.Background()).Result()
	if err != nil {
		return fmt.Errorf("failed to get cluster nodes: %v", err)
	}
	clusterNodesInfo := r.ParseClusterNodes(clusterNodesOutput)
	// get cluster instances simultaneously
	var (
		clusterInstances []*r.Instance
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
			if i, err := r.NewInstance(addr, vars.Password); err != nil {
				warnings.Store(fmt.Sprintf("%s,%s", addr, nodeId), err)
				errInstanceCount.Add(1)
				return
			} else {
				i.UpdateNodeIdAndSlots(clusterNodesInfo)
				mu.Lock()
				clusterInstances = append(clusterInstances, i)
				mu.Unlock()
			}
		}(addr, nodeId)
	}
	wg.Wait()
	// Print Cluster Basic Info
	fmt.Println(strings.Repeat("=", 155))
	fmt.Printf("%-20s", "Cluster Version:")
	fmt.Println(seedNode.Version)
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
		fmt.Printf("%-16s", "master")
		fmt.Printf("%-16s", fmt.Sprintf("%.2f/%.2f", m.UsedMemory, m.MaxMemory))
		fmt.Printf("%-16s", m.KeysCount)
		fmt.Printf("%-16s", fmt.Sprintf("%d/%d", m.ClientsCount, m.MaxClients))
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
			}
		}
		sortedSlaves := r.InstancesAscByAddr(slaves)
		sort.Sort(sortedSlaves)
		for _, s := range sortedSlaves {
			fmt.Printf("%-45s", s.NodeID)
			fmt.Printf("%-24s", s.Addr)
			if s.SlaveInit {
				fmt.Printf("%-16s", "-slave(init)")
			} else {
				fmt.Printf("%-16s", "-slave")
			}

			fmt.Printf("%-16s", fmt.Sprintf("%.2f/%.2f", s.UsedMemory, s.MaxMemory))
			fmt.Printf("%-16s", s.KeysCount)
			fmt.Printf("%-16s", fmt.Sprintf("%d/%d", s.ClientsCount, s.MaxClients))
			// if slave, skip slot info
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
		color.Red("Error nodes in cluster: %d\n", errInstanceCount.Load())
	}
	if slotsCount != 16384 {
		color.Red("Cluster slot count is not 16384. Some slots may be missing or migrating. Please check your cluster status.")
	}
	return nil
}
