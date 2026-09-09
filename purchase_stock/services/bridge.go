package services

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	_ "sumeru_addons/purchase"
	stocksvc "sumeru_addons/stock/services"

	"sumeru/core/orm"
)

func init() {
	orm.RegisterObjectAction("purchase.order", "action_confirm", actionConfirmPurchase)
	orm.RegisterObjectAction("purchase.order", "action_create_receipt", actionCreateReceipt)
	orm.RegisterObjectAction("purchase.order", "action_cancel", actionCancelPurchase)
}

func actionConfirmPurchase(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if model != "purchase.order" || id <= 0 {
		return "", fmt.Errorf("invalid order")
	}
	bypass := orm.ContextWithBypass(ctx, true)
	order, err := orm.SearchOne(bypass, "purchase.order", map[string]interface{}{"id": id})
	if err != nil {
		return "", err
	}
	if orm.AsString(order["state"]) == "cancel" {
		return "", fmt.Errorf("order cancelled")
	}

	name := orm.AsString(order["name"])
	if name == "" || name == "New" {
		name = nextDocName(bypass, "purchase.order", "PO")
	}
	if err := recomputePurchaseTotal(bypass, id); err != nil {
		return "", err
	}
	updates := map[string]interface{}{
		"state":          "purchase",
		"name":           name,
		"invoice_status": "to invoice",
	}
	if orm.AsString(order["date_order"]) == "" {
		updates["date_order"] = time.Now().Format("2006-01-02 15:04:05")
	}
	if err := orm.UpdateRecordByID(bypass, "purchase.order", id, updates); err != nil {
		return "", err
	}
	if _, err := CreateReceiptForOrder(ctx, id); err != nil {
		return "", err
	}
	return vals["next"], nil
}

func actionCreateReceipt(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if model != "purchase.order" || id <= 0 {
		return "", fmt.Errorf("invalid order")
	}
	if _, err := CreateReceiptForOrder(ctx, id); err != nil {
		return "", err
	}
	return vals["next"], nil
}

func actionCancelPurchase(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if model != "purchase.order" || id <= 0 {
		return "", fmt.Errorf("invalid order")
	}
	bypass := orm.ContextWithBypass(ctx, true)
	order, err := orm.SearchOne(bypass, "purchase.order", map[string]interface{}{"id": id})
	if err != nil {
		return "", err
	}
	if err := cancelReceiptForOrigin(ctx, orm.AsString(order["name"])); err != nil {
		return "", err
	}
	if err := orm.UpdateRecordByID(ctx, "purchase.order", id, map[string]interface{}{
		"state":          "cancel",
		"invoice_status": "no",
	}); err != nil {
		return "", err
	}
	return vals["next"], nil
}

// CreateReceiptForOrder creates the incoming transfer for stockable PO lines.
// The transfer origin makes repeated confirmation or manual creation idempotent.
func CreateReceiptForOrder(ctx context.Context, orderID int) (int, error) {
	if orderID <= 0 {
		return 0, fmt.Errorf("invalid order id")
	}
	bypass := orm.ContextWithBypass(ctx, true)
	order, err := orm.SearchOne(bypass, "purchase.order", map[string]interface{}{"id": orderID})
	if err != nil {
		return 0, err
	}
	if orm.AsString(order["state"]) != "purchase" {
		return 0, fmt.Errorf("confirm the order before creating a receipt")
	}

	lines, err := orm.Search(bypass, "purchase.order.line", [][]interface{}{{"order_id", "=", orderID}})
	if err != nil {
		return 0, err
	}
	stockable := stockableProductIDs(bypass, lines)
	partnerID, _ := orm.CoerceInt64(order["partner_id"])
	companyID, _ := orm.CoerceInt64(order["company_id"])
	spec := stocksvc.TransferSpec{
		Origin:    orm.AsString(order["name"]),
		PartnerID: partnerID,
		CompanyID: companyID,
		Code:      "incoming",
		Note:      "Receipt for " + orm.AsString(order["name"]),
	}
	for _, line := range lines {
		productID, ok := orm.CoerceInt64(line["product_id"])
		if !ok || productID <= 0 || !stockable[productID] {
			continue
		}
		qty := numericFloat(line["product_qty"])
		if qty <= 0 {
			continue
		}
		spec.Lines = append(spec.Lines, stocksvc.TransferLine{
			ProductID: productID,
			Name:      orm.AsString(line["name"]),
			Qty:       qty,
		})
	}
	if len(spec.Lines) == 0 {
		return 0, nil
	}
	return stocksvc.CreateTransfer(ctx, spec)
}

func cancelReceiptForOrigin(ctx context.Context, origin string) error {
	if strings.TrimSpace(origin) == "" {
		return nil
	}
	rows, err := orm.Search(orm.ContextWithBypass(ctx, true), "stock.picking", [][]interface{}{
		{"origin", "=", origin},
		{"picking_type_code", "=", "incoming"},
	})
	if err != nil {
		return err
	}
	for _, row := range rows {
		if orm.AsString(row["state"]) == "done" || orm.AsString(row["state"]) == "cancel" {
			continue
		}
		id, ok := orm.CoerceInt64(row["id"])
		if ok && id > 0 {
			if err := stocksvc.CancelPicking(ctx, int(id)); err != nil {
				return err
			}
		}
	}
	return nil
}

func stockableProductIDs(ctx context.Context, lines []map[string]interface{}) map[int64]bool {
	ids := make([]interface{}, 0, len(lines))
	for _, line := range lines {
		if id, ok := orm.CoerceInt64(line["product_id"]); ok && id > 0 {
			ids = append(ids, id)
		}
	}
	result := map[int64]bool{}
	if len(ids) == 0 {
		return result
	}
	products, err := orm.Search(ctx, "product.product", [][]interface{}{{"id", "in", ids}})
	if err != nil {
		return result
	}
	for _, product := range products {
		if orm.AsString(product["type"]) != "consu" {
			continue
		}
		if id, ok := orm.CoerceInt64(product["id"]); ok {
			result[id] = true
		}
	}
	return result
}

func recomputePurchaseTotal(ctx context.Context, orderID int) error {
	lines, err := orm.Search(ctx, "purchase.order.line", [][]interface{}{{"order_id", "=", orderID}})
	if err != nil {
		return err
	}
	total := 0.0
	for _, line := range lines {
		total += numericFloat(line["price_subtotal"])
	}
	return orm.UpdateRecordByID(ctx, "purchase.order", orderID, map[string]interface{}{"amount_total": round2(total)})
}

func nextDocName(ctx context.Context, code, fallbackPrefix string) string {
	if name, err := orm.NextSequence(ctx, code); err == nil && name != "" {
		return name
	}
	return fmt.Sprintf("%s/%05d", fallbackPrefix, time.Now().Unix()%100000)
}

func numericFloat(value interface{}) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int64:
		return float64(typed)
	case int:
		return float64(typed)
	default:
		return 0
	}
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}
