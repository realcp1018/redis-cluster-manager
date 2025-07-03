package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"redis-cluster-manager/vars"
)

var rootCmd = &cobra.Command{
	Use: vars.AppName,
	// no need to add short desc for root cmd
	Long: fmt.Sprintf("%s was designed to manage redis cluster.", vars.AppName),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Use %s -h or --help for details.", vars.AppName)
	},
}

func initAll() {
	initVersion()
	initCluster()
	rootCmd.PersistentFlags().StringVarP(&vars.Password, "password", "a", "", "Redis cluster password")
}

func Execute() {
	initAll()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
