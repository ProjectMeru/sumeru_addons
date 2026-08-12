package services

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"sumeru/core/orm"
)

const balanceEpsilon = 0.005

func modelOrErr(name string) (orm.Model, error) {
	m, ok := orm.Registry[name]
	if !ok {
		return nil, fmt.Errorf("model %s not registered", name)
	}
	return m, nil
}

func nextDocName(ctx context.Context, code, fallbackPrefix string) string {
	if name, err := orm.NextSequence(ctx, code); err == nil && name != "" {
		return name
	}
	return fmt.Sprintf("%s/%s/%05d", fallbackPrefix, time.Now().Format("2006"), time.Now().Unix()%100000)
}

// CreateCustomerInvoiceFromSale creates a draft out_invoice from a confirmed sale.order with product lines.
func CreateCustomerInvoiceFromSale(ctx context.Context, orderID int) (int, error) {
	if orderID <= 0 {
		return 0, fmt.Errorf("invalid sale order id")
	}
	bypass := orm.ContextWithBypass(ctx, true)
	order, err := orm.SearchOne(bypass, "sale.order", map[string]interface{}{"id": orderID})
	if err != nil {
		return 0, err
	}
	if orm.AsString(order["state"]) != "sale" {
		return 0, fmt.Errorf("sale order must be confirmed")
	}
	origin := orm.AsString(order["name"])
	if origin != "" {
		existing, _ := orm.Search(bypass, "account.move", [][]interface{}{
			{"invoice_origin", "=", origin},
			{"move_type", "=", "out_invoice"},
		})
		if len(existing) > 0 {
			id, _ := orm.CoerceInt64(existing[0]["id"])
			return int(id), nil
		}
	}
	partnerID, _ := orm.CoerceInt64(order["partner_id"])
	soLines, _ := orm.Search(bypass, "sale.order.line", [][]interface{}{
		{"order_id", "=", orderID},
	})
	untaxed := 0.0
	for _, ln := range soLines {
		sub := numericFloat(ln["price_subtotal"])
		if sub <= 0 {
			sub = numericFloat(ln["product_uom_qty"]) * numericFloat(ln["price_unit"])
		}
		untaxed += sub
	}
	if untaxed <= 0 {
		untaxed = numericFloat(order["amount_total"])
	}
	journalID := journalIDByType(bypass, "sale")
	incomeID := accountIDByType(bypass, "income")
	defaultTax := defaultTaxID(bypass, "sale")
	name := nextDocName(bypass, "account.move.out_invoice", "INV")
	today := time.Now().Format("2006-01-02")
	moveModel, err := modelOrErr("account.move")
	if err != nil {
		return 0, err
	}
	moveID, err := orm.Create(bypass, moveModel, map[string]interface{}{
		"name":             name,
		"move_type":        "out_invoice",
		"partner_id":       partnerID,
		"journal_id":       journalID,
		"date":             today,
		"invoice_date":     today,
		"state":            "draft",
		"invoice_origin":   origin,
		"amount_untaxed":   untaxed,
		"amount_tax":       0,
		"amount_total":     untaxed,
		"amount_residual":  untaxed,
		"payment_state":    "not_paid",
		"narration":        fmt.Sprintf("Invoice for %s", origin),
	})
	if err != nil {
		return 0, err
	}
	lineModel, err := modelOrErr("account.move.line")
	if err != nil {
		return 0, err
	}
	for _, ln := range soLines {
		qty := numericFloat(ln["product_uom_qty"])
		if qty <= 0 {
			qty = 1
		}
		price := numericFloat(ln["price_unit"])
		sub := numericFloat(ln["price_subtotal"])
		if sub <= 0 {
			sub = qty * price
		}
		productID, _ := orm.CoerceInt64(ln["product_id"])
		label := orm.AsString(ln["name"])
		acct := productAccountID(bypass, productID, true, incomeID)
		vals := map[string]interface{}{
			"move_id":        moveID,
			"account_id":     acct,
			"product_id":     productID,
			"name":           label,
			"partner_id":     partnerID,
			"quantity":       qty,
			"price_unit":     price,
			"price_subtotal": sub,
			"display_type":   "product",
			"debit":          0,
			"credit":         0,
			"balance":        0,
		}
		if defaultTax > 0 {
			vals["tax_id"] = defaultTax
		}
		_, _ = orm.Create(bypass, lineModel, vals)
	}
	_ = recomputeMoveAmounts(bypass, moveID)
	_ = orm.UpdateRecordByID(bypass, "sale.order", orderID, map[string]interface{}{
		"invoice_status": "invoiced",
	})
	return moveID, nil
}

