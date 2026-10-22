package storage

import (
	"strings"
	"testing"
	"time"

	"github.com/agusikprogramusik/kubetrim-hub/pkg/domain"
)

func TestMemoryStoreLifecycleAndKPIs(t *testing.T) {
	store := NewMemoryStore()

	if err := store.SaveSnapshot("", domain.AuditReport{}); err == nil {
		t.Fatal("expected error when both clusterID and report.ClusterName are empty")
	}

	t0 := time.Now().Add(-24 * time.Hour).UTC()
	t1 := time.Now().UTC()

	snap0 := domain.AuditReport{
		ReportID:    "snap-t0",
		ClusterName: "prod-eu",
		GeneratedAt: t0,
		Summary: domain.AuditSummary{
			TotalWorkloadsAudited: 1,
			TotalContainers:       1,
			CurrentMonthlyCostUSD: 500.0,
			TotalMonthlyWasteUSD:  300.0,
			ClusterEfficiency:     45.0,
		},
		Containers: []domain.ContainerAnalysis{
			{
				Namespace:       "payments",
				WorkloadKind:    "Deployment",
				WorkloadName:    "checkout",
				ContainerName:   "api",
				ConfigHash:      "hash-v1",
				CurrentReqCPU:   2000,
				TargetReqCPU:    500,
				CurrentReqMem:   2 * 1024 * 1024 * 1024,
				TargetReqMem:    1 * 1024 * 1024 * 1024,
				Replicas:        2,
				MonthlyWasteUSD: 300.0,
				RiskStatus:      "OVERPROVISIONED",
			},
		},
	}

	snap1 := domain.AuditReport{
		ReportID:    "snap-t1",
		ClusterName: "prod-eu",
		GeneratedAt: t1,
		Summary: domain.AuditSummary{
			TotalWorkloadsAudited: 1,
			TotalContainers:       1,
			CurrentMonthlyCostUSD: 200.0,
			TotalMonthlyWasteUSD:  20.0,
			ClusterEfficiency:     88.0,
			IdleGPUCount:          1,
		},
		Containers: []domain.ContainerAnalysis{
			{
				Namespace:       "payments",
				WorkloadKind:    "Deployment",
				WorkloadName:    "checkout",
				ContainerName:   "api",
				ConfigHash:      "hash-v2",
				CurrentReqCPU:   500,
				TargetReqCPU:    450,
				CurrentReqMem:   1 * 1024 * 1024 * 1024,
				TargetReqMem:    1 * 1024 * 1024 * 1024,
				Replicas:        2,
				MonthlyWasteUSD: 20.0,
				RiskStatus:      "SAFE",
				GPUCount:        1,
				GPUUtilization:  5.0,
			},
		},
	}

	if err := store.SaveSnapshot("prod-eu", snap0); err != nil {
		t.Fatalf("failed saving snap0: %v", err)
	}
	if err := store.SaveSnapshot("prod-eu", snap1); err != nil {
		t.Fatalf("failed saving snap1: %v", err)
	}

	snaps := store.GetClusterSnapshots("prod-eu")
	if len(snaps) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(snaps))
	}

	latest := store.GetLatestSnapshots()
	if len(latest) != 1 || latest[0].ReportID != "snap-t1" {
		t.Fatalf("expected latest snapshot snap-t1, got %+v", latest)
	}

	clusters := store.GetClusters()
	if len(clusters) != 1 || clusters[0].RealizedSavingsUSD <= 0 {
		t.Fatalf("expected positive realized savings on cluster overview, got %+v", clusters)
	}

	kpis := store.GetExecutiveKPIs()
	if kpis.TotalClusters != 1 || kpis.RealizedMonthlySavings <= 0 {
		t.Fatalf("unexpected executive KPIs: %+v", kpis)
	}

	attr, err := store.GetAttribution("prod-eu")
	if err != nil {
		t.Fatalf("unexpected error getting attribution: %v", err)
	}
	if attr.OptimizedContainersCount != 1 {
		t.Errorf("expected 1 optimized container, got %d", attr.OptimizedContainersCount)
	}

	store.RecordAnomaly(domain.AnomalyFinding{
		ClusterID:     "prod-eu",
		Namespace:     "payments",
		WorkloadName:  "worker",
		ContainerName: "app",
		AnomalyType:   domain.StatusLeakSuspected,
		Severity:      "CRITICAL",
	})
	anomalies := store.GetAnomalies("prod-eu")
	if len(anomalies) != 2 {
		t.Errorf("expected 2 anomalies (1 Idle GPU + 1 Recorded Leak), got %d", len(anomalies))
	}

	store.SavePullRequest(domain.PullRequestRecord{
		ID:        "pr-1",
		ClusterID: "prod-eu",
		Status:    "OPEN",
	})
	if len(store.ListPullRequests()) != 1 {
		t.Errorf("expected 1 pull request record")
	}

	ddl := PostgresSchemaSQL()
	if !strings.Contains(ddl, "vector(384)") || !strings.Contains(ddl, "create_hypertable") {
		t.Errorf("expected pgvector(384) and timescaledb hypertable in DDL")
	}
}
