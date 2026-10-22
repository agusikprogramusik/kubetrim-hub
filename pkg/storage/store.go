package storage

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/agusikprogramusik/kubetrim-hub/pkg/anomaly"
	"github.com/agusikprogramusik/kubetrim-hub/pkg/diff"
	"github.com/agusikprogramusik/kubetrim-hub/pkg/domain"
)

// Store defines the multi-cluster FinOps persistence and query contract
type Store interface {
	SaveSnapshot(clusterID string, report domain.AuditReport) error
	GetLatestSnapshots() []domain.AuditReport
	GetClusterSnapshots(clusterID string) []domain.AuditReport
	GetClusters() []domain.ClusterOverview
	GetExecutiveKPIs() domain.ExecutiveKPIs
	GetAttribution(clusterID string) (*domain.AttributionReport, error)
	GetAnomalies(clusterID string) []domain.AnomalyFinding
	RecordAnomaly(finding domain.AnomalyFinding)
	SavePullRequest(record domain.PullRequestRecord)
	ListPullRequests() []domain.PullRequestRecord
}

// MemoryStore is a thread-safe in-memory time-series store for multi-cluster FinOps snapshots
type MemoryStore struct {
	mu           sync.RWMutex
	snapshots    map[string][]domain.AuditReport
	customAlerts map[string][]domain.AnomalyFinding
	prs          []domain.PullRequestRecord
}

// NewMemoryStore initializes a new thread-safe MemoryStore
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		snapshots:    make(map[string][]domain.AuditReport),
		customAlerts: make(map[string][]domain.AnomalyFinding),
		prs:          make([]domain.PullRequestRecord, 0),
	}
}

// SaveSnapshot appends a new AuditReport snapshot for the specified cluster
func (s *MemoryStore) SaveSnapshot(clusterID string, report domain.AuditReport) error {
	if clusterID == "" {
		clusterID = report.ClusterName
	}
	if clusterID == "" {
		return fmt.Errorf("cluster_id is required")
	}
	report.ClusterName = clusterID
	if report.ReportID == "" {
		report.ReportID = fmt.Sprintf("snap-%d", time.Now().UnixNano())
	}
	if report.GeneratedAt.IsZero() {
		report.GeneratedAt = time.Now().UTC()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	list := append(s.snapshots[clusterID], report)
	sort.Slice(list, func(i, j int) bool {
		return list[i].GeneratedAt.Before(list[j].GeneratedAt)
	})
	s.snapshots[clusterID] = list
	return nil
}

// GetClusterSnapshots returns all chronological snapshots for a cluster
func (s *MemoryStore) GetClusterSnapshots(clusterID string) []domain.AuditReport {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := s.snapshots[clusterID]
	out := make([]domain.AuditReport, len(list))
	copy(out, list)
	return out
}

// GetLatestSnapshots returns the latest snapshot for each registered cluster
func (s *MemoryStore) GetLatestSnapshots() []domain.AuditReport {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var latest []domain.AuditReport
	for _, list := range s.snapshots {
		if len(list) > 0 {
			latest = append(latest, list[len(list)-1])
		}
	}
	sort.Slice(latest, func(i, j int) bool {
		return latest[i].ClusterName < latest[j].ClusterName
	})
	return latest
}

// GetClusters returns a summary of all audited clusters
func (s *MemoryStore) GetClusters() []domain.ClusterOverview {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]domain.ClusterOverview, 0, len(s.snapshots))
	for clusterID, list := range s.snapshots {
		if len(list) == 0 {
			continue
		}
		latest := list[len(list)-1]
		realized := 0.0
		if len(list) >= 2 {
			if attr, err := diff.EvaluateAttribution(&list[0], &latest); err == nil {
				realized = attr.TotalRealizedMonthlyUSD
			}
		}
		out = append(out, domain.ClusterOverview{
			ClusterID:             clusterID,
			LatestSnapshotID:      latest.ReportID,
			LastSeenAt:            latest.GeneratedAt,
			SnapshotCount:         len(list),
			TotalWorkloads:        latest.Summary.TotalWorkloadsAudited,
			TotalContainers:       latest.Summary.TotalContainers,
			CurrentMonthlyCostUSD: latest.Summary.CurrentMonthlyCostUSD,
			TotalMonthlyWasteUSD:  latest.Summary.TotalMonthlyWasteUSD,
			RealizedSavingsUSD:    realized,
			EfficiencyScore:       latest.Summary.ClusterEfficiency,
			IdleGPUCount:          latest.Summary.IdleGPUCount,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].ClusterID < out[j].ClusterID
	})
	return out
}

