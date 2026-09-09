package services

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"sumeru/core/orm"
)

const qtyEpsilon = 0.0001

// inventoryLossLocationID resolves the virtual counterpart location used by
// inventory adjustments (Odoo: "Inventory adjustment" / Inventory Loss).
func inventoryLossLocationID(ctx context.Context) int64 {
	if id, _, err := orm.ResolveXmlId(ctx, "stock.stock_location_inventory"); err == nil && id > 0 {
		return int64(id)
	}
	rows, err := orm.Search(ctx, "stock.location", [][]interface{}{{"usage", "=", "inventory"}})
	if err != nil || len(rows) == 0 {
		return 0
	}
	for _, r := range rows {
		if strings.Contains(strings.ToLower(orm.AsString(r["name"])), "inventory") {
			id, _ := orm.CoerceInt64(r["id"])
			return id
		}
	}
	id, _ := orm.CoerceInt64(rows[0]["id"])
	return id
}

// SetInventoryCount stores a counted quantity on a quant and refreshes the difference.
func SetInventoryCount(ctx context.Context, quantID int, counted float64) error {
	bypass := orm.ContextWithBypass(ctx, true)
	quant, err := orm.SearchOne(bypass, "stock.quant", map[string]interface{}{"id": quantID})
	if err != nil || quant == nil {
		return fmt.Errorf("quant not found")
	}
	return orm.UpdateRecordByID(bypass, "stock.quant", quantID, map[string]interface{}{
		"inventory_quantity":      counted,
		"inventory_diff_quantity": counted - numericFloat(quant["quantity"]),
		"inventory_quantity_set":  true,
	})
}

// ClearInventoryCount discards a pending count on a quant.
func ClearInventoryCount(ctx context.Context, quantID int) error {
	bypass := orm.ContextWithBypass(ctx, true)
	return orm.UpdateRecordByID(bypass, "stock.quant", quantID, map[string]interface{}{
		"inventory_quantity":      0,
		"inventory_diff_quantity": 0,
		"inventory_quantity_set":  false,
	})
}

// ApplyInventoryQuant applies the counted quantity of one quant: it books an
// inventory stock move against the Inventory Loss location, updates the quant
// on-hand quantity and clears the pending count.
func ApplyInventoryQuant(ctx context.Context, quantID int) error {
	bypass := orm.ContextWithBypass(ctx, true)
	quant, err := orm.SearchOne(bypass, "stock.quant", map[string]interface{}{"id": quantID})
	if err != nil || quant == nil {
		return fmt.Errorf("quant not found")
	}
	if !asBool(quant["inventory_quantity_set"]) {
		return fmt.Errorf("no counted quantity to apply")
	}
	onHand := numericFloat(quant["quantity"])
	counted := numericFloat(quant["inventory_quantity"])
	diff := counted - onHand
	productID, _ := orm.CoerceInt64(quant["product_id"])
	locationID, _ := orm.CoerceInt64(quant["location_id"])
	lotID, _ := orm.CoerceInt64(quant["lot_id"])
	companyID, _ := orm.CoerceInt64(quant["company_id"])
	reason := strings.TrimSpace(orm.AsString(quant["inventory_name"]))
	if reason == "" {
		reason = "Product Quantity Updated"
	}
	today := time.Now().Format("2006-01-02")

	if math.Abs(diff) > qtyEpsilon {
		lossID := inventoryLossLocationID(bypass)
		if lossID <= 0 {
			return fmt.Errorf("no inventory adjustment location configured")
		}
		srcID, destID := lossID, locationID
		if diff < 0 {
			srcID, destID = locationID, lossID
		}
		if err := createInventoryMove(bypass, inventoryMoveSpec{
			Name:       reason,
			ProductID:  productID,
			CompanyID:  companyID,
			LotID:      lotID,
			Qty:        math.Abs(diff),
			SourceID:   srcID,
			DestID:     destID,
			LocationID: locationID,
		}); err != nil {
			return err
		}
	}

	reserved := numericFloat(quant["reserved_quantity"])
	available := counted - reserved
	if available < 0 {
		available = 0
	}
	if err := orm.UpdateRecordByID(bypass, "stock.quant", quantID, map[string]interface{}{
		"quantity":                counted,
		"available_quantity":      available,
		"inventory_quantity":      0,
		"inventory_diff_quantity": 0,
		"inventory_quantity_set":  false,
		"last_count_date":         today,
	}); err != nil {
		return err
	}
	return recomputeLotQty(bypass, lotID)
}

// ApplyAllInventoryCounts applies every quant that has a pending count.
func ApplyAllInventoryCounts(ctx context.Context) (int, error) {
	bypass := orm.ContextWithBypass(ctx, true)
	rows, err := orm.Search(bypass, "stock.quant", [][]interface{}{
		{"inventory_quantity_set", "=", true},
	})
	if err != nil {
		return 0, err
	}
	applied := 0
	for _, r := range rows {
		id, ok := coerceID(r["id"])
		if !ok {
			continue
		}
		if err := ApplyInventoryQuant(ctx, id); err != nil {
			return applied, fmt.Errorf("quant %d: %w", id, err)
		}
		applied++
	}
	if applied == 0 {
		return 0, fmt.Errorf("no counted quantity to apply")
	}
	return applied, nil
}

type inventoryMoveSpec struct {
	Name       string
	ProductID  int64
	CompanyID  int64
	LotID      int64
	Qty        float64
	SourceID   int64
	DestID     int64
	LocationID int64
}

// createInventoryMove books a done stock.move (+ move line) flagged as inventory.
func createInventoryMove(ctx context.Context, spec inventoryMoveSpec) error {
	moveModel, err := modelOrErr("stock.move")
	if err != nil {
		return err
	}
	now := time.Now().Format(dtFormat)
	moveVals := map[string]interface{}{
		"name":             spec.Name,
		"product_id":       spec.ProductID,
		"product_qty":      spec.Qty,
		"quantity":         spec.Qty,
		"location_id":      spec.SourceID,
		"location_dest_id": spec.DestID,
		"state":            "done",
		"is_inventory":     true,
		"date":             now,
		"reference":        spec.Name,
		"origin":           "Inventory Adjustment",
	}
	if spec.CompanyID > 0 {
		moveVals["company_id"] = spec.CompanyID
	}
	moveID, err := orm.Create(ctx, moveModel, moveVals)
	if err != nil {
		return err
	}
	lineModel, err := modelOrErr("stock.move.line")
	if err != nil {
		return err
	}
	lineVals := map[string]interface{}{
		"move_id":          moveID,
		"product_id":       spec.ProductID,
		"quantity":         spec.Qty,
		"location_id":      spec.SourceID,
		"location_dest_id": spec.DestID,
		"state":            "done",
		"date":             now,
		"reference":        spec.Name,
	}
	if spec.CompanyID > 0 {
		lineVals["company_id"] = spec.CompanyID
	}
	if spec.LotID > 0 {
		lineVals["lot_id"] = spec.LotID
	}
	_, err = orm.Create(ctx, lineModel, lineVals)
	return err
}

func asBool(v interface{}) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		s := strings.ToLower(strings.TrimSpace(t))
		return s == "1" || s == "true" || s == "t" || s == "yes"
	case int, int32, int64, float32, float64:
		return numericFloat(v) != 0
	default:
		return false
	}
}
