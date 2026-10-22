package pricing

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/agusikprogramusik/kubetrim-hub/pkg/domain"
)

const (
	ProviderAWS    = "aws"
	ProviderGCP    = "gcp"
	ProviderAzure  = "azure"
	ProviderOnPrem = "on-prem"
	ProviderAuto   = "auto"
)

type vmSpec struct {
	Provider     string
	Arch         string
	VCPU         int
	MemoryGiB    float64
	GPUCount     int
	GPUModel     string
	BaseHourly   float64
	SpotDiscount float64 // ratio of On-Demand price paid on Spot (e.g. 0.30 means 70% discount)
}

var baseCatalog = map[string]vmSpec{
	// AWS x86 & ARM64 (Graviton) & GPU instances
	"t3.medium":     {Provider: ProviderAWS, Arch: "amd64", VCPU: 2, MemoryGiB: 4, BaseHourly: 0.0416, SpotDiscount: 0.30},
	"t3.large":      {Provider: ProviderAWS, Arch: "amd64", VCPU: 2, MemoryGiB: 8, BaseHourly: 0.0832, SpotDiscount: 0.30},
	"t3.xlarge":     {Provider: ProviderAWS, Arch: "amd64", VCPU: 4, MemoryGiB: 16, BaseHourly: 0.1664, SpotDiscount: 0.30},
	"m6i.large":     {Provider: ProviderAWS, Arch: "amd64", VCPU: 2, MemoryGiB: 8, BaseHourly: 0.0960, SpotDiscount: 0.32},
	"m6i.xlarge":    {Provider: ProviderAWS, Arch: "amd64", VCPU: 4, MemoryGiB: 16, BaseHourly: 0.1920, SpotDiscount: 0.32},
	"m6i.2xlarge":   {Provider: ProviderAWS, Arch: "amd64", VCPU: 8, MemoryGiB: 32, BaseHourly: 0.3840, SpotDiscount: 0.32},
	"m7g.large":     {Provider: ProviderAWS, Arch: "arm64", VCPU: 2, MemoryGiB: 8, BaseHourly: 0.0816, SpotDiscount: 0.28},
	"m7g.xlarge":    {Provider: ProviderAWS, Arch: "arm64", VCPU: 4, MemoryGiB: 16, BaseHourly: 0.1632, SpotDiscount: 0.28},
	"m7g.2xlarge":   {Provider: ProviderAWS, Arch: "arm64", VCPU: 8, MemoryGiB: 32, BaseHourly: 0.3264, SpotDiscount: 0.28},
	"c6i.xlarge":    {Provider: ProviderAWS, Arch: "amd64", VCPU: 4, MemoryGiB: 8, BaseHourly: 0.1700, SpotDiscount: 0.30},
	"c7g.xlarge":    {Provider: ProviderAWS, Arch: "arm64", VCPU: 4, MemoryGiB: 8, BaseHourly: 0.1445, SpotDiscount: 0.28},
	"r6i.xlarge":    {Provider: ProviderAWS, Arch: "amd64", VCPU: 4, MemoryGiB: 32, BaseHourly: 0.2520, SpotDiscount: 0.32},
	"g5.xlarge":     {Provider: ProviderAWS, Arch: "amd64", VCPU: 4, MemoryGiB: 16, GPUCount: 1, GPUModel: "A10G", BaseHourly: 1.0060, SpotDiscount: 0.35},
	"p4d.24xlarge":  {Provider: ProviderAWS, Arch: "amd64", VCPU: 96, MemoryGiB: 1152, GPUCount: 8, GPUModel: "A100", BaseHourly: 32.7726, SpotDiscount: 0.40},
	"p5.48xlarge":   {Provider: ProviderAWS, Arch: "amd64", VCPU: 192, MemoryGiB: 2048, GPUCount: 8, GPUModel: "H100", BaseHourly: 98.3200, SpotDiscount: 0.45},

	// GCP Compute Engine instances
	"e2-medium":       {Provider: ProviderGCP, Arch: "amd64", VCPU: 2, MemoryGiB: 4, BaseHourly: 0.0335, SpotDiscount: 0.30},
	"e2-standard-4":   {Provider: ProviderGCP, Arch: "amd64", VCPU: 4, MemoryGiB: 16, BaseHourly: 0.1340, SpotDiscount: 0.30},
	"n2-standard-4":   {Provider: ProviderGCP, Arch: "amd64", VCPU: 4, MemoryGiB: 16, BaseHourly: 0.1942, SpotDiscount: 0.29},
	"n2-standard-8":   {Provider: ProviderGCP, Arch: "amd64", VCPU: 8, MemoryGiB: 32, BaseHourly: 0.3884, SpotDiscount: 0.29},
	"c3-standard-4":   {Provider: ProviderGCP, Arch: "amd64", VCPU: 4, MemoryGiB: 16, BaseHourly: 0.2088, SpotDiscount: 0.30},
	"t2a-standard-4":  {Provider: ProviderGCP, Arch: "arm64", VCPU: 4, MemoryGiB: 16, BaseHourly: 0.1540, SpotDiscount: 0.28},
	"a2-highgpu-1g":   {Provider: ProviderGCP, Arch: "amd64", VCPU: 12, MemoryGiB: 85, GPUCount: 1, GPUModel: "A100", BaseHourly: 3.6730, SpotDiscount: 0.35},
	"a3-highgpu-8g":   {Provider: ProviderGCP, Arch: "amd64", VCPU: 208, MemoryGiB: 1872, GPUCount: 8, GPUModel: "H100", BaseHourly: 88.4900, SpotDiscount: 0.40},

	// Azure Virtual Machine instances
	"standard_b2s":            {Provider: ProviderAzure, Arch: "amd64", VCPU: 2, MemoryGiB: 4, BaseHourly: 0.0416, SpotDiscount: 0.30},
	"standard_d4s_v5":         {Provider: ProviderAzure, Arch: "amd64", VCPU: 4, MemoryGiB: 16, BaseHourly: 0.1920, SpotDiscount: 0.30},
	"standard_d8s_v5":         {Provider: ProviderAzure, Arch: "amd64", VCPU: 8, MemoryGiB: 32, BaseHourly: 0.3840, SpotDiscount: 0.30},
	"standard_d4ps_v5":        {Provider: ProviderAzure, Arch: "arm64", VCPU: 4, MemoryGiB: 16, BaseHourly: 0.1536, SpotDiscount: 0.28},
	"standard_e4s_v5":         {Provider: ProviderAzure, Arch: "amd64", VCPU: 4, MemoryGiB: 32, BaseHourly: 0.2520, SpotDiscount: 0.31},
	"standard_nc24ads_a100_v4": {Provider: ProviderAzure, Arch: "amd64", VCPU: 24, MemoryGiB: 220, GPUCount: 1, GPUModel: "A100", BaseHourly: 3.6700, SpotDiscount: 0.38},

	// On-Premise / Sovereign Bare-Metal nodes
	"baremetal-x86":   {Provider: ProviderOnPrem, Arch: "amd64", VCPU: 32, MemoryGiB: 128, BaseHourly: 0.9600, SpotDiscount: 1.00},
	"baremetal-arm64": {Provider: ProviderOnPrem, Arch: "arm64", VCPU: 32, MemoryGiB: 128, BaseHourly: 0.7680, SpotDiscount: 1.00},
}

