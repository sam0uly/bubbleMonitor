package system

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"

	"samouly.fun/bubbleMonitor/internal/data"
	"samouly.fun/bubbleMonitor/internal/msg"
)

var (
	dockerAvailable    bool
	k8sAvailable       bool
	containerCheckOnce sync.Once
)

func init() {
	containerCheckOnce.Do(checkContainerAvailability)
}

func checkContainerAvailability() {
	dockerAvailable = isDockerAvailable()
	k8sAvailable = isKubernetesAvailable()
}

func isDockerAvailable() bool {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("docker", "info")
		_ = cmd.Run()
		return cmd.ProcessState.ExitCode() == 0
	}
	conn, err := net.DialTimeout("unix", "/var/run/docker.sock", time.Second)
	if err != nil {
		conn, err = net.DialTimeout("unix", "/run/docker.sock", time.Second)
	}
	if err == nil {
		_ = conn.Close()
		return true
	}
	return false
}

func isKubernetesAvailable() bool {
	cmd := exec.Command("kubectl", "cluster-info")
	err := cmd.Run()
	return err == nil
}

func HasDocker() bool {
	return dockerAvailable
}

func HasKubernetes() bool {
	return k8sAvailable
}

func ContainerCmd() tea.Cmd {
	return func() tea.Msg {
		var (
			containers []data.ContainerInfo
			pods       []data.K8sPodInfo
			dockerErr  error
			k8sErr     error
		)

		if dockerAvailable {
			containers, dockerErr = fetchDockerContainers()
			if dockerErr != nil {
				containers = nil
			}
		}

		if k8sAvailable {
			pods, k8sErr = fetchKubernetesPods()
			if k8sErr != nil {
				pods = nil
			}
		}

		// Aggregate errors so the user can see both docker and k8s failures,
		// not just whichever the k8s call overwrote.
		var combinedErr error
		switch {
		case dockerErr != nil && k8sErr != nil:
			combinedErr = fmt.Errorf("docker: %v; k8s: %v", dockerErr, k8sErr)
		case dockerErr != nil:
			combinedErr = dockerErr
		case k8sErr != nil:
			combinedErr = k8sErr
		}

		return msg.ContainerInfoMsg{
			Err:           combinedErr,
			Containers:    containers,
			Pods:          pods,
			HasDocker:     dockerAvailable,
			HasKubernetes: k8sAvailable,
		}
	}
}

type dockerNetStats struct {
	RxBytes uint64 `json:"rx_bytes"`
	TxBytes uint64 `json:"tx_bytes"`
}

type dockerCPUStats struct {
	CPUUsage struct {
		TotalUsage uint64 `json:"total_usage"`
	} `json:"cpu_usage"`
	SystemCPUUsage uint64  `json:"system_cpu_usage"`
	OnlineCPUs     float64 `json:"online_cpus"`
}

type dockerMemStats struct {
	Usage uint64 `json:"usage"`
	Limit uint64 `json:"limit"`
}

// dockerStats mirrors the subset of the Docker stats API that we consume.
// Matches the JSON shape returned by "docker stats --no-stream --format
// '{{json .}}'" when --filter id= is used.
type dockerStats struct {
	Read     string                    `json:"read"`
	CPU      dockerCPUStats            `json:"cpu_stats"`
	PreCPU   dockerCPUStats            `json:"precpu_stats"`
	Memory   dockerMemStats            `json:"memory_stats"`
	Networks map[string]dockerNetStats `json:"networks"`
}

func calculateCPUPercent(stats *dockerStats) float64 {
	cpuDelta := float64(stats.CPU.CPUUsage.TotalUsage) - float64(stats.PreCPU.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPU.SystemCPUUsage) - float64(stats.PreCPU.SystemCPUUsage)

	if systemDelta > 0 && cpuDelta > 0 {
		onlineCPUs := stats.CPU.OnlineCPUs
		if onlineCPUs == 0 {
			onlineCPUs = 1
		}
		return (cpuDelta / systemDelta) * onlineCPUs * 100
	}

	return 0
}

func fetchDockerContainers() ([]data.ContainerInfo, error) {
	listCtx, listCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer listCancel()

	cmd := exec.CommandContext(listCtx, "docker", "ps", "-a", "--format", "{{.ID}}\t{{.Names}}\t{{.Image}}\t{{.Status}}\t{{.State}}\t{{.CreatedAt}}")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	result := make([]data.ContainerInfo, 0, len(lines))

	// Pre-build the container list synchronously (fast); then fan out stats
	// fetches in parallel so a single slow container doesn't blow the
	// 10-second overall budget for the whole list.
	type statsResult struct {
		id    string
		stats *dockerStats
		err   error
	}
	statsCh := make(chan statsResult, len(lines))
	pending := 0

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) < 6 {
			continue
		}

		ci := data.ContainerInfo{
			Name:    parts[1],
			ID:      parts[0],
			Status:  parts[3],
			State:   parts[4],
			Image:   parts[2],
			Created: parts[5],
			Type:    "docker",
		}

		// Per-container ctx: a 5s deadline that the parent list deadline
		// also bounds. We snapshot the ID so the goroutine closes over
		// the right value even though the loop variable is reused.
		id := ci.ID
		pending++
		go func() {
			statsCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			s, err := fetchDockerContainerStats(statsCtx, id)
			statsCh <- statsResult{id: id, stats: s, err: err}
		}()

		result = append(result, ci)
	}

	// Drain stats channel.
	statsByID := make(map[string]*dockerStats, pending)
	for range pending {
		r := <-statsCh
		if r.err == nil && r.stats != nil {
			statsByID[r.id] = r.stats
		}
	}

	for i := range result {
		stats, ok := statsByID[result[i].ID]
		if !ok {
			continue
		}
		result[i].CPUPercent = calculateCPUPercent(stats)
		result[i].MemUsage = stats.Memory.Usage
		result[i].MemLimit = stats.Memory.Limit
		if stats.Memory.Limit > 0 {
			result[i].MemPct = float64(stats.Memory.Usage) / float64(stats.Memory.Limit) * 100
		}
		// Sum across all network interfaces; do NOT collapse them into a
		// single fake "eth0" key. Multi-NIC containers retain per-NIC
		// accuracy in the underlying stats even if we only expose the
		// aggregate to the UI.
		for _, netStats := range stats.Networks {
			result[i].NetRx += netStats.RxBytes
			result[i].NetTx += netStats.TxBytes
		}
	}

	return result, nil
}