// GetExecutiveKPIs computes fleet-wide FinOps KPIs across all clusters
func (s *MemoryStore) GetExecutiveKPIs() domain.ExecutiveKPIs {
	s.mu.RLock()
	defer s.mu.RUnlock()

	kpis := domain.ExecutiveKPIs{
		RiskBreakdown: make(map[string]int),
	}
	var effSum float64

	for _, list := range s.snapshots {
		if len(list) == 0 {
			continue
		}
		kpis.TotalClusters++
		latest := list[len(list)-1]
		kpis.TotalWorkloadsAudited += latest.Summary.TotalWorkloadsAudited
		kpis.TotalContainersAudited += latest.Summary.TotalContainers
		kpis.FleetMonthlyCostUSD += latest.Summary.CurrentMonthlyCostUSD
		kpis.IdentifiedMonthlyWaste += latest.Summary.TotalMonthlyWasteUSD
		kpis.IdleGPUAlerts += latest.Summary.IdleGPUCount
		effSum += latest.Summary.ClusterEfficiency

		for risk, count := range latest.Summary.RiskBreakdown {
			kpis.RiskBreakdown[risk] += count
		}
		if len(latest.Summary.RiskBreakdown) == 0 {
			for _, c := range latest.Containers {
				if c.RiskStatus != "" {
					kpis.RiskBreakdown[c.RiskStatus]++
				}
			}
		}

		if len(list) >= 2 {
			if attr, err := diff.EvaluateAttribution(&list[0], &latest); err == nil {
				kpis.RealizedMonthlySavings += attr.TotalRealizedMonthlyUSD
				kpis.SLARegressionAlerts += attr.RegressionAlertsCount
			}
		}
	}

	kpis.FleetMonthlyCostUSD = math.Round(kpis.FleetMonthlyCostUSD*100) / 100
	kpis.IdentifiedMonthlyWaste = math.Round(kpis.IdentifiedMonthlyWaste*100) / 100
	kpis.IdentifiedYearlyWaste = math.Round(kpis.IdentifiedMonthlyWaste*12*100) / 100
	kpis.RealizedMonthlySavings = math.Round(kpis.RealizedMonthlySavings*100) / 100
	kpis.RealizedYearlySavings = math.Round(kpis.RealizedMonthlySavings*12*100) / 100
	if kpis.TotalClusters > 0 {
		kpis.AverageEfficiencyScore = math.Round((effSum/float64(kpis.TotalClusters))*10) / 10
	}

	return kpis
}

// GetAttribution evaluates the T0 -> T1 Before/After impact for a specific cluster
func (s *MemoryStore) GetAttribution(clusterID string) (*domain.AttributionReport, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := s.snapshots[clusterID]
	if len(list) == 0 {
		return nil, fmt.Errorf("no snapshots found for cluster %q", clusterID)
	}
	if len(list) == 1 {
		return diff.EvaluateAttribution(&list[0], &list[0])
	}
	return diff.EvaluateAttribution(&list[0], &list[len(list)-1])
}

// GetAnomalies returns all detected anomalies (Idle GPU + Sawtooth Memory Leaks) for a cluster (or all clusters if empty)
func (s *MemoryStore) GetAnomalies(clusterID string) []domain.AnomalyFinding {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var findings []domain.AnomalyFinding
	for cid, list := range s.snapshots {
		if clusterID != "" && cid != clusterID {
			continue
		}
		if len(list) > 0 {
			latest := list[len(list)-1]
			findings = append(findings, anomaly.DetectClusterAnomalies(&latest)...)
		}
		findings = append(findings, s.customAlerts[cid]...)
	}
	return findings
}

// RecordAnomaly stores an asynchronously detected anomaly (such as a 14-day sawtooth memory leak)
func (s *MemoryStore) RecordAnomaly(finding domain.AnomalyFinding) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.customAlerts[finding.ClusterID] = append(s.customAlerts[finding.ClusterID], finding)
}

// SavePullRequest stores a synthesized GitOps Pull Request record
func (s *MemoryStore) SavePullRequest(record domain.PullRequestRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.prs = append(s.prs, record)
}

// ListPullRequests returns all recorded GitOps Pull Requests
func (s *MemoryStore) ListPullRequests() []domain.PullRequestRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.PullRequestRecord, len(s.prs))
	copy(out, s.prs)
	return out
}

// PostgresSchemaSQL returns the PostgreSQL 18 + TimescaleDB + pgvector(384) + RLS DDL schema
func PostgresSchemaSQL() string {
	return `
CREATE EXTENSION IF NOT EXISTS timescaledb;
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS clusters (
    cluster_id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL DEFAULT 'default',
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS snapshots (
    report_id TEXT NOT NULL,
    cluster_id TEXT NOT NULL REFERENCES clusters(cluster_id),
    tenant_id TEXT NOT NULL DEFAULT 'default',
    generated_at TIMESTAMPTZ NOT NULL,
    git_commit TEXT,
    payload JSONB NOT NULL,
    embedding vector(384),
    PRIMARY KEY (cluster_id, generated_at, report_id)
);

SELECT create_hypertable('snapshots', 'generated_at', if_not_exists => TRUE);

ALTER TABLE snapshots ENABLE ROW LEVEL SECURITY;
`
}