// CreateVendorBillFromPurchase creates a draft in_invoice from a confirmed purchase.order.
func CreateVendorBillFromPurchase(ctx context.Context, poID int) (int, error) {
	if poID <= 0 {
		return 0, fmt.Errorf("invalid purchase order id")
	}
	bypass := orm.ContextWithBypass(ctx, true)
	po, err := orm.SearchOne(bypass, "purchase.order", map[string]interface{}{"id": poID})
	if err != nil {
		return 0, err
	}
	if orm.AsString(po["state"]) != "purchase" {
		return 0, fmt.Errorf("purchase order must be confirmed")
	}
	origin := orm.AsString(po["name"])
	if origin != "" {
		existing, _ := orm.Search(bypass, "account.move", [][]interface{}{
			{"invoice_origin", "=", origin},
			{"move_type", "=", "in_invoice"},
		})
		if len(existing) > 0 {
			id, _ := orm.CoerceInt64(existing[0]["id"])
			return int(id), nil
		}
	}
	partnerID, _ := orm.CoerceInt64(po["partner_id"])
	poLines, _ := orm.Search(bypass, "purchase.order.line", [][]interface{}{
		{"order_id", "=", poID},
	})
	untaxed := 0.0
	for _, ln := range poLines {
		sub := numericFloat(ln["price_subtotal"])
		if sub <= 0 {
			sub = numericFloat(ln["product_qty"]) * numericFloat(ln["price_unit"])
		}
		untaxed += sub
	}
	if untaxed <= 0 {
		untaxed = numericFloat(po["amount_total"])
	}
	journalID := journalIDByType(bypass, "purchase")
	expenseID := accountIDByType(bypass, "expense")
	defaultTax := defaultTaxID(bypass, "purchase")
	name := nextDocName(bypass, "account.move.in_invoice", "BILL")
	today := time.Now().Format("2006-01-02")
	moveModel, err := modelOrErr("account.move")
	if err != nil {
		return 0, err
	}
	moveID, err := orm.Create(bypass, moveModel, map[string]interface{}{
		"name":            name,
		"move_type":       "in_invoice",
		"partner_id":      partnerID,
		"journal_id":      journalID,
		"date":            today,
		"invoice_date":    today,
		"state":           "draft",
		"invoice_origin":  origin,
		"amount_untaxed":  untaxed,
		"amount_tax":      0,
		"amount_total":    untaxed,
		"amount_residual": untaxed,
		"payment_state":   "not_paid",
		"narration":       fmt.Sprintf("Bill for %s", origin),
	})
	if err != nil {
		return 0, err
	}
	lineModel, err := modelOrErr("account.move.line")
	if err != nil {
		return 0, err
	}
	for _, ln := range poLines {
		qty := numericFloat(ln["product_qty"])
		if qty <= 0 {
			qty = 1
		}
		price := numericFloat(ln["price_unit"])
		sub := numericFloat(ln["price_subtotal"])
		if sub <= 0 {
			sub = qty * price
		}
		productID, _ := orm.CoerceInt64(ln["product_id"])
		label := orm.AsString(ln["name"])
		acct := productAccountID(bypass, productID, false, expenseID)
		vals := map[string]interface{}{
			"move_id":        moveID,
			"account_id":     acct,
			"product_id":     productID,
			"name":           label,
			"partner_id":     partnerID,
			"quantity":       qty,
			"price_unit":     price,
			"price_subtotal": sub,
			"display_type":   "product",
			"debit":          0,
			"credit":         0,
			"balance":        0,
		}
		if defaultTax > 0 {
			vals["tax_id"] = defaultTax
		}
		_, _ = orm.Create(bypass, lineModel, vals)
	}
	_ = recomputeMoveAmounts(bypass, moveID)
	return moveID, nil
}

func productAccountID(ctx context.Context, productID int64, income bool, fallback int64) int64 {
	if productID <= 0 {
		return fallback
	}
	prod, err := orm.SearchOne(ctx, "product.product", map[string]interface{}{"id": productID})
	if err != nil {
		return fallback
	}
	field := "property_account_expense_id"
	if income {
		field = "property_account_income_id"
	}
	if id, ok := orm.CoerceInt64(prod[field]); ok && id > 0 {
		return id
	}
	return fallback
}

func defaultTaxID(ctx context.Context, use string) int64 {
	rows, err := orm.Search(ctx, "account.tax", [][]interface{}{
		{"type_tax_use", "=", use},
		{"active", "=", true},
	})
	if err != nil || len(rows) == 0 {
		return 0
	}
	id, _ := orm.CoerceInt64(rows[0]["id"])
	return id
}

