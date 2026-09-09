package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"sumeru/core/orm"
)

const dtFormat = "2006-01-02 15:04:05"

func modelOrErr(name string) (orm.Model, error) {
	m, ok := orm.Registry[name]
	if !ok {
		return nil, fmt.Errorf("model %s missing", name)
	}
	return m, nil
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
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return 0
		}
		var f float64
		if _, err := fmt.Sscanf(s, "%g", &f); err != nil {
			return 0
		}
		return f
	default:
		return 0
	}
}

func coerceID(v interface{}) (int, bool) {
	n, ok := orm.CoerceInt64(v)
	return int(n), ok && n > 0
}

func nextDocName(ctx context.Context, code, fallbackPrefix string) string {
	if name, err := orm.NextSequence(ctx, code); err == nil && name != "" {
		return name
	}
	return fmt.Sprintf("%s/%s/%05d", fallbackPrefix, time.Now().Format("2006"), time.Now().Unix()%100000)
}

// locationCompleteName builds "Parent/Child" style names walking up the tree.
func locationCompleteName(ctx context.Context, id int64) string {
	if id <= 0 {
		return ""
	}
	loc, err := orm.SearchOne(ctx, "stock.location", map[string]interface{}{"id": id})
	if err != nil || loc == nil {
		return ""
	}
	name := strings.TrimSpace(orm.AsString(loc["name"]))
	parentID, _ := orm.CoerceInt64(loc["location_id"])
	if parentID > 0 && parentID != id {
		if parent := locationCompleteName(ctx, parentID); parent != "" {
			return parent + "/" + name
		}
	}
	return name
}

// quantKey identifies a quant by product, location and optional lot.
func findQuant(ctx context.Context, productID, locationID, lotID int64) (int64, map[string]interface{}, error) {
	domain := [][]interface{}{
		{"product_id", "=", productID},
		{"location_id", "=", locationID},
	}
	if lotID > 0 {
		domain = append(domain, []interface{}{"lot_id", "=", lotID})
	}
	rows, err := orm.Search(ctx, "stock.quant", domain)
	if err != nil {
		return 0, nil, err
	}
	// Prefer a quant that actually holds stock so adjustments don't hit an
	// empty duplicate and leave the stocked quant untouched.
	var empty []map[string]interface{}
	for _, r := range rows {
		if lotID <= 0 {
			if l, _ := orm.CoerceInt64(r["lot_id"]); l > 0 {
				continue
			}
		}
		if numericFloat(r["quantity"]) > 0 || numericFloat(r["reserved_quantity"]) > 0 {
			id, _ := orm.CoerceInt64(r["id"])
			return id, r, nil
		}
		empty = append(empty, r)
	}
	if len(empty) > 0 {
		id, _ := orm.CoerceInt64(empty[0]["id"])
		return id, empty[0], nil
	}
	return 0, nil, nil
}

// adjustQuant moves delta units of product at a location, creating the quant if needed.
func adjustQuant(ctx context.Context, productID, locationID, lotID int64, delta float64) error {
	if productID <= 0 || locationID <= 0 || delta == 0 {
		return nil
	}
	id, row, err := findQuant(ctx, productID, locationID, lotID)
	if err != nil {
		return err
	}
	if id <= 0 {
		m, err := modelOrErr("stock.quant")
		if err != nil {
			return err
		}
		vals := map[string]interface{}{
			"product_id":         productID,
			"location_id":        locationID,
			"quantity":           delta,
			"available_quantity": delta,
			"in_date":            time.Now().Format(dtFormat),
		}
		if lotID > 0 {
			vals["lot_id"] = lotID
		}
		_, err = orm.Create(ctx, m, vals)
		return err
	}
	qty := numericFloat(row["quantity"]) + delta
	if qty < 0 {
		qty = 0
	}
	avail := qty - numericFloat(row["reserved_quantity"])
	if avail < 0 {
		avail = 0
	}
	return orm.UpdateRecordByID(ctx, "stock.quant", int(id), map[string]interface{}{
		"quantity":           qty,
		"available_quantity": avail,
	})
}

