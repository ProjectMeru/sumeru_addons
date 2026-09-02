package services

import (
	"context"
	"fmt"

	"sumeru/core/orm"
)

func init() {
	orm.RegisterObjectAction("crm.livechat.session", "action_create_lead", actionCreateLeadFromSession)
}

func actionCreateLeadFromSession(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if model != "crm.livechat.session" || id <= 0 {
		return "", fmt.Errorf("invalid session")
	}
	bypass := orm.ContextWithBypass(ctx, true)
	session, err := orm.SearchOne(bypass, "crm.livechat.session", map[string]interface{}{"id": id})
	if err != nil {
		return "", err
	}
	if leadID, ok := orm.CoerceInt64(session["lead_id"]); ok && leadID > 0 {
		return fmt.Sprintf("/web?action=crm.action_leads&view_type=form&id=%d", leadID), nil
	}
	leadModel, ok := orm.Registry["crm.lead"]
	if !ok {
		return "", fmt.Errorf("crm.lead missing")
	}
	name := orm.AsString(session["name"])
	if name == "" {
		name = fmt.Sprintf("Live Chat Session %d", id)
	}
	leadID, err := orm.Create(bypass, leadModel, map[string]interface{}{
		"name":        name,
		"type":        "lead",
		"description": "Created from live chat session (stub).",
	})
	if err != nil {
		return "", err
	}
	if err := orm.UpdateRecordByID(bypass, "crm.livechat.session", id, map[string]interface{}{
		"lead_id": leadID,
		"state":   "done",
	}); err != nil {
		return "", err
	}
	return fmt.Sprintf("/web?action=crm.action_leads&view_type=form&id=%d", leadID), nil
}
