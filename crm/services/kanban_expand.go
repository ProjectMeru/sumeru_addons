package services

import (
	"context"
	"sort"
	"strings"

	"sumeru/core/engine/swcmeta"
	"sumeru/core/orm"
)

func onLeadTeamChange(ctx context.Context, values map[string]interface{}, _ string) (orm.OnchangeResult, error) {
	_ = ctx
	result := orm.OnchangeResult{Value: map[string]interface{}{}}
	teamID, _ := orm.CoerceInt64(values["team_id"])
	if teamID <= 0 {
		result.Value["user_id"] = false
		return result, nil
	}
	result.Domain = map[string]interface{}{
		"user_id": [][]interface{}{
			{"active", "=", true},
		},
	}
	return result, nil
}

func expandCRMStageColumns(ctx context.Context, model, groupField string, records []map[string]interface{}) ([]swcmeta.KanbanColumn, error) {
	_ = model
	_ = groupField
	teamID := dominantTeamID(records)
	stageTeams, _ := loadStageTeamRestrictions(ctx)
	stages, err := orm.Search(ctx, "crm.stage", [][]interface{}{{"active", "=", true}})
	if err != nil {
		return nil, err
	}
	sort.SliceStable(stages, func(i, j int) bool {
		si, _ := orm.CoerceInt64(stages[i]["sequence"])
		sj, _ := orm.CoerceInt64(stages[j]["sequence"])
		if si != sj {
			return si < sj
		}
		return orm.AsString(stages[i]["name"]) < orm.AsString(stages[j]["name"])
	})
	stages = filterStagesForTeam(stages, teamID, stageTeams)
	enrichLeadNextActivities(ctx, records)

	buckets := map[int64][]map[string]interface{}{}
	var unassigned []map[string]interface{}
	for _, row := range records {
		sid, _ := orm.CoerceInt64(row["stage_id"])
		if sid <= 0 {
			unassigned = append(unassigned, row)
			continue
		}
		buckets[sid] = append(buckets[sid], row)
	}

	cols := make([]swcmeta.KanbanColumn, 0, len(stages)+1)
	for _, st := range stages {
		id, ok := orm.CoerceInt64(st["id"])
		if !ok || id <= 0 {
			continue
		}
		fold := asBool(st["fold"])
		recs := buckets[id]
		if fold && len(recs) == 0 {
			continue
		}
		rottingDays := int(coerceInt(st["rotting_threshold_days"]))
		cols = append(cols, swcmeta.KanbanColumn{
			Value:       id,
			Label:       orm.AsString(st["name"]),
			Sequence:    int(coerceInt(st["sequence"])),
			Color:       int(coerceInt(st["color"])),
			RottingDays: rottingDays,
			Fold:        fold,
			Records:     recs,
			ProgressSum: columnProratedSum(recs),
		})
	}
	if len(unassigned) > 0 {
		cols = append([]swcmeta.KanbanColumn{{
			Value: 0, Label: "New", Sequence: -1, Records: unassigned,
			ProgressSum: columnProratedSum(unassigned),
		}}, cols...)
	}
	applyColumnProgressMax(cols)
	return cols, nil
}

func expandCRMForecastColumns(ctx context.Context, model, groupField string, records []map[string]interface{}) ([]swcmeta.KanbanColumn, error) {
	_ = ctx
	_ = model
	_ = groupField
	buckets := map[string][]map[string]interface{}{}
	order := []string{}
	for _, row := range records {
		key := forecastBucketKey(row["date_deadline"])
		if _, ok := buckets[key]; !ok {
			order = append(order, key)
		}
		buckets[key] = append(buckets[key], row)
	}
	sort.Strings(order)
	cols := make([]swcmeta.KanbanColumn, 0, len(order))
	for i, key := range order {
		cols = append(cols, swcmeta.KanbanColumn{
			Value:       int64(i + 1),
			Label:       key,
			Sequence:    i,
			Records:     buckets[key],
			ProgressSum: columnProratedSum(buckets[key]),
		})
	}
	applyColumnProgressMax(cols)
	return cols, nil
}

func forecastBucketKey(raw interface{}) string {
	s := strings.TrimSpace(orm.AsString(raw))
	if s == "" {
		return "No Deadline"
	}
	if len(s) >= 7 {
		return s[:7]
	}
	return s
}

