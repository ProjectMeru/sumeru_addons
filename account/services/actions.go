package services

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"sumeru/core/orm"
)

func init() {
	orm.RegisterObjectAction("account.move", "action_post", actionPostMove)
	orm.RegisterObjectAction("account.move", "action_draft", actionDraftMove)
	orm.RegisterObjectAction("account.move", "action_cancel", actionCancelMove)
	orm.RegisterObjectAction("account.move", "action_register_payment", actionRegisterPayment)
	orm.RegisterObjectAction("account.move", "action_reverse", actionReverseWizard)
	orm.RegisterObjectAction("account.move", "action_print_invoice", actionPrintInvoice)
	orm.RegisterObjectAction("account.payment", "action_post", actionPostPayment)
	orm.RegisterObjectAction("account.payment.register", "action_create_payments", actionCreatePayments)
	orm.RegisterObjectAction("account.move.reversal", "action_reverse_moves", actionReverseMoves)
}

func actionPostMove(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if err := PostMove(ctx, id); err != nil {
		return "", err
	}
	return vals["next"], nil
}

func actionDraftMove(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if err := ResetMoveToDraft(ctx, id); err != nil {
		return "", err
	}
	return vals["next"], nil
}

func actionCancelMove(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if err := CancelMove(ctx, id); err != nil {
		return "", err
	}
	return vals["next"], nil
}

func actionPostPayment(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if err := PostPayment(ctx, id); err != nil {
		return "", err
	}
	return vals["next"], nil
}

func actionPrintInvoice(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	return fmt.Sprintf("/account/invoice/print?id=%d", id), nil
}

func actionRegisterPayment(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	bypass := orm.ContextWithBypass(ctx, true)
	move, err := orm.SearchOne(bypass, "account.move", map[string]interface{}{"id": id})
	if err != nil {
		return "", err
	}
	if orm.AsString(move["state"]) != "posted" {
		return "", fmt.Errorf("invoice not posted")
	}
	mt := orm.AsString(move["move_type"])
	paymentType := "inbound"
	if mt == "in_invoice" || mt == "out_refund" {
		paymentType = "outbound"
	}
	amount := numericFloat(move["amount_residual"])
	if amount <= 0 {
		amount = numericFloat(move["amount_total"])
	}
	partnerID, _ := orm.CoerceInt64(move["partner_id"])
	journalType := "bank"
	if paymentType == "outbound" {
		journalType = "cash"
	}
	journalID := journalIDByType(bypass, journalType)
	if journalID <= 0 {
		journalID = journalIDByType(bypass, "bank")
	}
	wizModel, err := modelOrErr("account.payment.register")
	if err != nil {
		return "", err
	}
	wizID, err := orm.Create(bypass, wizModel, map[string]interface{}{
		"invoice_id":    id,
		"partner_id":    partnerID,
		"amount":        amount,
		"journal_id":    journalID,
		"payment_date":  time.Now().Format("2006-01-02"),
		"communication": orm.AsString(move["name"]),
		"payment_type":  paymentType,
	})
	if err != nil {
		return "", err
	}
	return wizardFormURL("account.payment.register", wizID, vals["next"]), nil
}

func actionReverseWizard(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	bypass := orm.ContextWithBypass(ctx, true)
	move, err := orm.SearchOne(bypass, "account.move", map[string]interface{}{"id": id})
	if err != nil {
		return "", err
	}
	if orm.AsString(move["state"]) != "posted" {
		return "", fmt.Errorf("move not posted")
	}
	wizModel, err := modelOrErr("account.move.reversal")
	if err != nil {
		return "", err
	}
	wizID, err := orm.Create(bypass, wizModel, map[string]interface{}{
		"move_id": id,
		"date":    time.Now().Format("2006-01-02"),
		"reason":  "Credit note",
	})
	if err != nil {
		return "", err
	}
	return wizardFormURL("account.move.reversal", wizID, vals["next"]), nil
}

func actionCreatePayments(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	bypass := orm.ContextWithBypass(ctx, true)
	upd := map[string]interface{}{}
	if v := strings.TrimSpace(vals["amount"]); v != "" {
		upd["amount"] = numericFloat(v)
	}
	if v := strings.TrimSpace(vals["journal_id"]); v != "" {
		if jid, err := strconv.Atoi(v); err == nil && jid > 0 {
			upd["journal_id"] = jid
		}
	}
	if v := strings.TrimSpace(vals["payment_date"]); v != "" {
		upd["payment_date"] = v
	}
	if v := strings.TrimSpace(vals["communication"]); v != "" {
		upd["communication"] = v
	}
	if v := strings.TrimSpace(vals["payment_type"]); v != "" {
		upd["payment_type"] = v
	}
	if len(upd) > 0 {
		if err := orm.UpdateRecordByID(bypass, "account.payment.register", id, upd); err != nil {
			return "", err
		}
	}
	wiz, err := orm.SearchOne(bypass, "account.payment.register", map[string]interface{}{"id": id})
	if err != nil {
		return "", err
	}
	invoiceID, _ := orm.CoerceInt64(wiz["invoice_id"])
	amount := numericFloat(wiz["amount"])
	if invoiceID <= 0 || amount <= 0 {
		return "", fmt.Errorf("missing fields")
	}
	partnerID, _ := orm.CoerceInt64(wiz["partner_id"])
	journalID, _ := orm.CoerceInt64(wiz["journal_id"])
	paymentType := orm.AsString(wiz["payment_type"])
	if paymentType == "" {
		paymentType = "inbound"
	}
	partnerType := "customer"
	if paymentType == "outbound" {
		partnerType = "supplier"
	}
	date := orm.AsString(wiz["payment_date"])
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	date = normalizeDate(date)
	payModel, err := modelOrErr("account.payment")
	if err != nil {
		return "", err
	}
	name := nextDocName(bypass, "account.payment", "PAY")
	payID, err := orm.Create(bypass, payModel, map[string]interface{}{
		"name":         name,
		"payment_type": paymentType,
		"partner_type": partnerType,
		"partner_id":   partnerID,
		"amount":       amount,
		"date":         date,
		"journal_id":   journalID,
		"memo":         orm.AsString(wiz["communication"]),
		"state":        "draft",
		"invoice_id":   invoiceID,
	})
	if err != nil {
		return "", err
	}
	if err := PostPayment(ctx, payID); err != nil {
		return "", err
	}
	return moveFormURL(int(invoiceID)), nil
}

