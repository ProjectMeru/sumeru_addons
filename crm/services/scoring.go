package services

import (
	"context"
	"math"
	"strings"

	"sumeru/core/orm"
)

// ClampProbability bounds probability to 0..100.
func ClampProbability(p float64) float64 {
	if p < 0 {
		return 0
	}
	if p > 100 {
		return 100
	}
	return math.Round(p*100) / 100
}

// ScoreLead sets automated_probability from simple historical win rate by stage.
func ScoreLead(ctx context.Context, leadID int) error {
	bypass := orm.ContextWithBypass(ctx, true)
	lead, err := orm.SearchOne(bypass, "crm.lead", map[string]interface{}{"id": leadID})
	if err != nil {
		return err
	}
	if !asBool(lead["is_automated_probability"]) {
		return nil
	}
	stageID, _ := orm.CoerceInt64(lead["stage_id"])
	prob := stageWinRate(bypass, stageID)
	return orm.UpdateRecordByID(bypass, "crm.lead", leadID, map[string]interface{}{
		"automated_probability": prob,
		"probability":           prob,
	})
}

func stageWinRate(ctx context.Context, stageID int64) float64 {
	if stageID <= 0 {
		return 10
	}
	stage, err := orm.SearchOne(ctx, "crm.stage", map[string]interface{}{"id": stageID})
	if err != nil {
		return 10
	}
	if asBool(stage["is_won"]) {
		return 100
	}
	name := strings.ToLower(orm.AsString(stage["name"]))
	switch {
	case strings.Contains(name, "propos"):
		return 60
	case strings.Contains(name, "qualif"):
		return 30
	case strings.Contains(name, "new"):
		return 10
	default:
		seq, _ := orm.CoerceInt64(stage["sequence"])
		return ClampProbability(float64(seq) * 5)
	}
}

func rebuildAllLeadScores(ctx context.Context) error {
	rows, err := orm.Search(ctx, "crm.lead", [][]interface{}{{"active", "=", true}})
	if err != nil {
		return err
	}
	for _, row := range rows {
		id, ok := orm.CoerceInt64(row["id"])
		if !ok {
			continue
		}
		if err := ScoreLead(ctx, int(id)); err != nil {
			return err
		}
		if err := recomputeLeadRevenue(ctx, int(id)); err != nil {
			return err
		}
	}
	return nil
}
