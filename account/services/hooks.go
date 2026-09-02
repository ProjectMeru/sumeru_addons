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
	event.Subscribe("record.updated", onAccountMoveLineUpdated)
	orm.RegisterOnchange("account.move.line", "quantity", onAccountMoveLineSubtotalChange)
	orm.RegisterOnchange("account.move.line", "price_unit", onAccountMoveLineSubtotalChange)
}

// onAccountMoveLineSubtotalChange recomputes the untaxed line subtotal (qty × price)
// for product lines only. Section/note/tax/entry lines are left untouched.
func onAccountMoveLineSubtotalChange(_ context.Context, values map[string]interface{}, _ string) (orm.OnchangeResult, error) {
	result := orm.OnchangeResult{Value: map[string]interface{}{}}
	dt := orm.AsString(values["display_type"])
	if dt != "" && dt != "product" {
		return result, nil
	}
	qty := numericFloat(values["quantity"])
	if qty <= 0 {
		qty = 1
	}
	price := numericFloat(values["price_unit"])
	result.Value["price_subtotal"] = round2(qty * price)
	return result, nil
}

// onAccountMoveLineUpdated recomputes the parent move totals whenever a line
// changes, so amount_untaxed/amount_tax/amount_total stay in sync (Odoo-style).
func onAccountMoveLineUpdated(ctx context.Context, ev event.Event) error {
	model, _ := ev.Payload["model"].(string)
	if model != "account.move.line" {
		return nil
	}
	id, ok := coerceID(ev.Payload["id"])
	if !ok {
		return nil
	}
	bypass := orm.ContextWithBypass(ctx, true)
	line, err := orm.SearchOne(bypass, "account.move.line", map[string]interface{}{"id": id})
	if err != nil || line == nil {
		return nil
	}
	moveID, ok := orm.CoerceInt64(line["move_id"])
	if !ok || moveID <= 0 {
		return nil
	}
	return recomputeMoveAmounts(bypass, int(moveID))
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
		if err2 := orm.UpdateRecordByID(bypass, "account.move", id, map[string]interface{}{"state": "draft"}); err2 != nil {
			log.Printf("account: rollback move %d: %v", id, err2)
		}
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
	if err := orm.UpdateRecordByID(bypass, "account.payment", id, map[string]interface{}{"state": "draft"}); err != nil {
		log.Printf("account: draft payment %d: %v", id, err)
		return nil
	}
	if err := PostPayment(ctx, id); err != nil {
		log.Printf("account: post payment %d: %v", id, err)
		if err2 := orm.UpdateRecordByID(bypass, "account.payment", id, map[string]interface{}{"state": "draft"}); err2 != nil {
			log.Printf("account: rollback payment %d: %v", id, err2)
		}
	}
	return nil
}

func coerceID(v interface{}) (int, bool) {
	n, ok := orm.CoerceInt64(v)
	return int(n), ok && n > 0
}
