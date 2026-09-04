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

// ScoreLead sets automated_probability from PLS frequency tables when configured.
func ScoreLead(ctx context.Context, leadID int) error {
	bypass := orm.ContextWithBypass(ctx, true)
	lead, err := orm.SearchOne(bypass, "crm.lead", map[string]interface{}{"id": leadID})
	if err != nil {
		return err
	}
	if !asBool(lead["is_automated_probability"]) {
		return nil
	}
	prob := frequencyTableProbability(bypass, lead)
	return orm.UpdateRecordByID(bypass, "crm.lead", leadID, map[string]interface{}{
		"automated_probability": prob,
		"probability":           prob,
	})
}

func frequencyTableProbability(ctx context.Context, lead map[string]interface{}) float64 {
	fields := strings.Split(orm.GetConfigParam(ctx, "crm.pls_fields", "stage_id"), ",")
	start := orm.GetConfigParam(ctx, "crm.pls_start_date", "")
	if start != "" {
		if created := strings.TrimSpace(orm.AsString(lead["create_date"])); created != "" && created < start {
			return stageWinRate(ctx, coerceInt64Field(lead["stage_id"]))
		}
	}
	won := 0.0
	lost := 0.0
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		val := plsFieldValue(lead, field)
		if val == "" {
			continue
		}
		w, l := lookupFrequency(ctx, field, val)
		won += w
		lost += l
	}
	total := won + lost
	if total <= 0 {
		return stageWinRate(ctx, coerceInt64Field(lead["stage_id"]))
	}
	return ClampProbability(won / total * 100)
}

func plsFieldValue(lead map[string]interface{}, field string) string {
	switch field {
	case "stage_id":
		if id, ok := orm.CoerceInt64(lead["stage_id"]); ok && id > 0 {
			return orm.AsString(id)
		}
	case "team_id":
		if id, ok := orm.CoerceInt64(lead["team_id"]); ok && id > 0 {
			return orm.AsString(id)
		}
	default:
		return strings.TrimSpace(orm.AsString(lead[field]))
	}
	return ""
}

func lookupFrequency(ctx context.Context, field, value string) (won, lost float64) {
	rows, err := orm.Search(ctx, "crm.lead.scoring.frequency.field", [][]interface{}{
		{"field", "=", field},
		{"value", "=", value},
	})
	if err != nil || len(rows) == 0 {
		return 0, 0
	}
	w, _ := orm.CoerceInt64(rows[0]["won_count"])
	l, _ := orm.CoerceInt64(rows[0]["lost_count"])
	return float64(w), float64(l)
}

func coerceInt64Field(v interface{}) int64 {
	n, _ := orm.CoerceInt64(v)
	return n
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
	if err := rebuildFrequencyTables(ctx); err != nil {
		return err
	}
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

func rebuildFrequencyTables(ctx context.Context) error {
	fields := strings.Split(orm.GetConfigParam(ctx, "crm.pls_fields", "stage_id"), ",")
	start := orm.GetConfigParam(ctx, "crm.pls_start_date", "")
	domain := [][]interface{}{{"active", "=", true}}
	if start != "" {
		domain = append(domain, []interface{}{"create_date", ">=", start})
	}
	leads, err := orm.Search(ctx, "crm.lead", domain)
	if err != nil {
		return err
	}
	type bucket struct{ won, lost int64 }
	counts := map[string]bucket{}
	for _, lead := range leads {
		won := orm.AsString(lead["won_status"]) == "won"
		lost := orm.AsString(lead["won_status"]) == "lost" || !asBool(lead["active"])
		if !won && !lost {
			continue
		}
		for _, field := range fields {
			field = strings.TrimSpace(field)
			val := plsFieldValue(lead, field)
			if val == "" {
				continue
			}
			key := field + "\x00" + val
			b := counts[key]
			if won {
				b.won++
			}
			if lost {
				b.lost++
			}
			counts[key] = b
		}
	}
	bypass := orm.ContextWithBypass(ctx, true)
	for key, b := range counts {
		parts := strings.SplitN(key, "\x00", 2)
		if len(parts) != 2 {
			continue
		}
		field, val := parts[0], parts[1]
		existing, _ := orm.Search(bypass, "crm.lead.scoring.frequency.field", [][]interface{}{
			{"field", "=", field},
			{"value", "=", val},
		})
		if len(existing) > 0 {
			id, _ := orm.CoerceInt64(existing[0]["id"])
			_ = orm.UpdateRecordByID(bypass, "crm.lead.scoring.frequency.field", int(id), map[string]interface{}{
				"won_count": b.won, "lost_count": b.lost,
			})
			continue
		}
		freqRows, _ := orm.Search(bypass, "crm.lead.scoring.frequency", [][]interface{}{{"variable", "=", field}})
		var freqID int64
		if len(freqRows) > 0 {
			freqID, _ = orm.CoerceInt64(freqRows[0]["id"])
		} else if m, ok := orm.Registry["crm.lead.scoring.frequency"]; ok {
			createdID, err := orm.Create(bypass, m, map[string]interface{}{"variable": field})
			if err == nil {
				freqID = int64(createdID)
			}
		}
		if freqID <= 0 {
			continue
		}
		if m, ok := orm.Registry["crm.lead.scoring.frequency.field"]; ok {
			_, _ = orm.Create(bypass, m, map[string]interface{}{
				"frequency_id": freqID,
				"field":        field,
				"value":        val,
				"won_count":    b.won,
				"lost_count":   b.lost,
			})
		}
	}
	return nil
}
