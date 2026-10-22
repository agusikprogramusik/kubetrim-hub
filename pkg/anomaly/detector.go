package anomaly

import (
	"fmt"
	"math"

	"github.com/agusikprogramusik/kubetrim-hub/pkg/domain"
)

const (
	// DropThresholdRatio defines the relative drop (>30% decrease) that marks a sawtooth reset/OOM restart
	DropThresholdRatio = 0.30
	// MinMonotonicRatio defines the minimum fraction of rising steps between resets to qualify as a memory leak
	MinMonotonicRatio = 0.70
	// IdleGPUUtilizationThreshold flags GPUs below 20% average utilization
	IdleGPUUtilizationThreshold = 20.0
)

// AnalyzeMemorySawtooth inspects a 14-day series of container_memory_working_set_bytes
// to detect monotonic memory leaks interrupted by OOMKilled container restarts (sawtooth pattern).
func AnalyzeMemorySawtooth(clusterID, namespace, workload, container string, samples []domain.MemorySample) *domain.AnomalyFinding {
	if len(samples) < 6 {
		return nil
	}

	risingSteps := 0
	totalGrowthSteps := 0
	sawtoothDrops := 0
	var totalBytesDelta float64
	var totalHours float64

	for i := 1; i < len(samples); i++ {
		prev := samples[i-1]
		curr := samples[i]
		if prev.Bytes <= 0 {
			continue
		}

		delta := float64(curr.Bytes - prev.Bytes)
		dropRatio := -delta / float64(prev.Bytes)

		// Detect sharp drop (container restart / OOMKill)
		if curr.OOMKilled || dropRatio >= DropThresholdRatio {
			sawtoothDrops++
			continue
		}

		totalGrowthSteps++
		if delta > 0 {
			risingSteps++
			totalBytesDelta += delta
			hours := curr.Timestamp.Sub(prev.Timestamp).Hours()
			if hours <= 0 {
				hours = 1.0
			}
			totalHours += hours
		}
	}

	if totalGrowthSteps == 0 {
		return nil
	}

	monotonicRatio := float64(risingSteps) / float64(totalGrowthSteps)
	if sawtoothDrops >= 1 && monotonicRatio >= MinMonotonicRatio {
		slopePerHour := 0.0
		if totalHours > 0 {
			slopePerHour = math.Round((totalBytesDelta/totalHours)*100) / 100
		}
		return &domain.AnomalyFinding{
			ClusterID:           clusterID,
			Namespace:           namespace,
			WorkloadName:        workload,
			ContainerName:       container,
			AnomalyType:         domain.StatusLeakSuspected,
			Severity:            "CRITICAL",
			SawtoothCycles:      sawtoothDrops,
			SlopeBytesPerHour:   slopePerHour,
			BlockMemoryUpsizing: true,
			Recommendation: fmt.Sprintf(
				"Sawtooth memory leak pattern detected (%d OOM/restart cycles, %.0f%% monotonic growth). "+
					"Automatic memory request upsizing is BLOCKED. Capture a heap profile (pprof / heap dump) before modifying limits.",
				sawtoothDrops, monotonicRatio*100,
			),
		}
	}

	return nil
}

// DetectClusterAnomalies scans a cluster AuditReport for idle GPU accelerators and memory leak warnings
func DetectClusterAnomalies(report *domain.AuditReport) []domain.AnomalyFinding {
	if report == nil {
		return nil
	}

	var findings []domain.AnomalyFinding
	for _, c := range report.Containers {
		if c.GPUCount > 0 && c.GPUUtilization < IdleGPUUtilizationThreshold {
			findings = append(findings, domain.AnomalyFinding{
				ClusterID:           report.ClusterName,
				Namespace:           c.Namespace,
				WorkloadName:        c.WorkloadName,
				ContainerName:       c.ContainerName,
				AnomalyType:         domain.StatusIdleGPU,
				Severity:            "HIGH",
				BlockMemoryUpsizing: false,
				MonthlyWasteUSD:     c.MonthlyWasteUSD,
				Recommendation: fmt.Sprintf(
					"Container reserves %d GPU(s) with %.1f%% average utilization (< %.0f%% threshold). "+
						"Consider NVIDIA MIG partitioning, time-slicing, or migrating to Spot GPU instances.",
					c.GPUCount, c.GPUUtilization, IdleGPUUtilizationThreshold,
				),
			})
		}
	}
	return findings
}