var providerBaselineRates = map[string]domain.PriceRates{
	ProviderAWS: {
		VCPUHourlyX86:          0.0336,
		VCPUHourlyARM64:        0.0269,
		GiBRAMHourly:           0.0044,
		SpotDiscountMultiplier: 0.31,
		GPUHourlyA100:          3.6700,
		GPUHourlyH100:          10.8000,
		GPUHourlyGeneric:       1.0060,
	},
	ProviderGCP: {
		VCPUHourlyX86:          0.0318,
		VCPUHourlyARM64:        0.0254,
		GiBRAMHourly:           0.0042,
		SpotDiscountMultiplier: 0.29,
		GPUHourlyA100:          3.6730,
		GPUHourlyH100:          10.5000,
		GPUHourlyGeneric:       0.9500,
	},
	ProviderAzure: {
		VCPUHourlyX86:          0.0340,
		VCPUHourlyARM64:        0.0272,
		GiBRAMHourly:           0.0045,
		SpotDiscountMultiplier: 0.30,
		GPUHourlyA100:          3.6700,
		GPUHourlyH100:          10.9000,
		GPUHourlyGeneric:       1.0200,
	},
	ProviderOnPrem: {
		VCPUHourlyX86:          0.0210,
		VCPUHourlyARM64:        0.0168,
		GiBRAMHourly:           0.0028,
		SpotDiscountMultiplier: 1.00,
		GPUHourlyA100:          2.2000,
		GPUHourlyH100:          6.5000,
		GPUHourlyGeneric:       0.7500,
	},
}

