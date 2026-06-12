package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// problemWaitingReasons are container "waiting" reasons that indicate a real
// failure rather than a transient startup state (e.g. ContainerCreating).
var problemWaitingReasons = map[string]bool{
	"CrashLoopBackOff":           true,
	"ImagePullBackOff":           true,
	"ErrImagePull":               true,
	"ErrImageNeverPull":          true,
	"InvalidImageName":           true,
	"CreateContainerConfigError": true,
	"CreateContainerError":       true,
	"RunContainerError":          true,
}

func kubeconfigPath() string {
	return filepath.Join(os.Getenv("HOME"), ".kube", "config")
}

func getKubeClient() (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath())
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func getCurrentContext() (string, error) {
	config, err := clientcmd.LoadFromFile(kubeconfigPath())
	if err != nil {
		return "", err
	}
	return config.CurrentContext, nil
}

// displayUnhealthyPods lists pods (in the given namespace, or all namespaces
// when namespace is empty) and prints only the ones that look broken.
func displayUnhealthyPods(namespace string, restartThreshold int32) {
	clientset, err := getKubeClient()
	if err != nil {
		fmt.Printf("Error creating Kubernetes client: %v\n", err)
		return
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error listing pods: %v\n", err)
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAMESPACE", "POD", "STATUS", "RESTARTS", "AGE"})

	namespaces := map[string]struct{}{}
	unhealthy := 0

	for _, pod := range pods.Items {
		namespaces[pod.Namespace] = struct{}{}

		reason, restarts, healthy := podHealth(pod, restartThreshold)
		if healthy {
			continue
		}
		unhealthy++

		color := tablewriter.FgHiYellowColor
		if isHardFailure(reason) {
			color = tablewriter.FgHiRedColor
		}
		row := []string{
			pod.Namespace,
			pod.Name,
			reason,
			fmt.Sprintf("%d", restarts),
			podAge(pod),
		}
		table.Rich(row, []tablewriter.Colors{{}, {}, {color}, {}, {}})
	}

	if unhealthy > 0 {
		table.Render()
		fmt.Println()
	}

	fmt.Printf("Scanned %d pod(s) across %d namespace(s): %d unhealthy.\n",
		len(pods.Items), len(namespaces), unhealthy)
	if unhealthy == 0 {
		fmt.Println("All pods healthy.")
	}
}

// podHealth inspects a pod and returns a human-readable status reason, its total
// restart count, and whether the pod is considered healthy. Healthy pods are
// hidden from the report.
func podHealth(pod corev1.Pod, restartThreshold int32) (reason string, restarts int32, healthy bool) {
	if pod.Status.Phase == corev1.PodSucceeded {
		return "", 0, true
	}

	var (
		totalRestarts    int32
		waitingReason    string
		terminatedReason string
		notReady         bool
	)

	allStatuses := append([]corev1.ContainerStatus{}, pod.Status.InitContainerStatuses...)
	allStatuses = append(allStatuses, pod.Status.ContainerStatuses...)

	for _, cs := range allStatuses {
		totalRestarts += cs.RestartCount

		if w := cs.State.Waiting; w != nil && waitingReason == "" && problemWaitingReasons[w.Reason] {
			waitingReason = w.Reason
		}
		if t := cs.State.Terminated; t != nil && t.ExitCode != 0 && terminatedReason == "" {
			if t.Reason != "" {
				terminatedReason = t.Reason
			} else {
				terminatedReason = fmt.Sprintf("Exit:%d", t.ExitCode)
			}
		}
	}

	for _, cs := range pod.Status.ContainerStatuses {
		if !cs.Ready {
			notReady = true
		}
	}

	switch {
	case waitingReason != "":
		return waitingReason, totalRestarts, false
	case terminatedReason != "":
		return terminatedReason, totalRestarts, false
	case pod.Status.Phase == corev1.PodFailed:
		return "Failed", totalRestarts, false
	case pod.Status.Phase == corev1.PodPending:
		return "Pending", totalRestarts, false
	case totalRestarts > restartThreshold:
		return fmt.Sprintf("HighRestarts(%d)", totalRestarts), totalRestarts, false
	case pod.Status.Phase == corev1.PodRunning && notReady:
		return "NotReady", totalRestarts, false
	}

	return "", totalRestarts, true
}

// isHardFailure reports whether a status reason represents a hard failure
// (rendered red) as opposed to a softer warning state (rendered yellow).
func isHardFailure(reason string) bool {
	if problemWaitingReasons[reason] {
		return true
	}
	switch reason {
	case "Failed", "Error", "OOMKilled", "Evicted", "ContainerCannotRun", "DeadlineExceeded":
		return true
	}
	return strings.HasPrefix(reason, "Exit:")
}

func podAge(pod corev1.Pod) string {
	start := pod.CreationTimestamp.Time
	if start.IsZero() {
		return "-"
	}
	return humanizeDuration(time.Since(start))
}

func humanizeDuration(d time.Duration) string {
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
	default:
		return fmt.Sprintf("%dd%dh", int(d.Hours())/24, int(d.Hours())%24)
	}
}
