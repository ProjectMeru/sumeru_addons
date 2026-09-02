package services

import (
	"context"
	"fmt"
	"math"

	"sumeru/core/orm"
)

func reconcilePaymentToInvoice(ctx context.Context, invoiceID, paymentLineID int, amount float64) error {
	inv, err := orm.SearchOne(ctx, "account.move", map[string]interface{}{"id": invoiceID})
	if err != nil {
		return err
	}
	moveType := orm.AsString(inv["move_type"])
	lines, _ := orm.Search(ctx, "account.move.line", [][]interface{}{
		{"move_id", "=", invoiceID},
		{"display_type", "=", "entry"},
	})
	var invLine map[string]interface{}
	for _, ln := range lines {
		res := numericFloat(ln["amount_residual"])
		if math.Abs(res) < balanceEpsilon {
			continue
		}
		acctID, _ := orm.CoerceInt64(ln["account_id"])
		acct, _ := orm.SearchOne(ctx, "account.account", map[string]interface{}{"id": acctID})
		atype := orm.AsString(acct["account_type"])
		switch moveType {
		case "out_invoice", "out_refund":
			if atype == "asset_receivable" {
				invLine = ln
				break
			}
		case "in_invoice", "in_refund":
			if atype == "liability_payable" {
				invLine = ln
				break
			}
		}
		if invLine == nil && math.Abs(res) > balanceEpsilon {
			invLine = ln
		}
	}
	if invLine == nil {
		return fmt.Errorf("no open line")
	}
	invLineID, _ := orm.CoerceInt64(invLine["id"])
	invRes := numericFloat(invLine["amount_residual"])
	payLine, err := orm.SearchOne(ctx, "account.move.line", map[string]interface{}{"id": paymentLineID})
	if err != nil {
		return err
	}
	payRes := numericFloat(payLine["amount_residual"])

	reconcileAmt := math.Min(math.Abs(invRes), math.Min(math.Abs(payRes), amount))
	if reconcileAmt <= balanceEpsilon {
		return nil
	}

	debitID, creditID := int(invLineID), paymentLineID
	if invRes < 0 {
		debitID, creditID = paymentLineID, int(invLineID)
	}
	if numericFloat(invLine["debit"]) > 0 || invRes > 0 {
		debitID, creditID = int(invLineID), paymentLineID
	} else {
		debitID, creditID = paymentLineID, int(invLineID)
	}

	recModel, err := modelOrErr("account.partial.reconcile")
	if err != nil {
		return err
	}
	_, err = orm.Create(ctx, recModel, map[string]interface{}{
		"debit_move_id":  debitID,
		"credit_move_id": creditID,
		"amount":         round2(reconcileAmt),
	})
	if err != nil {
		return err
	}

	newInvRes := invRes
	newPayRes := payRes
	if invRes > 0 {
		newInvRes = invRes - reconcileAmt
	} else {
		newInvRes = invRes + reconcileAmt
	}
	if payRes > 0 {
		newPayRes = payRes - reconcileAmt
	} else {
		newPayRes = payRes + reconcileAmt
	}
	if err := orm.UpdateRecordByID(ctx, "account.move.line", int(invLineID), map[string]interface{}{
		"amount_residual": round2(newInvRes),
		"reconciled":      math.Abs(newInvRes) <= balanceEpsilon,
	}); err != nil {
		return err
	}
	if err := orm.UpdateRecordByID(ctx, "account.move.line", paymentLineID, map[string]interface{}{
		"amount_residual": round2(newPayRes),
		"reconciled":      math.Abs(newPayRes) <= balanceEpsilon,
	}); err != nil {
		return err
	}
	if math.Abs(newInvRes) <= balanceEpsilon {
		_ = maybeCreateFullReconcile(ctx, int(invLineID))
	}
	return recomputeInvoicePaymentState(ctx, invoiceID)
}

