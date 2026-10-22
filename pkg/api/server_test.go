package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/agusikprogramusik/kubetrim-hub/pkg/copilot"
	"github.com/agusikprogramusik/kubetrim-hub/pkg/domain"
)

func TestAPIServerEndToEnd(t *testing.T) {
	srv := NewServer(ServerConfig{
		APIKey: "secret-hub-key",
	})
	router := srv.Router()

	// 1. Health & Ready endpoints (no auth required)
	for _, path := range []string{"/healthz", "/ready"} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 on %s, got %d", path, w.Code)
		}
	}

	// 2. Unauthorized request should return 401
	wUnauth := httptest.NewRecorder()
	reqUnauth := httptest.NewRequest(http.MethodGet, "/api/v1/executive/kpis", nil)
	router.ServeHTTP(wUnauth, reqUnauth)
	if wUnauth.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without API key, got %d", wUnauth.Code)
	}

	// 3. Ingest T0 and T1 snapshots (including regression alert trigger)
	t0 := time.Now().Add(-2 * time.Hour).UTC()
	t1 := time.Now().UTC()

	snap0 := domain.AuditReport{
		SchemaVersion: "v1",
		ReportID:      "snap-01",
		ClusterName:   "prod-eu-central-1",
		GeneratedAt:   t0,
		Summary: domain.AuditSummary{
			TotalWorkloadsAudited: 1,
			TotalContainers:       1,
			CurrentMonthlyCostUSD: 400.0,
			TotalMonthlyWasteUSD:  250.0,
			ClusterEfficiency:     42.0,
		},
		Containers: []domain.ContainerAnalysis{
			{
				Namespace:       "payments",
				WorkloadKind:    "Deployment",
				WorkloadName:    "checkout",
				ContainerName:   "api",
				ConfigHash:      "h1",
				CurrentReqCPU:   2000,
				TargetReqCPU:    500,
				CurrentReqMem:   2 * 1024 * 1024 * 1024,
				TargetReqMem:    1 * 1024 * 1024 * 1024,
				Replicas:        2,
				MonthlyWasteUSD: 250.0,
				RiskStatus:      "OVERPROVISIONED",
			},
		},
	}

	snap1 := domain.AuditReport{
		SchemaVersion: "v1",
		ClusterName:   "", // test fallback to X-Cluster-ID header
		Summary: domain.AuditSummary{
			TotalWorkloadsAudited: 1,
			TotalContainers:       1,
			CurrentMonthlyCostUSD: 150.0,
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
				ConfigHash:      "h2",
				CurrentReqCPU:   500,
				TargetReqCPU:    450,
				CurrentReqMem:   1 * 1024 * 1024 * 1024,
				TargetReqMem:    1 * 1024 * 1024 * 1024,
				Replicas:        2,
				MonthlyWasteUSD: 20.0,
				RiskStatus:      "RISK_OOM",
				OOMCount:        2,
				GPUCount:        1,
				GPUUtilization:  4.0,
			},
		},
	}
	_ = t1

	b0, _ := json.Marshal(snap0)
	w0 := httptest.NewRecorder()
	req0 := httptest.NewRequest(http.MethodPost, "/api/v1/clusters/prod-eu-central-1/snapshots", bytes.NewReader(b0))
	req0.Header.Set("Authorization", "Bearer secret-hub-key")
	req0.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w0, req0)
	if w0.Code != http.StatusAccepted {
		t.Fatalf("expected 202 Accepted, got %d: %s", w0.Code, w0.Body.String())
	}

	b1, _ := json.Marshal(snap1)
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/ingest", bytes.NewReader(b1))
	req1.Header.Set("Authorization", "Bearer secret-hub-key")
	req1.Header.Set("X-Cluster-ID", "prod-eu-central-1")
	req1.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusAccepted {
		t.Fatalf("expected 202 Accepted on /api/v1/ingest, got %d: %s", w1.Code, w1.Body.String())
	}

	// Wait briefly for async post-ingest alert evaluation goroutine
	time.Sleep(25 * time.Millisecond)

	// 4. Verify Executive KPIs, Clusters, Attribution, Anomalies
	for _, ep := range []string{
		"/api/v1/executive/kpis",
		"/api/v1/clusters",
		"/api/v1/clusters/prod-eu-central-1/snapshots",
		"/api/v1/clusters/prod-eu-central-1/attribution",
		"/api/v1/clusters/prod-eu-central-1/anomalies",
	} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, ep, nil)
		req.Header.Set("X-API-Key", "secret-hub-key")
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 on %s, got %d: %s", ep, w.Code, w.Body.String())
		}
	}

	// 5. Create 1-Click GitOps PR (DryRun) and list PRs
	prReq := domain.PullRequestRequest{
		ClusterID:  "prod-eu-central-1",
		Repository: "agusikprogramusik/env",
		DryRun:     true,
		Container:  snap0.Containers[0],
	}
	prBytes, _ := json.Marshal(prReq)
	wPR := httptest.NewRecorder()
	reqPR := httptest.NewRequest(http.MethodPost, "/api/v1/gitops/pr", bytes.NewReader(prBytes))
	reqPR.Header.Set("Authorization", "Bearer secret-hub-key")
	router.ServeHTTP(wPR, reqPR)
	if wPR.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for GitOps PR, got %d: %s", wPR.Code, wPR.Body.String())
	}

	wListPR := httptest.NewRecorder()
	reqListPR := httptest.NewRequest(http.MethodGet, "/api/v1/gitops/prs", nil)
	reqListPR.Header.Set("Authorization", "Bearer secret-hub-key")
	router.ServeHTTP(wListPR, reqListPR)
	if wListPR.Code != http.StatusOK {
		t.Fatalf("expected 200 on /api/v1/gitops/prs, got %d", wListPR.Code)
	}

	// 6. Test FinOps Copilot Chat endpoint
	chatBytes, _ := json.Marshal(copilot.ChatRequest{
		Query:     "What is our realized savings on prod-eu-central-1?",
		ClusterID: "prod-eu-central-1",
	})
	wChat := httptest.NewRecorder()
	reqChat := httptest.NewRequest(http.MethodPost, "/api/v1/copilot/chat", bytes.NewReader(chatBytes))
	reqChat.Header.Set("Authorization", "Bearer secret-hub-key")
	router.ServeHTTP(wChat, reqChat)
	if wChat.Code != http.StatusOK {
		t.Fatalf("expected 200 on /api/v1/copilot/chat, got %d: %s", wChat.Code, wChat.Body.String())
	}

	// 7. Test validation & error branches (open mode, 400s, 404s)
	openSrv := NewServer(ServerConfig{APIKey: ""})
	openRouter := openSrv.Router()

	w404 := httptest.NewRecorder()
	openRouter.ServeHTTP(w404, httptest.NewRequest(http.MethodGet, "/api/v1/clusters/missing/attribution", nil))
	if w404.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing cluster attribution, got %d", w404.Code)
	}

	for _, badEp := range []string{"/api/v1/ingest", "/api/v1/gitops/pr", "/api/v1/copilot/chat"} {
		wBad := httptest.NewRecorder()
		openRouter.ServeHTTP(wBad, httptest.NewRequest(http.MethodPost, badEp, bytes.NewReader([]byte(`{}`))))
		if wBad.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 on %s with empty body, got %d", badEp, wBad.Code)
		}
	}
}