func actionReverseMoves(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	bypass := orm.ContextWithBypass(ctx, true)
	wiz, err := orm.SearchOne(bypass, "account.move.reversal", map[string]interface{}{"id": id})
	if err != nil {
		return "", err
	}
	moveID, _ := orm.CoerceInt64(wiz["move_id"])
	if moveID <= 0 {
		return "", fmt.Errorf("missing move")
	}
	src, err := orm.SearchOne(bypass, "account.move", map[string]interface{}{"id": moveID})
	if err != nil {
		return "", err
	}
	srcType := orm.AsString(src["move_type"])
	revType := map[string]string{
		"out_invoice": "out_refund",
		"out_refund":  "out_invoice",
		"in_invoice":  "in_refund",
		"in_refund":   "in_invoice",
		"entry":       "entry",
	}[srcType]
	if revType == "" {
		return "", fmt.Errorf("unsupported move_type %s", srcType)
	}
	date := orm.AsString(wiz["date"])
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	reason := orm.AsString(wiz["reason"])
	seqCode := "account.move.out_refund"
	prefix := "RINV"
	if revType == "in_refund" {
		seqCode = "account.move.in_refund"
		prefix = "RBILL"
	} else if revType == "out_invoice" {
		seqCode = "account.move.out_invoice"
		prefix = "INV"
	} else if revType == "in_invoice" {
		seqCode = "account.move.in_invoice"
		prefix = "BILL"
	}
	name := nextDocName(bypass, seqCode, prefix)
	moveModel, err := modelOrErr("account.move")
	if err != nil {
		return "", err
	}
	partnerID, _ := orm.CoerceInt64(src["partner_id"])
	journalID, _ := orm.CoerceInt64(src["journal_id"])
	newID, err := orm.Create(bypass, moveModel, map[string]interface{}{
		"name":              name,
		"move_type":         revType,
		"partner_id":        partnerID,
		"journal_id":        journalID,
		"date":              date,
		"invoice_date":      date,
		"state":             "draft",
		"invoice_origin":    orm.AsString(src["name"]),
		"ref":               reason,
		"reversed_entry_id": moveID,
		"amount_untaxed":    src["amount_untaxed"],
		"amount_tax":        src["amount_tax"],
		"amount_total":      src["amount_total"],
		"amount_residual":   src["amount_total"],
		"payment_state":     "not_paid",
		"payment_term_id":   src["payment_term_id"],
		"narration":         reason,
	})
	if err != nil {
		return "", err
	}
	srcLines, _ := orm.Search(bypass, "account.move.line", [][]interface{}{
		{"move_id", "=", moveID},
		{"display_type", "=", "product"},
	})
	lineModel, err := modelOrErr("account.move.line")
	if err != nil {
		return "", err
	}
	for _, ln := range srcLines {
		if _, err := orm.Create(bypass, lineModel, map[string]interface{}{
			"move_id":        newID,
			"account_id":     ln["account_id"],
			"product_id":     ln["product_id"],
			"tax_id":         ln["tax_id"],
			"name":           ln["name"],
			"partner_id":     partnerID,
			"quantity":       ln["quantity"],
			"price_unit":     ln["price_unit"],
			"price_subtotal": ln["price_subtotal"],
			"display_type":   "product",
		}); err != nil {
			return "", err
		}
	}
	if err := PostMove(ctx, newID); err != nil {
		return "", err
	}
	return moveFormURL(newID), nil
}

func wizardFormURL(model string, id int, next string) string {
	actionXML := ""
	switch model {
	case "account.payment.register":
		actionXML = "account.action_payment_register_wizard"
	case "account.move.reversal":
		actionXML = "account.action_move_reversal_wizard"
	}
	q := url.Values{}
	if actionXML != "" {
		q.Set("action", actionXML)
	}
	q.Set("view_type", "form")
	q.Set("id", strconv.Itoa(id))
	q.Set("edit", "1")
	if next != "" {
		q.Set("next", next)
	}
	return "/web?" + q.Encode()
}

func moveFormURL(id int) string {
	q := url.Values{}
	q.Set("action", "account.action_account_moves_out")
	q.Set("view_type", "form")
	q.Set("id", strconv.Itoa(id))
	return "/web?" + q.Encode()
}
