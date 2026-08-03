package sale_crm

import (
	"context"
	"fmt"
	"log"

	"sumeru/core/event"
	"sumeru/core/orm"
)

func init() {
	log.Println("Sale CRM Bridge Loaded")
	event.Subscribe("record.updated", onLeadWonCreateQuotation)
}

// CreateQuotationFromLead creates a draft sale.order linked to the opportunity.
func CreateQuotationFromLead(ctx context.Context, leadID int) (int, error) {
	if leadID <= 0 {
		return 0, fmt.Errorf("invalid lead id")
	}
	bypass := orm.ContextWithBypass(ctx, true)
	lead, err := orm.SearchOne(bypass, "crm.lead", map[string]interface{}{"id": leadID})
	if err != nil {
		return 0, err
	}
	existing, _ := orm.Search(bypass, "sale.order", [][]interface{}{
		{"opportunity_id", "=", leadID},
	})
	if len(existing) > 0 {
		id, _ := orm.CoerceInt64(existing[0]["id"])
		return int(id), nil
	}
	partnerID, _ := orm.CoerceInt64(lead["partner_id"])
	userID, _ := orm.CoerceInt64(lead["user_id"])
	name := orm.AsString(lead["name"])
	if name == "" {
		name = fmt.Sprintf("Opportunity %d", leadID)
	}
	if partnerID <= 0 {
		return 0, fmt.Errorf("opportunity %d has no customer", leadID)
	}
	orderModel, ok := orm.Registry["sale.order"]
	if !ok {
		return 0, fmt.Errorf("sale.order not registered")
	}
	vals := map[string]interface{}{
		"name":           fmt.Sprintf("Q-CRM-%d", leadID),
		"partner_id":     partnerID,
		"user_id":        userID,
		"opportunity_id": leadID,
		"state":          "draft",
		"invoice_status": "no",
		"note":           fmt.Sprintf("Created from opportunity: %s", name),
	}
	return orm.Create(bypass, orderModel, vals)
}

func onLeadWonCreateQuotation(ctx context.Context, ev event.Event) error {
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
	typ := orm.AsString(lead["type"])
	stageID, _ := orm.CoerceInt64(lead["stage_id"])
	won := false
	if stageID > 0 {
		if stage, err := orm.SearchOne(bypass, "crm.stage", map[string]interface{}{"id": stageID}); err == nil {
			won = asBool(stage["is_won"])
		}
	}
	if typ == "lead" && won {
		_ = orm.UpdateRecordByID(bypass, "crm.lead", id, map[string]interface{}{"type": "opportunity"})
	}
	if !won {
		return nil
	}
	if _, err := CreateQuotationFromLead(ctx, id); err != nil {
		log.Printf("sale_crm: quotation from lead %d: %v", id, err)
	}
	return nil
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
