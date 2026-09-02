// Package acceptance holds cross-module parity smoke checks (no DB required).
package acceptance

import (
	"testing"

	crm "sumeru_addons/crm/services"
)

func TestProratedRevenueFormula(t *testing.T) {
	got := crm.ProratedRevenue(10000, 40)
	if got != 4000 {
		t.Fatalf("expected 4000, got %v", got)
	}
}

func TestPLSProbabilityBounds(t *testing.T) {
	p := crm.ClampProbability(150)
	if p != 100 {
		t.Fatalf("expected 100, got %v", p)
	}
	p = crm.ClampProbability(-5)
	if p != 0 {
		t.Fatalf("expected 0, got %v", p)
	}
}

func TestWeightedPipelineIncludesRecurring(t *testing.T) {
	got := crm.WeightedPipeline(10000, 1000, 50, 12)
	if got != 11000 {
		t.Fatalf("expected 11000, got %v", got)
	}
}
