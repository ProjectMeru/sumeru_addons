package account

import (
	"context"
	"fmt"
	"time"

	"sumeru/core/orm"
)

// PostPayment posts a draft payment: liquidity journal entry + invoice payment_state.
func PostPayment(ctx context.Context, paymentID int) error {
	if paymentID <= 0 {
		return fmt.Errorf("invalid payment id")
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
		return fmt.Errorf("cannot post cancelled payment")
	}
	amount := numericFloat(pay["amount"])
	if amount <= 0 {
		return fmt.Errorf("payment amount must be positive")
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
	recvID := accountIDByType(bypass, "asset_receivable")
	payID := accountIDByType(bypass, "liability_payable")

	name := orm.AsString(pay["name"])
	if name == "" {
		name = nextDocName(bypass, "account.payment", "PAY")
		_ = orm.UpdateRecordByID(bypass, "account.payment", paymentID, map[string]interface{}{"name": name})
	}
	memo := orm.AsString(pay["memo"])
	if memo == "" {
		memo = name
	}
	date := orm.AsString(pay["date"])
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	moveModel, err := modelOrErr("account.move")
	if err != nil {
		return err
	}
	moveID, err := orm.Create(bypass, moveModel, map[string]interface{}{
		"name":          name,
		"move_type":     "entry",
		"partner_id":    partnerID,
		"journal_id":    journalID,
		"date":          date,
		"state":         "posted",
		"amount_total":  amount,
		"payment_state": "paid",
		"narration":     memo,
		"ref":           name,
	})
	if err != nil {
		return err
	}

	var debitAcct, creditAcct int64
	switch paymentType {
	case "outbound":
		debitAcct, creditAcct = payID, liquidityID
	default: // inbound
		debitAcct, creditAcct = liquidityID, recvID
	}
	if err := createEntryLine(bypass, moveID, debitAcct, partnerID, memo, amount, 0); err != nil {
		return err
	}
	if err := createEntryLine(bypass, moveID, creditAcct, partnerID, memo, 0, amount); err != nil {
		return err
	}

	_ = orm.UpdateRecordByID(bypass, "account.payment", paymentID, map[string]interface{}{
		"state":      "posted",
		"move_id":    moveID,
		"journal_id": journalID,
		"date":       date,
	})

	invoiceID, _ := orm.CoerceInt64(pay["invoice_id"])
	if invoiceID > 0 {
		_ = applyPaymentToInvoice(bypass, int(invoiceID), amount)
	}
	return nil
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

func applyPaymentToInvoice(ctx context.Context, invoiceID int, amount float64) error {
	inv, err := orm.SearchOne(ctx, "account.move", map[string]interface{}{"id": invoiceID})
	if err != nil {
		return err
	}
	total := numericFloat(inv["amount_total"])
	paid := amount
	// Sum other posted payments on this invoice.
	pays, _ := orm.Search(ctx, "account.payment", [][]interface{}{
		{"invoice_id", "=", invoiceID},
		{"state", "=", "posted"},
	})
	paid = 0
	for _, p := range pays {
		paid += numericFloat(p["amount"])
	}
	state := "partial"
	if paid+0.0001 >= total && total > 0 {
		state = "paid"
	} else if paid <= 0 {
		state = "not_paid"
	}
	return orm.UpdateRecordByID(ctx, "account.move", invoiceID, map[string]interface{}{
		"payment_state": state,
	})
}
