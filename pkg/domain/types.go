package domain

import "time"

const (
	Version = "0.2.2-dev"

	StatusVerifiedSavings = "VERIFIED_SAVINGS"
	StatusRegressionAlert = "REGRESSION_ALERT"
	StatusUnchanged       = "UNCHANGED"
	StatusLeakSuspected   = "LEAK_SUSPECTED"
	StatusIdleGPU         = "IDLE_GPU"
)

// AuditReport represents the canonical v1 FinOps snapshot ingested from KubeTrim CLI/Agent
type AuditReport struct {
	SchemaVersion string              `json:"schema_version"`
	ReportID      string              `json:"report_id"`
	ClusterName   string              `json:"cluster_name"`
	ClusterHash   string              `json:"cluster_hash"`
	GeneratedAt   time.Time           `json:"generated_at"`
	GitCommit     string              `json:"git_commit,omitempty"`
	Summary       AuditSummary        `json:"summary"`
	Containers    []ContainerAnalysis `json:"containers"`
}

// AuditSummary aggregates cluster-level FinOps metrics
type AuditSummary struct {
	TotalWorkloadsAudited int            `json:"total_workloads_audited"`
	TotalContainers       int            `json:"total_containers"`
	CurrentMonthlyCostUSD float64        `json:"current_monthly_cost_usd"`
	TargetMonthlyCostUSD  float64        `json:"target_monthly_cost_usd"`
	TotalMonthlyWasteUSD  float64        `json:"total_monthly_waste_usd"`
	TotalYearlyWasteUSD   float64        `json:"total_yearly_waste_usd"`
	ClusterEfficiency     float64        `json:"cluster_efficiency_score"`
	CriticalOOMCount      int            `json:"critical_oom_count"`
	ThrottledCount        int            `json:"throttled_count"`
	HPAManagedCount       int            `json:"hpa_managed_count"`
	IdleGPUCount          int            `json:"idle_gpu_count"`
	SafeCount             int            `json:"safe_count"`
	RiskBreakdown         map[string]int `json:"risk_breakdown"`
}

// ContainerAnalysis holds right-sizing telemetry and recommendations for a single container
type ContainerAnalysis struct {
	Namespace       string   `json:"namespace"`
	WorkloadName    string   `json:"workload_name"`
	WorkloadKind    string   `json:"workload_kind"`
	ContainerName   string   `json:"container_name"`
	ConfigHash      string   `json:"config_hash"`
	CurrentReqCPU   int64    `json:"current_req_cpu_milli"`
	TargetReqCPU    int64    `json:"target_req_cpu_milli"`
	CurrentLimitCPU int64    `json:"current_limit_cpu_milli"`
	TargetLimitCPU  int64    `json:"target_limit_cpu_milli"`
	CurrentReqMem   int64    `json:"current_req_mem_bytes"`
	TargetReqMem    int64    `json:"target_req_mem_bytes"`
	CurrentLimitMem int64    `json:"current_limit_mem_bytes"`
	TargetLimitMem  int64    `json:"target_limit_mem_bytes"`
	Replicas        int32    `json:"replicas"`
	MonthlyWasteUSD float64  `json:"monthly_waste_usd"`
	RiskStatus      string   `json:"risk_status"`
	CurrentQoS      string   `json:"current_qos"`
	ProjectedQoS    string   `json:"projected_qos"`
	HasHPA          bool     `json:"has_hpa"`
	HPAMetrics      []string `json:"hpa_metrics,omitempty"`
	IsSpotInstance  bool     `json:"is_spot_instance"`
	NodeArch        string   `json:"node_arch"`
	GPUCount        int64    `json:"gpu_count"`
	GPUUtilization  float64  `json:"gpu_utilization"`
	OOMCount        int      `json:"oom_count,omitempty"`
	ThrottleRatio   float64  `json:"throttle_ratio,omitempty"`
	Warnings        []string `json:"warnings,omitempty"`
}

// ContainerKey returns the canonical 4-tuple identifier within a cluster
func (c ContainerAnalysis) ContainerKey() string {
	return c.Namespace + "/" + c.WorkloadKind + "/" + c.WorkloadName + "/" + c.ContainerName
}