// productOnHand returns available quantity of a product at a location.
func productOnHand(ctx context.Context, productID, locationID int64) float64 {
	rows, err := orm.Search(ctx, "stock.quant", [][]interface{}{
		{"product_id", "=", productID},
		{"location_id", "=", locationID},
	})
	if err != nil {
		return 0
	}
	total := 0.0
	for _, r := range rows {
		q := numericFloat(r["quantity"]) - numericFloat(r["reserved_quantity"])
		if q > 0 {
			total += q
		}
	}
	return total
}

// productForecast adds incoming confirmed moves and subtracts outgoing confirmed moves.
func productForecast(ctx context.Context, productID, locationID int64) float64 {
	onHand := 0.0
	rows, err := orm.Search(ctx, "stock.quant", [][]interface{}{
		{"product_id", "=", productID},
		{"location_id", "=", locationID},
	})
	if err == nil {
		for _, r := range rows {
			onHand += numericFloat(r["quantity"])
		}
	}
	incoming, _ := orm.Search(ctx, "stock.move", [][]interface{}{
		{"product_id", "=", productID},
		{"location_dest_id", "=", locationID},
		{"state", "in", []interface{}{"draft", "waiting", "confirmed", "assigned"}},
	})
	outgoing, _ := orm.Search(ctx, "stock.move", [][]interface{}{
		{"product_id", "=", productID},
		{"location_id", "=", locationID},
		{"state", "in", []interface{}{"draft", "waiting", "confirmed", "assigned"}},
	})
	forecast := onHand
	for _, r := range incoming {
		forecast += numericFloat(r["product_qty"])
	}
	for _, r := range outgoing {
		forecast -= numericFloat(r["product_qty"])
	}
	return forecast
}

// recomputeLotQty updates stock.lot.product_qty from its quants.
func recomputeLotQty(ctx context.Context, lotID int64) error {
	if lotID <= 0 {
		return nil
	}
	rows, err := orm.Search(ctx, "stock.quant", [][]interface{}{{"lot_id", "=", lotID}})
	if err != nil {
		return err
	}
	total := 0.0
	for _, r := range rows {
		total += numericFloat(r["quantity"])
	}
	return orm.UpdateRecordByID(ctx, "stock.lot", int(lotID), map[string]interface{}{"product_qty": total})
}

// ConfirmPicking sets the picking and its moves to confirmed and assigns a name.
func ConfirmPicking(ctx context.Context, pickingID int) error {
	bypass := orm.ContextWithBypass(ctx, true)
	pick, err := orm.SearchOne(bypass, "stock.picking", map[string]interface{}{"id": pickingID})
	if err != nil || pick == nil {
		return fmt.Errorf("transfer not found")
	}
	if orm.AsString(pick["state"]) != "draft" {
		return fmt.Errorf("transfer already confirmed")
	}
	name := strings.TrimSpace(orm.AsString(pick["name"]))
	if name == "" {
		name = pickingDocName(bypass, pick)
	}
	updates := map[string]interface{}{
		"name":  name,
		"state": "confirmed",
	}
	if code := pickingCodeOf(bypass, pick); code != "" && code != orm.AsString(pick["picking_type_code"]) {
		updates["picking_type_code"] = code
	}
	if err := orm.UpdateRecordByID(bypass, "stock.picking", pickingID, updates); err != nil {
		return err
	}
	moves, _ := orm.Search(bypass, "stock.move", [][]interface{}{{"picking_id", "=", pickingID}})
	for _, mv := range moves {
		mvID, _ := orm.CoerceInt64(mv["id"])
		if orm.AsString(mv["state"]) == "draft" {
			if err := orm.UpdateRecordByID(bypass, "stock.move", int(mvID), map[string]interface{}{"state": "confirmed"}); err != nil {
				return err
			}
		}
	}
	return AssignPicking(ctx, pickingID)
}

func pickingDocName(ctx context.Context, pick map[string]interface{}) string {
	code := "stock.picking.internal"
	switch pickingCodeOf(ctx, pick) {
	case "incoming":
		code = "stock.picking.in"
	case "outgoing":
		code = "stock.picking.out"
	}
	prefix := "PICK"
	switch code {
	case "stock.picking.in":
		prefix = "IN"
	case "stock.picking.out":
		prefix = "OUT"
	}
	return nextDocName(ctx, code, prefix)
}

