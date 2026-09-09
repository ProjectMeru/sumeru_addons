package services

import (
	"context"
	"log"

	"sumeru/core/event"
	"sumeru/core/orm"
)

func init() {
	event.Subscribe("record.created", onLocationCreated)
	event.Subscribe("record.updated", onLocationUpdated)
	event.Subscribe("record.updated", onOrderpointUpdated)
	event.Subscribe("record.created", onPickingTypeSync)
	event.Subscribe("record.updated", onPickingTypeSync)
	event.Subscribe("record.created", onQuantCountWritten)
	event.Subscribe("record.updated", onQuantCountWritten)
	orm.RegisterOnchange("stock.picking", "picking_type_id", onPickingTypeChange)
	orm.RegisterOnchange("stock.move", "product_id", onMoveProductChange)
	orm.RegisterOnchange("stock.quant", "inventory_quantity", onQuantInventoryQuantityChange)
}

func onLocationCreated(ctx context.Context, ev event.Event) error {
	model, _ := ev.Payload["model"].(string)
	if model != "stock.location" {
		return nil
	}
	id, ok := coerceID(ev.Payload["id"])
	if !ok {
		return nil
	}
	return refreshLocationCompleteName(orm.ContextWithBypass(ctx, true), id)
}

func onLocationUpdated(ctx context.Context, ev event.Event) error {
	model, _ := ev.Payload["model"].(string)
	if model != "stock.location" {
		return nil
	}
	id, ok := coerceID(ev.Payload["id"])
	if !ok {
		return nil
	}
	bypass := orm.ContextWithBypass(ctx, true)
	if err := refreshLocationCompleteName(bypass, id); err != nil {
		return err
	}
	children, _ := orm.Search(bypass, "stock.location", [][]interface{}{{"location_id", "=", id}})
	for _, c := range children {
		if cid, ok := coerceID(c["id"]); ok {
			if err := refreshLocationCompleteName(bypass, cid); err != nil {
				return err
			}
		}
	}
	return nil
}

func refreshLocationCompleteName(ctx context.Context, locID int) error {
	loc, err := orm.SearchOne(ctx, "stock.location", map[string]interface{}{"id": locID})
	if err != nil || loc == nil {
		return nil
	}
	name := orm.AsString(loc["name"])
	complete := name
	parentID, _ := orm.CoerceInt64(loc["location_id"])
	if parentID > 0 {
		if p := locationCompleteName(ctx, parentID); p != "" {
			complete = p + "/" + name
		}
	}
	return orm.UpdateRecordByID(ctx, "stock.location", locID, map[string]interface{}{
		"complete_name": complete,
	})
}

func onOrderpointUpdated(ctx context.Context, ev event.Event) error {
	model, _ := ev.Payload["model"].(string)
	if model != "stock.warehouse.orderpoint" {
		return nil
	}
	id, ok := coerceID(ev.Payload["id"])
	if !ok {
		return nil
	}
	if err := RecomputeOrderpoint(ctx, id); err != nil {
		log.Printf("stock: recompute orderpoint %d: %v", id, err)
	}
	return nil
}

func onPickingTypeChange(ctx context.Context, values map[string]interface{}, field string) (orm.OnchangeResult, error) {
	result := orm.OnchangeResult{Value: map[string]interface{}{}}
	typeID, ok := coerceID(values["picking_type_id"])
	if !ok || typeID <= 0 {
		return result, nil
	}
	pt, err := orm.SearchOne(orm.ContextWithBypass(ctx, true), "stock.picking.type", map[string]interface{}{"id": typeID})
	if err != nil || pt == nil {
		return result, nil
	}
	if src, ok := orm.CoerceInt64(pt["default_location_src_id"]); ok && src > 0 {
		result.Value["location_id"] = src
	}
	if dest, ok := orm.CoerceInt64(pt["default_location_dest_id"]); ok && dest > 0 {
		result.Value["location_dest_id"] = dest
	}
	if code := orm.AsString(pt["code"]); code != "" {
		result.Value["picking_type_code"] = code
	}
	return result, nil
}

