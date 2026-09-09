package services

import (
	"context"
	"fmt"
	"strings"

	"sumeru/core/orm"
)

// TransferLine is one product line requested on a transfer.
type TransferLine struct {
	ProductID int64
	Name      string
	Qty       float64
}

// TransferSpec describes a transfer to create from another document
// (sales order, purchase order, reordering rule, …).
type TransferSpec struct {
	Origin        string
	PartnerID     int64
	CompanyID     int64
	Code          string // incoming | outgoing | internal
	WarehouseID   int64
	ScheduledDate string
	Note          string
	Lines         []TransferLine
}

// PickingTypeForCode resolves the operation type to use for a code, preferring
// the warehouse configuration when one is given.
func PickingTypeForCode(ctx context.Context, code string, warehouseID int64) int64 {
	code = strings.TrimSpace(code)
	if code == "" {
		code = "internal"
	}
	if warehouseID > 0 {
		if wh, err := orm.SearchOne(ctx, "stock.warehouse", map[string]interface{}{"id": warehouseID}); err == nil && wh != nil {
			field := "int_type_id"
			switch code {
			case "incoming":
				field = "in_type_id"
			case "outgoing":
				field = "out_type_id"
			}
			if id, ok := orm.CoerceInt64(wh[field]); ok && id > 0 {
				return id
			}
		}
	}
	rows, err := orm.Search(ctx, "stock.picking.type", [][]interface{}{{"code", "=", code}})
	if err != nil || len(rows) == 0 {
		return 0
	}
	id, _ := orm.CoerceInt64(rows[0]["id"])
	return id
}

// pickingTypeLocations returns the default source/destination of an operation type.
func pickingTypeLocations(ctx context.Context, typeID int64) (srcID, destID int64) {
	if typeID <= 0 {
		return 0, 0
	}
	pt, err := orm.SearchOne(ctx, "stock.picking.type", map[string]interface{}{"id": typeID})
	if err != nil || pt == nil {
		return 0, 0
	}
	srcID, _ = orm.CoerceInt64(pt["default_location_src_id"])
	destID, _ = orm.CoerceInt64(pt["default_location_dest_id"])
	return srcID, destID
}

// fallbackLocationByUsage finds any location of the requested usage.
func fallbackLocationByUsage(ctx context.Context, usage string) int64 {
	rows, err := orm.Search(ctx, "stock.location", [][]interface{}{{"usage", "=", usage}})
	if err != nil || len(rows) == 0 {
		return 0
	}
	id, _ := orm.CoerceInt64(rows[0]["id"])
	return id
}

// TransferForOrigin returns an existing non-cancelled transfer created for origin.
func TransferForOrigin(ctx context.Context, origin, code string) (int, bool) {
	origin = strings.TrimSpace(origin)
	if origin == "" {
		return 0, false
	}
	rows, err := orm.Search(ctx, "stock.picking", [][]interface{}{
		{"origin", "=", origin},
		{"picking_type_code", "=", code},
	})
	if err != nil {
		return 0, false
	}
	for _, r := range rows {
		if orm.AsString(r["state"]) == "cancel" {
			continue
		}
		if id, ok := coerceID(r["id"]); ok {
			return id, true
		}
	}
	return 0, false
}

// CreateTransfer creates a confirmed transfer with one move per line.
// It is the entry point used by the sale_stock / purchase_stock bridges.
func CreateTransfer(ctx context.Context, spec TransferSpec) (int, error) {
	if len(spec.Lines) == 0 {
		return 0, fmt.Errorf("no lines to transfer")
	}
	bypass := orm.ContextWithBypass(ctx, true)
	code := strings.TrimSpace(spec.Code)
	if code == "" {
		code = "internal"
	}
	if id, ok := TransferForOrigin(bypass, spec.Origin, code); ok {
		return id, nil
	}
	typeID := PickingTypeForCode(bypass, code, spec.WarehouseID)
	srcID, destID := pickingTypeLocations(bypass, typeID)
	if srcID <= 0 {
		switch code {
		case "incoming":
			srcID = fallbackLocationByUsage(bypass, "supplier")
		default:
			srcID = fallbackLocationByUsage(bypass, "internal")
		}
	}
	if destID <= 0 {
		switch code {
		case "outgoing":
			destID = fallbackLocationByUsage(bypass, "customer")
		default:
			destID = fallbackLocationByUsage(bypass, "internal")
		}
	}
	pickModel, err := modelOrErr("stock.picking")
	if err != nil {
		return 0, err
	}
	pickVals := map[string]interface{}{
		"origin":            spec.Origin,
		"state":             "draft",
		"location_id":       srcID,
		"location_dest_id":  destID,
		"picking_type_id":   typeID,
		"picking_type_code": code,
		"move_type":         "direct",
		"note":              spec.Note,
	}
	if spec.PartnerID > 0 {
		pickVals["partner_id"] = spec.PartnerID
	}
	if spec.CompanyID > 0 {
		pickVals["company_id"] = spec.CompanyID
	}
	if spec.ScheduledDate != "" {
		pickVals["scheduled_date"] = spec.ScheduledDate
	}
	pickID, err := orm.Create(bypass, pickModel, pickVals)
	if err != nil {
		return 0, err
	}
	moveModel, err := modelOrErr("stock.move")
	if err != nil {
		return 0, err
	}
	for _, ln := range spec.Lines {
		if ln.ProductID <= 0 || ln.Qty <= 0 {
			continue
		}
		name := strings.TrimSpace(ln.Name)
		if name == "" {
			name = productName(bypass, ln.ProductID)
		}
		if name == "" {
			name = "Product"
		}
		moveVals := map[string]interface{}{
			"name":             name,
			"product_id":       ln.ProductID,
			"product_qty":      ln.Qty,
			"location_id":      srcID,
			"location_dest_id": destID,
			"picking_id":       pickID,
			"picking_type_id":  typeID,
			"state":            "draft",
			"origin":           spec.Origin,
		}
		if spec.PartnerID > 0 {
			moveVals["partner_id"] = spec.PartnerID
		}
		if spec.CompanyID > 0 {
			moveVals["company_id"] = spec.CompanyID
		}
		if spec.ScheduledDate != "" {
			moveVals["date"] = spec.ScheduledDate
		}
		if _, err := orm.Create(bypass, moveModel, moveVals); err != nil {
			return 0, err
		}
	}
	if err := ConfirmPicking(ctx, pickID); err != nil {
		return 0, err
	}
	return pickID, nil
}
