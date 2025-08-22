package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"redis-cluster-manager/vars"
	"time"
)

var rootCmd = &cobra.Command{
	Use: vars.AppName,
	// no need to add short desc for root cmd
	Long: fmt.Sprintf("%s was designed to manage redis cluster.", vars.AppName),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Use %s -h or --help for details.\n", vars.AppName)
	},
}

func initAll() {
	initVersion()
	initCluster()
	rootCmd.PersistentFlags().DurationVarP(&vars.Timeout, "timeout", "t", time.Second*3, "timeout setting, default 3s, can be any of time.Duration format(10ms,1s,1m,... )")
	rootCmd.PersistentFlags().StringVarP(&vars.Password, "password", "a", "", "Redis cluster password")
	rootCmd.PersistentFlags().BoolVar(&vars.CPUProfiler, "cpupprof", false, "write cpu performance profiler to cpu.pprof")
	rootCmd.PersistentFlags().BoolVar(&vars.MEMProfiler, "mempprof", false, "write memory performance profiler to mem.pprof")
}

func Execute() {
	initAll()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
