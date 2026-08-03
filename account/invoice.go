package account

import (
	"context"
	"fmt"

	"sumeru/core/orm"
)

func modelOrErr(name string) (orm.Model, error) {
	m, ok := orm.Registry[name]
	if !ok {
		return nil, fmt.Errorf("model %s not registered", name)
	}
	return m, nil
}

// CreateCustomerInvoiceFromSale creates a draft out_invoice from a confirmed sale.order.
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
	total := numericFloat(order["amount_total"])
	if total <= 0 {
		lines, _ := orm.Search(bypass, "sale.order.line", [][]interface{}{
			{"order_id", "=", orderID},
		})
		for _, ln := range lines {
			total += numericFloat(ln["price_subtotal"])
		}
	}
	journalID := journalIDByType(bypass, "sale")
	name := fmt.Sprintf("INV/%s", origin)
	if origin == "" {
		name = fmt.Sprintf("INV/SO-%d", orderID)
	}
	moveModel, err := modelOrErr("account.move")
	if err != nil {
		return 0, err
	}
	moveID, err := orm.Create(bypass, moveModel, map[string]interface{}{
		"name":           name,
		"move_type":      "out_invoice",
		"partner_id":     partnerID,
		"journal_id":     journalID,
		"state":          "draft",
		"invoice_origin": origin,
		"amount_total":   total,
		"payment_state":  "not_paid",
		"narration":      fmt.Sprintf("Invoice for %s", origin),
	})
	if err != nil {
		return 0, err
	}
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
	total := numericFloat(po["amount_total"])
	if total <= 0 {
		lines, _ := orm.Search(bypass, "purchase.order.line", [][]interface{}{
			{"order_id", "=", poID},
		})
		for _, ln := range lines {
			total += numericFloat(ln["price_subtotal"])
		}
	}
	journalID := journalIDByType(bypass, "purchase")
	name := fmt.Sprintf("BILL/%s", origin)
	if origin == "" {
		name = fmt.Sprintf("BILL/PO-%d", poID)
	}
	moveModel, err := modelOrErr("account.move")
	if err != nil {
		return 0, err
	}
	return orm.Create(bypass, moveModel, map[string]interface{}{
		"name":           name,
		"move_type":      "in_invoice",
		"partner_id":     partnerID,
		"journal_id":     journalID,
		"state":          "draft",
		"invoice_origin": origin,
		"amount_total":   total,
		"payment_state":  "not_paid",
		"narration":      fmt.Sprintf("Bill for %s", origin),
	})
}

// PostMove writes balanced journal lines and marks the move posted.
func PostMove(ctx context.Context, moveID int) error {
	if moveID <= 0 {
		return fmt.Errorf("invalid move id")
	}
	bypass := orm.ContextWithBypass(ctx, true)
	move, err := orm.SearchOne(bypass, "account.move", map[string]interface{}{"id": moveID})
	if err != nil {
		return err
	}
	if orm.AsString(move["state"]) == "posted" {
		lines, _ := orm.Search(bypass, "account.move.line", [][]interface{}{
			{"move_id", "=", moveID},
		})
		if len(lines) > 0 {
			return nil
		}
	}
	if orm.AsString(move["state"]) == "cancel" {
		return fmt.Errorf("cannot post cancelled move")
	}
	total := numericFloat(move["amount_total"])
	if total <= 0 {
		return fmt.Errorf("amount_total must be positive to post")
	}
	partnerID, _ := orm.CoerceInt64(move["partner_id"])
	moveType := orm.AsString(move["move_type"])
	label := orm.AsString(move["name"])

	oldLines, _ := orm.Search(bypass, "account.move.line", [][]interface{}{
		{"move_id", "=", moveID},
	})
	for _, ln := range oldLines {
		if lid, ok := orm.CoerceInt64(ln["id"]); ok {
			_ = orm.Unlink(bypass, "account.move.line", int(lid))
		}
	}

	recvID := accountIDByType(bypass, "asset_receivable")
	payID := accountIDByType(bypass, "liability_payable")
	incomeID := accountIDByType(bypass, "income")
	expenseID := accountIDByType(bypass, "expense")

	switch moveType {
	case "out_invoice":
		if err := createLine(bypass, moveID, recvID, partnerID, label, total, 0); err != nil {
			return err
		}
		if err := createLine(bypass, moveID, incomeID, partnerID, label, 0, total); err != nil {
			return err
		}
	case "in_invoice":
		if err := createLine(bypass, moveID, expenseID, partnerID, label, total, 0); err != nil {
			return err
		}
		if err := createLine(bypass, moveID, payID, partnerID, label, 0, total); err != nil {
			return err
		}
	default:
		return fmt.Errorf("posting move_type %q not supported in v1", moveType)
	}
	return orm.UpdateRecordByID(bypass, "account.move", moveID, map[string]interface{}{"state": "posted"})
}

func createLine(ctx context.Context, moveID int, accountID, partnerID int64, name string, debit, credit float64) error {
	if accountID <= 0 {
		return fmt.Errorf("missing chart account for posting")
	}
	lineModel, err := modelOrErr("account.move.line")
	if err != nil {
		return err
	}
	_, err = orm.Create(ctx, lineModel, map[string]interface{}{
		"move_id":    moveID,
		"account_id": accountID,
		"partner_id": partnerID,
		"name":       name,
		"debit":      debit,
		"credit":     credit,
		"balance":    debit - credit,
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
