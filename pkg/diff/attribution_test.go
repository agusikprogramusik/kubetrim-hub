package diff

import (
	"testing"
	"time"

	"github.com/agusikprogramusik/kubetrim-hub/pkg/domain"
)

func TestEvaluateAttribution_VerifiedSavingsAndRegression(t *testing.T) {
	t0Time := time.Date(2026, 10, 15, 10, 0, 0, 0, time.UTC)
	t1Time := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)

	baseline := &domain.AuditReport{
		SchemaVersion: "v1",
		ReportID:      "snap-t0",
		ClusterName:   "eks-prod",
		GeneratedAt:   t0Time,
		Containers: []domain.ContainerAnalysis{
			{
				Namespace:     "payments",
				WorkloadKind:  "Deployment",
				WorkloadName:  "payment-api",
				ContainerName: "api",
				ConfigHash:    "hash-old-1",
				CurrentReqCPU: 2000,
				CurrentReqMem: 4 * 1024 * 1024 * 1024,
				Replicas:      3,
				RiskStatus:    "SAFE",
				OOMCount:      0,
				ThrottleRatio: 0.01,
				NodeArch:      "amd64",
			},
			{
				Namespace:     "orders",
				WorkloadKind:  "Deployment",
				WorkloadName:  "order-worker",
				ContainerName: "worker",
				ConfigHash:    "hash-old-2",
				CurrentReqCPU: 1000,
				CurrentReqMem: 2 * 1024 * 1024 * 1024,
				Replicas:      2,
				RiskStatus:    "SAFE",
				OOMCount:      0,
				ThrottleRatio: 0.02,
				NodeArch:      "arm64",
				IsSpotInstance: true,
			},
			{
				Namespace:     "auth",
				WorkloadKind:  "StatefulSet",
				WorkloadName:  "session-store",
				ContainerName: "redis",
				ConfigHash:    "hash-unchanged",
				CurrentReqCPU: 500,
				CurrentReqMem: 1024 * 1024 * 1024,
				Replicas:      1,
				RiskStatus:    "SAFE",
			},
		},
	}

	current := &domain.AuditReport{
		SchemaVersion: "v1",
		ReportID:      "snap-t1",
		ClusterName:   "eks-prod",
		GeneratedAt:   t1Time,
		Containers: []domain.ContainerAnalysis{
			{
				// Right-sized safely: 2000m -> 500m, 4Gi -> 1Gi
				Namespace:     "payments",
				WorkloadKind:  "Deployment",
				WorkloadName:  "payment-api",
				ContainerName: "api",
				ConfigHash:    "hash-new-1",
				CurrentReqCPU: 500,
				CurrentReqMem: 1 * 1024 * 1024 * 1024,
				Replicas:      3,
				RiskStatus:    "SAFE",
				OOMCount:      0,
				ThrottleRatio: 0.03,
				NodeArch:      "amd64",
			},
			{
				// Regressed after downsizing: OOMCount increased
				Namespace:     "orders",
				WorkloadKind:  "Deployment",
				WorkloadName:  "order-worker",
				ContainerName: "worker",
				ConfigHash:    "hash-new-2",
				CurrentReqCPU: 200,
				CurrentReqMem: 256 * 1024 * 1024,
				Replicas:      2,
				RiskStatus:    "RISK_OOM",
				OOMCount:      2,
				ThrottleRatio: 0.25,
				NodeArch:      "arm64",
				IsSpotInstance: true,
			},
			{
				// Unchanged
				Namespace:     "auth",
				WorkloadKind:  "StatefulSet",
				WorkloadName:  "session-store",
				ContainerName: "redis",
				ConfigHash:    "hash-unchanged",
				CurrentReqCPU: 500,
				CurrentReqMem: 1024 * 1024 * 1024,
				Replicas:      1,
				RiskStatus:    "SAFE",
			},
		},
	}

	attr, err := EvaluateAttribution(baseline, current)
	if err != nil {
		t.Fatalf("EvaluateAttribution failed: %v", err)
	}

	if attr.OptimizedContainersCount != 1 {
		t.Errorf("expected 1 optimized container, got %d", attr.OptimizedContainersCount)
	}
	if attr.RegressionAlertsCount != 1 {
		t.Errorf("expected 1 regression alert, got %d", attr.RegressionAlertsCount)
	}
	if attr.TotalRealizedMonthlyUSD <= 100.0 {
		t.Errorf("expected significant realized monthly savings > $100, got %.2f", attr.TotalRealizedMonthlyUSD)
	}

	// Verify individual container statuses
	statusByKey := make(map[string]domain.ContainerAttribution)
	for _, a := range attr.Attributions {
		statusByKey[a.Key] = a
	}

	payAttr := statusByKey["payments/Deployment/payment-api/api"]
	if payAttr.Status != domain.StatusVerifiedSavings || payAttr.RollbackRecommended {
		t.Errorf("unexpected payment-api attribution: %+v", payAttr)
	}

	ordAttr := statusByKey["orders/Deployment/order-worker/worker"]
	if ordAttr.Status != domain.StatusRegressionAlert || !ordAttr.RollbackRecommended {
		t.Errorf("expected REGRESSION_ALERT and rollback for order-worker, got: %+v", ordAttr)
	}

	authAttr := statusByKey["auth/StatefulSet/session-store/redis"]
	if authAttr.Status != domain.StatusUnchanged || authAttr.ConfigChanged {
		t.Errorf("expected UNCHANGED for session-store, got: %+v", authAttr)
	}
}

func TestEvaluateAttribution_NilInputsAndThrottleRegression(t *testing.T) {
	if _, err := EvaluateAttribution(nil, nil); err == nil {
		t.Fatal("expected error for nil inputs")
	}

	t0 := &domain.AuditReport{
		Containers: []domain.ContainerAnalysis{
			{
				Namespace: "default", WorkloadKind: "Deployment", WorkloadName: "app", ContainerName: "c",
				CurrentReqCPU: 1000, CurrentReqMem: 512 * 1024 * 1024, Replicas: 0, RiskStatus: "SAFE", ThrottleRatio: 0.01,
			},
		},
	}
	t1 := &domain.AuditReport{
		Containers: []domain.ContainerAnalysis{
			{
				Namespace: "default", WorkloadKind: "Deployment", WorkloadName: "app", ContainerName: "c",
				CurrentReqCPU: 200, CurrentReqMem: 512 * 1024 * 1024, Replicas: 0, RiskStatus: "SAFE", ThrottleRatio: 0.30,
			},
		},
	}

	res, err := EvaluateAttribution(t0, t1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.RegressionAlertsCount != 1 || !res.Attributions[0].RollbackRecommended {
		t.Errorf("expected throttle delta regression alert, got %+v", res.Attributions)
	}
}
