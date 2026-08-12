package account

import (
	"context"
	"fmt"
	"time"

	"sumeru/core/orm"
)

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
	total := 0.0
	for _, ln := range soLines {
		sub := numericFloat(ln["price_subtotal"])
		if sub <= 0 {
			sub = numericFloat(ln["product_uom_qty"]) * numericFloat(ln["price_unit"])
		}
		total += sub
	}
	if total <= 0 {
		total = numericFloat(order["amount_total"])
	}
	journalID := journalIDByType(bypass, "sale")
	incomeID := accountIDByType(bypass, "income")
	name := nextDocName(bypass, "account.move.out_invoice", "INV")
	moveModel, err := modelOrErr("account.move")
	if err != nil {
		return 0, err
	}
	moveID, err := orm.Create(bypass, moveModel, map[string]interface{}{
		"name":           name,
		"move_type":      "out_invoice",
		"partner_id":     partnerID,
		"journal_id":     journalID,
		"date":           time.Now().Format("2006-01-02"),
		"state":          "draft",
		"invoice_origin": origin,
		"amount_total":   total,
		"payment_state":  "not_paid",
		"narration":      fmt.Sprintf("Invoice for %s", origin),
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
		_, _ = orm.Create(bypass, lineModel, map[string]interface{}{
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
		})
	}
	_ = orm.UpdateRecordByID(bypass, "sale.order", orderID, map[string]interface{}{
		"invoice_status": "invoiced",
	})
	return moveID, nil
}

// CreateVendorBillFromPurchase creates a draft in_invoice from a confirmed purchase.order with product lines.
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
	total := 0.0
	for _, ln := range poLines {
		sub := numericFloat(ln["price_subtotal"])
		if sub <= 0 {
			sub = numericFloat(ln["product_qty"]) * numericFloat(ln["price_unit"])
		}
		total += sub
	}
	if total <= 0 {
		total = numericFloat(po["amount_total"])
	}
	journalID := journalIDByType(bypass, "purchase")
	expenseID := accountIDByType(bypass, "expense")
	name := nextDocName(bypass, "account.move.in_invoice", "BILL")
	moveModel, err := modelOrErr("account.move")
	if err != nil {
		return 0, err
	}
	moveID, err := orm.Create(bypass, moveModel, map[string]interface{}{
		"name":           name,
		"move_type":      "in_invoice",
		"partner_id":     partnerID,
		"journal_id":     journalID,
		"date":           time.Now().Format("2006-01-02"),
		"state":          "draft",
		"invoice_origin": origin,
		"amount_total":   total,
		"payment_state":  "not_paid",
		"narration":      fmt.Sprintf("Bill for %s", origin),
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
		_, _ = orm.Create(bypass, lineModel, map[string]interface{}{
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
		})
	}
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

func productLinesTotal(lines []map[string]interface{}) float64 {
	total := 0.0
	for _, ln := range lines {
		dt := orm.AsString(ln["display_type"])
		if dt != "" && dt != "product" {
			continue
		}
		// Skip pure accounting entry lines (have debit/credit, no product subtotal intent).
		if dt == "entry" {
			continue
		}
		sub := numericFloat(ln["price_subtotal"])
		if sub <= 0 {
			sub = numericFloat(ln["quantity"]) * numericFloat(ln["price_unit"])
		}
		total += sub
	}
	return total
}

func entryLinesExist(lines []map[string]interface{}) bool {
	for _, ln := range lines {
		if orm.AsString(ln["display_type"]) == "entry" {
			return true
		}
	}
	return false
}

func deleteEntryLines(ctx context.Context, lines []map[string]interface{}) {
	for _, ln := range lines {
		if orm.AsString(ln["display_type"]) != "entry" {
			continue
		}
		if lid, ok := orm.CoerceInt64(ln["id"]); ok {
			_ = orm.Unlink(ctx, "account.move.line", int(lid))
		}
	}
}

// PostMove writes balanced journal entry lines and marks the move posted.
// Product invoice lines (display_type=product) are preserved.
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
	if orm.AsString(move["state"]) == "posted" && entryLinesExist(allLines) {
		return nil
	}

	total := productLinesTotal(allLines)
	if total <= 0 {
		total = numericFloat(move["amount_total"])
	}
	if total <= 0 {
		return fmt.Errorf("amount_total must be positive to post")
	}
	partnerID, _ := orm.CoerceInt64(move["partner_id"])
	moveType := orm.AsString(move["move_type"])
	label := orm.AsString(move["name"])

	deleteEntryLines(bypass, allLines)

	recvID := accountIDByType(bypass, "asset_receivable")
	payID := accountIDByType(bypass, "liability_payable")
	incomeID := accountIDByType(bypass, "income")
	expenseID := accountIDByType(bypass, "expense")

	switch moveType {
	case "out_invoice":
		if err := createEntryLine(bypass, moveID, recvID, partnerID, label, total, 0); err != nil {
			return err
		}
		if err := createEntryLine(bypass, moveID, incomeID, partnerID, label, 0, total); err != nil {
			return err
		}
	case "out_refund":
		if err := createEntryLine(bypass, moveID, incomeID, partnerID, label, total, 0); err != nil {
			return err
		}
		if err := createEntryLine(bypass, moveID, recvID, partnerID, label, 0, total); err != nil {
			return err
		}
	case "in_invoice":
		if err := createEntryLine(bypass, moveID, expenseID, partnerID, label, total, 0); err != nil {
			return err
		}
		if err := createEntryLine(bypass, moveID, payID, partnerID, label, 0, total); err != nil {
			return err
		}
	case "in_refund":
		if err := createEntryLine(bypass, moveID, payID, partnerID, label, total, 0); err != nil {
			return err
		}
		if err := createEntryLine(bypass, moveID, expenseID, partnerID, label, 0, total); err != nil {
			return err
		}
	default:
		return fmt.Errorf("posting move_type %q not supported in v1", moveType)
	}
	return orm.UpdateRecordByID(bypass, "account.move", moveID, map[string]interface{}{
		"state":        "posted",
		"amount_total": total,
	})
}

func createEntryLine(ctx context.Context, moveID int, accountID, partnerID int64, name string, debit, credit float64) error {
	if accountID <= 0 {
		return fmt.Errorf("missing chart account for posting")
	}
	lineModel, err := modelOrErr("account.move.line")
	if err != nil {
		return err
	}
	_, err = orm.Create(ctx, lineModel, map[string]interface{}{
		"move_id":      moveID,
		"account_id":   accountID,
		"partner_id":   partnerID,
		"name":         name,
		"display_type": "entry",
		"debit":        debit,
		"credit":       credit,
		"balance":      debit - credit,
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
	case string:
		var f float64
		_, _ = fmt.Sscanf(t, "%f", &f)
		return f
	default:
		return 0
	}
}
