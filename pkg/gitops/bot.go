package gitops

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/agusikprogramusik/kubetrim-hub/pkg/domain"
	"github.com/google/uuid"
)

// BotConfig configures the 1-Click GitOps Pull Request automation service
type BotConfig struct {
	GitHubAPIBaseURL string
	GitHubToken      string
	HTTPClient       *http.Client
}

// Bot synthesizes Kustomize v5 strategic merge patches and opens GitHub Pull Requests
type Bot struct {
	apiBaseURL string
	token      string
	httpClient *http.Client
}

// NewBot constructs a new GitOps PR automation bot
func NewBot(cfg BotConfig) *Bot {
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.GitHubAPIBaseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &Bot{
		apiBaseURL: baseURL,
		token:      strings.TrimSpace(cfg.GitHubToken),
		httpClient: client,
	}
}

// SynthesizeKustomizePatch generates a Kustomize v5 strategic-merge patch for a container
func SynthesizeKustomizePatch(c domain.ContainerAnalysis) string {
	kind := c.WorkloadKind
	if kind == "" {
		kind = "Deployment"
	}
	ns := c.Namespace
	if ns == "" {
		ns = "default"
	}

	reqMemMi := max(int64(64), c.TargetReqMem/(1024*1024))
	limMemMi := max(reqMemMi, c.TargetLimitMem/(1024*1024))
	reqCPU := max(int64(50), c.TargetReqCPU)
	limCPU := max(reqCPU, c.TargetLimitCPU)

	return fmt.Sprintf(`apiVersion: apps/v1
kind: %s
metadata:
  name: %s
  namespace: %s
spec:
  template:
    spec:
      containers:
        - name: %s
          resources:
            requests:
              cpu: "%dm"
              memory: "%dMi"
            limits:
              cpu: "%dm"
              memory: "%dMi"
`, kind, c.WorkloadName, ns, c.ContainerName, reqCPU, reqMemMi, limCPU, limMemMi)
}

// SynthesizePRDescription creates an executive and engineering justification for the GitOps PR
func SynthesizePRDescription(clusterID string, c domain.ContainerAnalysis) string {
	curMemMi := c.CurrentReqMem / (1024 * 1024)
	tgtMemMi := max(int64(64), c.TargetReqMem/(1024*1024))
	yearlySavings := c.MonthlyWasteUSD * 12.0

	return fmt.Sprintf(`## KubeTrim Automated Right-Sizing Pull Request

**Cluster:** `+"`%s`"+` | **Namespace:** `+"`%s`"+` | **Workload:** `+"`%s/%s`"+` (`+"`%s`"+`)

### 1. Executive Financial Impact
| Metric | Before ($T_0$) | After ($T_1$) | Projected Savings |
| :--- | :--- | :--- | :--- |
| **CPU Request** | `+"`%dm`"+` | `+"`%dm`"+` | **1.25x p95 Headroom Preserved** |
| **Memory Request** | `+"`%dMi`"+` | `+"`%dMi`"+` | **1.30x Peak Buffer Preserved** |
| **QoS Class** | `+"`%s`"+` | `+"`%s`"+` | **Kernel OOM Score Protected** |
| **Financial Impact** | — | — | **$%.2f / month ($%.2f / year)** |

### 2. Safety & Stability Verification
- **HPA Anti-Flapping Guard:** `+"`has_hpa=%t`"+` (Risk Status: `+"`%s`"+`)
- **Deterministic Fingerprint:** `+"`ConfigHash=%s`"+`

### 3. Rollback Procedure
If post-deployment SLA guard detects elevated throttling or OOM events, revert via:
`+"```bash\ngit revert HEAD && git push origin main\n```"+`
`,
		clusterID, c.Namespace, c.WorkloadKind, c.WorkloadName, c.ContainerName,
		c.CurrentReqCPU, c.TargetReqCPU,
		curMemMi, tgtMemMi,
		c.CurrentQoS, c.ProjectedQoS,
		c.MonthlyWasteUSD, yearlySavings,
		c.HasHPA, c.RiskStatus,
		c.ConfigHash,
	)
}