// pickingCodeOf returns the operation type code of a transfer row.
func pickingCodeOf(ctx context.Context, pick map[string]interface{}) string {
	if code := strings.TrimSpace(orm.AsString(pick["picking_type_code"])); code != "" {
		return code
	}
	typeID, _ := orm.CoerceInt64(pick["picking_type_id"])
	if typeID <= 0 {
		return ""
	}
	pt, err := orm.SearchOne(ctx, "stock.picking.type", map[string]interface{}{"id": typeID})
	if err != nil || pt == nil {
		return ""
	}
	return strings.TrimSpace(orm.AsString(pt["code"]))
}

// AssignReport describes the outcome of a reservation pass.
type AssignReport struct {
	// Unreserved lists moves that could not be fully reserved as
	// "Product (X available / Y needed)".
	Unreserved []string
	// Reserved counts moves that are now fully reserved.
	Reserved int
}

// AssignPicking checks availability and reserves quantities on the source
// quants (reserved_quantity/available_quantity) and maintains the move lines
// for the reserved amount (Odoo: action_assign). Reservation is partial: when
// stock is insufficient, whatever is available gets reserved and the rest
// stays unreserved. Internal callers (ConfirmPicking) ignore the report.
func AssignPicking(ctx context.Context, pickingID int) error {
	_, err := AssignPickingReport(ctx, pickingID)
	return err
}

// AssignPickingReport runs the reservation pass and reports which moves could
// not be fully reserved.
func AssignPickingReport(ctx context.Context, pickingID int) (AssignReport, error) {
	var report AssignReport
	bypass := orm.ContextWithBypass(ctx, true)
	pick, err := orm.SearchOne(bypass, "stock.picking", map[string]interface{}{"id": pickingID})
	if err != nil || pick == nil {
		return report, fmt.Errorf("transfer not found")
	}
	state := orm.AsString(pick["state"])
	if state != "confirmed" && state != "assigned" {
		return report, nil
	}
	moves, _ := orm.Search(bypass, "stock.move", [][]interface{}{{"picking_id", "=", pickingID}})
	allReserved := true
	for _, mv := range moves {
		mvID, _ := orm.CoerceInt64(mv["id"])
		if orm.AsString(mv["state"]) == "cancel" || orm.AsString(mv["state"]) == "done" {
			continue
		}
		productID, _ := orm.CoerceInt64(mv["product_id"])
		locID, _ := orm.CoerceInt64(mv["location_id"])
		demand := numericFloat(mv["product_qty"])
		if demand <= 0 {
			demand = numericFloat(mv["quantity"])
		}
		if demand <= 0 {
			continue
		}
		lines, _ := orm.Search(bypass, "stock.move.line", [][]interface{}{{"move_id", "=", mvID}})
		already := 0.0
		var lineID int64
		for _, ln := range lines {
			already += numericFloat(ln["quantity"])
			if lineID <= 0 {
				lineID, _ = orm.CoerceInt64(ln["id"])
			}
		}
		if already >= demand {
			if err := orm.UpdateRecordByID(bypass, "stock.move", int(mvID), map[string]interface{}{"state": "assigned"}); err != nil {
				return report, err
			}
			report.Reserved++
			continue
		}
		avail := productOnHand(bypass, productID, locID)
		needed := demand - already
		if avail <= 0 {
			allReserved = false
			report.Unreserved = append(report.Unreserved,
				fmt.Sprintf("%s (%.0f available / %.0f needed)", productName(bypass, productID), avail, demand))
			continue
		}
		reserved := reserveProduct(bypass, productID, locID, 0, needed)
		if reserved <= 0 {
			allReserved = false
			report.Unreserved = append(report.Unreserved,
				fmt.Sprintf("%s (%.0f available / %.0f needed)", productName(bypass, productID), avail, demand))
			continue
		}
		already += reserved
		if lineID > 0 {
			if err := orm.UpdateRecordByID(bypass, "stock.move.line", int(lineID), map[string]interface{}{
				"quantity": already,
				"state":    "assigned",
			}); err != nil {
				return report, err
			}
		} else {
			if err := createAssignedMoveLine(bypass, mv, already); err != nil {
				return report, err
			}
		}
		if already >= demand {
			if err := orm.UpdateRecordByID(bypass, "stock.move", int(mvID), map[string]interface{}{"state": "assigned"}); err != nil {
				return report, err
			}
			report.Reserved++
		} else {
			allReserved = false
			report.Unreserved = append(report.Unreserved,
				fmt.Sprintf("%s (%.0f available / %.0f needed)", productName(bypass, productID), avail, demand))
		}
	}
	newState := "confirmed"
	if allReserved && report.Reserved > 0 {
		newState = "assigned"
	}
	if err := orm.UpdateRecordByID(bypass, "stock.picking", pickingID, map[string]interface{}{"state": newState}); err != nil {
		return report, err
	}
	return report, nil
}

