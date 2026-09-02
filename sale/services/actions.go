package services

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"sumeru/core/event"
	"sumeru/core/orm"
)

func init() {
	orm.RegisterObjectAction("sale.order", "action_confirm", actionConfirmSale)
	orm.RegisterObjectAction("sale.order", "action_quotation_send", actionSendQuotation)
	orm.RegisterObjectAction("sale.order", "action_cancel", actionCancelSale)
	orm.RegisterObjectAction("sale.order", "action_draft", actionDraftSale)
	orm.RegisterObjectAction("sale.order", "action_view_invoices", actionViewInvoices)

	event.Subscribe("record.updated", onSaleLineUpdated)
	event.Subscribe("record.created", onSaleLineCreated)
}

func actionConfirmSale(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if model != "sale.order" || id <= 0 {
		return "", fmt.Errorf("invalid order")
	}
	bypass := orm.ContextWithBypass(ctx, true)
	order, err := orm.SearchOne(bypass, "sale.order", map[string]interface{}{"id": id})
	if err != nil {
		return "", err
	}
	if orm.AsString(order["state"]) == "cancel" {
		return "", fmt.Errorf("order cancelled")
	}
	name := orm.AsString(order["name"])
	if name == "" || name == "New" {
		name = nextDocName(bypass, "sale.order", "SO")
	}
	if err := recomputeOrderTotals(bypass, id); err != nil {
		return "", err
	}
	upd := map[string]interface{}{
		"state":          "sale",
		"name":           name,
		"invoice_status": "to invoice",
	}
	if orm.AsString(order["date_order"]) == "" {
		upd["date_order"] = time.Now().Format("2006-01-02 15:04:05")
	}
	if err := orm.UpdateRecordByID(bypass, "sale.order", id, upd); err != nil {
		return "", err
	}
	return vals["next"], nil
}

func actionSendQuotation(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if err := orm.UpdateRecordByID(ctx, "sale.order", id, map[string]interface{}{"state": "sent"}); err != nil {
		return "", err
	}
	return vals["next"], nil
}

func actionCancelSale(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	upd := map[string]interface{}{"state": "cancel", "invoice_status": "no"}
	if err := orm.UpdateRecordByID(ctx, "sale.order", id, upd); err != nil {
		return "", err
	}
	return vals["next"], nil
}

func actionDraftSale(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	upd := map[string]interface{}{"state": "draft", "invoice_status": "no"}
	if err := orm.UpdateRecordByID(ctx, "sale.order", id, upd); err != nil {
		return "", err
	}
	return vals["next"], nil
}

func actionViewInvoices(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	order, err := orm.SearchOne(ctx, "sale.order", map[string]interface{}{"id": id})
	if err != nil {
		return "", err
	}
	origin := orm.AsString(order["name"])
	moves, err := orm.Search(ctx, "account.move", [][]interface{}{
		{"invoice_origin", "=", origin},
		{"move_type", "in", []interface{}{"out_invoice", "out_refund"}},
	})
	if err != nil || len(moves) == 0 {
		return "", fmt.Errorf("no invoices")
	}
	mid, _ := orm.CoerceInt64(moves[0]["id"])
	return fmt.Sprintf("/web?action=account.action_account_moves_out&view_type=form&id=%d", mid), nil
}

func onSaleLineUpdated(ctx context.Context, ev event.Event) error {
	return onSaleLineChange(ctx, ev)
}

func onSaleLineCreated(ctx context.Context, ev event.Event) error {
	return onSaleLineChange(ctx, ev)
}

func onSaleLineChange(ctx context.Context, ev event.Event) error {
	model, _ := ev.Payload["model"].(string)
	if model != "sale.order.line" {
		return nil
	}
	id, ok := coerceID(ev.Payload["id"])
	if !ok {
		return nil
	}
	bypass := orm.ContextWithBypass(ctx, true)
	line, err := orm.SearchOne(bypass, "sale.order.line", map[string]interface{}{"id": id})
	if err != nil {
		return nil
	}
	orderID, _ := orm.CoerceInt64(line["order_id"])
	if orderID <= 0 {
		return nil
	}
	qty := numericFloat(line["product_uom_qty"])
	if qty <= 0 {
		qty = 1
	}
	price := numericFloat(line["price_unit"])
	sub := qty * price
	if err := orm.UpdateRecordByID(bypass, "sale.order.line", id, map[string]interface{}{
		"price_subtotal": round2(sub),
	}); err != nil {
		return err
	}
	return recomputeOrderTotals(bypass, int(orderID))
}

func recomputeOrderTotals(ctx context.Context, orderID int) error {
	lines, err := orm.Search(ctx, "sale.order.line", [][]interface{}{{"order_id", "=", orderID}})
	if err != nil {
		return err
	}
	untaxed := 0.0
	for _, ln := range lines {
		untaxed += numericFloat(ln["price_subtotal"])
	}
	return orm.UpdateRecordByID(ctx, "sale.order", orderID, map[string]interface{}{
		"amount_untaxed": round2(untaxed),
		"amount_total":   round2(untaxed),
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
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case int32:
		return float64(t)
	default:
		s := strings.TrimSpace(orm.AsString(v))
		if s == "" {
			return 0
		}
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0
		}
		return f
	}
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
