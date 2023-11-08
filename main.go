package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/olekukonko/tablewriter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func getKubeClient() (*kubernetes.Clientset, error) {
	kubeconfigPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func displayKubernetesStatus() {
	clientset, err := getKubeClient()
	if err != nil {
		fmt.Printf("Error creating Kubernetes client: %v\n", err)
		return
	}

	pods, err := clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error listing pods: %v\n", err)
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "NAME", "STATUS"})

	for _, pod := range pods.Items {
		id := string(pod.ObjectMeta.UID)
		name := pod.ObjectMeta.Name
		status := strings.ToLower(string(pod.Status.Phase))
		table.Append([]string{id, name, status})
	}

	table.Render()
}

func getCurrentContext() (string, error) {
	kubeconfigPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		return "", err
	}
	return config.CurrentContext, nil
}

func main() {
	flag.Parse()

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
