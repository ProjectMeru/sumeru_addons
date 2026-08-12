package services

import (
	"context"
	"log"

	"sumeru/core/event"
	"sumeru/core/orm"
)

func init() {
	event.Subscribe("record.updated", onSaleOrderToInvoice)
	event.Subscribe("record.updated", onPurchaseOrderToBill)
	event.Subscribe("record.updated", onMovePost)
	event.Subscribe("record.updated", onPaymentPost)
}

func onSaleOrderToInvoice(ctx context.Context, ev event.Event) error {
	model, _ := ev.Payload["model"].(string)
	if model != "sale.order" {
		return nil
	}
	id, ok := coerceID(ev.Payload["id"])
	if !ok {
		return nil
	}
	if _, ok := orm.Registry["sale.order"]; !ok {
		return nil
	}
	bypass := orm.ContextWithBypass(ctx, true)
	order, err := orm.SearchOne(bypass, "sale.order", map[string]interface{}{"id": id})
	if err != nil {
		return nil
	}
	if orm.AsString(order["state"]) != "sale" {
		return nil
	}
	if orm.AsString(order["invoice_status"]) == "invoiced" {
		return nil
	}
	if _, err := CreateCustomerInvoiceFromSale(ctx, id); err != nil {
		log.Printf("account: invoice from SO %d: %v", id, err)
	}
	return nil
}

func onPurchaseOrderToBill(ctx context.Context, ev event.Event) error {
	model, _ := ev.Payload["model"].(string)
	if model != "purchase.order" {
		return nil
	}
	id, ok := coerceID(ev.Payload["id"])
	if !ok {
		return nil
	}
	if _, ok := orm.Registry["purchase.order"]; !ok {
		return nil
	}
	bypass := orm.ContextWithBypass(ctx, true)
	po, err := orm.SearchOne(bypass, "purchase.order", map[string]interface{}{"id": id})
	if err != nil {
		return nil
	}
	if orm.AsString(po["state"]) != "purchase" {
		return nil
	}
	if _, err := CreateVendorBillFromPurchase(ctx, id); err != nil {
		log.Printf("account: bill from PO %d: %v", id, err)
	}
	return nil
}

func onMovePost(ctx context.Context, ev event.Event) error {
	model, _ := ev.Payload["model"].(string)
	if model != "account.move" {
		return nil
	}
	id, ok := coerceID(ev.Payload["id"])
	if !ok {
		return nil
	}
	bypass := orm.ContextWithBypass(ctx, true)
	move, err := orm.SearchOne(bypass, "account.move", map[string]interface{}{"id": id})
	if err != nil {
		return nil
	}
	if orm.AsString(move["state"]) != "posted" {
		return nil
	}
	all, _ := orm.Search(bypass, "account.move.line", [][]interface{}{{"move_id", "=", id}})
	for _, ln := range all {
		dt := orm.AsString(ln["display_type"])
		if dt == "entry" || dt == "tax" {
			return nil
		}
	}
	moveType := orm.AsString(move["move_type"])
	if moveType == "entry" {
		return nil
	}
	if err := PostMove(ctx, id); err != nil {
		log.Printf("account: post move %d: %v", id, err)
		_ = orm.UpdateRecordByID(bypass, "account.move", id, map[string]interface{}{"state": "draft"})
	}
	return nil
}

func onPaymentPost(ctx context.Context, ev event.Event) error {
	model, _ := ev.Payload["model"].(string)
	if model != "account.payment" {
		return nil
	}
	id, ok := coerceID(ev.Payload["id"])
	if !ok {
		return nil
	}
	bypass := orm.ContextWithBypass(ctx, true)
	pay, err := orm.SearchOne(bypass, "account.payment", map[string]interface{}{"id": id})
	if err != nil {
		return nil
	}
	if orm.AsString(pay["state"]) != "posted" {
		return nil
	}
	if mid, ok := orm.CoerceInt64(pay["move_id"]); ok && mid > 0 {
		return nil
	}
	_ = orm.UpdateRecordByID(bypass, "account.payment", id, map[string]interface{}{"state": "draft"})
	if err := PostPayment(ctx, id); err != nil {
		log.Printf("account: post payment %d: %v", id, err)
		_ = orm.UpdateRecordByID(bypass, "account.payment", id, map[string]interface{}{"state": "draft"})
	}
	return nil
}

func coerceID(v interface{}) (int, bool) {
	n, ok := orm.CoerceInt64(v)
	return int(n), ok && n > 0
}
