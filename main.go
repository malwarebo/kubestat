package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var namespace string

func main() {
	var rootCmd = &cobra.Command{
		Use:   "kubestat",
		Short: "Kubestat displays Kubernetes pod status",
		Run: func(cmd *cobra.Command, args []string) {
			runKubestat(cmd, args, namespace)
		},
	}

	rootCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Kubernetes namespace to display pods from")
	rootCmd.MarkFlagRequired("namespace")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
