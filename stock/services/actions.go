package services

import (
	"context"
	"fmt"

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
}

func actionConfirmPicking(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if err := ConfirmPicking(ctx, id); err != nil {
		return "", err
	}
	return vals["next"], nil
}

func actionAssignPicking(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if err := AssignPicking(ctx, id); err != nil {
		return "", err
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
