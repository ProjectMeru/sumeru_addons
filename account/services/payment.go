package services

import (
	"context"
	"fmt"
	"time"

	"sumeru/core/orm"
)

func PostPayment(ctx context.Context, paymentID int) error {
	if paymentID <= 0 {
		return fmt.Errorf("bad payment id")
	}
	bypass := orm.ContextWithBypass(ctx, true)
	pay, err := orm.SearchOne(bypass, "account.payment", map[string]interface{}{"id": paymentID})
	if err != nil {
		return err
	}
	if orm.AsString(pay["state"]) == "posted" {
		return nil
	}
	if orm.AsString(pay["state"]) == "cancelled" {
		return fmt.Errorf("payment cancelled")
	}
	amount := numericFloat(pay["amount"])
	if amount <= 0 {
		return fmt.Errorf("amount zero")
	}
	partnerID, _ := orm.CoerceInt64(pay["partner_id"])
	paymentType := orm.AsString(pay["payment_type"])
	journalID, _ := orm.CoerceInt64(pay["journal_id"])
	if journalID <= 0 {
		if paymentType == "outbound" {
			journalID = journalIDByType(bypass, "cash")
			if journalID <= 0 {
				journalID = journalIDByType(bypass, "bank")
			}
		} else {
			journalID = journalIDByType(bypass, "bank")
			if journalID <= 0 {
				journalID = journalIDByType(bypass, "cash")
			}
		}
	}
	liquidityID := liquidityAccountID(bypass, journalID)
	if liquidityID <= 0 {
		liquidityID = accountIDByType(bypass, "asset_cash")
	}
	recvID := partnerAccountID(bypass, partnerID, true, accountIDByType(bypass, "asset_receivable"))
	payAcctID := partnerAccountID(bypass, partnerID, false, accountIDByType(bypass, "liability_payable"))

	name := orm.AsString(pay["name"])
	if name == "" {
		name = nextDocName(bypass, "account.payment", "PAY")
		if err := orm.UpdateRecordByID(bypass, "account.payment", paymentID, map[string]interface{}{"name": name}); err != nil {
			return err
		}
	}
	memo := orm.AsString(pay["memo"])
	if memo == "" {
		memo = name
	}
	date := orm.AsString(pay["date"])
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	date = normalizeDate(date)

	moveModel, err := modelOrErr("account.move")
	if err != nil {
		return err
	}
	moveID, err := orm.Create(bypass, moveModel, map[string]interface{}{
		"name":            name,
		"move_type":       "entry",
		"partner_id":      partnerID,
		"journal_id":      journalID,
		"date":            date,
		"state":           "posted",
		"amount_total":    amount,
		"amount_residual": 0,
		"payment_state":   "paid",
		"narration":       memo,
		"ref":             name,
	})
	if err != nil {
		return err
	}

	var payLineID int
	switch paymentType {
	case "outbound":
		payLineID, err = createEntryLineID(bypass, moveID, payAcctID, partnerID, memo, amount, 0, amount)
		if err != nil {
			return err
		}
		if err := createEntryLine(bypass, moveID, liquidityID, partnerID, memo, 0, amount, 0); err != nil {
			return err
		}
	default:
		if err := createEntryLine(bypass, moveID, liquidityID, partnerID, memo, amount, 0, 0); err != nil {
			return err
		}
		payLineID, err = createEntryLineID(bypass, moveID, recvID, partnerID, memo, 0, amount, -amount)
		if err != nil {
			return err
		}
	}

	if err := orm.UpdateRecordByID(bypass, "account.payment", paymentID, map[string]interface{}{
		"state":      "posted",
		"move_id":    moveID,
		"journal_id": journalID,
		"date":       date,
	}); err != nil {
		return err
	}
	invoiceID, _ := orm.CoerceInt64(pay["invoice_id"])
	if invoiceID > 0 && payLineID > 0 {
		if err := reconcilePaymentToInvoice(bypass, int(invoiceID), payLineID, amount); err != nil {
			return err
		}
	}
	return nil
}

func createEntryLineID(ctx context.Context, moveID int, accountID, partnerID int64, name string, debit, credit, residual float64) (int, error) {
	if accountID <= 0 {
		return 0, fmt.Errorf("missing account")
	}
	lineModel, err := modelOrErr("account.move.line")
	if err != nil {
		return 0, err
	}
	id, err := orm.Create(ctx, lineModel, map[string]interface{}{
		"move_id":         moveID,
		"account_id":      accountID,
		"partner_id":      partnerID,
		"name":            name,
		"display_type":    "entry",
		"debit":           round2(debit),
		"credit":          round2(credit),
		"balance":         round2(debit - credit),
		"amount_residual": round2(residual),
		"reconciled":      false,
	})
	return id, err
}

func liquidityAccountID(ctx context.Context, journalID int64) int64 {
	if journalID <= 0 {
		return 0
	}
	j, err := orm.SearchOne(ctx, "account.journal", map[string]interface{}{"id": journalID})
	if err != nil {
		return 0
	}
	if id, ok := orm.CoerceInt64(j["default_account_id"]); ok && id > 0 {
		return id
	}
	return 0
}
