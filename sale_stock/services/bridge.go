package services

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	// Ensure sale/services registers its action_confirm handler first, so the
	// handler registered below replaces it (Odoo: sale_stock overrides
	// sale.order._action_confirm).
	_ "sumeru_addons/sale"
	stocksvc "sumeru_addons/stock/services"

	"sumeru/core/event"
	"sumeru/core/orm"
)

func init() {
	orm.RegisterObjectAction("sale.order", "action_confirm", actionConfirmSale)
	orm.RegisterObjectAction("sale.order", "action_create_delivery", actionCreateDelivery)
	orm.RegisterObjectAction("sale.order", "action_view_deliveries", actionViewDeliveries)

	event.Subscribe("record.deleted", onPickingDeleted)
}

// onPickingDeleted reacts to a deleted transfer: it resets the linked sale
// order back to draft so the Confirm button reappears. Reservation cleanup
// happens earlier, in stock's unlink guard, while moves are still linked.
func onPickingDeleted(ctx context.Context, ev event.Event) error {
	model, _ := ev.Payload["model"].(string)
	if model != "stock.picking" {
		return nil
	}
	before, _ := ev.Payload["before"].(map[string]interface{})
	if before == nil {
		return nil
	}
	code := orm.AsString(before["picking_type_code"])
	if code != "outgoing" {
		return nil
	}
	origin := strings.TrimSpace(orm.AsString(before["origin"]))
	if origin == "" {
		return nil
	}
	bypass := orm.ContextWithBypass(ctx, true)
	orders, err := orm.Search(bypass, "sale.order", [][]interface{}{{"name", "=", origin}})
	if err != nil {
		return err
	}
	for _, r := range orders {
		id, _ := orm.CoerceInt64(r["id"])
		if id <= 0 || orm.AsString(r["state"]) != "sale" {
			continue
		}
		if err := orm.UpdateRecordByID(bypass, "sale.order", int(id), map[string]interface{}{
			"state":          "draft",
			"invoice_status": "no",
		}); err != nil {
			return err
		}
	}
	return nil
}

// actionConfirmSale mirrors sale/services actionConfirmSale and, once the order
// is confirmed, launches the stock delivery (Odoo: sale_stock's
// _action_confirm -> order_line._action_launch_stock_rule).
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
	state := orm.AsString(order["state"])
	if state != "draft" && state != "sent" {
		return "", fmt.Errorf("order already confirmed")
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
	if _, err := CreateDeliveryForOrder(ctx, id); err != nil {
		return "", err
	}
	return vals["next"], nil
}

// actionCreateDelivery backs the manual "Delivery" button on the sale order:
// it creates the outgoing transfer when missing (idempotent otherwise) and
// opens the delivery form so the click always has a visible result.
func actionCreateDelivery(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if model != "sale.order" || id <= 0 {
		return "", fmt.Errorf("invalid order")
	}
	order, err := orm.SearchOne(ctx, "sale.order", map[string]interface{}{"id": id})
	if err != nil {
		return "", err
	}
	pickID, err := CreateDeliveryForOrder(ctx, id)
	if err != nil {
		return "", err
	}
	if pickID <= 0 {
		// No storable lines to create a transfer — reuse an existing delivery.
		if pid, ok := stocksvc.TransferForOrigin(ctx, orm.AsString(order["name"]), "outgoing"); ok {
			pickID = pid
		}
	}
	if pickID <= 0 {
		return "", fmt.Errorf("no storable lines to deliver")
	}
	actionID, _, err := orm.ResolveXmlId(ctx, "stock.action_stock_picking_out")
	if err != nil || actionID <= 0 {
		return "", fmt.Errorf("delivery action not found")
	}
	return fmt.Sprintf("/web?action=%d&view_type=form&id=%d", actionID, pickID), nil
}

// actionViewDeliveries backs the "Deliveries" smart button: it opens the
// delivery order (outgoing picking) created for the sale order.
func actionViewDeliveries(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if model != "sale.order" || id <= 0 {
		return "", fmt.Errorf("invalid order")
	}
	order, err := orm.SearchOne(ctx, "sale.order", map[string]interface{}{"id": id})
	if err != nil {
		return "", err
	}
	picks, err := orm.Search(ctx, "stock.picking", [][]interface{}{
		{"origin", "=", orm.AsString(order["name"])},
		{"picking_type_code", "=", "outgoing"},
	})
	if err != nil || len(picks) == 0 {
		return "", fmt.Errorf("no deliveries for this order")
	}
	pickID, _ := orm.CoerceInt64(picks[0]["id"])
	actionID, _, err := orm.ResolveXmlId(ctx, "stock.action_stock_picking_out")
	if err != nil || actionID <= 0 {
		return "", fmt.Errorf("delivery action not found")
	}
	return fmt.Sprintf("/web?action=%d&view_type=form&id=%d", actionID, pickID), nil
}

// CreateDeliveryForOrder creates an outgoing transfer (delivery order) for the
// storable lines of a confirmed sale order. It is idempotent: the stock
// transfer is keyed on the order reference.
func CreateDeliveryForOrder(ctx context.Context, orderID int) (int, error) {
	if orderID <= 0 {
		return 0, fmt.Errorf("invalid order id")
	}
	bypass := orm.ContextWithBypass(ctx, true)
	order, err := orm.SearchOne(bypass, "sale.order", map[string]interface{}{"id": orderID})
	if err != nil {
		return 0, err
	}
	if orm.AsString(order["state"]) != "sale" {
		return 0, fmt.Errorf("confirm the order before creating a delivery")
	}
	lines, err := orm.Search(bypass, "sale.order.line", [][]interface{}{{"order_id", "=", orderID}})
	if err != nil {
		return 0, err
	}
	storable := storableProductIDs(bypass, lines)

	partnerID, _ := orm.CoerceInt64(order["partner_id"])
	companyID, _ := orm.CoerceInt64(order["company_id"])
	spec := stocksvc.TransferSpec{
		Origin:    orm.AsString(order["name"]),
		Code:      "outgoing",
		PartnerID: partnerID,
		CompanyID: companyID,
		Note:      "Delivery for " + orm.AsString(order["name"]),
	}
	for _, ln := range lines {
		pid, ok := orm.CoerceInt64(ln["product_id"])
		if !ok || pid <= 0 || !storable[pid] {
			continue
		}
		qty := numericFloat(ln["product_uom_qty"])
		if qty <= 0 {
			continue
		}
		spec.Lines = append(spec.Lines, stocksvc.TransferLine{
			ProductID: pid,
			Name:      orm.AsString(ln["name"]),
			Qty:       qty,
		})
	}
	if len(spec.Lines) == 0 {
		return 0, nil
	}
	return stocksvc.CreateTransfer(ctx, spec)
}

// storableProductIDs maps product ids whose type is "consu" (storable/consumable).
func storableProductIDs(ctx context.Context, lines []map[string]interface{}) map[int64]bool {
	out := map[int64]bool{}
	var ids []interface{}
	for _, ln := range lines {
		if pid, ok := orm.CoerceInt64(ln["product_id"]); ok && pid > 0 {
			ids = append(ids, pid)
		}
	}
	if len(ids) == 0 {
		return out
	}
	products, err := orm.Search(ctx, "product.product", [][]interface{}{{"id", "in", ids}})
	if err != nil {
		return out
	}
	for _, p := range products {
		if orm.AsString(p["type"]) != "consu" {
			continue
		}
		if pid, ok := orm.CoerceInt64(p["id"]); ok {
			out[pid] = true
		}
	}
	return out
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
