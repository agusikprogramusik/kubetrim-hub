package domain

import "testing"

func TestContainerKey(t *testing.T) {
	c := ContainerAnalysis{
		Namespace:     "payments",
		WorkloadKind:  "Deployment",
		WorkloadName:  "checkout",
		ContainerName: "api",
	}
	want := "payments/Deployment/checkout/api"
	if got := c.ContainerKey(); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
