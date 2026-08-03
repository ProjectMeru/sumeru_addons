package sale

import (
	"context"
	"log"

	"sumeru/core/event"
	"sumeru/core/orm"

	_ "sumeru_addons/sale/models"
)

func init() {
	log.Println("Sale Addon Loaded")
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
		return nil
	}
	state := orm.AsString(row["state"])
	inv := orm.AsString(row["invoice_status"])
	if state == "sale" && (inv == "" || inv == "no") {
		_ = orm.UpdateRecordByID(bypass, "sale.order", id, map[string]interface{}{
			"invoice_status": "to invoice",
		})
	}
	if state == "cancel" && inv != "invoiced" && inv != "no" {
		_ = orm.UpdateRecordByID(bypass, "sale.order", id, map[string]interface{}{
			"invoice_status": "no",
		})
	}
	return nil
}

func coerceID(v interface{}) (int, bool) {
	n, ok := orm.CoerceInt64(v)
	return int(n), ok && n > 0
}
