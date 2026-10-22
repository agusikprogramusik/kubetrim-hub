package pricing

import (
	"testing"

	"github.com/agusikprogramusik/kubetrim-hub/pkg/domain"
)

func TestResolveQuote_MultiCloudAndAutoDetection(t *testing.T) {
	// 1. AWS with explicit SKUs (m6i.xlarge + m7g.large + unknown m6i.4xlarge)
	awsQuote, err := ResolveQuote(domain.PricingQuoteRequest{
		CloudProvider: "aws",
		Region:        "eu-central-1",
		InstanceTypes: []string{"m6i.xlarge", "m7g.large", "m6i.4xlarge"},
	})
	if err != nil {
		t.Fatalf("unexpected error for AWS quote: %v", err)
	}
	if awsQuote.CloudProvider != ProviderAWS || len(awsQuote.InstancePrices) != 3 {
		t.Fatalf("expected 3 AWS instance prices, got %+v", awsQuote)
	}
	m6i := awsQuote.InstancePrices["m6i.xlarge"]
	m7g := awsQuote.InstancePrices["m7g.large"]
	if m6i.OnDemandHourlyUSD <= 0 || m7g.Architecture != "arm64" {
		t.Errorf("unexpected SKU metadata: m6i=%+v, m7g=%+v", m6i, m7g)
	}
	if awsQuote.Rates.VCPUHourlyARM64 >= awsQuote.Rates.VCPUHourlyX86 {
		t.Errorf("expected ARM64 vCPU rate (%f) < x86 vCPU rate (%f)", awsQuote.Rates.VCPUHourlyARM64, awsQuote.Rates.VCPUHourlyX86)
	}

	// 2. Auto-detect GCP from n2-standard-4
	gcpQuote, err := ResolveQuote(domain.PricingQuoteRequest{
		CloudProvider: "auto",
		InstanceTypes: []string{"n2-standard-4", "e2-medium"},
	})
	if err != nil {
		t.Fatalf("unexpected error for GCP auto quote: %v", err)
	}
	if gcpQuote.CloudProvider != ProviderGCP {
		t.Errorf("expected auto-detected provider gcp, got %s", gcpQuote.CloudProvider)
	}

	// 3. Auto-detect Azure from Standard_D4s_v5
	azQuote, err := ResolveQuote(domain.PricingQuoteRequest{
		CloudProvider: "auto",
		Region:        "eastus",
		InstanceTypes: []string{"Standard_D4s_v5", "Standard_D8s_v5"},
	})
	if err != nil {
		t.Fatalf("unexpected error for Azure auto quote: %v", err)
	}
	if azQuote.CloudProvider != ProviderAzure {
		t.Errorf("expected auto-detected provider azure, got %s", azQuote.CloudProvider)
	}

	// 4. On-Prem provider
	onPremQuote, err := ResolveQuote(domain.PricingQuoteRequest{
		CloudProvider: "on-prem",
		InstanceTypes: []string{"baremetal-x86"},
	})
	if err != nil {
		t.Fatalf("unexpected error for on-prem quote: %v", err)
	}
	if onPremQuote.Rates.SpotDiscountMultiplier != 1.0 {
		t.Errorf("expected on-prem spot multiplier 1.0, got %f", onPremQuote.Rates.SpotDiscountMultiplier)
	}

	// 5. Invalid provider returns error
	if _, err := ResolveQuote(domain.PricingQuoteRequest{CloudProvider: "digitalocean-legacy"}); err == nil {
		t.Fatal("expected error for unsupported cloud provider")
	}
}