// ClusterOverview summarizes the latest state of a registered Kubernetes cluster
type ClusterOverview struct {
	ClusterID             string    `json:"cluster_id"`
	LatestSnapshotID      string    `json:"latest_snapshot_id"`
	LastSeenAt            time.Time `json:"last_seen_at"`
	SnapshotCount         int       `json:"snapshot_count"`
	TotalWorkloads        int       `json:"total_workloads"`
	TotalContainers       int       `json:"total_containers"`
	CurrentMonthlyCostUSD float64   `json:"current_monthly_cost_usd"`
	TotalMonthlyWasteUSD  float64   `json:"total_monthly_waste_usd"`
	RealizedSavingsUSD    float64   `json:"realized_savings_usd"`
	EfficiencyScore       float64   `json:"efficiency_score"`
	IdleGPUCount          int       `json:"idle_gpu_count"`
}

// ExecutiveKPIs aggregates multi-cluster FinOps metrics for the Executive Dashboard
type ExecutiveKPIs struct {
	TotalClusters            int            `json:"total_clusters"`
	TotalWorkloadsAudited    int            `json:"total_workloads_audited"`
	TotalContainersAudited   int            `json:"total_containers_audited"`
	FleetMonthlyCostUSD      float64        `json:"fleet_monthly_cost_usd"`
	IdentifiedMonthlyWaste   float64        `json:"identified_monthly_waste_usd"`
	IdentifiedYearlyWaste    float64        `json:"identified_yearly_waste_usd"`
	RealizedMonthlySavings   float64        `json:"realized_monthly_savings_usd"`
	RealizedYearlySavings    float64        `json:"realized_yearly_savings_usd"`
	AverageEfficiencyScore   float64        `json:"average_efficiency_score"`
	IdleGPUAlerts            int            `json:"idle_gpu_alerts"`
	SLARegressionAlerts      int            `json:"sla_regression_alerts"`
	RiskBreakdown            map[string]int `json:"risk_breakdown"`
}

// ContainerAttribution tracks T0 -> T1 Before/After optimization impact for a container
type ContainerAttribution struct {
	Key                       string  `json:"key"`
	Namespace                 string  `json:"namespace"`
	WorkloadKind              string  `json:"workload_kind"`
	WorkloadName              string  `json:"workload_name"`
	ContainerName             string  `json:"container_name"`
	BeforeConfigHash          string  `json:"before_config_hash"`
	AfterConfigHash           string  `json:"after_config_hash"`
	ConfigChanged             bool    `json:"config_changed"`
	BeforeCPUReqMilli         int64   `json:"before_cpu_req_milli"`
	AfterCPUReqMilli          int64   `json:"after_cpu_req_milli"`
	BeforeMemReqBytes         int64   `json:"before_mem_req_bytes"`
	AfterMemReqBytes          int64   `json:"after_mem_req_bytes"`
	RealizedMonthlySavingsUSD float64 `json:"realized_monthly_savings_usd"`
	OOMDelta                  int     `json:"oom_delta"`
	ThrottleDelta             float64 `json:"throttle_delta"`
	Status                    string  `json:"status"` // VERIFIED_SAVINGS, REGRESSION_ALERT, UNCHANGED
	RollbackRecommended       bool    `json:"rollback_recommended"`
}

// AttributionReport summarizes cluster-wide T0 -> T1 impact verification
type AttributionReport struct {
	ClusterID                 string                 `json:"cluster_id"`
	BaselineSnapshotID        string                 `json:"baseline_snapshot_id"`
	CurrentSnapshotID         string                 `json:"current_snapshot_id"`
	BaselineTimestamp         time.Time              `json:"baseline_timestamp"`
	CurrentTimestamp          time.Time              `json:"current_timestamp"`
	TotalRealizedMonthlyUSD   float64                `json:"total_realized_monthly_usd"`
	TotalRealizedYearlyUSD    float64                `json:"total_realized_yearly_usd"`
	OptimizedContainersCount  int                    `json:"optimized_containers_count"`
	RegressionAlertsCount     int                    `json:"regression_alerts_count"`
	Attributions              []ContainerAttribution `json:"attributions"`
}

// MemorySample represents a single time-series point of container_memory_working_set_bytes
type MemorySample struct {
	Timestamp time.Time `json:"timestamp"`
	Bytes     int64     `json:"bytes"`
	OOMKilled bool      `json:"oom_killed,omitempty"`
}