// CreatePullRequest synthesizes the Kustomize patch and opens a GitHub Pull Request (or executes in DryRun mode)
func (b *Bot) CreatePullRequest(ctx context.Context, req domain.PullRequestRequest) (*domain.PullRequestRecord, error) {
	if req.Container.WorkloadName == "" || req.Container.ContainerName == "" {
		return nil, fmt.Errorf("workload_name and container_name are required")
	}

	repo := strings.TrimSpace(req.Repository)
	if repo == "" {
		repo = "agusikprogramusik/env"
	}
	baseBranch := strings.TrimSpace(req.BaseBranch)
	if baseBranch == "" {
		baseBranch = "main"
	}

	patch := SynthesizeKustomizePatch(req.Container)
	desc := SynthesizePRDescription(req.ClusterID, req.Container)
	ts := time.Now().UTC().Format("20060102-150405")
	branchName := fmt.Sprintf("kubetrim/optimize-%s-%s", strings.ToLower(req.Container.WorkloadName), ts)
	title := fmt.Sprintf("feat(finops): right-size %s/%s (-$%.2f/mo)", req.Container.Namespace, req.Container.WorkloadName, req.Container.MonthlyWasteUSD)

	record := &domain.PullRequestRecord{
		ID:             uuid.NewString(),
		ClusterID:      req.ClusterID,
		Repository:     repo,
		BranchName:     branchName,
		Title:          title,
		Description:    desc,
		KustomizePatch: patch,
		MonthlySavings: req.Container.MonthlyWasteUSD,
		Status:         "DRY_RUN",
		PRURL:          fmt.Sprintf("https://github.com/%s/pull/dry-run", repo),
		CreatedAt:      time.Now().UTC(),
	}

	if req.DryRun || b.token == "" {
		return record, nil
	}

	prURL, err := b.openGitHubPR(ctx, repo, baseBranch, branchName, req.TargetPath, title, desc, patch, req.Container)
	if err != nil {
		return nil, err
	}

	record.Status = "OPEN"
	record.PRURL = prURL
	return record, nil
}

func (b *Bot) openGitHubPR(ctx context.Context, repo, baseBranch, branchName, targetPath, title, body, patch string, c domain.ContainerAnalysis) (string, error) {
	if targetPath == "" {
		targetPath = fmt.Sprintf("patches/patch-%s-%s.yaml", c.Namespace, c.WorkloadName)
	}

	// 1. Commit file to branch via GitHub Contents API
	contentURL := fmt.Sprintf("%s/repos/%s/contents/%s", b.apiBaseURL, repo, strings.TrimLeft(targetPath, "/"))
	putPayload := map[string]string{
		"message": title,
		"content": base64.StdEncoding.EncodeToString([]byte(patch)),
		"branch":  branchName,
	}
	if err := b.doJSON(ctx, http.MethodPut, contentURL, putPayload, nil); err != nil {
		return "", fmt.Errorf("failed to commit patch file to GitHub: %w", err)
	}

	// 2. Open Pull Request
	pullsURL := fmt.Sprintf("%s/repos/%s/pulls", b.apiBaseURL, repo)
	prPayload := map[string]string{
		"title": title,
		"head":  branchName,
		"base":  baseBranch,
		"body":  body,
	}
	var prResp struct {
		HTMLURL string `json:"html_url"`
	}
	if err := b.doJSON(ctx, http.MethodPost, pullsURL, prPayload, &prResp); err != nil {
		return "", fmt.Errorf("failed to create GitHub Pull Request: %w", err)
	}

	if prResp.HTMLURL == "" {
		prResp.HTMLURL = fmt.Sprintf("https://github.com/%s/pulls", repo)
	}
	return prResp.HTMLURL, nil
}

func (b *Bot) doJSON(ctx context.Context, method, url string, payload any, out any) error {
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github+json")
	if b.token != "" {
		req.Header.Set("Authorization", "Bearer "+b.token)
	}

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	respBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("GitHub API %s %s returned HTTP %d: %s", method, url, resp.StatusCode, strings.TrimSpace(string(respBytes)))
	}
	if out != nil && len(bytes.TrimSpace(respBytes)) > 0 {
		return json.Unmarshal(respBytes, out)
	}
	return nil
}