func partnerAccountID(ctx context.Context, partnerID int64, receivable bool, fallback int64) int64 {
	if partnerID <= 0 {
		return fallback
	}
	p, err := orm.SearchOne(ctx, "core.partner", map[string]interface{}{"id": partnerID})
	if err != nil {
		return fallback
	}
	field := "property_account_payable_id"
	if receivable {
		field = "property_account_receivable_id"
	}
	if id, ok := orm.CoerceInt64(p[field]); ok && id > 0 {
		return id
	}
	return fallback
}

func productLines(lines []map[string]interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(lines))
	for _, ln := range lines {
		dt := orm.AsString(ln["display_type"])
		if dt == "" || dt == "product" {
			out = append(out, ln)
		}
	}
	return out
}

func deleteAccountingLines(ctx context.Context, lines []map[string]interface{}) {
	for _, ln := range lines {
		dt := orm.AsString(ln["display_type"])
		if dt != "entry" && dt != "tax" {
			continue
		}
		if lid, ok := orm.CoerceInt64(ln["id"]); ok {
			_ = orm.Unlink(ctx, "account.move.line", int(lid))
		}
	}
}

func taxAmountPercent(ctx context.Context, taxID int64) (float64, int64) {
	if taxID <= 0 {
		return 0, 0
	}
	tax, err := orm.SearchOne(ctx, "account.tax", map[string]interface{}{"id": taxID})
	if err != nil {
		return 0, 0
	}
	acct, _ := orm.CoerceInt64(tax["account_id"])
	if acct <= 0 {
		acct = accountIDByType(ctx, "liability_current")
	}
	return numericFloat(tax["amount"]), acct
}

func recomputeMoveAmounts(ctx context.Context, moveID int) error {
	lines, _ := orm.Search(ctx, "account.move.line", [][]interface{}{{"move_id", "=", moveID}})
	untaxed := 0.0
	taxTotal := 0.0
	for _, ln := range productLines(lines) {
		sub := numericFloat(ln["price_subtotal"])
		if sub <= 0 {
			sub = numericFloat(ln["quantity"]) * numericFloat(ln["price_unit"])
		}
		untaxed += sub
		taxID, _ := orm.CoerceInt64(ln["tax_id"])
		pct, _ := taxAmountPercent(ctx, taxID)
		taxTotal += sub * pct / 100.0
	}
	total := untaxed + taxTotal
	return orm.UpdateRecordByID(ctx, "account.move", moveID, map[string]interface{}{
		"amount_untaxed":  round2(untaxed),
		"amount_tax":      round2(taxTotal),
		"amount_total":    round2(total),
		"amount_residual": round2(total),
	})
}

func applyPaymentTermDue(ctx context.Context, move map[string]interface{}, moveID int) {
	termID, _ := orm.CoerceInt64(move["payment_term_id"])
	invDate := orm.AsString(move["invoice_date"])
	if invDate == "" {
		invDate = orm.AsString(move["date"])
	}
	if invDate == "" {
		invDate = time.Now().Format("2006-01-02")
	}
	days := 0
	if termID > 0 {
		term, err := orm.SearchOne(ctx, "account.payment.term", map[string]interface{}{"id": termID})
		if err == nil {
			if d, ok := orm.CoerceInt64(term["days"]); ok {
				days = int(d)
			}
		}
	}
	t, err := time.Parse("2006-01-02", invDate)
	if err != nil {
		t = time.Now()
	}
	due := t.AddDate(0, 0, days).Format("2006-01-02")
	_ = orm.UpdateRecordByID(ctx, "account.move", moveID, map[string]interface{}{
		"invoice_date":     invDate,
		"invoice_date_due": due,
	})
}

