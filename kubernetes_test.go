package main

import (
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
)

func waiting(reason string) corev1.ContainerState {
	return corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: reason}}
}

func terminated(exitCode int32, reason string) corev1.ContainerState {
	return corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{ExitCode: exitCode, Reason: reason}}
}

func runningState() corev1.ContainerState {
	return corev1.ContainerState{Running: &corev1.ContainerStateRunning{}}
}

func status(state corev1.ContainerState, ready bool, restarts int32) corev1.ContainerStatus {
	return corev1.ContainerStatus{State: state, Ready: ready, RestartCount: restarts}
}

func newPod(phase corev1.PodPhase, statuses ...corev1.ContainerStatus) corev1.Pod {
	return corev1.Pod{Status: corev1.PodStatus{Phase: phase, ContainerStatuses: statuses}}
}

func TestPodHealth(t *testing.T) {
	tests := []struct {
		name         string
		pod          corev1.Pod
		threshold    int32
		wantReason   string
		wantRestarts int32
		wantHealthy  bool
	}{
		{
			name:        "running and ready is healthy",
			pod:         newPod(corev1.PodRunning, status(runningState(), true, 0)),
			threshold:   5,
			wantHealthy: true,
		},
		{
			name:        "succeeded is healthy",
			pod:         newPod(corev1.PodSucceeded, status(terminated(0, "Completed"), false, 0)),
			threshold:   5,
			wantHealthy: true,
		},
		{
			name:         "crashloopbackoff is reported",
			pod:          newPod(corev1.PodRunning, status(waiting("CrashLoopBackOff"), false, 7)),
			threshold:    5,
			wantReason:   "CrashLoopBackOff",
			wantRestarts: 7,
		},
		{
			name:         "imagepullbackoff is reported",
			pod:          newPod(corev1.PodPending, status(waiting("ImagePullBackOff"), false, 0)),
			threshold:    5,
			wantReason:   "ImagePullBackOff",
			wantRestarts: 0,
		},
		{
			name:         "terminated with reason uses the reason",
			pod:          newPod(corev1.PodRunning, status(terminated(137, "OOMKilled"), false, 2)),
			threshold:    5,
			wantReason:   "OOMKilled",
			wantRestarts: 2,
		},
		{
			name:         "terminated without reason uses exit code",
			pod:          newPod(corev1.PodRunning, status(terminated(137, ""), false, 0)),
			threshold:    5,
			wantReason:   "Exit:137",
			wantRestarts: 0,
		},
		{
			name:        "failed phase is reported",
			pod:         newPod(corev1.PodFailed),
			threshold:   5,
			wantReason:  "Failed",
			wantHealthy: false,
		},
		{
			name:        "pending phase is reported",
			pod:         newPod(corev1.PodPending, status(waiting("ContainerCreating"), false, 0)),
			threshold:   5,
			wantReason:  "Pending",
			wantHealthy: false,
		},
		{
			name:         "restarts over threshold are reported",
			pod:          newPod(corev1.PodRunning, status(runningState(), true, 8)),
			threshold:    5,
			wantReason:   "HighRestarts(8)",
			wantRestarts: 8,
		},
		{
			name:         "restarts equal to threshold are healthy",
			pod:          newPod(corev1.PodRunning, status(runningState(), true, 5)),
			threshold:    5,
			wantRestarts: 5,
			wantHealthy:  true,
		},
		{
			name:        "running but not ready is reported",
			pod:         newPod(corev1.PodRunning, status(runningState(), false, 0)),
			threshold:   5,
			wantReason:  "NotReady",
			wantHealthy: false,
		},
		{
			name: "restart count sums across containers",
			pod: newPod(corev1.PodRunning,
				status(waiting("CrashLoopBackOff"), false, 3),
				status(runningState(), true, 4),
			),
			threshold:    5,
			wantReason:   "CrashLoopBackOff",
			wantRestarts: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reason, restarts, healthy := podHealth(tt.pod, tt.threshold)
			if healthy != tt.wantHealthy {
				t.Errorf("healthy = %v, want %v", healthy, tt.wantHealthy)
			}
			if reason != tt.wantReason {
				t.Errorf("reason = %q, want %q", reason, tt.wantReason)
			}
			if restarts != tt.wantRestarts {
				t.Errorf("restarts = %d, want %d", restarts, tt.wantRestarts)
			}
		})
	}
}

func TestIsHardFailure(t *testing.T) {
	tests := []struct {
		reason string
		want   bool
	}{
		{"CrashLoopBackOff", true},
		{"ImagePullBackOff", true},
		{"Failed", true},
		{"OOMKilled", true},
		{"Exit:1", true},
		{"Pending", false},
		{"NotReady", false},
		{"HighRestarts(9)", false},
	}

	for _, tt := range tests {
		t.Run(tt.reason, func(t *testing.T) {
			if got := isHardFailure(tt.reason); got != tt.want {
				t.Errorf("isHardFailure(%q) = %v, want %v", tt.reason, got, tt.want)
			}
		})
	}
}

func TestHumanizeDuration(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"seconds", 45 * time.Second, "45s"},
		{"minutes", 5 * time.Minute, "5m"},
		{"hours", 2*time.Hour + 30*time.Minute, "2h30m"},
		{"days", 49 * time.Hour, "2d1h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := humanizeDuration(tt.d); got != tt.want {
				t.Errorf("humanizeDuration(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}
