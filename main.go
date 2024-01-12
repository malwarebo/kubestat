// main.go
package main

import (
	"fmt"
	"os"

	"github.com/malwarebo/kubestat/cmd"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{Use: "kubestat", Run: cmd.kubestat}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
