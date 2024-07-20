package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func runKubestat(cmd *cobra.Command, args []string, namespace string) {
	fmt.Println("Checking Kubernetes Status...")
	currentContext, err := getCurrentContext()
	if err != nil {
		fmt.Printf("Error reading current context: %v\n", err)
		return
	}
	fmt.Printf("Current context: %s\n", currentContext)
	fmt.Printf("Namespace: %s\n", namespace)
	fmt.Println("Running pods:")
	displayKubernetesStatus(namespace)
}