func onMoveProductChange(ctx context.Context, values map[string]interface{}, field string) (orm.OnchangeResult, error) {
	result := orm.OnchangeResult{Value: map[string]interface{}{}}
	productID, ok := coerceID(values["product_id"])
	if !ok || productID <= 0 {
		return result, nil
	}
	if name := productName(orm.ContextWithBypass(ctx, true), int64(productID)); name != "" {
		result.Value["name"] = name
	}
	return result, nil
}

// onQuantInventoryQuantityChange previews the difference while counting in the form.
func onQuantInventoryQuantityChange(_ context.Context, values map[string]interface{}, _ string) (orm.OnchangeResult, error) {
	result := orm.OnchangeResult{Value: map[string]interface{}{}}
	counted := numericFloat(values["inventory_quantity"])
	onHand := numericFloat(values["quantity"])
	result.Value["inventory_diff_quantity"] = counted - onHand
	result.Value["inventory_quantity_set"] = true
	return result, nil
}

// onQuantCountWritten keeps inventory_diff_quantity in sync with user writes.
func onQuantCountWritten(ctx context.Context, ev event.Event) error {
	model, _ := ev.Payload["model"].(string)
	if model != "stock.quant" {
		return nil
	}
	id, ok := coerceID(ev.Payload["id"])
	if !ok {
		return nil
	}
	bypass := orm.ContextWithBypass(ctx, true)
	quant, err := orm.SearchOne(bypass, "stock.quant", map[string]interface{}{"id": id})
	if err != nil || quant == nil {
		return nil
	}
	updates := map[string]interface{}{}
	onHand := numericFloat(quant["quantity"])
	reserved := numericFloat(quant["reserved_quantity"])
	available := onHand - reserved
	if available < 0 {
		available = 0
	}
	if numericFloat(quant["available_quantity"]) != available {
		updates["available_quantity"] = available
	}
	counted := numericFloat(quant["inventory_quantity"])
	countSet := asBool(quant["inventory_quantity_set"])
	if !countSet && counted != 0 {
		// API/import writes may set the counted quantity without the flag.
		countSet = true
		updates["inventory_quantity_set"] = true
	}
	if countSet {
		diff := counted - onHand
		if numericFloat(quant["inventory_diff_quantity"]) != diff {
			updates["inventory_diff_quantity"] = diff
		}
	}
	if len(updates) == 0 {
		return nil
	}
	if err := orm.UpdateRecordByID(bypass, "stock.quant", id, updates); err != nil {
		log.Printf("stock: sync quant %d: %v", id, err)
	}
	if lotID, ok := orm.CoerceInt64(quant["lot_id"]); ok && lotID > 0 {
		if err := recomputeLotQty(bypass, lotID); err != nil {
			log.Printf("stock: recompute lot %d: %v", lotID, err)
		}
	}
	return nil
}

// onPickingTypeSync stores the operation type code on the transfer so list
// actions can filter Receipts / Delivery Orders / Internal Transfers.
func onPickingTypeSync(ctx context.Context, ev event.Event) error {
	model, _ := ev.Payload["model"].(string)
	if model != "stock.picking" {
		return nil
	}
	id, ok := coerceID(ev.Payload["id"])
	if !ok {
		return nil
	}
	bypass := orm.ContextWithBypass(ctx, true)
	pick, err := orm.SearchOne(bypass, "stock.picking", map[string]interface{}{"id": id})
	if err != nil || pick == nil {
		return nil
	}
	typeID, _ := orm.CoerceInt64(pick["picking_type_id"])
	code := ""
	if typeID > 0 {
		if pt, err := orm.SearchOne(bypass, "stock.picking.type", map[string]interface{}{"id": typeID}); err == nil && pt != nil {
			code = orm.AsString(pt["code"])
		}
	}
	if code == "" || code == orm.AsString(pick["picking_type_code"]) {
		return nil
	}
	if err := orm.UpdateRecordByID(bypass, "stock.picking", id, map[string]interface{}{
		"picking_type_code": code,
	}); err != nil {
		log.Printf("stock: sync picking type code %d: %v", id, err)
	}
	return nil
}
