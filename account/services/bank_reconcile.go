package services

import (
	"context"
	"math"

	"sumeru/core/orm"
)

// ReconcileStatementLine marks a move line reconciled against a bank statement amount.
func ReconcileStatementLine(ctx context.Context, moveLineID int, amount float64) error {
	line, err := orm.SearchOne(ctx, "account.move.line", map[string]interface{}{"id": moveLineID})
	if err != nil {
		return err
	}
	res := numericFloat(line["amount_residual"])
	if math.Abs(res) <= balanceEpsilon {
		return nil
	}
	newRes := res
	if res > 0 {
		newRes = res - math.Abs(amount)
	} else {
		newRes = res + math.Abs(amount)
	}
	if err := orm.UpdateRecordByID(ctx, "account.move.line", moveLineID, map[string]interface{}{
		"amount_residual": round2(newRes),
		"reconciled":      math.Abs(newRes) <= balanceEpsilon,
	}); err != nil {
		return err
	}
	moveID, _ := orm.CoerceInt64(line["move_id"])
	if moveID > 0 {
		_ = recomputeInvoicePaymentState(ctx, int(moveID))
	}
	return nil
}