// fetchDockerContainerStats invokes "docker stats --no-stream" filtered to
// the requested container. It is CPU-correct: it parses both cpu_stats and
// precpu_stats from the Docker API JSON so calculateCPUPercent has a real
// delta. The previous implementation piped the container ID to stdin (which
// docker stats ignores) and stuffed a fake single-sample into the struct.
func fetchDockerContainerStats(ctx context.Context, containerID string) (*dockerStats, error) {
	args := []string{"stats", "--no-stream", "--format", "{{json .}}", "--filter", "id=" + containerID}
	cmd := exec.CommandContext(ctx, "docker", args...)

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return nil, fmt.Errorf("docker stats failed: %w: %s", err, stderr.String())
		}
		return nil, err
	}

	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return nil, fmt.Errorf("empty stats output for container %s", containerID)
	}

	stats := &dockerStats{}
	if err := json.Unmarshal([]byte(output), stats); err != nil {
		return nil, fmt.Errorf("parsing docker stats: %w", err)
	}

	return stats, nil
}

func fetchKubernetesPods() ([]data.K8sPodInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "kubectl", "get", "pods", "-o", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var k8sResult struct {
		Items []struct {
			Metadata struct {
				Name              string `json:"name"`
				Namespace         string `json:"namespace"`
				CreationTimestamp string `json:"creationTimestamp"`
			} `json:"metadata"`
			Status struct {
				Phase      string `json:"phase"`
				PodIP      string `json:"podIP"`
				HostIP     string `json:"hostIP"`
				Conditions []struct {
					Type   string `json:"type"`
					Status string `json:"status"`
				} `json:"conditions"`
				ContainerStatuses []struct {
					Name         string `json:"name"`
					Ready        bool   `json:"ready"`
					RestartCount int    `json:"restartCount"`
					State        struct {
						Running struct{} `json:"running"`
						Waiting struct {
							Reason string `json:"reason"`
						} `json:"waiting"`
						Terminated struct {
							Reason string `json:"reason"`
						} `json:"terminated"`
					} `json:"state"`
				} `json:"containerStatuses"`
			} `json:"status"`
			Spec struct {
				NodeName   string `json:"nodeName"`
				Containers []struct {
					Name      string `json:"name"`
					Resources struct {
						Requests struct {
							CPU    string `json:"cpu"`
							Memory string `json:"memory"`
						} `json:"requests"`
						Limits struct {
							CPU    string `json:"cpu"`
							Memory string `json:"memory"`
						} `json:"limits"`
					} `json:"resources"`
				} `json:"containers"`
			} `json:"spec"`
		} `json:"items"`
	}

	if err := json.Unmarshal(out, &k8sResult); err != nil {
		return nil, err
	}

	pods := make([]data.K8sPodInfo, 0, len(k8sResult.Items))

	for _, item := range k8sResult.Items {
		readyCount := 0
		totalCount := len(item.Status.ContainerStatuses)
		for _, cs := range item.Status.ContainerStatuses {
			if cs.Ready {
				readyCount++
			}
		}

		status := item.Status.Phase
		if len(item.Status.ContainerStatuses) > 0 {
			cs := item.Status.ContainerStatuses[0]
			if cs.State.Waiting.Reason != "" {
				status = cs.State.Waiting.Reason
			} else if cs.State.Terminated.Reason != "" {
				status = cs.State.Terminated.Reason
			}
		}

		cpuReq := ""
		memReq := ""
		cpuLim := ""
		memLim := ""
		if len(item.Spec.Containers) > 0 {
			c := item.Spec.Containers[0]
			cpuReq = c.Resources.Requests.CPU
			memReq = c.Resources.Requests.Memory
			cpuLim = c.Resources.Limits.CPU
			memLim = c.Resources.Limits.Memory
		}

		pod := data.K8sPodInfo{
			Name:      item.Metadata.Name,
			Namespace: item.Metadata.Namespace,
			Status:    status,
			Ready:     fmt.Sprintf("%d/%d", readyCount, totalCount),
			Node:      item.Spec.NodeName,
			CPUReq:    cpuReq,
			MemReq:    memReq,
			CPULim:    cpuLim,
			MemLim:    memLim,
			Age:       item.Metadata.CreationTimestamp,
		}

		if totalCount > 0 {
			pod.Restarts = item.Status.ContainerStatuses[0].RestartCount
		}

		pods = append(pods, pod)
	}

	return pods, nil
}
