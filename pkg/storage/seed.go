package storage

import (
	"time"

	"github.com/agusikprogramusik/kubetrim-hub/pkg/domain"
)

// SeedDemoFleet populates the store with realistic multi-cluster T0 -> T1 snapshots and sawtooth anomalies
func SeedDemoFleet(store Store) {
	now := time.Now().UTC()
	t0 := now.Add(-72 * time.Hour)
	t1 := now.Add(-5 * time.Minute)

	// Cluster 1: prod-eu-central-1 (AWS EKS - T0 baseline before optimization)
	awsT0 := domain.AuditReport{
		SchemaVersion: "v1",
		ReportID:      "snap-aws-t0-8f91a",
		ClusterName:   "prod-eu-central-1",
		ClusterHash:   "sha256:4a8b99e1c02f",
		GeneratedAt:   t0,
		GitCommit:     "7cd1db7",
		Summary: domain.AuditSummary{
			TotalWorkloadsAudited: 4,
			TotalContainers:       4,
			CurrentMonthlyCostUSD: 2840.00,
			TargetMonthlyCostUSD:  1290.00,
			TotalMonthlyWasteUSD:  1550.00,
			TotalYearlyWasteUSD:   18600.00,
			ClusterEfficiency:     45.4,
			CriticalOOMCount:      1,
			ThrottledCount:        0,
			HPAManagedCount:       1,
			IdleGPUCount:          1,
			SafeCount:             2,
		},
		Containers: []domain.ContainerAnalysis{
			{
				Namespace:       "payments",
				WorkloadName:    "checkout-engine",
				WorkloadKind:    "Deployment",
				ContainerName:   "api",
				ConfigHash:      "sha256:9c11b0a4",
				CurrentReqCPU:   2000,
				TargetReqCPU:    450,
				CurrentLimitCPU: 4000,
				TargetLimitCPU:  900,
				CurrentReqMem:   4 * 1024 * 1024 * 1024,
				TargetReqMem:    1 * 1024 * 1024 * 1024,
				CurrentLimitMem: 4 * 1024 * 1024 * 1024,
				TargetLimitMem:  1 * 1024 * 1024 * 1024,
				Replicas:        6,
				MonthlyWasteUSD: 412.80,
				RiskStatus:      "SAFE",
				CurrentQoS:      "Guaranteed",
				ProjectedQoS:    "Guaranteed",
				NodeArch:        "amd64",
			},
			{
				Namespace:       "ml-inference",
				WorkloadName:    "qwen-rag-worker",
				WorkloadKind:    "Deployment",
				ContainerName:   "vllm-shard",
				ConfigHash:      "sha256:3e77d192",
				CurrentReqCPU:   4000,
				TargetReqCPU:    1000,
				CurrentLimitCPU: 8000,
				TargetLimitCPU:  2000,
				CurrentReqMem:   16 * 1024 * 1024 * 1024,
				TargetReqMem:    8 * 1024 * 1024 * 1024,
				CurrentLimitMem: 16 * 1024 * 1024 * 1024,
				TargetLimitMem:  8 * 1024 * 1024 * 1024,
				Replicas:        1,
				MonthlyWasteUSD: 794.50,
				RiskStatus:      "IDLE_GPU",
				CurrentQoS:      "Burstable",
				ProjectedQoS:    "Burstable",
				NodeArch:        "amd64",
				GPUCount:        1,
				GPUUtilization:  6.4,
			},
			{
				Namespace:       "billing",
				WorkloadName:    "invoice-aggregator",
				WorkloadKind:    "StatefulSet",
				ContainerName:   "worker",
				ConfigHash:      "sha256:6b20f44e",
				CurrentReqCPU:   1500,
				TargetReqCPU:    400,
				CurrentLimitCPU: 3000,
				TargetLimitCPU:  800,
				CurrentReqMem:   4 * 1024 * 1024 * 1024,
				TargetReqMem:    4 * 1024 * 1024 * 1024,
				CurrentLimitMem: 8 * 1024 * 1024 * 1024,
				TargetLimitMem:  8 * 1024 * 1024 * 1024,
				Replicas:        3,
				MonthlyWasteUSD: 218.40,
				RiskStatus:      "RISK_OOM",
				CurrentQoS:      "Burstable",
				ProjectedQoS:    "Burstable",
				NodeArch:        "arm64",
				OOMCount:        4,
				Warnings:        []string{"LEAK_SUSPECTED: 14-day sawtooth memory pattern detected (4 OOM resets). Memory upsizing blocked."},
			},
			{
				Namespace:       "core-gateway",
				WorkloadName:    "envoy-edge",
				WorkloadKind:    "Deployment",
				ContainerName:   "proxy",
				ConfigHash:      "sha256:1d84e20c",
				CurrentReqCPU:   1000,
				TargetReqCPU:    1000,
				CurrentLimitCPU: 2000,
				TargetLimitCPU:  2000,
				CurrentReqMem:   2 * 1024 * 1024 * 1024,
				TargetReqMem:    768 * 1024 * 1024,
				CurrentLimitMem: 4 * 1024 * 1024 * 1024,
				TargetLimitMem:  1536 * 1024 * 1024,
				Replicas:        4,
				MonthlyWasteUSD: 124.30,
				RiskStatus:      "HPA_MANAGED",
				CurrentQoS:      "Burstable",
				ProjectedQoS:    "Burstable",
				HasHPA:          true,
				HPAMetrics:      []string{"cpu"},
				IsSpotInstance:  true,
				NodeArch:        "arm64",
				Warnings:        []string{"HPA_GUARD: CPU request downsizing suppressed to prevent autoscaler flapping"},
			},
		},
	}

	// Cluster 1: prod-eu-central-1 (T1 after checkout-engine optimization merged via GitOps)
	awsT1 := awsT0
	awsT1.ReportID = "snap-aws-t1-c402e"
	awsT1.GeneratedAt = t1
	awsT1.Summary.CurrentMonthlyCostUSD = 2427.20
	awsT1.Summary.TotalMonthlyWasteUSD = 1137.20
	awsT1.Summary.TotalYearlyWasteUSD = 13646.40
	awsT1.Summary.ClusterEfficiency = 78.4
	awsT1Containers := make([]domain.ContainerAnalysis, len(awsT0.Containers))
	copy(awsT1Containers, awsT0.Containers)
	awsT1Containers[0].ConfigHash = "sha256:2f88e91b"
	awsT1Containers[0].CurrentReqCPU = 450
	awsT1Containers[0].CurrentReqMem = 1 * 1024 * 1024 * 1024
	awsT1Containers[0].MonthlyWasteUSD = 18.50
	awsT1.Containers = awsT1Containers

	_ = store.SaveSnapshot("prod-eu-central-1", awsT0)
	_ = store.SaveSnapshot("prod-eu-central-1", awsT1)

	// Record 14-day sawtooth memory leak anomaly for billing/invoice-aggregator
	store.RecordAnomaly(domain.AnomalyFinding{
		ClusterID:           "prod-eu-central-1",
		Namespace:           "billing",
		WorkloadName:        "invoice-aggregator",
		ContainerName:       "worker",
		AnomalyType:         domain.StatusLeakSuspected,
		Severity:            "CRITICAL",
		SawtoothCycles:      4,
		SlopeBytesPerHour:   48.5 * 1024 * 1024,
		BlockMemoryUpsizing: true,
		MonthlyWasteUSD:     218.40,
		Recommendation:      "Sawtooth memory leak pattern detected (4 OOM/restart cycles, 92% monotonic growth). Automatic memory request upsizing is BLOCKED. Capture a Go pprof heap profile before modifying limits.",
	})

	// Cluster 2: gke-analytics-west1 (GCP GKE)
	gcpSnap := domain.AuditReport{
		SchemaVersion: "v1",
		ReportID:      "snap-gcp-t1-910a",
		ClusterName:   "gke-analytics-west1",
		ClusterHash:   "sha256:77bc1209",
		GeneratedAt:   t1,
		Summary: domain.AuditSummary{
			TotalWorkloadsAudited: 2,
			TotalContainers:       2,
			CurrentMonthlyCostUSD: 980.00,
			TargetMonthlyCostUSD:  690.00,
			TotalMonthlyWasteUSD:  290.00,
			TotalYearlyWasteUSD:   3480.00,
			ClusterEfficiency:     81.2,
			SafeCount:             2,
		},
		Containers: []domain.ContainerAnalysis{
			{
				Namespace:       "data-pipeline",
				WorkloadName:    "kafka-ingest-stream",
				WorkloadKind:    "StatefulSet",
				ContainerName:   "consumer",
				ConfigHash:      "sha256:55a190bc",
				CurrentReqCPU:   1500,
				TargetReqCPU:    600,
				CurrentLimitCPU: 3000,
				TargetLimitCPU:  1200,
				CurrentReqMem:   4 * 1024 * 1024 * 1024,
				TargetReqMem:    2 * 1024 * 1024 * 1024,
				Replicas:        3,
				MonthlyWasteUSD: 290.00,
				RiskStatus:      "SAFE",
				CurrentQoS:      "Burstable",
				ProjectedQoS:    "Burstable",
				NodeArch:        "amd64",
			},
		},
	}
	_ = store.SaveSnapshot("gke-analytics-west1", gcpSnap)
}
