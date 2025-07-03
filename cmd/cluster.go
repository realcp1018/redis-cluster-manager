package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"redis-cluster-manager/cmd/subcmd/cluster"
	"redis-cluster-manager/vars"
)

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Cluster operations root cmd",
	Long:  `Cluster operations root cmd`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Run `%s cluster --help` for details.\n", vars.AppName)
	},
}

func initCluster() {
	rootCmd.AddCommand(clusterCmd)
	// add status subcmd
	cluster.InitStatus()
	clusterCmd.AddCommand(cluster.StatusCmd)
	// add exec subcmd
	cluster.InitExecParams()
	clusterCmd.AddCommand(cluster.ExecCmd)
}
