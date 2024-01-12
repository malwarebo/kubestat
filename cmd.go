package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func runKubestat(cmd *cobra.Command, args []string) {
	fmt.Println("Checking Kubernetes Status...")

	currentContext, err := getCurrentContext()
	if err != nil {
		fmt.Printf("Error reading current context: %v\n", err)
		return
	}

	fmt.Printf("Current context: %s\n", currentContext)

	fmt.Println("Running pods:")
	displayKubernetesStatus()
}
