package services

import (
	"context"

	"sumeru/core/event"
	"sumeru/core/orm"
)

func init() {
	event.Subscribe("record.updated", onPurchaseOrderUpdated)
}

func onPurchaseOrderUpdated(ctx context.Context, ev event.Event) error {
	model, _ := ev.Payload["model"].(string)
	if model != "purchase.order" {
		return nil
	}
	id, ok := coerceID(ev.Payload["id"])
	if !ok {
		return nil
	}
	bypass := orm.ContextWithBypass(ctx, true)
	po, err := orm.SearchOne(bypass, "purchase.order", map[string]interface{}{"id": id})
	if err != nil {
		return err
	}
	state := orm.AsString(po["state"])
	inv := orm.AsString(po["invoice_status"])
	if state == "purchase" && (inv == "" || inv == "no") {
		return orm.UpdateRecordByID(bypass, "purchase.order", id, map[string]interface{}{
			"invoice_status": "to invoice",
		})
	}
	if state == "purchase" && inv == "to invoice" {
		origin := orm.AsString(po["name"])
		moves, _ := orm.Search(bypass, "account.move", [][]interface{}{
			{"invoice_origin", "=", origin},
			{"move_type", "=", "in_invoice"},
		})
		if len(moves) > 0 {
			return orm.UpdateRecordByID(bypass, "purchase.order", id, map[string]interface{}{
				"invoice_status": "invoiced",
			})
		}
	}
	return nil
}