// ResolveQuote resolves cloud provider baseline rates and exact VM SKU prices for a cluster
func ResolveQuote(req domain.PricingQuoteRequest) (domain.PricingQuoteResponse, error) {
	provider := normalizeProvider(req.CloudProvider, req.InstanceTypes)
	if provider == "" {
		return domain.PricingQuoteResponse{}, fmt.Errorf("unsupported cloud_provider %q (allowed: aws, gcp, azure, on-prem, auto)", req.CloudProvider)
	}

	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = defaultRegionForProvider(provider)
	}
	regMult := regionMultiplier(region)

	baseRates := providerBaselineRates[provider]
	rates := domain.PriceRates{
		VCPUHourlyX86:          round4(baseRates.VCPUHourlyX86 * regMult),
		VCPUHourlyARM64:        round4(baseRates.VCPUHourlyARM64 * regMult),
		GiBRAMHourly:           round4(baseRates.GiBRAMHourly * regMult),
		SpotDiscountMultiplier: baseRates.SpotDiscountMultiplier,
		GPUHourlyA100:          round4(baseRates.GPUHourlyA100 * regMult),
		GPUHourlyH100:          round4(baseRates.GPUHourlyH100 * regMult),
		GPUHourlyGeneric:       round4(baseRates.GPUHourlyGeneric * regMult),
	}

	instancePrices := make(map[string]domain.VMInstancePrice)
	var sumVCPUX86, sumVCPUARM, sumRAM float64
	var countX86, countARM, countRAM int

	for _, rawSKU := range req.InstanceTypes {
		sku := strings.TrimSpace(rawSKU)
		if sku == "" {
			continue
		}
		vmPrice := resolveSKU(provider, region, regMult, sku, rates)
		instancePrices[sku] = vmPrice

		if vmPrice.Architecture == "arm64" {
			sumVCPUARM += vmPrice.DerivedVCPUHourly
			countARM++
		} else {
			sumVCPUX86 += vmPrice.DerivedVCPUHourly
			countX86++
		}
		sumRAM += vmPrice.DerivedGiBRAMHourly
		countRAM++
	}

	// If cluster nodes reported specific VM SKUs, blend cluster-specific VM rates into top-level PriceRates
	if countX86 > 0 {
		rates.VCPUHourlyX86 = round4(sumVCPUX86 / float64(countX86))
	}
	if countARM > 0 {
		rates.VCPUHourlyARM64 = round4(sumVCPUARM / float64(countARM))
	} else if countX86 > 0 {
		rates.VCPUHourlyARM64 = round4(rates.VCPUHourlyX86 * 0.80)
	}
	if countRAM > 0 {
		rates.GiBRAMHourly = round4(sumRAM / float64(countRAM))
	}

	return domain.PricingQuoteResponse{
		CloudProvider:  provider,
		Region:         region,
		Currency:       "USD",
		HoursPerMonth:  730.0,
		Source:         fmt.Sprintf("kubetrim-hub-pricing-catalog/%s/%s", provider, region),
		UpdatedAt:      time.Now().UTC(),
		Rates:          rates,
		InstancePrices: instancePrices,
	}, nil
}

func resolveSKU(provider, region string, regMult float64, sku string, fallbackRates domain.PriceRates) domain.VMInstancePrice {
	lookupKey := strings.ToLower(sku)
	spec, ok := baseCatalog[lookupKey]
	if !ok {
		spec = synthesizeUnknownSKU(provider, sku, fallbackRates)
	}

	onDemand := round4(spec.BaseHourly * regMult)
	spot := round4(onDemand * spec.SpotDiscount)

	// Standard FinOps CPU/Memory cost attribution split: 70% compute (vCPU), 30% memory (GiB)
	vcpuCount := max(1, spec.VCPU)
	memGiB := math.Max(1.0, spec.MemoryGiB)
	derivedVCPU := round4((onDemand * 0.70) / float64(vcpuCount))
	derivedRAM := round4((onDemand * 0.30) / memGiB)

	return domain.VMInstancePrice{
		InstanceType:        sku,
		CloudProvider:       provider,
		Region:              region,
		Architecture:        spec.Arch,
		VCPU:                vcpuCount,
		MemoryGiB:           memGiB,
		GPUCount:            spec.GPUCount,
		GPUModel:            spec.GPUModel,
		OnDemandHourlyUSD:   onDemand,
		SpotHourlyUSD:       spot,
		DerivedVCPUHourly:   derivedVCPU,
		DerivedGiBRAMHourly: derivedRAM,
	}
}

