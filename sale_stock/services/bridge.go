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

	"sumeru/core/orm"
)

func init() {
	orm.RegisterObjectAction("sale.order", "action_confirm", actionConfirmSale)
	orm.RegisterObjectAction("sale.order", "action_create_delivery", actionCreateDelivery)
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

// actionCreateDelivery backs the manual "Delivery" button on the sale order.
func actionCreateDelivery(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if model != "sale.order" || id <= 0 {
		return "", fmt.Errorf("invalid order")
	}
	if _, err := CreateDeliveryForOrder(ctx, id); err != nil {
		return "", err
	}
	return vals["next"], nil
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
