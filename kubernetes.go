package main

import (
	"context"
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

func displayKubernetesStatus(namespace string) {
	clientset, err := getKubeClient()
	if err != nil {
		fmt.Printf("Error creating Kubernetes client: %v\n", err)
		return
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error listing pods in namespace %s: %v\n", namespace, err)
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "NAME", "STATUS", "NAMESPACE"})

	for _, pod := range pods.Items {
		id := string(pod.ObjectMeta.UID)
		name := pod.ObjectMeta.Name
		status := strings.ToLower(string(pod.Status.Phase))
		color := tablewriter.FgCyanColor
		if status == "running" {
			color = tablewriter.FgHiGreenColor
		} else if status == "terminated" || status == "failed" {
			color = tablewriter.FgHiRedColor
		}
		table.Rich([]string{id, name, status, namespace}, []tablewriter.Colors{{color}})
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