// PostMove writes balanced journal entry lines (AR/AP + income/expense + tax) and marks posted.
func PostMove(ctx context.Context, moveID int) error {
	if moveID <= 0 {
		return fmt.Errorf("invalid move id")
	}
	bypass := orm.ContextWithBypass(ctx, true)
	move, err := orm.SearchOne(bypass, "account.move", map[string]interface{}{"id": moveID})
	if err != nil {
		return err
	}
	if orm.AsString(move["state"]) == "cancel" {
		return fmt.Errorf("cannot post cancelled move")
	}
	allLines, _ := orm.Search(bypass, "account.move.line", [][]interface{}{
		{"move_id", "=", moveID},
	})
	moveType := orm.AsString(move["move_type"])

	if moveType == "entry" {
		return postManualEntry(bypass, moveID, move, allLines)
	}

	_ = recomputeMoveAmounts(bypass, moveID)
	move, _ = orm.SearchOne(bypass, "account.move", map[string]interface{}{"id": moveID})
	applyPaymentTermDue(bypass, move, moveID)

	prods := productLines(allLines)
	untaxed := 0.0
	taxByAcct := map[int64]float64{}
	taxByTaxID := map[int64]float64{}
	incomeExpenseByAcct := map[int64]float64{}
	for _, ln := range prods {
		sub := numericFloat(ln["price_subtotal"])
		if sub <= 0 {
			sub = numericFloat(ln["quantity"]) * numericFloat(ln["price_unit"])
		}
		untaxed += sub
		acct, _ := orm.CoerceInt64(ln["account_id"])
		if acct <= 0 {
			if moveType == "out_invoice" || moveType == "out_refund" {
				acct = accountIDByType(bypass, "income")
			} else {
				acct = accountIDByType(bypass, "expense")
			}
		}
		incomeExpenseByAcct[acct] += sub
		taxID, _ := orm.CoerceInt64(ln["tax_id"])
		pct, taxAcct := taxAmountPercent(bypass, taxID)
		if pct != 0 && taxAcct > 0 {
			amt := sub * pct / 100.0
			taxByAcct[taxAcct] += amt
			taxByTaxID[taxID] += amt
		}
	}
	taxTotal := 0.0
	for _, a := range taxByAcct {
		taxTotal += a
	}
	total := untaxed + taxTotal
	if total <= 0 {
		total = numericFloat(move["amount_total"])
		untaxed = total
	}
	if total <= 0 {
		return fmt.Errorf("amount_total must be positive to post")
	}

	partnerID, _ := orm.CoerceInt64(move["partner_id"])
	label := orm.AsString(move["name"])
	deleteAccountingLines(bypass, allLines)

	recvFallback := accountIDByType(bypass, "asset_receivable")
	payFallback := accountIDByType(bypass, "liability_payable")
	recvID := partnerAccountID(bypass, partnerID, true, recvFallback)
	payID := partnerAccountID(bypass, partnerID, false, payFallback)

	switch moveType {
	case "out_invoice":
		if err := createEntryLine(bypass, moveID, recvID, partnerID, label, total, 0, total); err != nil {
			return err
		}
		for acct, amt := range incomeExpenseByAcct {
			if err := createEntryLine(bypass, moveID, acct, partnerID, label, 0, amt, 0); err != nil {
				return err
			}
		}
		for taxID, amt := range taxByTaxID {
			_, taxAcct := taxAmountPercent(bypass, taxID)
			if err := createTaxLine(bypass, moveID, taxAcct, partnerID, taxID, label+" Tax", 0, amt); err != nil {
				return err
			}
		}
	case "out_refund":
		if err := createEntryLine(bypass, moveID, recvID, partnerID, label, 0, total, -total); err != nil {
			return err
		}
		for acct, amt := range incomeExpenseByAcct {
			if err := createEntryLine(bypass, moveID, acct, partnerID, label, amt, 0, 0); err != nil {
				return err
			}
		}
		for taxID, amt := range taxByTaxID {
			_, taxAcct := taxAmountPercent(bypass, taxID)
			if err := createTaxLine(bypass, moveID, taxAcct, partnerID, taxID, label+" Tax", amt, 0); err != nil {
				return err
			}
		}
	case "in_invoice":
		if err := createEntryLine(bypass, moveID, payID, partnerID, label, 0, total, -total); err != nil {
			return err
		}
		for acct, amt := range incomeExpenseByAcct {
			if err := createEntryLine(bypass, moveID, acct, partnerID, label, amt, 0, 0); err != nil {
				return err
			}
		}
		for taxID, amt := range taxByTaxID {
			_, taxAcct := taxAmountPercent(bypass, taxID)
			if err := createTaxLine(bypass, moveID, taxAcct, partnerID, taxID, label+" Tax", amt, 0); err != nil {
				return err
			}
		}
	case "in_refund":
		if err := createEntryLine(bypass, moveID, payID, partnerID, label, total, 0, total); err != nil {
			return err
		}
		for acct, amt := range incomeExpenseByAcct {
			if err := createEntryLine(bypass, moveID, acct, partnerID, label, 0, amt, 0); err != nil {
				return err
			}
		}
		for taxID, amt := range taxByTaxID {
			_, taxAcct := taxAmountPercent(bypass, taxID)
			if err := createTaxLine(bypass, moveID, taxAcct, partnerID, taxID, label+" Tax", 0, amt); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("posting move_type %q not supported", moveType)
	}

	postedLines, _ := orm.Search(bypass, "account.move.line", [][]interface{}{{"move_id", "=", moveID}})
	if err := assertBalanced(postedLines); err != nil {
		return err
	}
	_ = upsertInvoiceReport(bypass, moveID)
	return orm.UpdateRecordByID(bypass, "account.move", moveID, map[string]interface{}{
		"state":           "posted",
		"amount_untaxed":  round2(untaxed),
		"amount_tax":      round2(taxTotal),
		"amount_total":    round2(total),
		"amount_residual": round2(total),
		"payment_state":   "not_paid",
	})
}

func postManualEntry(ctx context.Context, moveID int, move map[string]interface{}, lines []map[string]interface{}) error {
	if len(lines) == 0 {
		return fmt.Errorf("manual entry requires journal items")
	}
	if err := assertBalanced(lines); err != nil {
		return err
	}
	total := 0.0
	for _, ln := range lines {
		total += numericFloat(ln["debit"])
	}
	return orm.UpdateRecordByID(ctx, "account.move", moveID, map[string]interface{}{
		"state":           "posted",
		"amount_total":    round2(total),
		"amount_residual": 0,
		"payment_state":   "paid",
	})
}

func assertBalanced(lines []map[string]interface{}) error {
	debit, credit := 0.0, 0.0
	for _, ln := range lines {
		dt := orm.AsString(ln["display_type"])
		if dt == "product" || dt == "line_section" || dt == "line_note" {
			continue
		}
		debit += numericFloat(ln["debit"])
		credit += numericFloat(ln["credit"])
	}
	if math.Abs(debit-credit) > balanceEpsilon {
		return fmt.Errorf("unbalanced entry: debit=%.2f credit=%.2f", debit, credit)
	}
	return nil
}

func createEntryLine(ctx context.Context, moveID int, accountID, partnerID int64, name string, debit, credit, residual float64) error {
	if accountID <= 0 {
		return fmt.Errorf("missing chart account for posting")
	}
	lineModel, err := modelOrErr("account.move.line")
	if err != nil {
		return err
	}
	_, err = orm.Create(ctx, lineModel, map[string]interface{}{
		"move_id":         moveID,
		"account_id":      accountID,
		"partner_id":      partnerID,
		"name":            name,
		"display_type":    "entry",
		"debit":           round2(debit),
		"credit":          round2(credit),
		"balance":         round2(debit - credit),
		"amount_residual": round2(residual),
		"reconciled":      residual == 0 && (debit > 0 || credit > 0),
	})
	return err
}

func createTaxLine(ctx context.Context, moveID int, accountID, partnerID, taxID int64, name string, debit, credit float64) error {
	if accountID <= 0 {
		return fmt.Errorf("missing tax account for posting")
	}
	lineModel, err := modelOrErr("account.move.line")
	if err != nil {
		return err
	}
	_, err = orm.Create(ctx, lineModel, map[string]interface{}{
		"move_id":      moveID,
		"account_id":   accountID,
		"partner_id":   partnerID,
		"tax_line_id":  taxID,
		"name":         name,
		"display_type": "tax",
		"debit":        round2(debit),
		"credit":       round2(credit),
		"balance":      round2(debit - credit),
	})
	return err
}

func journalIDByType(ctx context.Context, typ string) int64 {
	rows, err := orm.Search(ctx, "account.journal", [][]interface{}{
		{"type", "=", typ},
		{"active", "=", true},
	})
	if err != nil || len(rows) == 0 {
		return 0
	}
	id, _ := orm.CoerceInt64(rows[0]["id"])
	return id
}

func accountIDByType(ctx context.Context, typ string) int64 {
	rows, err := orm.Search(ctx, "account.account", [][]interface{}{
		{"account_type", "=", typ},
		{"active", "=", true},
	})
	if err != nil || len(rows) == 0 {
		return 0
	}
	id, _ := orm.CoerceInt64(rows[0]["id"])
	return id
}

func numericFloat(v interface{}) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int64:
		return float64(t)
	case int:
		return float64(t)
	case int32:
		return float64(t)
	case string:
		var f float64
		_, _ = fmt.Sscanf(t, "%f", &f)
		return f
	case []byte:
		var f float64
		_, _ = fmt.Sscanf(string(t), "%f", &f)
		return f
	default:
		s := strings.TrimSpace(fmt.Sprint(v))
		if s == "" || s == "<nil>" {
			return 0
		}
		var f float64
		_, _ = fmt.Sscanf(s, "%f", &f)
		return f
	}
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