// AnomalyFinding describes a detected memory leak (sawtooth) or idle GPU accelerator
type AnomalyFinding struct {
	ClusterID           string  `json:"cluster_id"`
	Namespace           string  `json:"namespace"`
	WorkloadName        string  `json:"workload_name"`
	ContainerName       string  `json:"container_name"`
	AnomalyType         string  `json:"anomaly_type"` // LEAK_SUSPECTED, IDLE_GPU
	Severity            string  `json:"severity"`     // CRITICAL, HIGH
	SawtoothCycles      int     `json:"sawtooth_cycles,omitempty"`
	SlopeBytesPerHour   float64 `json:"slope_bytes_per_hour,omitempty"`
	BlockMemoryUpsizing bool    `json:"block_memory_upsizing"`
	MonthlyWasteUSD     float64 `json:"monthly_waste_usd"`
	Recommendation      string  `json:"recommendation"`
}

// PullRequestRequest triggers 1-Click GitOps PR creation for a workload
type PullRequestRequest struct {
	ClusterID     string            `json:"cluster_id"`
	Repository    string            `json:"repository"` // e.g. "agusikprogramusik/env"
	BaseBranch    string            `json:"base_branch"`
	TargetPath    string            `json:"target_path"`
	Container     ContainerAnalysis `json:"container"`
	DryRun        bool              `json:"dry_run"`
}

// PullRequestRecord stores the synthesized GitOps PR and Kustomize patch
type PullRequestRecord struct {
	ID              string    `json:"id"`
	ClusterID       string    `json:"cluster_id"`
	Repository      string    `json:"repository"`
	BranchName      string    `json:"branch_name"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	KustomizePatch  string    `json:"kustomize_patch"`
	MonthlySavings  float64   `json:"monthly_savings_usd"`
	PRURL           string    `json:"pr_url"`
	Status          string    `json:"status"` // OPEN, DRY_RUN
	CreatedAt       time.Time `json:"created_at"`
}

// PriceRates defines cloud provider hourly compute rates returned to KubeTrim CLI/Agent
type PriceRates struct {
	VCPUHourlyX86          float64 `json:"vcpu_hourly_x86"`
	VCPUHourlyARM64        float64 `json:"vcpu_hourly_arm64"`
	GiBRAMHourly           float64 `json:"gib_ram_hourly"`
	SpotDiscountMultiplier float64 `json:"spot_discount_multiplier"`
	GPUHourlyA100          float64 `json:"gpu_hourly_a100"`
	GPUHourlyH100          float64 `json:"gpu_hourly_h100"`
	GPUHourlyGeneric       float64 `json:"gpu_hourly_generic"`
}

// VMInstancePrice describes the hourly cost and hardware profile of a specific cloud VM SKU
type VMInstancePrice struct {
	InstanceType       string  `json:"instance_type"`
	CloudProvider      string  `json:"cloud_provider"`
	Region             string  `json:"region"`
	Architecture       string  `json:"architecture"` // amd64, arm64
	VCPU               int     `json:"vcpu"`
	MemoryGiB          float64 `json:"memory_gib"`
	GPUCount           int     `json:"gpu_count,omitempty"`
	GPUModel           string  `json:"gpu_model,omitempty"`
	OnDemandHourlyUSD  float64 `json:"on_demand_hourly_usd"`
	SpotHourlyUSD      float64 `json:"spot_hourly_usd"`
	DerivedVCPUHourly  float64 `json:"derived_vcpu_hourly_usd"`
	DerivedGiBRAMHourly float64 `json:"derived_gib_ram_hourly_usd"`
}

// PricingQuoteRequest is sent by KubeTrim in-cluster agent/CLI to resolve dynamic VM pricing
type PricingQuoteRequest struct {
	ClusterID     string   `json:"cluster_id,omitempty"`
	CloudProvider string   `json:"cloud_provider"`           // aws, gcp, azure, on-prem, auto
	Region        string   `json:"region,omitempty"`         // e.g. eu-central-1, us-east-1, europe-west1
	InstanceTypes []string `json:"instance_types,omitempty"` // e.g. ["m6i.xlarge", "m7g.large"]
}

// PricingQuoteResponse returns dynamic cloud provider rates and per-VM SKU pricing to KubeTrim
type PricingQuoteResponse struct {
	CloudProvider  string                     `json:"cloud_provider"`
	Region         string                     `json:"region"`
	Currency       string                     `json:"currency"`
	HoursPerMonth  float64                    `json:"hours_per_month"`
	Source         string                     `json:"source"`
	UpdatedAt      time.Time                  `json:"updated_at"`
	Rates          PriceRates                 `json:"rates"`
	InstancePrices map[string]VMInstancePrice `json:"instance_prices"`
}
