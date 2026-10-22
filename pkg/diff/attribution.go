package diff

import (
	"fmt"
	"math"

	"github.com/agusikprogramusik/kubetrim-hub/pkg/domain"
)

const (
	// DefaultCPUHourlyRateUSD is the standard baseline rate per vCPU-hour ($0.035)
	DefaultCPUHourlyRateUSD = 0.035
	// DefaultMemGiBHourlyRateUSD is the standard baseline rate per GiB-hour ($0.0045)
	DefaultMemGiBHourlyRateUSD = 0.0045
	// HoursPerMonth is the standard FinOps 730h billing month
	HoursPerMonth = 730.0
	// MaxSafeThrottleDelta is the maximum allowed increase in CFS CPU throttling ratio (15%)
	MaxSafeThrottleDelta = 0.15
)

// EvaluateAttribution compares a baseline snapshot (T0) against a current snapshot (T1)
// to verify realized financial savings and detect SLA regressions (OOMKills / CPU Throttling).
func EvaluateAttribution(baseline, current *domain.AuditReport) (*domain.AttributionReport, error) {
	if baseline == nil || current == nil {
		return nil, fmt.Errorf("both baseline (T0) and current (T1) AuditReports are required")
	}

	clusterID := current.ClusterName
	if clusterID == "" {
		clusterID = baseline.ClusterName
	}

	baseMap := make(map[string]domain.ContainerAnalysis, len(baseline.Containers))
	for _, c := range baseline.Containers {
		baseMap[c.ContainerKey()] = c
	}

	report := &domain.AttributionReport{
		ClusterID:          clusterID,
		BaselineSnapshotID: baseline.ReportID,
		CurrentSnapshotID:  current.ReportID,
		BaselineTimestamp:  baseline.GeneratedAt,
		CurrentTimestamp:   current.GeneratedAt,
		Attributions:       make([]domain.ContainerAttribution, 0, len(current.Containers)),
	}

	for _, t1 := range current.Containers {
		key := t1.ContainerKey()
		t0, existed := baseMap[key]
		if !existed {
			continue
		}

		changed := t0.ConfigHash != "" && t1.ConfigHash != "" && t0.ConfigHash != t1.ConfigHash
		if !changed && (t0.CurrentReqCPU != t1.CurrentReqCPU || t0.CurrentReqMem != t1.CurrentReqMem || t0.Replicas != t1.Replicas) {
			changed = true
		}

		costT0 := estimateContainerMonthlyCost(t0)
		costT1 := estimateContainerMonthlyCost(t1)
		savings := 0.0
		if changed {
			savings = math.Round((costT0-costT1)*100) / 100
		}

		oomDelta := t1.OOMCount - t0.OOMCount
		throttleDelta := math.Round((t1.ThrottleRatio-t0.ThrottleRatio)*10000) / 10000

		status := domain.StatusUnchanged
		rollback := false

		if changed {
			if isSLARegression(t0, t1, oomDelta, throttleDelta) {
				status = domain.StatusRegressionAlert
				rollback = true
				report.RegressionAlertsCount++
			} else if savings > 0 {
				status = domain.StatusVerifiedSavings
				report.OptimizedContainersCount++
				report.TotalRealizedMonthlyUSD += savings
			}
		}

		report.Attributions = append(report.Attributions, domain.ContainerAttribution{
			Key:                       key,
			Namespace:                 t1.Namespace,
			WorkloadKind:              t1.WorkloadKind,
			WorkloadName:              t1.WorkloadName,
			ContainerName:             t1.ContainerName,
			BeforeConfigHash:          t0.ConfigHash,
			AfterConfigHash:           t1.ConfigHash,
			ConfigChanged:             changed,
			BeforeCPUReqMilli:         t0.CurrentReqCPU,
			AfterCPUReqMilli:          t1.CurrentReqCPU,
			BeforeMemReqBytes:         t0.CurrentReqMem,
			AfterMemReqBytes:          t1.CurrentReqMem,
			RealizedMonthlySavingsUSD: savings,
			OOMDelta:                  oomDelta,
			ThrottleDelta:             throttleDelta,
			Status:                    status,
			RollbackRecommended:       rollback,
		})
	}

	report.TotalRealizedMonthlyUSD = math.Round(report.TotalRealizedMonthlyUSD*100) / 100
	report.TotalRealizedYearlyUSD = math.Round(report.TotalRealizedMonthlyUSD*12*100) / 100

	return report, nil
}

func isSLARegression(t0, t1 domain.ContainerAnalysis, oomDelta int, throttleDelta float64) bool {
	if oomDelta > 0 {
		return true
	}
	if throttleDelta > MaxSafeThrottleDelta {
		return true
	}
	if t0.RiskStatus == "SAFE" && (t1.RiskStatus == "RISK_OOM" || t1.RiskStatus == "THROTTLED") {
		return true
	}
	return false
}

func estimateContainerMonthlyCost(c domain.ContainerAnalysis) float64 {
	replicas := float64(c.Replicas)
	if replicas <= 0 {
		replicas = 1
	}

	vCPU := float64(c.CurrentReqCPU) / 1000.0
	memGiB := float64(c.CurrentReqMem) / (1024.0 * 1024.0 * 1024.0)

	archMultiplier := 1.0
	if c.NodeArch == "arm64" {
		archMultiplier = 0.80
	}
	spotMultiplier := 1.0
	if c.IsSpotInstance {
		spotMultiplier = 0.30
	}

	hourly := (vCPU*DefaultCPUHourlyRateUSD + memGiB*DefaultMemGiBHourlyRateUSD) * archMultiplier * spotMultiplier
	return hourly * HoursPerMonth * replicas
}
