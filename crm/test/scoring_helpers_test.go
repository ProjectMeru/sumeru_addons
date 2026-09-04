package test

import (
	"testing"

	crm "sumeru_addons/crm/services"
)

func TestNormalizeEmail(t *testing.T) {
	if got := crm.NormalizeEmail("  Test@Example.COM "); got != "test@example.com" {
		t.Fatalf("expected normalized email, got %q", got)
	}
}

func TestColumnProratedSum(t *testing.T) {
	sum := crm.ColumnProratedSum([]map[string]interface{}{
		{"prorated_revenue": 100.0},
		{"prorated_revenue": 50.5},
	})
	if sum != 150.5 {
		t.Fatalf("expected 150.5, got %v", sum)
	}
}
