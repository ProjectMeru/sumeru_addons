package services

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"sumeru/core/event"
	"sumeru/core/orm"
)

func onLeadCreated(ctx context.Context, ev event.Event) error {
	model, _ := ev.Payload["model"].(string)
	if model != "crm.lead" {
		return nil
	}
	id, ok := coerceID(ev.Payload["id"])
	if !ok {
		return nil
	}
	lead, err := orm.SearchOne(ctx, "crm.lead", map[string]interface{}{"id": id})
	if err != nil {
		return nil
	}
	updates := map[string]interface{}{}
	if sid, _ := orm.CoerceInt64(lead["stage_id"]); sid <= 0 {
		if stageID := firstStageID(ctx); stageID > 0 {
			updates["stage_id"] = stageID
		}
	}
	if orm.AsString(lead["date_open"]) == "" {
		updates["date_open"] = time.Now().Format("2006-01-02")
	}
	if uid, _ := orm.CoerceInt64(lead["user_id"]); uid <= 0 {
		if cur := orm.SecurityUID(ctx); cur > 0 {
			updates["user_id"] = cur
		}
	}
	now := time.Now().Format(time.RFC3339)
	updates["date_last_stage_update"] = now
	if len(updates) > 0 {
		bypass := orm.ContextWithBypass(ctx, true)
		_ = orm.UpdateRecordByID(bypass, "crm.lead", id, updates)
	}
	_ = ScoreLead(ctx, id)
	_ = syncLeadPartner(ctx, id)
	_ = refreshDuplicateLeads(ctx, id)
	return recomputeLeadRevenue(ctx, id)
}

func onLeadUpdated(ctx context.Context, ev event.Event) error {
	model, _ := ev.Payload["model"].(string)
	if model != "crm.lead" {
		return nil
	}
	id, ok := coerceID(ev.Payload["id"])
	if !ok {
		return nil
	}
	bypass := orm.ContextWithBypass(ctx, true)
	lead, err := orm.SearchOne(bypass, "crm.lead", map[string]interface{}{"id": id})
	if err != nil {
		return nil
	}
	stageID, _ := orm.CoerceInt64(lead["stage_id"])
	updates := map[string]interface{}{"date_last_stage_update": time.Now().Format(time.RFC3339)}
	if stageID > 0 {
		if stage, err := orm.SearchOne(bypass, "crm.stage", map[string]interface{}{"id": stageID}); err == nil {
			if asBool(stage["is_won"]) {
				if prob, _ := orm.CoerceInt64(lead["probability"]); prob != 100 {
					updates["probability"] = 100.0
				}
				if orm.AsString(lead["won_status"]) != "won" {
					updates["won_status"] = "won"
				}
				if orm.AsString(lead["date_closed"]) == "" {
					updates["date_closed"] = time.Now().Format("2006-01-02")
				}
			}
		}
	}
	if len(updates) > 0 {
		if err := orm.UpdateRecordByID(bypass, "crm.lead", id, updates); err != nil {
			return err
		}
	}
	if orm.AsString(lead["type"]) == "opportunity" && orm.AsString(lead["date_conversion"]) == "" {
		_ = orm.UpdateRecordByID(bypass, "crm.lead", id, map[string]interface{}{
			"date_conversion": time.Now().Format("2006-01-02"),
		})
	}
	if kpi := updateLeadDayKPIs(ctx, lead); len(kpi) > 0 {
		_ = orm.UpdateRecordByID(bypass, "crm.lead", id, kpi)
	}
	_ = ScoreLead(ctx, id)
	_ = syncLeadPartner(ctx, id)
	_ = refreshDuplicateLeads(ctx, id)
	return recomputeLeadRevenue(ctx, id)
}

