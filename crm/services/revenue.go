package services

import (
	"context"
	"math"
	"strconv"

	"sumeru/core/orm"
)

// ProratedRevenue returns expected revenue weighted by win probability.
func ProratedRevenue(expectedRevenue, probability float64) float64 {
	if expectedRevenue <= 0 {
		return 0
	}
	if probability < 0 {
		probability = 0
	}
	if probability > 100 {
		probability = 100
	}
	return roundMoney(expectedRevenue * probability / 100)
}

// WeightedPipeline includes recurring revenue scaled by plan months.
func WeightedPipeline(expected, recurring, probability, planMonths float64) float64 {
	base := expected
	if recurring > 0 && planMonths > 0 {
		base += recurring * planMonths
	}
	return ProratedRevenue(base, probability)
}

func roundMoney(v float64) float64 {
	return math.Round(v*100) / 100
}

func recomputeLeadRevenue(ctx context.Context, leadID int) error {
	bypass := orm.ContextWithBypass(ctx, true)
	lead, err := orm.SearchOne(bypass, "crm.lead", map[string]interface{}{"id": leadID})
	if err != nil {
		return err
	}
	expected := numericFloat(lead["expected_revenue"])
	recurring := numericFloat(lead["recurring_revenue"])
	prob := numericFloat(lead["probability"])
	if asBool(lead["is_automated_probability"]) {
		prob = numericFloat(lead["automated_probability"])
	}
	planMonths := 1.0
	if planID, ok := orm.CoerceInt64(lead["recurring_plan"]); ok && planID > 0 {
		if plan, err := orm.SearchOne(bypass, "crm.recurring.plan", map[string]interface{}{"id": planID}); err == nil {
			if m, ok := orm.CoerceInt64(plan["number_of_months"]); ok && m > 0 {
				planMonths = float64(m)
			}
		}
	}
	return orm.UpdateRecordByID(bypass, "crm.lead", leadID, map[string]interface{}{
		"prorated_revenue": WeightedPipeline(expected, recurring, prob, planMonths),
	})
}

func numericFloat(v interface{}) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int64:
		return float64(t)
	case int:
		return float64(t)
	default:
		s := orm.AsString(v)
		if s == "" {
			return 0
		}
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0
		}
		return f
	}
}
