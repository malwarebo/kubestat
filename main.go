package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	namespace        string
	restartThreshold int32
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "kubestat",
		Short: "Show only the unhealthy Kubernetes pods across your namespaces",
		Run: func(cmd *cobra.Command, args []string) {
			runKubestat(namespace, restartThreshold)
		},
	}

	rootCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Limit the scan to a single namespace (default: all namespaces)")
	rootCmd.Flags().Int32Var(&restartThreshold, "restarts", 5, "Flag pods whose total restart count exceeds this threshold")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
