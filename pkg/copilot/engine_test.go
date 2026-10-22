package copilot

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/agusikprogramusik/kubetrim-hub/pkg/domain"
)

func sampleFleetContext() FleetContext {
	return FleetContext{
		KPIs: domain.ExecutiveKPIs{
			TotalClusters:          1,
			TotalWorkloadsAudited:  2,
			TotalContainersAudited: 2,
			FleetMonthlyCostUSD:    1200.0,
			IdentifiedMonthlyWaste: 450.0,
			IdentifiedYearlyWaste:  5400.0,
			RealizedMonthlySavings: 180.0,
			AverageEfficiencyScore: 68.5,
		},
		Snapshots: []domain.AuditReport{
			{
				ReportID:    "rep-001",
				ClusterName: "prod-eu",
				Containers: []domain.ContainerAnalysis{
					{
						Namespace:       "payments",
						WorkloadKind:    "Deployment",
						WorkloadName:    "checkout-api",
						ContainerName:   "api",
						CurrentReqCPU:   2000,
						TargetReqCPU:    400,
						MonthlyWasteUSD: 320.0,
						RiskStatus:      "OVERPROVISIONED",
					},
				},
			},
		},
		Anomalies: []domain.AnomalyFinding{
			{
				ClusterID:       "prod-eu",
				Namespace:       "ml",
				WorkloadName:    "gpu-infer",
				ContainerName:   "vllm",
				AnomalyType:     domain.StatusIdleGPU,
				Severity:        "CRITICAL",
				MonthlyWasteUSD: 600.0,
				Recommendation:  "Scale idle GPU workload to 0.",
			},
		},
	}
}

func TestCopilotDeterministicAndLLM(t *testing.T) {
	fleet := sampleFleetContext()

	// 1. Test empty query validation
	eng := NewEngine(Config{EnableLLM: false})
	if _, err := eng.Ask(context.Background(), ChatRequest{Query: "   "}, fleet); err == nil {
		t.Fatal("expected error for empty query")
	}

	// 2. Test deterministic RAG synthesis
	resp, err := eng.Ask(context.Background(), ChatRequest{
		Query:     "Which namespace generates the highest CPU waste?",
		ClusterID: "prod-eu",
	}, fleet)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(resp.Answer, "checkout-api") || !strings.Contains(resp.Answer, "$320.00/mo") {
		t.Errorf("expected deterministic answer to cite checkout-api and $320.00/mo, got: %s", resp.Answer)
	}
	if len(resp.Citations) < 2 {
		t.Errorf("expected at least 2 citations, got %v", resp.Citations)
	}

	// 3. Test OpenAI-compatible llama-server integration (Qwen3.5-9B-GGUF:Q8_0)
	llmSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{
					"message": map[string]string{
						"content": "Namespace `payments` (`checkout-api`) generates the highest waste at $320.00/mo.",
					},
				},
			},
		})
	}))
	defer llmSrv.Close()

	llmEng := NewEngine(Config{
		Endpoint:  llmSrv.URL + "/v1",
		ModelID:   DefaultModelID,
		APIKey:    "local-key",
		EnableLLM: true,
	})
	llmResp, err := llmEng.Ask(context.Background(), ChatRequest{
		Query: "Top wasted workload?",
	}, fleet)
	if err != nil {
		t.Fatalf("unexpected error from LLM engine: %v", err)
	}
	if llmResp.ModelUsed != DefaultModelID {
		t.Errorf("expected model %s, got %s", DefaultModelID, llmResp.ModelUsed)
	}
	if !strings.Contains(llmResp.Answer, "payments") {
		t.Errorf("unexpected LLM answer: %s", llmResp.Answer)
	}
}