// reserveProduct reserves demand units across the product's quants at a
// location, updating reserved_quantity and available_quantity. It returns the
// amount actually reserved.
func reserveProduct(ctx context.Context, productID, locationID, lotID int64, demand float64) float64 {
	rows, err := orm.Search(ctx, "stock.quant", quantDomain(productID, locationID, lotID))
	if err != nil {
		return 0
	}
	remaining := demand
	for _, r := range rows {
		if remaining <= qtyEpsilon {
			break
		}
		if lotID <= 0 {
			if l, _ := orm.CoerceInt64(r["lot_id"]); l > 0 {
				continue
			}
		}
		avail := numericFloat(r["quantity"]) - numericFloat(r["reserved_quantity"])
		if avail <= qtyEpsilon {
			continue
		}
		take := avail
		if take > remaining {
			take = remaining
		}
		qid, _ := orm.CoerceInt64(r["id"])
		reserved := numericFloat(r["reserved_quantity"]) + take
		if err := orm.UpdateRecordByID(ctx, "stock.quant", int(qid), map[string]interface{}{
			"reserved_quantity":  reserved,
			"available_quantity": numericFloat(r["quantity"]) - reserved,
		}); err != nil {
			return demand - remaining
		}
		remaining -= take
	}
	return demand - remaining
}

// releaseProduct releases reserved units of product at a location, restoring
// available quantity.
func releaseProduct(ctx context.Context, productID, locationID, lotID int64, amount float64) {
	if amount <= qtyEpsilon {
		return
	}
	rows, err := orm.Search(ctx, "stock.quant", quantDomain(productID, locationID, lotID))
	if err != nil {
		return
	}
	remaining := amount
	for _, r := range rows {
		if remaining <= qtyEpsilon {
			break
		}
		if lotID <= 0 {
			if l, _ := orm.CoerceInt64(r["lot_id"]); l > 0 {
				continue
			}
		}
		reserved := numericFloat(r["reserved_quantity"])
		if reserved <= qtyEpsilon {
			continue
		}
		take := reserved
		if take > remaining {
			take = remaining
		}
		qid, _ := orm.CoerceInt64(r["id"])
		newReserved := reserved - take
		if err := orm.UpdateRecordByID(ctx, "stock.quant", int(qid), map[string]interface{}{
			"reserved_quantity":  newReserved,
			"available_quantity": numericFloat(r["quantity"]) - newReserved,
		}); err != nil {
			return
		}
		remaining -= take
	}
}

// ReleasePickingData releases the reservations of a transfer's moves and
// removes its moves and move lines. Call it while the transfer still exists
// (e.g. from the unlink guard): deleting the picking nulls picking_id on its
// moves via FK ON DELETE SET NULL, which would make them unfindable.
func ReleasePickingData(ctx context.Context, pickingID int) {
	if pickingID <= 0 {
		return
	}
	bypass := orm.ContextWithBypass(ctx, true)
	moves, err := orm.Search(bypass, "stock.move", [][]interface{}{{"picking_id", "=", pickingID}})
	if err != nil {
		return
	}
	for _, mv := range moves {
		mvID, _ := orm.CoerceInt64(mv["id"])
		if mvID <= 0 {
			continue
		}
		productID, _ := orm.CoerceInt64(mv["product_id"])
		srcID, _ := orm.CoerceInt64(mv["location_id"])
		lines, _ := orm.Search(bypass, "stock.move.line", [][]interface{}{{"move_id", "=", mvID}})
		reserved := 0.0
		for _, ln := range lines {
			reserved += numericFloat(ln["quantity"])
		}
		if reserved > 0 {
			releaseProduct(bypass, productID, srcID, 0, reserved)
		}
		for _, ln := range lines {
			lnID, _ := orm.CoerceInt64(ln["id"])
			if lnID > 0 {
				_ = orm.Unlink(bypass, "stock.move.line", int(lnID))
			}
		}
		_ = orm.Unlink(bypass, "stock.move", int(mvID))
	}
}

