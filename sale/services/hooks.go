package services

import (
	"context"

	"sumeru/core/event"
	"sumeru/core/orm"
)

func init() {
	event.Subscribe("record.updated", onSaleOrderUpdated)
}

func onSaleOrderUpdated(ctx context.Context, ev event.Event) error {
	model, _ := ev.Payload["model"].(string)
	if model != "sale.order" {
		return nil
	}
	id, ok := coerceID(ev.Payload["id"])
	if !ok || id <= 0 {
		return nil
	}
	bypass := orm.ContextWithBypass(ctx, true)
	row, err := orm.SearchOne(bypass, "sale.order", map[string]interface{}{"id": id})
	if err != nil {
		return err
	}
	state := orm.AsString(row["state"])
	inv := orm.AsString(row["invoice_status"])
	if state == "sale" && (inv == "" || inv == "no") {
		if err := orm.UpdateRecordByID(bypass, "sale.order", id, map[string]interface{}{
			"invoice_status": "to invoice",
		}); err != nil {
			return err
		}
	}
	if state == "cancel" && inv != "invoiced" && inv != "no" {
		if err := orm.UpdateRecordByID(bypass, "sale.order", id, map[string]interface{}{
			"invoice_status": "no",
		}); err != nil {
			return err
		}
	}
	return nil
}
