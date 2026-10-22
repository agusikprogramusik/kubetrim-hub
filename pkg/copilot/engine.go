package copilot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/agusikprogramusik/kubetrim-hub/pkg/domain"
)

const (
	DefaultLlamaServerEndpoint = "http://localhost:8080/v1"
	DefaultModelID             = "unsloth/Qwen3.5-9B-GGUF:Q8_0"
)

// Config configures the FinOps RAG & Copilot Engine
type Config struct {
	Endpoint   string
	ModelID    string
	APIKey     string
	EnableLLM  bool
	HTTPClient *http.Client
}

// ChatRequest represents a natural-language FinOps query from the user
type ChatRequest struct {
	Query     string `json:"query"`
	ClusterID string `json:"cluster_id,omitempty"`
}

// ChatResponse returns a grounded FinOps answer along with telemetry citations
type ChatResponse struct {
	Answer     string   `json:"answer"`
	ModelUsed  string   `json:"model_used"`
	Citations  []string `json:"citations"`
	Confidence float64  `json:"confidence"`
}

// FleetContext holds the grounded snapshot data passed into the Copilot engine
type FleetContext struct {
	KPIs       domain.ExecutiveKPIs
	Clusters   []domain.ClusterOverview
	Snapshots  []domain.AuditReport
	Anomalies  []domain.AnomalyFinding
}

// Engine executes grounded FinOps RAG and natural-language queries
type Engine struct {
	cfg    Config
	client *http.Client
}

// NewEngine creates a new FinOps Copilot Engine
func NewEngine(cfg Config) *Engine {
	if cfg.Endpoint == "" {
		cfg.Endpoint = DefaultLlamaServerEndpoint
	}
	if cfg.ModelID == "" {
		cfg.ModelID = DefaultModelID
	}
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	return &Engine{
		cfg:    cfg,
		client: client,
	}
}

// Ask answers a natural-language FinOps question grounded in ingested cluster snapshots
func (e *Engine) Ask(ctx context.Context, req ChatRequest, fleet FleetContext) (ChatResponse, error) {
	if strings.TrimSpace(req.Query) == "" {
		return ChatResponse{}, fmt.Errorf("query cannot be empty")
	}

	groundedPrompt, citations := buildGroundedContext(req, fleet)

	if e.cfg.EnableLLM {
		ans, err := e.callOpenAICompatible(ctx, req.Query, groundedPrompt)
		if err == nil && strings.TrimSpace(ans) != "" {
			return ChatResponse{
				Answer:     strings.TrimSpace(ans),
				ModelUsed:  e.cfg.ModelID,
				Citations:  citations,
				Confidence: 0.94,
			}, nil
		}
	}

	// Deterministic FinOps synthesis fallback grounded in real telemetry
	return ChatResponse{
		Answer:     synthesizeDeterministicAnswer(req, fleet),
		ModelUsed:  "kubetrim-deterministic-rag-v1 (" + e.cfg.ModelID + ")",
		Citations:  citations,
		Confidence: 0.98,
	}, nil
}

func buildGroundedContext(req ChatRequest, fleet FleetContext) (string, []string) {
	var sb strings.Builder
	var citations []string

	sb.WriteString(fmt.Sprintf(
		"Fleet Summary: %d clusters, %d workloads, Fleet Spend: $%.2f/mo, Identified Waste: $%.2f/mo ($%.2f/yr), Realized Savings: $%.2f/mo, Efficiency: %.1f%%.\n",
		fleet.KPIs.TotalClusters,
		fleet.KPIs.TotalWorkloadsAudited,
		fleet.KPIs.FleetMonthlyCostUSD,
		fleet.KPIs.IdentifiedMonthlyWaste,
		fleet.KPIs.IdentifiedYearlyWaste,
		fleet.KPIs.RealizedMonthlySavings,
		fleet.KPIs.AverageEfficiencyScore,
	))

	type offender struct {
		Cluster string
		Cont    domain.ContainerAnalysis
	}
	var offenders []offender
	for _, snap := range fleet.Snapshots {
		if req.ClusterID != "" && snap.ClusterName != req.ClusterID {
			continue
		}
		citations = append(citations, fmt.Sprintf("snapshot:%s@%s", snap.ClusterName, snap.ReportID))
		for _, c := range snap.Containers {
			offenders = append(offenders, offender{Cluster: snap.ClusterName, Cont: c})
		}
	}

	sort.Slice(offenders, func(i, j int) bool {
		return offenders[i].Cont.MonthlyWasteUSD > offenders[j].Cont.MonthlyWasteUSD
	})

	limit := 5
	if len(offenders) < limit {
		limit = len(offenders)
	}
	if limit > 0 {
		sb.WriteString("Top Wasted Containers:\n")
		for i := 0; i < limit; i++ {
			o := offenders[i]
			sb.WriteString(fmt.Sprintf(
				"- [%s] %s/%s (%s): $%.2f/mo waste, Risk=%s, CPU %dm->%dm\n",
				o.Cluster, o.Cont.Namespace, o.Cont.WorkloadName, o.Cont.ContainerName,
				o.Cont.MonthlyWasteUSD, o.Cont.RiskStatus, o.Cont.CurrentReqCPU, o.Cont.TargetReqCPU,
			))
			citations = append(citations, fmt.Sprintf("workload:%s/%s/%s", o.Cluster, o.Cont.Namespace, o.Cont.WorkloadName))
		}
	}

	for _, a := range fleet.Anomalies {
		if req.ClusterID != "" && a.ClusterID != req.ClusterID {
			continue
		}
		sb.WriteString(fmt.Sprintf(
			"Anomaly [%s] on %s/%s/%s: $%.2f/mo waste — %s\n",
			a.AnomalyType, a.ClusterID, a.Namespace, a.WorkloadName, a.MonthlyWasteUSD, a.Recommendation,
		))
	}

	return sb.String(), citations
}

func synthesizeDeterministicAnswer(req ChatRequest, fleet FleetContext) string {
	contextSummary, _ := buildGroundedContext(req, fleet)
	return fmt.Sprintf(
		"Based on KubeTrim Hub telemetry for query %q:\n\n%s",
		req.Query,
		strings.TrimSpace(contextSummary),
	)
}

func (e *Engine) callOpenAICompatible(ctx context.Context, query, groundedContext string) (string, error) {
	url := strings.TrimRight(e.cfg.Endpoint, "/") + "/chat/completions"
	payload := map[string]any{
		"model":       e.cfg.ModelID,
		"temperature": 0.1,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are KubeTrim FinOps Copilot. Answer concisely in English using ONLY the provided Kubernetes FinOps telemetry context.",
			},
			{
				"role":    "user",
				"content": "Telemetry Context:\n" + groundedContext + "\n\nQuestion: " + query,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if e.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+e.cfg.APIKey)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("llm status %d", resp.StatusCode)
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("empty choices")
	}
	return parsed.Choices[0].Message.Content, nil
}
