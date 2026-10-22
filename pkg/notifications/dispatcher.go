package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/agusikprogramusik/kubetrim-hub/pkg/domain"
)

// Config configures Slack and MS Teams notification webhooks
type Config struct {
	SlackWebhookURL string
	TeamsWebhookURL string
	HTTPClient      *http.Client
}

// Dispatcher sends FinOps alerts and weekly digests to Slack and MS Teams
type Dispatcher struct {
	cfg    Config
	client *http.Client
}

// NewDispatcher creates a new notifications Dispatcher
func NewDispatcher(cfg Config) *Dispatcher {
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &Dispatcher{
		cfg:    cfg,
		client: client,
	}
}

// FormatWeeklyDigest builds the human-readable weekly executive FinOps summary
func FormatWeeklyDigest(kpis domain.ExecutiveKPIs) string {
	return fmt.Sprintf(
		"*[KubeTrim Hub // Weekly Executive FinOps Digest]*\n"+
			"• *Audited Fleet:* %d clusters / %d workloads (%d containers)\n"+
			"• *Fleet Monthly Spend:* $%.2f/mo | *Efficiency Score:* %.1f%%\n"+
			"• *Realized Savings:* $%.2f/mo ($%.2f/yr)\n"+
			"• *Remaining Identified Waste:* $%.2f/mo ($%.2f/yr)\n"+
			"• *Active Alerts:* %d Idle GPU(s), %d SLA Regression(s)",
		kpis.TotalClusters,
		kpis.TotalWorkloadsAudited,
		kpis.TotalContainersAudited,
		kpis.FleetMonthlyCostUSD,
		kpis.AverageEfficiencyScore,
		kpis.RealizedMonthlySavings,
		kpis.RealizedYearlySavings,
		kpis.IdentifiedMonthlyWaste,
		kpis.IdentifiedYearlyWaste,
		kpis.IdleGPUAlerts,
		kpis.SLARegressionAlerts,
	)
}

// FormatAnomalyAlert builds an alert message for IDLE_GPU or LEAK_SUSPECTED findings
func FormatAnomalyAlert(finding domain.AnomalyFinding) string {
	return fmt.Sprintf(
		"*[KubeTrim Hub // %s Alert (%s)]*\n"+
			"• *Cluster:* `%s` | *Workload:* `%s/%s` (`%s`)\n"+
			"• *Estimated Waste:* $%.2f/mo\n"+
			"• *Recommendation:* %s",
		finding.AnomalyType,
		finding.Severity,
		finding.ClusterID,
		finding.Namespace,
		finding.WorkloadName,
		finding.ContainerName,
		finding.MonthlyWasteUSD,
		finding.Recommendation,
	)
}

// FormatRegressionAlert builds an urgent SLA regression alert when OOMKills or throttling spike post-trim
func FormatRegressionAlert(clusterID string, attr domain.ContainerAttribution) string {
	return fmt.Sprintf(
		"*[KubeTrim Hub // SLA REGRESSION ALERT]*\n"+
			"• *Cluster:* `%s` | *Container:* `%s`\n"+
			"• *Telemetry Delta:* OOMKills `+%d`, CPU Throttling `+%.1f%%`\n"+
			"• *Action:* Automatic rollback PR recommended (`RollbackRecommended=%v`)",
		clusterID,
		attr.Key,
		attr.OOMDelta,
		attr.ThrottleDelta*100.0,
		attr.RollbackRecommended,
	)
}

// SendWeeklyDigest dispatches the weekly executive digest to configured webhooks
func (d *Dispatcher) SendWeeklyDigest(ctx context.Context, kpis domain.ExecutiveKPIs) error {
	return d.dispatchText(ctx, "KubeTrim Weekly FinOps Digest", FormatWeeklyDigest(kpis))
}

// SendAnomalyAlert dispatches an anomaly finding to configured webhooks
func (d *Dispatcher) SendAnomalyAlert(ctx context.Context, finding domain.AnomalyFinding) error {
	title := fmt.Sprintf("KubeTrim %s Alert", finding.AnomalyType)
	return d.dispatchText(ctx, title, FormatAnomalyAlert(finding))
}

// SendRegressionAlert dispatches an SLA regression alert to configured webhooks
func (d *Dispatcher) SendRegressionAlert(ctx context.Context, clusterID string, attr domain.ContainerAttribution) error {
	return d.dispatchText(ctx, "KubeTrim SLA Regression Alert", FormatRegressionAlert(clusterID, attr))
}

func (d *Dispatcher) dispatchText(ctx context.Context, title, text string) error {
	if d.cfg.SlackWebhookURL != "" {
		payload := map[string]string{"text": text}
		if err := d.postJSON(ctx, d.cfg.SlackWebhookURL, payload); err != nil {
			return fmt.Errorf("slack webhook failed: %w", err)
		}
	}
	if d.cfg.TeamsWebhookURL != "" {
		payload := map[string]string{
			"@type":      "MessageCard",
			"@context":   "http://schema.org/extensions",
			"title":      title,
			"text":       text,
			"themeColor": "FF4F00",
		}
		if err := d.postJSON(ctx, d.cfg.TeamsWebhookURL, payload); err != nil {
			return fmt.Errorf("teams webhook failed: %w", err)
		}
	}
	return nil
}

func (d *Dispatcher) postJSON(ctx context.Context, url string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected webhook status %d", resp.StatusCode)
	}
	return nil
}
