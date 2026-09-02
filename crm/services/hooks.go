package services

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"sumeru/core/engine/swcmeta"
	"sumeru/core/event"
	"sumeru/core/orm"
)

func init() {
	swcmeta.RegisterKanbanGroupExpander("crm.lead", "stage_id", expandCRMStageColumns)
	swcmeta.RegisterKanbanGroupExpander("crm.lead", "date_deadline", expandCRMForecastColumns)

	orm.RegisterOnchange("crm.lead", "team_id", onLeadTeamChange)

	event.Subscribe("record.created", onLeadCreated)
	event.Subscribe("record.updated", onLeadUpdated)
	event.Subscribe("record.created", onActivityCreated)
	event.Subscribe("record.updated", onActivityUpdated)
	event.Subscribe("crm.cron_assign_leads", onCronAssignLeads)
	event.Subscribe("crm.cron_pls_rebuild", onCronPLSRebuild)

	registerObjectActions()
}

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
		cols = append(cols, swcmeta.KanbanColumn{
			Value:    id,
			Label:    orm.AsString(st["name"]),
			Sequence: int(coerceInt(st["sequence"])),
			Color:    int(coerceInt(st["color"])),
			Fold:     fold,
			Records:  recs,
		})
	}
	if len(unassigned) > 0 {
		cols = append([]swcmeta.KanbanColumn{{
			Value: 0, Label: "New", Sequence: -1, Records: unassigned,
		}}, cols...)
	}
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
			Value:    int64(i + 1),
			Label:    key,
			Sequence: i,
			Records:  buckets[key],
		})
	}
	return cols, nil
}

func forecastBucketKey(raw interface{}) string {
	s := strings.TrimSpace(orm.AsString(raw))
	if s == "" {
		return "No Deadline"
	}
	if len(s) >= 7 {
		return s[:7] // YYYY-MM
	}
	return s
}

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
	if len(updates) > 0 {
		bypass := orm.ContextWithBypass(ctx, true)
		_ = orm.UpdateRecordByID(bypass, "crm.lead", id, updates)
	}
	_ = ScoreLead(ctx, id)
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
	updates := map[string]interface{}{}
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
	_ = ScoreLead(ctx, id)
	return recomputeLeadRevenue(ctx, id)
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

func coerceID(v interface{}) (int, bool) {
	n, ok := orm.CoerceInt64(v)
	return int(n), ok && n > 0
}

func coerceInt(v interface{}) int64 {
	n, _ := orm.CoerceInt64(v)
	return n
}

func asBool(v interface{}) bool {
	switch t := v.(type) {
	case bool:
		return t
	case int64:
		return t != 0
	case int:
		return t != 0
	case string:
		return t == "true" || t == "t" || t == "1"
	default:
		return false
	}
}

func wonStageID(ctx context.Context) int64 {
	rows, err := orm.Search(ctx, "crm.stage", [][]interface{}{{"is_won", "=", true}})
	if err != nil || len(rows) == 0 {
		return 0
	}
	id, _ := orm.CoerceInt64(rows[0]["id"])
	return id
}

func actionSetWon(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if model != "crm.lead" || id <= 0 {
		return "", fmt.Errorf("invalid lead")
	}
	sid := wonStageID(ctx)
	upd := map[string]interface{}{"won_status": "won", "probability": 100.0}
	if sid > 0 {
		upd["stage_id"] = sid
	}
	if err := orm.UpdateRecordByID(ctx, "crm.lead", id, upd); err != nil {
		return "", err
	}
	return vals["next"], nil
}

func actionRestore(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if model != "crm.lead" || id <= 0 {
		return "", fmt.Errorf("invalid lead")
	}
	upd := map[string]interface{}{
		"active":         true,
		"won_status":     "pending",
		"lost_reason_id": nil,
	}
	if sid := firstStageID(ctx); sid > 0 {
		upd["stage_id"] = sid
	}
	if err := orm.UpdateRecordByID(ctx, "crm.lead", id, upd); err != nil {
		return "", err
	}
	return vals["next"], nil
}

func actionConvertOpportunity(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if model != "crm.lead" || id <= 0 {
		return "", fmt.Errorf("invalid lead")
	}
	if err := orm.UpdateRecordByID(ctx, "crm.lead", id, map[string]interface{}{"type": "opportunity"}); err != nil {
		return "", err
	}
	return vals["next"], nil
}