func synthesizeUnknownSKU(provider, sku string, rates domain.PriceRates) vmSpec {
	lower := strings.ToLower(sku)
	arch := "amd64"
	if strings.Contains(lower, "7g") || strings.Contains(lower, "6g") || strings.Contains(lower, "t2a") || strings.Contains(lower, "ps_v") || strings.Contains(lower, "arm") {
		arch = "arm64"
	}

	vcpu := 4
	memGiB := 16.0
	switch {
	case strings.Contains(lower, "medium"):
		vcpu, memGiB = 2, 4
	case strings.Contains(lower, "2xlarge") || strings.HasSuffix(lower, "-8") || strings.Contains(lower, "d8"):
		vcpu, memGiB = 8, 32
	case strings.Contains(lower, "4xlarge") || strings.HasSuffix(lower, "-16") || strings.Contains(lower, "d16"):
		vcpu, memGiB = 16, 64
	case strings.Contains(lower, "large") && !strings.Contains(lower, "xlarge"):
		vcpu, memGiB = 2, 8
	}

	vcpuRate := rates.VCPUHourlyX86
	if arch == "arm64" {
		vcpuRate = rates.VCPUHourlyARM64
	}
	baseHourly := float64(vcpu)*vcpuRate + memGiB*rates.GiBRAMHourly

	return vmSpec{
		Provider:     provider,
		Arch:         arch,
		VCPU:         vcpu,
		MemoryGiB:    memGiB,
		BaseHourly:   baseHourly,
		SpotDiscount: rates.SpotDiscountMultiplier,
	}
}

func normalizeProvider(raw string, instanceTypes []string) string {
	p := strings.ToLower(strings.TrimSpace(raw))
	switch p {
	case "", ProviderAuto:
		for _, sku := range instanceTypes {
			s := strings.ToLower(strings.TrimSpace(sku))
			switch {
			case strings.HasPrefix(s, "standard_"):
				return ProviderAzure
			case strings.HasPrefix(s, "e2-") || strings.HasPrefix(s, "n2-") || strings.HasPrefix(s, "c3-") || strings.HasPrefix(s, "t2a-") || strings.HasPrefix(s, "a2-") || strings.HasPrefix(s, "a3-"):
				return ProviderGCP
			case strings.Contains(s, "."):
				return ProviderAWS
			case strings.HasPrefix(s, "baremetal"):
				return ProviderOnPrem
			}
		}
		return ProviderAWS
	case "aws", "eks", "amazon":
		return ProviderAWS
	case "gcp", "gke", "google":
		return ProviderGCP
	case "azure", "aks", "microsoft":
		return ProviderAzure
	case "on-prem", "onprem", "baremetal", "sovereign":
		return ProviderOnPrem
	default:
		return ""
	}
}

func defaultRegionForProvider(provider string) string {
	switch provider {
	case ProviderGCP:
		return "europe-west1"
	case ProviderAzure:
		return "westeurope"
	case ProviderOnPrem:
		return "dc-eu-central"
	default:
		return "eu-central-1"
	}
}

func regionMultiplier(region string) float64 {
	r := strings.ToLower(strings.TrimSpace(region))
	switch {
	case strings.HasPrefix(r, "us-east") || r == "us-central1" || r == "eastus":
		return 1.00
	case strings.HasPrefix(r, "eu-") || strings.HasPrefix(r, "europe-") || r == "westeurope" || r == "northeurope":
		return 1.08
	case strings.HasPrefix(r, "ap-"):
		return 1.12
	default:
		return 1.00
	}
}

func round4(v float64) float64 {
	return math.Round(v*10000) / 10000
}
