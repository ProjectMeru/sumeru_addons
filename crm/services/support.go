package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"sumeru/core/orm"
)

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
	case string:
		f, err := strconv.ParseFloat(t, 64)
		if err == nil {
			return f
		}
	}
	return 0
}

func joinIntIDs(ids []int) string {
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		parts = append(parts, strconv.Itoa(id))
	}
	return strings.Join(parts, ",")
}

func refreshDuplicateLeads(ctx context.Context, leadID int) error {
	bypass := orm.ContextWithBypass(ctx, true)
	lead, err := orm.SearchOne(bypass, "crm.lead", map[string]interface{}{"id": leadID})
	if err != nil {
		return err
	}
	email := normalizeEmail(orm.AsString(lead["email_from"]))
	phone := normalizePhone(orm.AsString(lead["phone"]))
	domain := make([][]interface{}, 0, 4)
	domain = append(domain, []interface{}{"id", "!=", leadID})
	domain = append(domain, []interface{}{"active", "=", true})
	clauses := make([][]interface{}, 0, 2)
	if email != "" {
		clauses = append(clauses, []interface{}{"email_from", "ilike", email})
	}
	if phone != "" {
		clauses = append(clauses, []interface{}{"phone", "=", phone})
	}
	if len(clauses) == 0 {
		return orm.UpdateRecordByID(bypass, "crm.lead", leadID, map[string]interface{}{
			"duplicate_lead_count": 0,
			"duplicate_lead_ids":   "",
		})
	}
	var dupes []map[string]interface{}
	for _, clause := range clauses {
		q := append(append([][]interface{}{}, domain...), clause)
		rows, err := orm.Search(bypass, "crm.lead", q)
		if err != nil {
			return err
		}
		dupes = append(dupes, rows...)
	}
	seen := map[int64]struct{}{}
	var ids []int
	for _, row := range dupes {
		id, ok := orm.CoerceInt64(row["id"])
		if !ok || id <= 0 {
			continue
		}
		if _, hit := seen[id]; hit {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, int(id))
	}
	return orm.UpdateRecordByID(bypass, "crm.lead", leadID, map[string]interface{}{
		"duplicate_lead_count": int64(len(ids)),
		"duplicate_lead_ids":   joinIntIDs(ids),
	})
}