// quantDomain builds the quant lookup domain for a product/location/lot.
func quantDomain(productID, locationID, lotID int64) [][]interface{} {
	domain := [][]interface{}{
		{"product_id", "=", productID},
		{"location_id", "=", locationID},
	}
	if lotID > 0 {
		domain = append(domain, []interface{}{"lot_id", "=", lotID})
	}
	return domain
}

// createAssignedMoveLine creates the stock.move.line backing a reservation.
func createAssignedMoveLine(ctx context.Context, mv map[string]interface{}, qty float64) error {
	lineModel, err := modelOrErr("stock.move.line")
	if err != nil {
		return err
	}
	now := time.Now().Format(dtFormat)
	vals := map[string]interface{}{
		"move_id":          mv["id"],
		"product_id":       mv["product_id"],
		"quantity":         qty,
		"location_id":      mv["location_id"],
		"location_dest_id": mv["location_dest_id"],
		"state":            "assigned",
		"date":             now,
		"reference":        orm.AsString(mv["name"]),
	}
	if cid, ok := orm.CoerceInt64(mv["company_id"]); ok && cid > 0 {
		vals["company_id"] = mv["company_id"]
	}
	_, err = orm.Create(ctx, lineModel, vals)
	return err
}

// ValidatePicking applies move quantities to quants and closes the transfer.
func ValidatePicking(ctx context.Context, pickingID int) error {
	bypass := orm.ContextWithBypass(ctx, true)
	pick, err := orm.SearchOne(bypass, "stock.picking", map[string]interface{}{"id": pickingID})
	if err != nil || pick == nil {
		return fmt.Errorf("transfer not found")
	}
	if orm.AsString(pick["state"]) == "done" {
		return nil
	}
	moves, _ := orm.Search(bypass, "stock.move", [][]interface{}{{"picking_id", "=", pickingID}})
	now := time.Now().Format(dtFormat)
	for _, mv := range moves {
		mvID, _ := orm.CoerceInt64(mv["id"])
		if orm.AsString(mv["state"]) == "cancel" {
			continue
		}
		productID, _ := orm.CoerceInt64(mv["product_id"])
		srcID, _ := orm.CoerceInt64(mv["location_id"])
		destID, _ := orm.CoerceInt64(mv["location_dest_id"])
		doneQty := numericFloat(mv["quantity"])
		if doneQty <= 0 {
			doneQty = numericFloat(mv["product_qty"])
		}
		tracking := productTracking(bypass, productID)
		lines, _ := orm.Search(bypass, "stock.move.line", [][]interface{}{{"move_id", "=", mvID}})
		if len(lines) > 0 {
			doneQty = 0
			for _, ln := range lines {
				lq := numericFloat(ln["quantity"])
				lotID, _ := orm.CoerceInt64(ln["lot_id"])
				if lq <= 0 {
					continue
				}
				if lotID <= 0 {
					if lotName := strings.TrimSpace(orm.AsString(ln["lot_name"])); lotName != "" {
						if lotID, err = getOrCreateLot(bypass, productID, lotName); err != nil {
							return err
						}
					}
				}
				if (tracking == "lot" || tracking == "serial") && lotID <= 0 {
					return fmt.Errorf("product %q requires a lot/serial number", productName(bypass, productID))
				}
				if tracking == "serial" && lq != 1 {
					return fmt.Errorf("serial tracked product %q must have quantity 1 per line", productName(bypass, productID))
				}
				doneQty += lq
				if err := adjustQuant(bypass, productID, srcID, lotID, -lq); err != nil {
					return err
				}
				// Consume the reservation made by Check Availability on the source quants.
				releaseProduct(bypass, productID, srcID, lotID, lq)
				if err := adjustQuant(bypass, productID, destID, lotID, lq); err != nil {
					return err
				}
				if lotID > 0 {
					if err := recomputeLotQty(bypass, lotID); err != nil {
						return err
					}
				}
				lnID, _ := orm.CoerceInt64(ln["id"])
				upd := map[string]interface{}{"state": "done"}
				if lotID > 0 {
					upd["lot_id"] = lotID
				}
				if err := orm.UpdateRecordByID(bypass, "stock.move.line", int(lnID), upd); err != nil {
					return err
				}
			}
		} else if doneQty > 0 {
			if tracking == "lot" || tracking == "serial" {
				return fmt.Errorf("product %q requires a lot/serial number", productName(bypass, productID))
			}
			if err := adjustQuant(bypass, productID, srcID, 0, -doneQty); err != nil {
				return err
			}
			if err := adjustQuant(bypass, productID, destID, 0, doneQty); err != nil {
				return err
			}
		}
		if err := orm.UpdateRecordByID(bypass, "stock.move", int(mvID), map[string]interface{}{
			"state":    "done",
			"quantity": doneQty,
			"date":     now,
		}); err != nil {
			return err
		}
		if err := ApplyPushRulesForMove(bypass, int(mvID)); err != nil {
			log.Printf("stock: apply push rule for move %d: %v", mvID, err)
		}
	}
	return orm.UpdateRecordByID(bypass, "stock.picking", pickingID, map[string]interface{}{
		"state":     "done",
		"date_done": now,
	})
}

