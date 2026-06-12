package main

import (
	"fmt"
)

func runKubestat(namespace string, restartThreshold int32) {
	currentContext, err := getCurrentContext()
	if err != nil {
		fmt.Printf("Error reading current context: %v\n", err)
		return
	}

	scope := "all namespaces"
	if namespace != "" {
		scope = fmt.Sprintf("namespace %q", namespace)
	}
	fmt.Printf("Context: %s  |  Scanning %s for unhealthy pods...\n\n", currentContext, scope)

	displayUnhealthyPods(namespace, restartThreshold)
}
