package gitops

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/agusikprogramusik/kubetrim-hub/pkg/domain"
)

func sampleContainer() domain.ContainerAnalysis {
	return domain.ContainerAnalysis{
		Namespace:       "payments",
		WorkloadKind:    "Deployment",
		WorkloadName:    "payment-api",
		ContainerName:   "api",
		ConfigHash:      "sha256-xyz",
		CurrentReqCPU:   1000,
		TargetReqCPU:    250,
		CurrentLimitCPU: 2000,
		TargetLimitCPU:  500,
		CurrentReqMem:   1024 * 1024 * 1024,
		TargetReqMem:    512 * 1024 * 1024,
		CurrentLimitMem: 2048 * 1024 * 1024,
		TargetLimitMem:  1024 * 1024 * 1024,
		MonthlyWasteUSD: 420.00,
		RiskStatus:      "SAFE",
		CurrentQoS:      "Burstable",
		ProjectedQoS:    "Burstable",
	}
}

func TestSynthesizeKustomizePatchAndDescription(t *testing.T) {
	c := sampleContainer()
	patch := SynthesizeKustomizePatch(c)
	if !strings.Contains(patch, `cpu: "250m"`) || !strings.Contains(patch, `memory: "512Mi"`) {
		t.Errorf("unexpected kustomize patch:\n%s", patch)
	}

	desc := SynthesizePRDescription("eks-prod", c)
	if !strings.Contains(desc, "$420.00 / month") || !strings.Contains(desc, "sha256-xyz") {
		t.Errorf("unexpected PR description:\n%s", desc)
	}
}

func TestBot_CreatePullRequest_DryRunAndLiveGitHub(t *testing.T) {
	t.Run("validation error on missing workload", func(t *testing.T) {
		bot := NewBot(BotConfig{})
		if _, err := bot.CreatePullRequest(context.Background(), domain.PullRequestRequest{}); err == nil {
			t.Fatal("expected error for empty container")
		}
	})

	t.Run("dry run mode succeeds without network call", func(t *testing.T) {
		bot := NewBot(BotConfig{})
		rec, err := bot.CreatePullRequest(context.Background(), domain.PullRequestRequest{
			ClusterID:  "eks-prod",
			Repository: "agusikprogramusik/env",
			Container:  sampleContainer(),
			DryRun:     true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Status != "DRY_RUN" || !strings.Contains(rec.KustomizePatch, "payment-api") {
			t.Errorf("unexpected dry-run record: %+v", rec)
		}
	})

	t.Run("live GitHub API creates commit and opens PR", func(t *testing.T) {
		var putCalled, postPullsCalled bool
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPut && strings.Contains(r.URL.Path, "/contents/") {
				putCalled = true
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"content":{"path":"patches/patch-payments-payment-api.yaml"}}`))
				return
			}
			if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/pulls") {
				postPullsCalled = true
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"html_url":"https://github.com/agusikprogramusik/env/pull/42"}`))
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()

		bot := NewBot(BotConfig{
			GitHubAPIBaseURL: srv.URL,
			GitHubToken:      "ghp_test_token",
		})
		rec, err := bot.CreatePullRequest(context.Background(), domain.PullRequestRequest{
			ClusterID:  "eks-prod",
			Repository: "agusikprogramusik/env",
			Container:  sampleContainer(),
			DryRun:     false,
		})
		if err != nil {
			t.Fatalf("CreatePullRequest failed: %v", err)
		}
		if !putCalled || !postPullsCalled {
			t.Errorf("expected both PUT contents and POST pulls to be called (put=%t, post=%t)", putCalled, postPullsCalled)
		}
		if rec.Status != "OPEN" || rec.PRURL != "https://github.com/agusikprogramusik/env/pull/42" {
			t.Errorf("unexpected PR record: %+v", rec)
		}
	})
}