// CancelPicking cancels the transfer and its moves.
func CancelPicking(ctx context.Context, pickingID int) error {
	bypass := orm.ContextWithBypass(ctx, true)
	pick, err := orm.SearchOne(bypass, "stock.picking", map[string]interface{}{"id": pickingID})
	if err != nil || pick == nil {
		return fmt.Errorf("transfer not found")
	}
	if orm.AsString(pick["state"]) == "done" {
		return fmt.Errorf("done transfer cannot be cancelled")
	}
	moves, _ := orm.Search(bypass, "stock.move", [][]interface{}{{"picking_id", "=", pickingID}})
	for _, mv := range moves {
		mvID, _ := orm.CoerceInt64(mv["id"])
		if orm.AsString(mv["state"]) == "done" {
			continue
		}
		// Release the reservation made by Check Availability before cancelling.
		productID, _ := orm.CoerceInt64(mv["product_id"])
		srcID, _ := orm.CoerceInt64(mv["location_id"])
		lines, _ := orm.Search(bypass, "stock.move.line", [][]interface{}{{"move_id", "=", mvID}})
		reserved := 0.0
		for _, ln := range lines {
			if orm.AsString(ln["state"]) != "cancel" {
				reserved += numericFloat(ln["quantity"])
			}
		}
		if reserved > 0 {
			releaseProduct(bypass, productID, srcID, 0, reserved)
		}
		for _, ln := range lines {
			if orm.AsString(ln["state"]) == "cancel" {
				continue
			}
			lnID, _ := orm.CoerceInt64(ln["id"])
			if err := orm.UpdateRecordByID(bypass, "stock.move.line", int(lnID), map[string]interface{}{"state": "cancel"}); err != nil {
				return err
			}
		}
		if err := orm.UpdateRecordByID(bypass, "stock.move", int(mvID), map[string]interface{}{"state": "cancel"}); err != nil {
			return err
		}
	}
	return orm.UpdateRecordByID(bypass, "stock.picking", pickingID, map[string]interface{}{"state": "cancel"})
}

// ValidateScrap moves scrap_qty from source location to scrap location.
func ValidateScrap(ctx context.Context, scrapID int) error {
	bypass := orm.ContextWithBypass(ctx, true)
	scrap, err := orm.SearchOne(bypass, "stock.scrap", map[string]interface{}{"id": scrapID})
	if err != nil || scrap == nil {
		return fmt.Errorf("scrap not found")
	}
	if orm.AsString(scrap["state"]) == "done" {
		return nil
	}
	productID, _ := orm.CoerceInt64(scrap["product_id"])
	locID, _ := orm.CoerceInt64(scrap["location_id"])
	scrapLocID, _ := orm.CoerceInt64(scrap["scrap_location_id"])
	lotID, _ := orm.CoerceInt64(scrap["lot_id"])
	qty := numericFloat(scrap["scrap_qty"])
	if qty <= 0 {
		return fmt.Errorf("quantity must be positive")
	}
	if err := adjustQuant(bypass, productID, locID, lotID, -qty); err != nil {
		return err
	}
	if err := adjustQuant(bypass, productID, scrapLocID, lotID, qty); err != nil {
		return err
	}
	if err := recomputeLotQty(bypass, lotID); err != nil {
		return err
	}
	return orm.UpdateRecordByID(bypass, "stock.scrap", scrapID, map[string]interface{}{
		"state":     "done",
		"date_done": time.Now().Format(dtFormat),
	})
}