func normalizeEmail(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func normalizePhone(raw string) string {
	var b strings.Builder
	for _, r := range raw {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func duplicateLeadIDs(raw string) []int {
	var out []int
	for _, part := range strings.Split(raw, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err == nil && n > 0 {
			out = append(out, n)
		}
	}
	return out
}

func syncLeadPartner(ctx context.Context, leadID int) error {
	bypass := orm.ContextWithBypass(ctx, true)
	lead, err := orm.SearchOne(bypass, "crm.lead", map[string]interface{}{"id": leadID})
	if err != nil {
		return err
	}
	partnerID, _ := orm.CoerceInt64(lead["partner_id"])
	if partnerID <= 0 {
		return nil
	}
	partner, err := orm.SearchOne(bypass, "core.partner", map[string]interface{}{"id": partnerID})
	if err != nil {
		return err
	}
	upd := map[string]interface{}{}
	if email := strings.TrimSpace(orm.AsString(lead["email_from"])); email != "" && orm.AsString(partner["email"]) == "" {
		upd["email"] = email
	}
	if phone := strings.TrimSpace(orm.AsString(lead["phone"])); phone != "" && orm.AsString(partner["phone"]) == "" {
		upd["phone"] = phone
	}
	if mobile := strings.TrimSpace(orm.AsString(lead["mobile"])); mobile != "" && orm.AsString(partner["mobile"]) == "" {
		upd["mobile"] = mobile
	}
	for _, pair := range [][2]string{
		{"street", "street"},
		{"street2", "street2"},
		{"city", "city"},
		{"zip", "zip"},
	} {
		if v := strings.TrimSpace(orm.AsString(lead[pair[0]])); v != "" && orm.AsString(partner[pair[1]]) == "" {
			upd[pair[1]] = v
		}
	}
	if sid, ok := orm.CoerceInt64(lead["state_id"]); ok && sid > 0 && partner["state_id"] == nil {
		upd["state_id"] = sid
	}
	if cid, ok := orm.CoerceInt64(lead["country_id"]); ok && cid > 0 && partner["country_id"] == nil {
		upd["country_id"] = cid
	}
	if len(upd) == 0 {
		return nil
	}
	return orm.UpdateRecordByID(bypass, "core.partner", int(partnerID), upd)
}

const (
	cfgAutoAssignment     = "crm.lead.auto.assignment"
	cfgMembershipMulti    = "sales_team.membership_multi"
	cfgLeadEnrichAuto     = "crm.iap.lead.enrich.setting"
	cfgLeadMiningPipeline = "crm.lead_mining_in_pipeline"
	cfgPLSStartDate       = "crm.pls_start_date"
	cfgPLSFields          = "crm.pls_fields"
)

func actionSaveCRMSettings(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if model != "res.config.settings" || id <= 0 {
		return "", fmt.Errorf("invalid settings")
	}
	row, err := orm.SearchOne(ctx, "res.config.settings", map[string]interface{}{"id": id})
	if err != nil {
		return "", err
	}
	bypass := orm.ContextWithBypass(ctx, true)
	if err := applyCRMSettings(bypass, row); err != nil {
		return "", err
	}
	return orm.ActionCloseToken(), nil
}

func applyCRMSettings(ctx context.Context, row map[string]interface{}) error {
	if err := orm.SetConfigParam(ctx, cfgAutoAssignment, boolParam(asBool(row["crm_use_auto_assignment"]))); err != nil {
		return err
	}
	if err := orm.SetConfigParam(ctx, cfgMembershipMulti, boolParam(asBool(row["is_membership_multi"]))); err != nil {
		return err
	}
	if err := orm.SetConfigParam(ctx, cfgLeadEnrichAuto, orm.AsString(row["lead_enrich_auto"])); err != nil {
		return err
	}
	if err := orm.SetConfigParam(ctx, cfgLeadMiningPipeline, boolParam(asBool(row["lead_mining_in_pipeline"]))); err != nil {
		return err
	}
	if start := strings.TrimSpace(orm.AsString(row["predictive_lead_scoring_start_date"])); start != "" {
		if err := orm.SetConfigParam(ctx, cfgPLSStartDate, start); err != nil {
			return err
		}
	}
	if fields := strings.TrimSpace(orm.AsString(row["predictive_lead_scoring_fields_str"])); fields != "" {
		if err := orm.SetConfigParam(ctx, cfgPLSFields, fields); err != nil {
			return err
		}
	}
	if asBool(row["group_use_lead"]) {
		teams, _ := orm.Search(ctx, "crm.team", [][]interface{}{{"use_opportunities", "=", true}})
		for _, team := range teams {
			tid, ok := orm.CoerceInt64(team["id"])
			if !ok {
				continue
			}
			_ = orm.UpdateRecordByID(ctx, "crm.team", int(tid), map[string]interface{}{"use_leads": true})
		}
	}
	return syncAssignmentCron(ctx, row)
}

func boolParam(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func syncAssignmentCron(ctx context.Context, row map[string]interface{}) error {
	rows, err := orm.Search(ctx, "sys.cron", [][]interface{}{{"event_name", "=", "crm.cron_assign_leads"}})
	if err != nil || len(rows) == 0 {
		return err
	}
	cronID, _ := orm.CoerceInt64(rows[0]["id"])
	active := asBool(row["crm_use_auto_assignment"]) && orm.AsString(row["crm_auto_assignment_action"]) == "auto"
	interval := coerceInt(row["crm_auto_assignment_interval_number"])
	if interval <= 0 {
		interval = 1
	}
	unit := orm.AsString(row["crm_auto_assignment_interval_type"])
	minutes := cronIntervalMinutes(interval, unit)
	upd := map[string]interface{}{
		"active":            active,
		"interval_number": minutes,
	}
	if next := strings.TrimSpace(orm.AsString(row["crm_auto_assignment_run_datetime"])); next != "" {
		upd["next_call"] = next
	}
	return orm.UpdateRecordByID(orm.ContextWithBypass(ctx, true), "sys.cron", int(cronID), upd)
}

func cronIntervalMinutes(n int64, unit string) int64 {
	switch unit {
	case "minutes":
		return n
	case "hours":
		return n * 60
	case "weeks":
		return n * 7 * 24 * 60
	default:
		return n * 24 * 60
	}
}

// DefaultCRMSettingsRow builds default values for a new settings record.
func DefaultCRMSettingsRow(ctx context.Context) map[string]interface{} {
	start := time.Now().AddDate(0, 0, -8).Format("2006-01-02")
	return map[string]interface{}{
		"group_use_lead":                       true,
		"group_use_recurring_revenues":         false,
		"is_membership_multi":                  orm.ConfigParamBool(ctx, cfgMembershipMulti, false),
		"crm_use_auto_assignment":              orm.ConfigParamBool(ctx, cfgAutoAssignment, false),
		"crm_auto_assignment_action":           "manual",
		"crm_auto_assignment_interval_number":  int64(1),
		"crm_auto_assignment_interval_type":    "days",
		"lead_enrich_auto":                     orm.GetConfigParam(ctx, cfgLeadEnrichAuto, "auto"),
		"lead_mining_in_pipeline":              orm.ConfigParamBool(ctx, cfgLeadMiningPipeline, false),
		"predictive_lead_scoring_start_date":   orm.GetConfigParam(ctx, cfgPLSStartDate, start),
		"predictive_lead_scoring_fields_str":   orm.GetConfigParam(ctx, cfgPLSFields, ""),
		"predictive_lead_scoring_field_labels": "Stage",
	}
}
