package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"sumeru/core/orm"
)

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

func actionMassConvertWizard(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	ids := collectMergeLeadIDs(id, vals["active_ids"])
	if len(ids) == 0 {
		return "", fmt.Errorf("no leads selected")
	}
	wid, err := orm.Create(ctx, orm.Registry["crm.lead2opportunity.mass"], map[string]interface{}{
		"lead_ids": joinIntIDs(ids),
	})
	if err != nil {
		return "", err
	}
	return orm.ActionOpenURL("crm.lead2opportunity.mass", wid), nil
}

func actionApplyMassConvert(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if model != "crm.lead2opportunity.mass" || id <= 0 {
		return "", fmt.Errorf("invalid wizard")
	}
	wiz, err := orm.SearchOne(ctx, "crm.lead2opportunity.mass", map[string]interface{}{"id": id})
	if err != nil {
		return "", err
	}
	for _, lid := range duplicateLeadIDs(orm.AsString(wiz["lead_ids"])) {
		upd := map[string]interface{}{"type": "opportunity", "date_conversion": time.Now().Format("2006-01-02")}
		if uid, ok := orm.CoerceInt64(wiz["user_id"]); ok && uid > 0 {
			upd["user_id"] = uid
		}
		if tid, ok := orm.CoerceInt64(wiz["team_id"]); ok && tid > 0 {
			upd["team_id"] = tid
		}
		if err := orm.UpdateRecordByID(ctx, "crm.lead", lid, upd); err != nil {
			return "", err
		}
	}
	return orm.ActionCloseToken(), nil
}
