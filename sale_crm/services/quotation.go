package services

import (
	"context"
	"fmt"

	"sumeru/core/event"
	"sumeru/core/orm"
)

func init() {
	event.Subscribe("record.updated", onLeadWonCreateQuotation)
	orm.RegisterObjectAction("crm.lead", "action_create_quotation", actionCreateQuotation)
	orm.RegisterObjectAction("crm.lead", "action_quotation_wizard", actionQuotationWizard)
	orm.RegisterObjectAction("crm.opportunity.to.quotation", "action_apply_quotation", actionApplyQuotation)
}

func actionCreateQuotation(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if model != "crm.lead" || id <= 0 {
		return "", fmt.Errorf("invalid lead")
	}
	orderID, err := CreateQuotationFromLead(ctx, id, nil)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("/web?action=sale.action_sale_quotations&view_type=form&id=%d", orderID), nil
}

func actionQuotationWizard(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if model != "crm.lead" || id <= 0 {
		return "", fmt.Errorf("invalid lead")
	}
	lead, err := orm.SearchOne(ctx, "crm.lead", map[string]interface{}{"id": id})
	if err != nil {
		return "", err
	}
	wid, err := orm.Create(ctx, orm.Registry["crm.opportunity.to.quotation"], map[string]interface{}{
		"lead_id": id,
		"partner_id": lead["partner_id"],
	})
	if err != nil {
		return "", err
	}
	return orm.ActionOpenURL("crm.opportunity.to.quotation", wid), nil
}

func actionApplyQuotation(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if model != "crm.opportunity.to.quotation" || id <= 0 {
		return "", fmt.Errorf("invalid wizard")
	}
	wiz, err := orm.SearchOne(ctx, "crm.opportunity.to.quotation", map[string]interface{}{"id": id})
	if err != nil {
		return "", err
	}
	lid, _ := orm.CoerceInt64(wiz["lead_id"])
	if lid <= 0 {
		return "", fmt.Errorf("missing opportunity")
	}
	line := map[string]interface{}{}
	if pid, ok := orm.CoerceInt64(wiz["product_id"]); ok && pid > 0 {
		line["product_id"] = pid
		line["product_uom_qty"] = numericFloat(wiz["quantity"])
	}
	orderID, err := CreateQuotationFromLead(ctx, int(lid), line)
	if err != nil {
		return "", err
	}
	if asBool(wiz["mark_won"]) {
		_, _ = orm.RunObjectAction(ctx, "crm.lead", int(lid), "action_set_won", nil)
	}
	return fmt.Sprintf("/web?action=sale.action_sale_quotations&view_type=form&id=%d", orderID), nil
}

func CreateQuotationFromLead(ctx context.Context, leadID int, line map[string]interface{}) (int, error) {
	if leadID <= 0 {
		return 0, fmt.Errorf("invalid lead id")
	}
	bypass := orm.ContextWithBypass(ctx, true)
	lead, err := orm.SearchOne(bypass, "crm.lead", map[string]interface{}{"id": leadID})
	if err != nil {
		return 0, err
	}
	existing, err := orm.Search(bypass, "sale.order", [][]interface{}{
		{"opportunity_id", "=", leadID},
		{"state", "=", "draft"},
	})
	if err != nil {
		return 0, err
	}
	if len(existing) > 0 {
		id, _ := orm.CoerceInt64(existing[0]["id"])
		return int(id), nil
	}
	partnerID, _ := orm.CoerceInt64(lead["partner_id"])
	userID, _ := orm.CoerceInt64(lead["user_id"])
	teamID, _ := orm.CoerceInt64(lead["team_id"])
	name := orm.AsString(lead["name"])
	if name == "" {
		name = fmt.Sprintf("Opportunity %d", leadID)
	}
	if partnerID <= 0 {
		return 0, fmt.Errorf("no customer")
	}
	orderModel, ok := orm.Registry["sale.order"]
	if !ok {
		return 0, fmt.Errorf("sale.order missing")
	}
	vals := map[string]interface{}{
		"name":           fmt.Sprintf("Q-CRM-%d", leadID),
		"partner_id":     partnerID,
		"user_id":        userID,
		"team_id":        teamID,
		"opportunity_id": leadID,
		"state":          "draft",
		"invoice_status": "no",
		"note":           fmt.Sprintf("Created from opportunity: %s", name),
	}
	orderID, err := orm.Create(bypass, orderModel, vals)
	if err != nil {
		return 0, err
	}
	if len(line) > 0 && orm.Registry["sale.order.line"] != nil {
		line["order_id"] = orderID
		_, _ = orm.Create(bypass, orm.Registry["sale.order.line"], line)
	}
	_ = refreshQuotationCount(bypass, int64(leadID))
	return orderID, nil
}

func refreshQuotationCount(ctx context.Context, leadID int64) error {
	rows, err := orm.Search(ctx, "sale.order", [][]interface{}{{"opportunity_id", "=", leadID}})
	if err != nil {
		return err
	}
	return orm.UpdateRecordByID(ctx, "crm.lead", int(leadID), map[string]interface{}{
		"quotation_count": int64(len(rows)),
	})
}

func onLeadWonCreateQuotation(ctx context.Context, ev event.Event) error {
	if !orm.ConfigParamBool(ctx, "sale_crm.auto_quotation", true) {
		return nil
	}
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
		return err
	}
	typ := orm.AsString(lead["type"])
	stageID, _ := orm.CoerceInt64(lead["stage_id"])
	won := orm.AsString(lead["won_status"]) == "won"
	if !won && stageID > 0 {
		if stage, err := orm.SearchOne(bypass, "crm.stage", map[string]interface{}{"id": stageID}); err == nil {
			won = asBool(stage["is_won"])
		}
	}
	if typ == "lead" && won {
		if err := orm.UpdateRecordByID(bypass, "crm.lead", id, map[string]interface{}{"type": "opportunity"}); err != nil {
			return err
		}
	}
	if !won {
		return nil
	}
	_, err = CreateQuotationFromLead(ctx, id, nil)
	return err
}

func coerceID(v interface{}) (int, bool) {
	n, ok := orm.CoerceInt64(v)
	return int(n), ok && n > 0
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
	case int64:
		return float64(t)
	default:
		return 1
	}
}