func updateLeadDayKPIs(ctx context.Context, lead map[string]interface{}) map[string]interface{} {
	_ = ctx
	out := map[string]interface{}{}
	if open := strings.TrimSpace(orm.AsString(lead["date_open"])); open != "" {
		if t, err := time.Parse("2006-01-02", open); err == nil {
			out["day_open"] = int64(time.Since(t).Hours() / 24)
		}
	}
	if closed := strings.TrimSpace(orm.AsString(lead["date_closed"])); closed != "" {
		if t, err := time.Parse("2006-01-02", closed); err == nil {
			out["day_close"] = int64(time.Since(t).Hours() / 24)
		}
	}
	return out
}

func firstStageID(ctx context.Context) int64 {
	rows, err := orm.Search(ctx, "crm.stage", [][]interface{}{{"active", "=", true}})
	if err != nil || len(rows) == 0 {
		return 0
	}
	sort.SliceStable(rows, func(i, j int) bool {
		si, _ := orm.CoerceInt64(rows[i]["sequence"])
		sj, _ := orm.CoerceInt64(rows[j]["sequence"])
		return si < sj
	})
	id, _ := orm.CoerceInt64(rows[0]["id"])
	return id
}

func wonStageID(ctx context.Context) int64 {
	rows, err := orm.Search(ctx, "crm.stage", [][]interface{}{{"is_won", "=", true}})
	if err != nil || len(rows) == 0 {
		return 0
	}
	id, _ := orm.CoerceInt64(rows[0]["id"])
	return id
}

func onCronAssignLeads(ctx context.Context, ev event.Event) error {
	if !orm.ConfigParamBool(ctx, "crm.lead.auto.assignment", false) {
		return nil
	}
	teams, err := orm.Search(ctx, "crm.team", [][]interface{}{
		{"assignment_enabled", "=", true},
		{"active", "=", true},
	})
	if err != nil {
		return err
	}
	for _, team := range teams {
		tid, ok := orm.CoerceInt64(team["id"])
		if !ok || tid <= 0 {
			continue
		}
		_ = assignTeamLeads(ctx, int(tid))
	}
	return nil
}

func onCronPLSRebuild(ctx context.Context, ev event.Event) error {
	return rebuildAllLeadScores(ctx)
}

func onActivityCreated(ctx context.Context, ev event.Event) error {
	model, _ := ev.Payload["model"].(string)
	if model != "mail.activity" {
		return nil
	}
	id, ok := coerceID(ev.Payload["id"])
	if !ok {
		return nil
	}
	return syncActivityReport(ctx, id)
}

func onActivityUpdated(ctx context.Context, ev event.Event) error {
	return onActivityCreated(ctx, ev)
}

func syncActivityReport(ctx context.Context, activityID int) error {
	act, err := orm.SearchOne(ctx, "mail.activity", map[string]interface{}{"id": activityID})
	if err != nil || orm.AsString(act["model"]) != "crm.lead" {
		return nil
	}
	lid, _ := orm.CoerceInt64(act["res_id"])
	if lid <= 0 {
		return nil
	}
	lead, err := orm.SearchOne(ctx, "crm.lead", map[string]interface{}{"id": lid})
	if err != nil {
		return nil
	}
	vals := map[string]interface{}{
		"activity_id":   activityID,
		"lead_id":       lid,
		"user_id":       act["user_id"],
		"team_id":       lead["team_id"],
		"stage_id":      lead["stage_id"],
		"summary":       act["summary"],
		"date_deadline": act["date_deadline"],
		"state":         act["state"],
	}
	existing, _ := orm.Search(ctx, "crm.activity.report", [][]interface{}{{"activity_id", "=", activityID}})
	if len(existing) > 0 {
		rid, _ := orm.CoerceInt64(existing[0]["id"])
		return orm.UpdateRecordByID(ctx, "crm.activity.report", int(rid), vals)
	}
	m, ok := orm.Registry["crm.activity.report"]
	if !ok {
		return fmt.Errorf("crm.activity.report not registered")
	}
	_, err = orm.Create(ctx, m, vals)
	return err
}
