package notifications

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/agusikprogramusik/kubetrim-hub/pkg/domain"
)

func TestDispatcherWebhooksAndFormatting(t *testing.T) {
	var slackCalls int32
	var teamsCalls int32

	slackSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&slackCalls, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer slackSrv.Close()

	teamsSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&teamsCalls, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer teamsSrv.Close()

	d := NewDispatcher(Config{
		SlackWebhookURL: slackSrv.URL,
		TeamsWebhookURL: teamsSrv.URL,
	})

	kpis := domain.ExecutiveKPIs{
		TotalClusters:          3,
		TotalWorkloadsAudited:  42,
		TotalContainersAudited: 68,
		FleetMonthlyCostUSD:    4200.0,
		AverageEfficiencyScore: 76.5,
		RealizedMonthlySavings: 850.0,
		RealizedYearlySavings:  10200.0,
		IdentifiedMonthlyWaste: 620.0,
		IdentifiedYearlyWaste:  7440.0,
		IdleGPUAlerts:          1,
		SLARegressionAlerts:    0,
	}

	if err := d.SendWeeklyDigest(context.Background(), kpis); err != nil {
		t.Fatalf("unexpected error sending weekly digest: %v", err)
	}

	finding := domain.AnomalyFinding{
		ClusterID:       "prod-eu-central-1",
		Namespace:       "ml-ops",
		WorkloadName:    "vllm-worker",
		ContainerName:   "inference",
		AnomalyType:     domain.StatusIdleGPU,
		Severity:        "CRITICAL",
		MonthlyWasteUSD: 720.0,
		Recommendation:  "Scale GPU deployment to 0 or migrate to time-sliced MIG profile.",
	}
	if err := d.SendAnomalyAlert(context.Background(), finding); err != nil {
		t.Fatalf("unexpected error sending anomaly alert: %v", err)
	}

	attr := domain.ContainerAttribution{
		Key:                 "payments/Deployment/checkout/api",
		OOMDelta:            3,
		ThrottleDelta:       0.18,
		Status:              domain.StatusRegressionAlert,
		RollbackRecommended: true,
	}
	if err := d.SendRegressionAlert(context.Background(), "prod-eu-central-1", attr); err != nil {
		t.Fatalf("unexpected error sending regression alert: %v", err)
	}

	if atomic.LoadInt32(&slackCalls) != 3 {
		t.Errorf("expected 3 slack calls, got %d", slackCalls)
	}
	if atomic.LoadInt32(&teamsCalls) != 3 {
		t.Errorf("expected 3 teams calls, got %d", teamsCalls)
	}

	digestText := FormatWeeklyDigest(kpis)
	if !strings.Contains(digestText, "$850.00/mo") {
		t.Errorf("expected realized savings in digest text, got %s", digestText)
	}
}

func TestDispatcherErrorHandling(t *testing.T) {
	failSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer failSrv.Close()

	dSlack := NewDispatcher(Config{SlackWebhookURL: failSrv.URL})
	if err := dSlack.SendWeeklyDigest(context.Background(), domain.ExecutiveKPIs{}); err == nil {
		t.Fatal("expected error when slack webhook returns 502")
	}

	dTeams := NewDispatcher(Config{TeamsWebhookURL: failSrv.URL})
	if err := dTeams.SendWeeklyDigest(context.Background(), domain.ExecutiveKPIs{}); err == nil {
		t.Fatal("expected error when teams webhook returns 502")
	}
}
