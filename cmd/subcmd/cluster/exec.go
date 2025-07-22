package cluster

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"net/netip"
	r "redis-cluster-manager/redis"
	"redis-cluster-manager/vars"
	"strings"
	"sync"
	"sync/atomic"
)

var (
	redisCmd string // the redis command to be executed
	nodes    string // comma separated nodeID or ip:port
	role     string // master/slave/all
)

var ExecCmd = &cobra.Command{
	Use:   "exec",
	Short: "Execute cmd on specified cluster nodes or role",
	Long: `You can specify comma separated nodeID or ip:port; or master/slave/all.
If master/slave/all specified, cmd will be run on masters/slaves/all nodes.
If no nodes&&roles specified, cmd will run on the seed node itself.
The -n and -r options are mutually exclusive.`,
	Args:    cobra.ExactArgs(1),
	Example: fmt.Sprintf("%s cluster exec <seed-node> -a \"xxx\" -c \"PING\" [-n=<nodeID/ip:port,...> | -r=<master/slave/all>]", vars.AppName),
	RunE: func(cmd *cobra.Command, args []string) error {
		vars.HostPort = args[0]
		if len(role) > 0 {
			if role != "master" && role != "slave" && role != "all" {
				return fmt.Errorf("role must be `master` or `slave` or `all` when specified")
			}
		}
		if err := PrintExecuteResult(vars.HostPort, vars.Password); err != nil {
			return err
		}
		return nil
	},
}

func InitExecParams() {
	ExecCmd.Flags().StringVarP(&redisCmd, "cmd", "c", "PING", "redis command to be executed")
	ExecCmd.Flags().StringVarP(&nodes, "nodes", "n", "", "nodes to be executed on")
	ExecCmd.Flags().StringVarP(&role, "role", "r", "", "role to be executed on, one of [master, slave, all]")
	ExecCmd.MarkFlagsMutuallyExclusive("nodes", "role")
}

func PrintExecuteResult(hostPort string, password string) error {
	// validate redisCmd
	cmdFields := strings.Fields(redisCmd)
	_, exists := vars.ForbiddenCmds[strings.ToUpper(cmdFields[0])]
	if exists {
		return fmt.Errorf("command `%s` is forbidden to execute", cmdFields[0])
	}
	// we call the provided node as `the seed node`
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

	// get cluster instances, then filter it by nodes or role
	var (
		clusterInstances []*r.Instance
		execInstances    []*r.Instance
		mu               sync.Mutex
		wg               sync.WaitGroup
		warnings         sync.Map     // instances can not be created err messages
		errInstanceCount atomic.Int32 // instances can not be created counter
	)
	for _, nodeInfo := range clusterNodesInfo {
		addr := nodeInfo[1]
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			if i, err := r.NewInstance(addr, vars.Password); err != nil {
				warnings.Store(addr, err)
				errInstanceCount.Add(1)
				return
			} else {
				i.UpdateNodeIdAndSlots(clusterNodesInfo)
				mu.Lock()
				clusterInstances = append(clusterInstances, i)
				mu.Unlock()
			}
		}(addr)
	}
	wg.Wait()

	filterType, execInstances, err := filterInstances(clusterInstances)
	if err != nil {
		return fmt.Errorf("failed to parse nodes/role: %v", err)
	}

	// execute the command on the filtered instances
	var results sync.Map
	var wgExec sync.WaitGroup
	for _, instance := range execInstances {
		wgExec.Add(1)
		go func(i *r.Instance) {
			defer wgExec.Done()
			fields := strings.Fields(redisCmd)
			if len(fields) == 0 {
				results.Store(i.Addr, "")
				return
			}
			args := make([]interface{}, 0, len(fields)-1)
			for _, f := range fields {
				args = append(args, f)
			}
			var addrDisplayed = i.Addr
			if filterType == vars.FILTER_NODEID {
				addrDisplayed = fmt.Sprintf("%s(%s)", i.Addr, i.NodeID)
			}
			stdout, err := i.Client.Do(context.Background(), args...).Result()
			if err != nil {
				results.Store(addrDisplayed, fmt.Sprintf("Error executing command: %v", err))
			} else {
				results.Store(addrDisplayed, stdout)
			}
		}(instance)
	}
	wgExec.Wait()
	// print results
	results.Range(func(addr, stdout interface{}) bool {
		color.Yellow("Output of `%s` on %s:\n", redisCmd, addr)
		fmt.Println(stdout)
		return true
	})
	if errInstanceCount.Load() != 0 {
		color.Cyan("Warnings:")
		warnings.Range(func(addr, stdout interface{}) bool {
			color.Red("failed to create instance for node %s: %v\n", addr, err)
			return true
		})
	}
	color.Cyan("Done!")
	return nil
}

// filterInstances filters the cluster instances based on the provided nodes or role flags.
func filterInstances(clusterInstances []*r.Instance) (int, []*r.Instance, error) {
	var filterType int
	var execInstances []*r.Instance
	// nodes/role have been marked to MarkFlagsMutuallyExclusive and checked in RunE,
	// so we can safely check them with `else if`
	if len(nodes) > 0 {
		nodeList := strings.Split(nodes, ",")
		_, err := netip.ParseAddrPort(nodeList[0])
		if err == nil {
			filterType = vars.FILTER_ADDR
			// if the first node is an ip:port, we treat all nodes as ip:port
			for _, addrPort := range nodeList {
				for _, i := range clusterInstances {
					if i.Addr == addrPort {
						execInstances = append(execInstances, i)
					}
				}
			}
			if len(execInstances) != len(nodeList) {
				return filterType, nil, fmt.Errorf("some nodes not found in cluster")
			}
		} else {
			filterType = vars.FILTER_NODEID
			// otherwise we treat all nodes as nodeID
			for _, nodeID := range nodeList {
				for _, instance := range clusterInstances {
					if instance.NodeID == nodeID {
						execInstances = append(execInstances, instance)
					}
				}
			}
			if len(execInstances) != len(nodeList) {
				return filterType, nil, fmt.Errorf("some nodes not found in cluster")
			}
		}
	} else if role == vars.MASTER {
		filterType = vars.FILTER_ROLE
		for _, i := range clusterInstances {
			if i.Role == "master" {
				execInstances = append(execInstances, i)
			}
		}
	} else if role == vars.SLAVE {
		filterType = vars.FILTER_ROLE
		for _, i := range clusterInstances {
			if i.Role == "slave" {
				execInstances = append(execInstances, i)
			}
		}
	} else if role == vars.ALL {
		filterType = vars.FILTER_ROLE
		execInstances = clusterInstances // no filter, use all instances
	} else {
		// if no nodes or role specified, we run the command on the seed node only
		filterType = vars.FILTER_NONE
		for _, i := range clusterInstances {
			if i.Addr == vars.HostPort {
				execInstances = append(execInstances, i)
				break
			}
		}
	}
	return filterType, execInstances, nil
}