func actionAssignTeamLeads(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if model != "crm.team" || id <= 0 {
		return "", fmt.Errorf("invalid team")
	}
	if err := assignTeamLeads(ctx, id); err != nil {
		return "", err
	}
	return vals["next"], nil
}

func actionLostWizard(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if model != "crm.lead" || id <= 0 {
		return "", fmt.Errorf("invalid lead")
	}
	wid, err := orm.Create(ctx, orm.Registry["crm.lead.lost"], map[string]interface{}{"lead_id": id})
	if err != nil {
		return "", err
	}
	return orm.ActionOpenURL("crm.lead.lost", wid), nil
}

func actionApplyLost(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if model != "crm.lead.lost" || id <= 0 {
		return "", fmt.Errorf("invalid wizard")
	}
	wiz, err := orm.SearchOne(ctx, "crm.lead.lost", map[string]interface{}{"id": id})
	if err != nil {
		return "", err
	}
	lid, _ := orm.CoerceInt64(wiz["lead_id"])
	if lid <= 0 {
		return "", fmt.Errorf("missing lead")
	}
	upd := map[string]interface{}{"active": false, "won_status": "lost"}
	if rid, ok := orm.CoerceInt64(wiz["lost_reason_id"]); ok && rid > 0 {
		upd["lost_reason_id"] = rid
	}
	upd["lost_feedback"] = orm.AsString(wiz["lost_feedback"])
	if err := orm.UpdateRecordByID(ctx, "crm.lead", int(lid), upd); err != nil {
		return "", err
	}
	if next := strings.TrimSpace(vals["next"]); next != "" {
		return next, nil
	}
	return orm.ActionCloseToken(), nil
}

func actionConvertWizard(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if model != "crm.lead" || id <= 0 {
		return "", fmt.Errorf("invalid lead")
	}
	lead, err := orm.SearchOne(ctx, "crm.lead", map[string]interface{}{"id": id})
	if err != nil {
		return "", err
	}
	wid, err := orm.Create(ctx, orm.Registry["crm.lead2opportunity"], map[string]interface{}{
		"lead_id": id, "partner_id": lead["partner_id"], "name": lead["name"],
		"user_id": lead["user_id"], "team_id": lead["team_id"],
	})
	if err != nil {
		return "", err
	}
	return orm.ActionOpenURL("crm.lead2opportunity", wid), nil
}

func actionMergeWizard(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if model != "crm.lead" || id <= 0 {
		return "", fmt.Errorf("invalid lead")
	}
	ids := collectMergeLeadIDs(id, vals["active_ids"])
	if len(ids) == 0 {
		return "", fmt.Errorf("no leads selected")
	}
	name := ""
	if lead, err := orm.SearchOne(ctx, "crm.lead", map[string]interface{}{"id": ids[0]}); err == nil {
		name = orm.AsString(lead["name"])
	}
	wid, err := orm.Create(ctx, orm.Registry["crm.merge.opportunity"], map[string]interface{}{
		"lead_ids": joinIntIDs(ids),
		"name":     name,
	})
	if err != nil {
		return "", err
	}
	return orm.ActionOpenURL("crm.merge.opportunity", wid), nil
}

func collectMergeLeadIDs(current int, activeCSV string) []int {
	seen := map[int]struct{}{}
	var ids []int
	add := func(n int) {
		if n <= 0 {
			return
		}
		if _, ok := seen[n]; ok {
			return
		}
		seen[n] = struct{}{}
		ids = append(ids, n)
	}
	for _, part := range strings.Split(activeCSV, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err == nil {
			add(n)
		}
	}
	add(current)
	return ids
}

func joinIntIDs(ids []int) string {
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		parts = append(parts, strconv.Itoa(id))
	}
	return strings.Join(parts, ",")
}

