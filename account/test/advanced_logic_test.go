package test

import (
	"testing"
	"time"

	"sumeru_addons/account/services"
)

func TestAdvanceRecurringDateMonthly(t *testing.T) {
	got := services.AdvanceRecurringDateForTest("2026-03-01", "monthly")
	if got != "2026-04-01" {
		t.Fatalf("got %s", got)
	}
	if services.AdvanceRecurringDateForTest("2026-03-01", "weekly") != "2026-03-08" {
		t.Fatal("weekly")
	}
}

func TestMonthsBetween(t *testing.T) {
	start, _ := time.Parse("2006-01-02", "2026-01-01")
	end, _ := time.Parse("2006-01-02", "2026-12-31")
	if n := services.MonthsBetweenForTest(start, end); n != 11 {
		t.Fatalf("months=%d", n)
	}
}
