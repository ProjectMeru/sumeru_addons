package services

import (
	"context"
	"fmt"
	"math"
	"time"

	"sumeru/core/event"
	"sumeru/core/orm"
)

func init() {
	orm.RegisterObjectAction("purchase.order", "action_confirm", actionConfirmPurchase)
	orm.RegisterObjectAction("purchase.order", "action_rfq_send", actionSendRFQ)
	orm.RegisterObjectAction("purchase.order", "action_cancel", actionCancelPurchase)
	orm.RegisterObjectAction("purchase.order", "action_draft", actionDraftPurchase)
	orm.RegisterObjectAction("purchase.order", "action_view_bills", actionViewBills)

	event.Subscribe("record.updated", onPurchaseOrderUpdated)
	event.Subscribe("record.updated", onPurchaseLineUpdated)
	event.Subscribe("record.created", onPurchaseLineUpdated)
}

func actionConfirmPurchase(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if model != "purchase.order" || id <= 0 {
		return "", fmt.Errorf("invalid order")
	}
	bypass := orm.ContextWithBypass(ctx, true)
	po, err := orm.SearchOne(bypass, "purchase.order", map[string]interface{}{"id": id})
	if err != nil {
		return "", err
	}
	name := orm.AsString(po["name"])
	if name == "" || name == "New" {
		name = nextDocName(bypass, "purchase.order", "PO")
	}
	if err := recomputePOTotals(bypass, id); err != nil {
		return "", err
	}
	upd := map[string]interface{}{
		"state":          "purchase",
		"name":           name,
		"invoice_status": "to invoice",
	}
	if orm.AsString(po["date_order"]) == "" {
		upd["date_order"] = time.Now().Format("2006-01-02 15:04:05")
	}
	if err := orm.UpdateRecordByID(bypass, "purchase.order", id, upd); err != nil {
		return "", err
	}
	return vals["next"], nil
}

func actionSendRFQ(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if err := orm.UpdateRecordByID(ctx, "purchase.order", id, map[string]interface{}{"state": "sent"}); err != nil {
		return "", err
	}
	return vals["next"], nil
}

func actionCancelPurchase(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	upd := map[string]interface{}{"state": "cancel", "invoice_status": "no"}
	if err := orm.UpdateRecordByID(ctx, "purchase.order", id, upd); err != nil {
		return "", err
	}
	return vals["next"], nil
}

func actionDraftPurchase(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	upd := map[string]interface{}{"state": "draft", "invoice_status": "no"}
	if err := orm.UpdateRecordByID(ctx, "purchase.order", id, upd); err != nil {
		return "", err
	}
	return vals["next"], nil
}

func actionViewBills(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	po, err := orm.SearchOne(ctx, "purchase.order", map[string]interface{}{"id": id})
	if err != nil {
		return "", err
	}
	origin := orm.AsString(po["name"])
	moves, err := orm.Search(ctx, "account.move", [][]interface{}{
		{"invoice_origin", "=", origin},
		{"move_type", "in", []interface{}{"in_invoice", "in_refund"}},
	})
	if err != nil || len(moves) == 0 {
		return "", fmt.Errorf("no bills")
	}
	mid, _ := orm.CoerceInt64(moves[0]["id"])
	return fmt.Sprintf("/web?action=account.action_account_moves_in&view_type=form&id=%d", mid), nil
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

func onPurchaseLineUpdated(ctx context.Context, ev event.Event) error {
	model, _ := ev.Payload["model"].(string)
	if model != "purchase.order.line" {
		return nil
	}
	id, ok := coerceID(ev.Payload["id"])
	if !ok {
		return nil
	}
	bypass := orm.ContextWithBypass(ctx, true)
	line, err := orm.SearchOne(bypass, "purchase.order.line", map[string]interface{}{"id": id})
	if err != nil {
		return nil
	}
	orderID, _ := orm.CoerceInt64(line["order_id"])
	if orderID <= 0 {
		return nil
	}
	qty := numericFloat(line["product_qty"])
	if qty <= 0 {
		qty = 1
	}
	price := numericFloat(line["price_unit"])
	sub := qty * price
	if err := orm.UpdateRecordByID(bypass, "purchase.order.line", id, map[string]interface{}{
		"price_subtotal": round2(sub),
	}); err != nil {
		return err
	}
	return recomputePOTotals(bypass, int(orderID))
}

func recomputePOTotals(ctx context.Context, orderID int) error {
	lines, err := orm.Search(ctx, "purchase.order.line", [][]interface{}{{"order_id", "=", orderID}})
	if err != nil {
		return err
	}
	total := 0.0
	for _, ln := range lines {
		total += numericFloat(ln["price_subtotal"])
	}
	return orm.UpdateRecordByID(ctx, "purchase.order", orderID, map[string]interface{}{
		"amount_total": round2(total),
	})
}

func nextDocName(ctx context.Context, code, fallbackPrefix string) string {
	if name, err := orm.NextSequence(ctx, code); err == nil && name != "" {
		return name
	}
	return fmt.Sprintf("%s/%05d", fallbackPrefix, time.Now().Unix()%100000)
}

func coerceID(v interface{}) (int, bool) {
	n, ok := orm.CoerceInt64(v)
	return int(n), ok && n > 0
}

func numericFloat(v interface{}) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case int64:
		return float64(t)
	case int:
		return float64(t)
	default:
		return 0
	}
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
