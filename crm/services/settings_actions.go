package services

import (
	"context"
	"fmt"
	"strings"

	"sumeru/core/orm"
)

func actionOpenPLSUpdate(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	_ = model
	_ = id
	start := orm.GetConfigParam(ctx, "crm.pls_start_date", "")
	fields := orm.GetConfigParam(ctx, "crm.pls_fields", "stage_id")
	wid, err := orm.Create(ctx, orm.Registry["crm.lead.pls.update"], map[string]interface{}{
		"start_date": start,
		"field_list": fields,
	})
	if err != nil {
		return "", err
	}
	return orm.ActionOpenURL("crm.lead.pls.update", wid), nil
}

func actionApplyPLSUpdate(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if model != "crm.lead.pls.update" || id <= 0 {
		return "", fmt.Errorf("invalid wizard")
	}
	wiz, err := orm.SearchOne(ctx, "crm.lead.pls.update", map[string]interface{}{"id": id})
	if err != nil {
		return "", err
	}
	bypass := orm.ContextWithBypass(ctx, true)
	if start := strings.TrimSpace(orm.AsString(wiz["start_date"])); start != "" {
		_ = orm.SetConfigParam(bypass, "crm.pls_start_date", start)
	}
	if fields := strings.TrimSpace(orm.AsString(wiz["field_list"])); fields != "" {
		_ = orm.SetConfigParam(bypass, "crm.pls_fields", fields)
	}
	if err := rebuildAllLeadScores(bypass); err != nil {
		return "", err
	}
	return orm.ActionCloseToken(), nil
}