// RecomputeOrderpoint refreshes on-hand / forecast / to-order values.
func RecomputeOrderpoint(ctx context.Context, opID int) error {
	bypass := orm.ContextWithBypass(ctx, true)
	op, err := orm.SearchOne(bypass, "stock.warehouse.orderpoint", map[string]interface{}{"id": opID})
	if err != nil || op == nil {
		return fmt.Errorf("reordering rule not found")
	}
	productID, _ := orm.CoerceInt64(op["product_id"])
	locID, _ := orm.CoerceInt64(op["location_id"])
	onHand := productOnHand(bypass, productID, locID)
	forecast := productForecast(bypass, productID, locID)
	toOrder := 0.0
	if forecast < numericFloat(op["product_min_qty"]) {
		toOrder = numericFloat(op["product_max_qty"]) - forecast
		if toOrder < 0 {
			toOrder = 0
		}
	}
	return orm.UpdateRecordByID(bypass, "stock.warehouse.orderpoint", opID, map[string]interface{}{
		"qty_on_hand":  onHand,
		"qty_forecast": forecast,
		"qty_to_order": toOrder,
	})
}

// ReplenishOrderpoint creates a receipt for the suggested quantity.
func ReplenishOrderpoint(ctx context.Context, opID int) (int, error) {
	bypass := orm.ContextWithBypass(ctx, true)
	op, err := orm.SearchOne(bypass, "stock.warehouse.orderpoint", map[string]interface{}{"id": opID})
	if err != nil || op == nil {
		return 0, fmt.Errorf("reordering rule not found")
	}
	if err := RecomputeOrderpoint(bypass, opID); err != nil {
		return 0, err
	}
	op, _ = orm.SearchOne(bypass, "stock.warehouse.orderpoint", map[string]interface{}{"id": opID})
	qty := numericFloat(op["qty_to_order"])
	if qty <= 0 {
		return 0, fmt.Errorf("no replenishment needed")
	}
	productID, _ := orm.CoerceInt64(op["product_id"])
	whID, _ := orm.CoerceInt64(op["warehouse_id"])
	companyID, _ := orm.CoerceInt64(op["company_id"])
	origin := strings.TrimSpace(orm.AsString(op["name"]))
	if origin == "" {
		origin = fmt.Sprintf("Reordering Rule %d", opID)
	}
	return CreateTransfer(ctx, TransferSpec{
		Origin:      origin,
		CompanyID:   companyID,
		Code:        "incoming",
		WarehouseID: whID,
		Note:        "Automatic replenishment",
		Lines: []TransferLine{{
			ProductID: productID,
			Name:      productName(bypass, productID),
			Qty:       qty,
		}},
	})
}

func productName(ctx context.Context, productID int64) string {
	if productID <= 0 {
		return ""
	}
	p, err := orm.SearchOne(ctx, "product.product", map[string]interface{}{"id": productID})
	if err != nil || p == nil {
		return ""
	}
	return orm.AsString(p["name"])
}

// productTracking returns the tracking mode of a product: none, lot, or serial.
func productTracking(ctx context.Context, productID int64) string {
	if productID <= 0 {
		return "none"
	}
	p, err := orm.SearchOne(ctx, "product.product", map[string]interface{}{"id": productID})
	if err != nil || p == nil {
		return "none"
	}
	t := strings.TrimSpace(orm.AsString(p["tracking"]))
	if t == "" {
		return "none"
	}
	return t
}

// getOrCreateLot finds an existing lot/serial for a product by name or creates one.
func getOrCreateLot(ctx context.Context, productID int64, name string) (int64, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, fmt.Errorf("lot/serial number is required")
	}
	rows, err := orm.Search(ctx, "stock.lot", [][]interface{}{
		{"product_id", "=", productID},
		{"name", "=", name},
	})
	if err != nil {
		return 0, err
	}
	for _, r := range rows {
		if id, ok := coerceID(r["id"]); ok {
			return int64(id), nil
		}
	}
	m, err := modelOrErr("stock.lot")
	if err != nil {
		return 0, err
	}
	id, err := orm.Create(ctx, m, map[string]interface{}{
		"name":       name,
		"product_id": productID,
	})
	if err != nil {
		return 0, err
	}
	return int64(id), nil
}
