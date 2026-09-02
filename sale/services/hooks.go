package services

import (
	"context"
	"math"
	"strconv"
	"strings"

	"sumeru/core/event"
	"sumeru/core/orm"
)

func init() {
	event.Subscribe("record.updated", onSaleOrderUpdated)
	orm.RegisterOnchange("sale.order.line", "product_id", onSaleOrderLineProductChange)
	orm.RegisterOnchange("sale.order.line", "product_uom_qty", onSaleOrderLineSubtotalChange)
	orm.RegisterOnchange("sale.order.line", "price_unit", onSaleOrderLineSubtotalChange)
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
<<<<<<< HEAD

// onSaleOrderLineProductChange updates the line description from the product name.
func onSaleOrderLineProductChange(ctx context.Context, values map[string]interface{}, _ string) (orm.OnchangeResult, error) {
	result := orm.OnchangeResult{Value: map[string]interface{}{}}
	productID, ok := orm.CoerceInt64(values["product_id"])
	if !ok || productID <= 0 {
		return result, nil
	}
	product, err := orm.SearchOne(ctx, "product.product", map[string]interface{}{"id": int(productID)})
	if err != nil || product == nil {
		return result, nil
	}
	result.Value["name"] = orm.AsString(product["name"])
	return result, nil
}

// onSaleOrderLineSubtotalChange recomputes the untaxed line subtotal (qty × price).
func onSaleOrderLineSubtotalChange(_ context.Context, values map[string]interface{}, _ string) (orm.OnchangeResult, error) {
	result := orm.OnchangeResult{Value: map[string]interface{}{}}
	qty := numericFloat(values["product_uom_qty"])
	if qty <= 0 {
		qty = 1
	}
	price := numericFloat(values["price_unit"])
	result.Value["price_subtotal"] = round2(qty * price)
	return result, nil
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

func coerceID(v interface{}) (int, bool) {
	n, ok := orm.CoerceInt64(v)
	return int(n), ok && n > 0
}
=======
>>>>>>> 1a55542 (feat(sale): add order actions, sequences, and typed model fields)