func columnProratedSum(records []map[string]interface{}) float64 {
	var sum float64
	for _, row := range records {
		sum += numericFloat(row["prorated_revenue"])
	}
	return sum
}

func applyColumnProgressMax(cols []swcmeta.KanbanColumn) {
	var max float64
	for i := range cols {
		if cols[i].ProgressSum > max {
			max = cols[i].ProgressSum
		}
	}
	for i := range cols {
		cols[i].ProgressMax = max
	}
}

func dominantTeamID(records []map[string]interface{}) int64 {
	counts := map[int64]int{}
	for _, row := range records {
		tid, ok := orm.CoerceInt64(row["team_id"])
		if ok && tid > 0 {
			counts[tid]++
		}
	}
	var best int64
	max := 0
	for tid, n := range counts {
		if n > max {
			max = n
			best = tid
		}
	}
	return best
}

func loadStageTeamRestrictions(ctx context.Context) (map[int64][]int64, error) {
	out := map[int64][]int64{}
	if orm.DB == nil {
		return out, nil
	}
	rows, err := orm.DB.QueryContext(ctx, `SELECT stage_id, team_id FROM crm_stage_team_rel`)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var stageID, teamID int64
		if err := rows.Scan(&stageID, &teamID); err != nil {
			continue
		}
		out[stageID] = append(out[stageID], teamID)
	}
	return out, rows.Err()
}

func filterStagesForTeam(stages []map[string]interface{}, teamID int64, stageTeams map[int64][]int64) []map[string]interface{} {
	if teamID <= 0 || len(stageTeams) == 0 {
		return stages
	}
	filtered := make([]map[string]interface{}, 0, len(stages))
	for _, st := range stages {
		sid, ok := orm.CoerceInt64(st["id"])
		if !ok {
			continue
		}
		teams := stageTeams[sid]
		if len(teams) == 0 || stageTeamIncludes(teams, teamID) {
			filtered = append(filtered, st)
		}
	}
	return filtered
}

func stageTeamIncludes(teams []int64, teamID int64) bool {
	for _, t := range teams {
		if t == teamID {
			return true
		}
	}
	return false
}

func enrichLeadNextActivities(ctx context.Context, records []map[string]interface{}) {
	if len(records) == 0 || orm.DB == nil {
		return
	}
	leadIDs := make([]interface{}, 0, len(records))
	leadIndex := map[int64]map[string]interface{}{}
	for _, row := range records {
		id, ok := orm.CoerceInt64(row["id"])
		if !ok || id <= 0 {
			continue
		}
		leadIDs = append(leadIDs, id)
		leadIndex[id] = row
	}
	if len(leadIDs) == 0 {
		return
	}
	acts, err := orm.Search(ctx, "mail.activity", [][]interface{}{
		{"model", "=", "crm.lead"},
		{"res_id", "in", leadIDs},
		{"state", "=", "planned"},
	})
	if err != nil || len(acts) == 0 {
		return
	}
	nextByLead := map[int64]map[string]interface{}{}
	for _, act := range acts {
		lid, ok := orm.CoerceInt64(act["res_id"])
		if !ok || lid <= 0 {
			continue
		}
		prev, exists := nextByLead[lid]
		if !exists || activityDeadlineBefore(act, prev) {
			nextByLead[lid] = act
		}
	}
	for lid, act := range nextByLead {
		row, ok := leadIndex[lid]
		if !ok {
			continue
		}
		summary := strings.TrimSpace(orm.AsString(act["summary"]))
		if summary == "" {
			summary = strings.TrimSpace(orm.AsString(act["name"]))
		}
		if summary != "" {
			row["activity_summary"] = summary
		}
		if deadline := strings.TrimSpace(orm.AsString(act["date_deadline"])); deadline != "" {
			row["activity_deadline"] = deadline
		}
	}
}

func activityDeadlineBefore(candidate, current map[string]interface{}) bool {
	c := strings.TrimSpace(orm.AsString(candidate["date_deadline"]))
	p := strings.TrimSpace(orm.AsString(current["date_deadline"]))
	if c == "" {
		return false
	}
	if p == "" {
		return true
	}
	return c < p
}
