package anomaly

import (
	"testing"
	"time"

	"github.com/agusikprogramusik/kubetrim-hub/pkg/domain"
)

func TestAnalyzeMemorySawtooth_DetectsLeakAndBlocksUpsizing(t *testing.T) {
	baseTime := time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC)
	const mb = 1024 * 1024

	// Simulate 2 sawtooth cycles: steady climb from 200MB -> 950MB -> OOM reset to 210MB -> climb to 960MB -> OOM reset
	samples := []domain.MemorySample{
		{Timestamp: baseTime.Add(0 * time.Hour), Bytes: 200 * mb},
		{Timestamp: baseTime.Add(6 * time.Hour), Bytes: 420 * mb},
		{Timestamp: baseTime.Add(12 * time.Hour), Bytes: 680 * mb},
		{Timestamp: baseTime.Add(18 * time.Hour), Bytes: 950 * mb},
		{Timestamp: baseTime.Add(19 * time.Hour), Bytes: 210 * mb, OOMKilled: true},
		{Timestamp: baseTime.Add(25 * time.Hour), Bytes: 480 * mb},
		{Timestamp: baseTime.Add(31 * time.Hour), Bytes: 760 * mb},
		{Timestamp: baseTime.Add(37 * time.Hour), Bytes: 960 * mb},
		{Timestamp: baseTime.Add(38 * time.Hour), Bytes: 205 * mb, OOMKilled: true},
	}

	finding := AnalyzeMemorySawtooth("eks-prod", "checkout", "cart-service", "app", samples)
	if finding == nil {
		t.Fatal("expected sawtooth memory leak finding, got nil")
	}
	if finding.AnomalyType != domain.StatusLeakSuspected {
		t.Errorf("expected %s, got %s", domain.StatusLeakSuspected, finding.AnomalyType)
	}
	if !finding.BlockMemoryUpsizing {
		t.Error("expected BlockMemoryUpsizing=true for suspected memory leak")
	}
	if finding.SawtoothCycles != 2 {
		t.Errorf("expected 2 sawtooth cycles, got %d", finding.SawtoothCycles)
	}
	if finding.SlopeBytesPerHour <= 0 {
		t.Errorf("expected positive SlopeBytesPerHour, got %f", finding.SlopeBytesPerHour)
	}
}

func TestAnalyzeMemorySawtooth_StableWorkloadReturnsNil(t *testing.T) {
	if f := AnalyzeMemorySawtooth("c", "ns", "w", "c", nil); f != nil {
		t.Fatalf("expected nil for empty samples, got %+v", f)
	}

	baseTime := time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC)
	const mb = 1024 * 1024
	stable := []domain.MemorySample{
		{Timestamp: baseTime.Add(1 * time.Hour), Bytes: 300 * mb},
		{Timestamp: baseTime.Add(2 * time.Hour), Bytes: 310 * mb},
		{Timestamp: baseTime.Add(3 * time.Hour), Bytes: 295 * mb},
		{Timestamp: baseTime.Add(4 * time.Hour), Bytes: 305 * mb},
		{Timestamp: baseTime.Add(5 * time.Hour), Bytes: 298 * mb},
		{Timestamp: baseTime.Add(6 * time.Hour), Bytes: 302 * mb},
	}

	if f := AnalyzeMemorySawtooth("eks-prod", "default", "stable-api", "api", stable); f != nil {
		t.Fatalf("expected nil for stable oscillating memory, got %+v", f)
	}
}

func TestDetectClusterAnomalies_IdleGPU(t *testing.T) {
	if res := DetectClusterAnomalies(nil); res != nil {
		t.Fatalf("expected nil for nil report")
	}

	report := &domain.AuditReport{
		ClusterName: "gpu-cluster",
		Containers: []domain.ContainerAnalysis{
			{
				Namespace:       "ml-inference",
				WorkloadName:    "llm-worker",
				ContainerName:   "vllm",
				GPUCount:        2,
				GPUUtilization:  6.4,
				MonthlyWasteUSD: 2450.00,
			},
			{
				Namespace:      "ml-training",
				WorkloadName:   "trainer",
				ContainerName:  "pytorch",
				GPUCount:       4,
				GPUUtilization: 88.5,
			},
		},
	}

	findings := DetectClusterAnomalies(report)
	if len(findings) != 1 {
		t.Fatalf("expected 1 IDLE_GPU finding, got %d", len(findings))
	}
	if findings[0].AnomalyType != domain.StatusIdleGPU || findings[0].WorkloadName != "llm-worker" {
		t.Errorf("unexpected finding: %+v", findings[0])
	}
}