func actionApplyConvert(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if model != "crm.lead2opportunity" || id <= 0 {
		return "", fmt.Errorf("invalid wizard")
	}
	wiz, err := orm.SearchOne(ctx, "crm.lead2opportunity", map[string]interface{}{"id": id})
	if err != nil {
		return "", err
	}
	lid, _ := orm.CoerceInt64(wiz["lead_id"])
	if lid <= 0 {
		return "", fmt.Errorf("missing lead")
	}
	upd := map[string]interface{}{"type": "opportunity"}
	if n := strings.TrimSpace(orm.AsString(wiz["name"])); n != "" {
		upd["name"] = n
	}
	if pid, ok := orm.CoerceInt64(wiz["partner_id"]); ok && pid > 0 {
		upd["partner_id"] = pid
	}
	if uid, ok := orm.CoerceInt64(wiz["user_id"]); ok && uid > 0 {
		upd["user_id"] = uid
	}
	if tid, ok := orm.CoerceInt64(wiz["team_id"]); ok && tid > 0 {
		upd["team_id"] = tid
	}
	if err := orm.UpdateRecordByID(ctx, "crm.lead", int(lid), upd); err != nil {
		return "", err
	}
	if next := strings.TrimSpace(vals["next"]); next != "" {
		return next, nil
	}
	return orm.ActionCloseToken(), nil
}

func actionApplyMerge(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if model != "crm.merge.opportunity" || id <= 0 {
		return "", fmt.Errorf("invalid wizard")
	}
	wiz, err := orm.SearchOne(ctx, "crm.merge.opportunity", map[string]interface{}{"id": id})
	if err != nil {
		return "", err
	}
	rawIDs := strings.TrimSpace(orm.AsString(wiz["lead_ids"]))
	if rawIDs == "" {
		return "", fmt.Errorf("no leads selected")
	}
	var ids []int
	for _, part := range strings.Split(rawIDs, ",") {
		if n, err := strconv.Atoi(strings.TrimSpace(part)); err == nil && n > 0 {
			ids = append(ids, n)
		}
	}
	if len(ids) < 2 {
		return "", fmt.Errorf("select at least two leads")
	}
	keepID := ids[0]
	merged := mergeLeadFields(ctx, ids)
	if name := strings.TrimSpace(orm.AsString(wiz["name"])); name != "" {
		merged["name"] = name
	}
	if err := orm.UpdateRecordByID(ctx, "crm.lead", keepID, merged); err != nil {
		return "", err
	}
	for _, lid := range ids[1:] {
		if err := orm.UpdateRecordByID(ctx, "crm.lead", lid, map[string]interface{}{"active": false}); err != nil {
			return "", err
		}
	}
	if next := strings.TrimSpace(vals["next"]); next != "" {
		return next, nil
	}
	return orm.ActionCloseToken(), nil
}

func onCronAssignLeads(ctx context.Context, ev event.Event) error {
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

func mergeLeadFields(ctx context.Context, ids []int) map[string]interface{} {
	out := map[string]interface{}{}
	maxRevenue := 0.0
	maxRecurring := 0.0
	for _, id := range ids {
		lead, err := orm.SearchOne(ctx, "crm.lead", map[string]interface{}{"id": id})
		if err != nil {
			continue
		}
		if rev := numericFloat(lead["expected_revenue"]); rev > maxRevenue {
			maxRevenue = rev
		}
		if rec := numericFloat(lead["recurring_revenue"]); rec > maxRecurring {
			maxRecurring = rec
		}
		if out["partner_id"] == nil {
			if pid, ok := orm.CoerceInt64(lead["partner_id"]); ok && pid > 0 {
				out["partner_id"] = pid
			}
		}
		if out["description"] == nil && orm.AsString(lead["description"]) != "" {
			out["description"] = lead["description"]
		}
	}
	if maxRevenue > 0 {
		out["expected_revenue"] = maxRevenue
	}
	if maxRecurring > 0 {
		out["recurring_revenue"] = maxRecurring
	}
	return out
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

func registerObjectActions() {
	orm.RegisterObjectAction("crm.lead", "action_set_won", actionSetWon)
	orm.RegisterObjectAction("crm.lead", "action_restore", actionRestore)
	orm.RegisterObjectAction("crm.lead", "action_convert_opportunity", actionConvertOpportunity)
	orm.RegisterObjectAction("crm.lead", "action_lost_wizard", actionLostWizard)
	orm.RegisterObjectAction("crm.lead", "action_convert_wizard", actionConvertWizard)
	orm.RegisterObjectAction("crm.lead", "action_merge_wizard", actionMergeWizard)
	orm.RegisterObjectAction("crm.lead.lost", "action_apply_lost", actionApplyLost)
	orm.RegisterObjectAction("crm.lead2opportunity", "action_apply_convert", actionApplyConvert)
	orm.RegisterObjectAction("crm.merge.opportunity", "action_apply_merge", actionApplyMerge)
	orm.RegisterObjectAction("crm.team", "action_assign_leads", actionAssignTeamLeads)
}
