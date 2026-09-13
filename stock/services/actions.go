package services

import (
	"context"
	"fmt"
	"strings"

	"sumeru/core/orm"
)

func init() {
	orm.RegisterObjectAction("stock.picking", "action_confirm", actionConfirmPicking)
	orm.RegisterObjectAction("stock.picking", "action_assign", actionAssignPicking)
	orm.RegisterObjectAction("stock.picking", "action_done", actionDonePicking)
	orm.RegisterObjectAction("stock.picking", "action_cancel", actionCancelPicking)
	orm.RegisterObjectAction("stock.picking", "action_print", actionPrintPicking)
	orm.RegisterObjectAction("stock.scrap", "action_validate", actionValidateScrap)
	orm.RegisterObjectAction("stock.warehouse.orderpoint", "action_replenish", actionReplenishOrderpoint)
	orm.RegisterObjectAction("stock.quant", "action_apply_inventory", actionApplyInventory)
	orm.RegisterObjectAction("stock.quant", "action_apply_all_inventory", actionApplyAllInventory)
	orm.RegisterObjectAction("stock.quant", "action_set_inventory_zero", actionSetInventoryZero)
	orm.RegisterObjectAction("stock.quant", "action_clear_inventory", actionClearInventory)

	orm.RegisterWriteGuard("stock.picking", guardPickingWrite)
	orm.RegisterUnlinkGuard("stock.picking", guardPickingUnlink)
}

// guardPickingWrite locks done and cancelled transfers; printed (set by the
// print route) stays writable so delivery notes can still be marked printed.
func guardPickingWrite(_ context.Context, _ string, before map[string]interface{}, values map[string]interface{}) error {
	state := orm.AsString(before["state"])
	if state != "done" && state != "cancel" {
		return nil
	}
	for k := range values {
		if k != "printed" {
			return fmt.Errorf("cannot modify a %s transfer", state)
		}
	}
	return nil
}

// guardPickingUnlink blocks deleting only done transfers. Deleting a delivery
// (confirmed/assigned) is allowed; the guard cleans up the transfer's
// reservations and moves first, while they are still linked to the picking.
func guardPickingUnlink(ctx context.Context, _ string, record map[string]interface{}) error {
	state := orm.AsString(record["state"])
	if state == "done" {
		return fmt.Errorf("cannot delete a done transfer")
	}
	if id, ok := orm.CoerceInt64(record["id"]); ok && id > 0 {
		ReleasePickingData(ctx, int(id))
	}
	return nil
}

func actionConfirmPicking(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if err := ConfirmPicking(ctx, id); err != nil {
		return "", err
	}
	return vals["next"], nil
}

func actionAssignPicking(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	report, err := AssignPickingReport(ctx, id)
	if err != nil {
		return "", err
	}
	if len(report.Unreserved) > 0 {
		return "", fmt.Errorf("not enough stock to reserve: %s", strings.Join(report.Unreserved, "; "))
	}
	return vals["next"], nil
}

func actionDonePicking(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if err := ValidatePicking(ctx, id); err != nil {
		return "", err
	}
	return vals["next"], nil
}

func actionCancelPicking(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if err := CancelPicking(ctx, id); err != nil {
		return "", err
	}
	return vals["next"], nil
}

func actionPrintPicking(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	return fmt.Sprintf("/stock/picking/print?id=%d", id), nil
}

func actionValidateScrap(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if err := ValidateScrap(ctx, id); err != nil {
		return "", err
	}
	return vals["next"], nil
}

func actionReplenishOrderpoint(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if _, err := ReplenishOrderpoint(ctx, id); err != nil {
		return "", err
	}
	return vals["next"], nil
}

func actionApplyInventory(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if err := ApplyInventoryQuant(ctx, id); err != nil {
		return "", err
	}
	return vals["next"], nil
}

func actionApplyAllInventory(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if _, err := ApplyAllInventoryCounts(ctx); err != nil {
		return "", err
	}
	return vals["next"], nil
}

func actionSetInventoryZero(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if err := SetInventoryCount(ctx, id, 0); err != nil {
		return "", err
	}
	return vals["next"], nil
}

func actionClearInventory(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if err := ClearInventoryCount(ctx, id); err != nil {
		return "", err
	}
	return vals["next"], nil
}