func recomputeInvoicePaymentState(ctx context.Context, invoiceID int) error {
	inv, err := orm.SearchOne(ctx, "account.move", map[string]interface{}{"id": invoiceID})
	if err != nil {
		return err
	}
	total := numericFloat(inv["amount_total"])
	lines, _ := orm.Search(ctx, "account.move.line", [][]interface{}{
		{"move_id", "=", invoiceID},
		{"display_type", "=", "entry"},
	})
	residual := 0.0
	found := false
	for _, ln := range lines {
		acctID, _ := orm.CoerceInt64(ln["account_id"])
		acct, _ := orm.SearchOne(ctx, "account.account", map[string]interface{}{"id": acctID})
		atype := orm.AsString(acct["account_type"])
		if atype == "asset_receivable" || atype == "liability_payable" {
			residual += math.Abs(numericFloat(ln["amount_residual"]))
			found = true
		}
	}
	if !found {
		residual = total
	}
	state := "partial"
	if residual <= balanceEpsilon {
		state = "paid"
		residual = 0
	} else if residual+balanceEpsilon >= total {
		state = "not_paid"
	}
	if err := orm.UpdateRecordByID(ctx, "account.move", invoiceID, map[string]interface{}{
		"amount_residual": round2(residual),
		"payment_state":   state,
	}); err != nil {
		return err
	}
	return upsertInvoiceReport(ctx, invoiceID)
}

func upsertInvoiceReport(ctx context.Context, moveID int) error {
	move, err := orm.SearchOne(ctx, "account.move", map[string]interface{}{"id": moveID})
	if err != nil {
		return err
	}
	mt := orm.AsString(move["move_type"])
	if mt == "entry" {
		return nil
	}
	existing, _ := orm.Search(ctx, "account.invoice.report", [][]interface{}{
		{"move_id", "=", moveID},
	})
	vals := map[string]interface{}{
		"move_id":         moveID,
		"name":            orm.AsString(move["name"]),
		"partner_id":      move["partner_id"],
		"move_type":       mt,
		"invoice_date":    move["invoice_date"],
		"state":           move["state"],
		"payment_state":   move["payment_state"],
		"amount_untaxed":  move["amount_untaxed"],
		"amount_tax":      move["amount_tax"],
		"amount_total":    move["amount_total"],
		"amount_residual": move["amount_residual"],
	}
	if len(existing) > 0 {
		id, _ := orm.CoerceInt64(existing[0]["id"])
		return orm.UpdateRecordByID(ctx, "account.invoice.report", int(id), vals)
	}
	m, err := modelOrErr("account.invoice.report")
	if err != nil {
		return nil
	}
	_, err = orm.Create(ctx, m, vals)
	return err
}

func CancelMove(ctx context.Context, moveID int) error {
	bypass := orm.ContextWithBypass(ctx, true)
	move, err := orm.SearchOne(bypass, "account.move", map[string]interface{}{"id": moveID})
	if err != nil {
		return err
	}
	if orm.AsString(move["payment_state"]) == "paid" || orm.AsString(move["payment_state"]) == "partial" {
		if numericFloat(move["amount_residual"])+balanceEpsilon < numericFloat(move["amount_total"]) {
			return fmt.Errorf("invoice reconciled")
		}
	}
	return orm.UpdateRecordByID(bypass, "account.move", moveID, map[string]interface{}{
		"state": "cancel",
	})
}

func ResetMoveToDraft(ctx context.Context, moveID int) error {
	bypass := orm.ContextWithBypass(ctx, true)
	move, err := orm.SearchOne(bypass, "account.move", map[string]interface{}{"id": moveID})
	if err != nil {
		return err
	}
	if orm.AsString(move["state"]) != "posted" && orm.AsString(move["state"]) != "cancel" {
		return nil
	}
	residual := numericFloat(move["amount_residual"])
	total := numericFloat(move["amount_total"])
	if orm.AsString(move["move_type"]) != "entry" && residual+balanceEpsilon < total {
		return fmt.Errorf("payments exist")
	}
	return orm.UpdateRecordByID(bypass, "account.move", moveID, map[string]interface{}{
		"state":         "draft",
		"payment_state": "not_paid",
	})
}
