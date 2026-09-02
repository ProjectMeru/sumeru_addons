package services

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"sumeru/core/orm"
)

const balanceEpsilon = 0.005

type orderBillSpec struct {
	orderModel      string
	orderID         int
	confirmedState  string
	lineModel       string
	qtyField        string
	moveType        string
	seqCode         string
	docPrefix       string
	journalType     string
	incomeSide      bool
	taxUse          string
	narrationPrefix string
	afterCreate     func(context.Context, int) error
}

type invoicePosting struct {
	partnerPayable bool
	partnerDR      bool
	productDR      bool
	taxDR          bool
}

var postingByMoveType = map[string]invoicePosting{
	"out_invoice": {partnerDR: true},
	"out_refund":  {productDR: true, taxDR: true},
	"in_invoice":  {partnerPayable: true, productDR: true, taxDR: true},
	"in_refund":   {partnerPayable: true, partnerDR: true},
}

func modelOrErr(name string) (orm.Model, error) {
	m, ok := orm.Registry[name]
	if !ok {
		return nil, fmt.Errorf("model %s missing", name)
	}
	return m, nil
}

func nextDocName(ctx context.Context, code, fallbackPrefix string) string {
	if name, err := orm.NextSequence(ctx, code); err == nil && name != "" {
		return name
	}
	return fmt.Sprintf("%s/%s/%05d", fallbackPrefix, time.Now().Format("2006"), time.Now().Unix()%100000)
}

func CreateCustomerInvoiceFromSale(ctx context.Context, orderID int) (int, error) {
	return createBillFromOrder(ctx, orderBillSpec{
		orderModel:      "sale.order",
		orderID:         orderID,
		confirmedState:  "sale",
		lineModel:       "sale.order.line",
		qtyField:        "product_uom_qty",
		moveType:        "out_invoice",
		seqCode:         "account.move.out_invoice",
		docPrefix:       "INV",
		journalType:     "sale",
		incomeSide:      true,
		taxUse:          "sale",
		narrationPrefix: "Invoice for",
		afterCreate: func(ctx context.Context, id int) error {
			return orm.UpdateRecordByID(ctx, "sale.order", id, map[string]interface{}{
				"invoice_status": "invoiced",
			})
		},
	})
}

func CreateVendorBillFromPurchase(ctx context.Context, poID int) (int, error) {
	return createBillFromOrder(ctx, orderBillSpec{
		orderModel:      "purchase.order",
		orderID:         poID,
		confirmedState:  "purchase",
		lineModel:       "purchase.order.line",
		qtyField:        "product_qty",
		moveType:        "in_invoice",
		seqCode:         "account.move.in_invoice",
		docPrefix:       "BILL",
		journalType:     "purchase",
		incomeSide:      false,
		taxUse:          "purchase",
		narrationPrefix: "Bill for",
		afterCreate: func(ctx context.Context, id int) error {
			return orm.UpdateRecordByID(ctx, "purchase.order", id, map[string]interface{}{
				"invoice_status": "invoiced",
			})
		},
	})
}

func createBillFromOrder(ctx context.Context, spec orderBillSpec) (int, error) {
	if spec.orderID <= 0 {
		return 0, fmt.Errorf("invalid order id")
	}
	bypass := orm.ContextWithBypass(ctx, true)
	order, err := orm.SearchOne(bypass, spec.orderModel, map[string]interface{}{"id": spec.orderID})
	if err != nil {
		return 0, err
	}
	if orm.AsString(order["state"]) != spec.confirmedState {
		return 0, fmt.Errorf("order not confirmed")
	}
	origin := orm.AsString(order["name"])
	if origin != "" {
		existing, err := orm.Search(bypass, "account.move", [][]interface{}{
			{"invoice_origin", "=", origin},
			{"move_type", "=", spec.moveType},
		})
		if err != nil {
			return 0, err
		}
		if len(existing) > 0 {
			id, _ := orm.CoerceInt64(existing[0]["id"])
			return int(id), nil
		}
	}
	orderLines, err := orm.Search(bypass, spec.lineModel, [][]interface{}{
		{"order_id", "=", spec.orderID},
	})
	if err != nil {
		return 0, err
	}
	untaxed := 0.0
	for _, ln := range orderLines {
		_, _, sub := lineAmounts(ln, spec.qtyField)
		untaxed += sub
	}
	if untaxed <= 0 {
		untaxed = numericFloat(order["amount_total"])
	}
	partnerID, _ := orm.CoerceInt64(order["partner_id"])
	journalID := journalIDByType(bypass, spec.journalType)
	fallbackAcct := accountIDByType(bypass, "expense")
	if spec.incomeSide {
		fallbackAcct = accountIDByType(bypass, "income")
	}
	defaultTax := defaultTaxID(bypass, spec.taxUse)
	name := nextDocName(bypass, spec.seqCode, spec.docPrefix)
	today := time.Now().Format("2006-01-02")
	moveModel, err := modelOrErr("account.move")
	if err != nil {
		return 0, err
	}
	moveID, err := orm.Create(bypass, moveModel, map[string]interface{}{
		"name":            name,
		"move_type":       spec.moveType,
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
		"narration":       fmt.Sprintf("%s %s", spec.narrationPrefix, origin),
	})
	if err != nil {
		return 0, err
	}
	lineModel, err := modelOrErr("account.move.line")
	if err != nil {
		return 0, err
	}
	for _, ln := range orderLines {
		qty, price, sub := lineAmounts(ln, spec.qtyField)
		productID, _ := orm.CoerceInt64(ln["product_id"])
		acct := productAccountID(bypass, productID, spec.incomeSide, fallbackAcct)
		vals := map[string]interface{}{
			"move_id":        moveID,
			"account_id":     acct,
			"product_id":     productID,
			"name":           orm.AsString(ln["name"]),
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
		if _, err := orm.Create(bypass, lineModel, vals); err != nil {
			return 0, err
		}
	}
	if err := recomputeMoveAmounts(bypass, moveID); err != nil {
		return 0, err
	}
	if spec.afterCreate != nil {
		if err := spec.afterCreate(bypass, spec.orderID); err != nil {
			return 0, err
		}
	}
	return moveID, nil
}

func lineAmounts(ln map[string]interface{}, qtyField string) (qty, price, sub float64) {
	qty = numericFloat(ln[qtyField])
	if qty <= 0 {
		qty = 1
	}
	price = numericFloat(ln["price_unit"])
	sub = numericFloat(ln["price_subtotal"])
	if sub <= 0 {
		sub = qty * price
	}
	return qty, price, sub
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
		lid, ok := orm.CoerceInt64(ln["id"])
		if !ok {
			continue
		}
		if err := orm.Unlink(ctx, "account.move.line", int(lid)); err != nil {
			log.Printf("account: unlink line %d: %v", lid, err)
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
	lines, err := orm.Search(ctx, "account.move.line", [][]interface{}{{"move_id", "=", moveID}})
	if err != nil {
		return err
	}
	untaxed := 0.0
	taxTotal := 0.0
	for _, ln := range productLines(lines) {
		_, _, sub := lineAmounts(ln, "quantity")
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
		return fmt.Errorf("move cancelled")
	}
	allLines, err := orm.Search(bypass, "account.move.line", [][]interface{}{
		{"move_id", "=", moveID},
	})
	if err != nil {
		return err
	}
	moveType := orm.AsString(move["move_type"])
	if moveType == "entry" {
		return postManualEntry(bypass, moveID, move, allLines)
	}
	if err := recomputeMoveAmounts(bypass, moveID); err != nil {
		return err
	}
	move, err = orm.SearchOne(bypass, "account.move", map[string]interface{}{"id": moveID})
	if err != nil {
		return err
	}
	applyPaymentTermDue(bypass, move, moveID)

	prods := productLines(allLines)
	partnerID, _ := orm.CoerceInt64(move["partner_id"])
	untaxed := 0.0
	taxByTaxID := map[int64]float64{}
	incomeExpenseByAcct := map[int64]float64{}
	for _, ln := range prods {
		_, _, sub := lineAmounts(ln, "quantity")
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
		taxID = mapFiscalTax(bypass, partnerID, taxID)
		_, taxAcct, factor := taxRepartition(bypass, taxID)
		pct, fallbackAcct := taxAmountPercent(bypass, taxID)
		if taxAcct <= 0 {
			taxAcct = fallbackAcct
		}
		if pct != 0 && taxAcct > 0 {
			taxByTaxID[taxID] += sub * pct * factor / 10000.0
		}
	}
	taxTotal := 0.0
	for _, amt := range taxByTaxID {
		taxTotal += amt
	}
	total := untaxed + taxTotal
	if total <= 0 {
		total = numericFloat(move["amount_total"])
		untaxed = total
	}
	if total <= 0 {
		return fmt.Errorf("amount_total zero")
	}

	label := orm.AsString(move["name"])
	deleteAccountingLines(bypass, allLines)

	recvID := partnerAccountID(bypass, partnerID, true, accountIDByType(bypass, "asset_receivable"))
	payID := partnerAccountID(bypass, partnerID, false, accountIDByType(bypass, "liability_payable"))
	if err := postInvoiceMove(bypass, moveID, moveType, recvID, payID, partnerID, label, total, incomeExpenseByAcct, taxByTaxID); err != nil {
		return err
	}

	postedLines, err := orm.Search(bypass, "account.move.line", [][]interface{}{{"move_id", "=", moveID}})
	if err != nil {
		return err
	}
	if err := assertBalanced(postedLines); err != nil {
		return err
	}
	if err := upsertInvoiceReport(bypass, moveID); err != nil {
		log.Printf("account: invoice report move %d: %v", moveID, err)
	}
	return orm.UpdateRecordByID(bypass, "account.move", moveID, map[string]interface{}{
		"state":           "posted",
		"amount_untaxed":  round2(untaxed),
		"amount_tax":      round2(taxTotal),
		"amount_total":    round2(total),
		"amount_residual": round2(total),
		"payment_state":   "not_paid",
	})
}

func postInvoiceMove(ctx context.Context, moveID int, moveType string, recvID, payID, partnerID int64, label string, total float64, byAcct map[int64]float64, taxByTaxID map[int64]float64) error {
	cfg, ok := postingByMoveType[moveType]
	if !ok {
		return fmt.Errorf("unsupported move_type %q", moveType)
	}
	partnerAcct := recvID
	if cfg.partnerPayable {
		partnerAcct = payID
	}
	var pDebit, pCredit, pRes float64
	if cfg.partnerDR {
		pDebit, pCredit, pRes = total, 0, total
	} else {
		pDebit, pCredit, pRes = 0, total, -total
	}
	if err := createEntryLine(ctx, moveID, partnerAcct, partnerID, label, pDebit, pCredit, pRes); err != nil {
		return err
	}
	for acct, amt := range byAcct {
		var debit, credit float64
		if cfg.productDR {
			debit, credit = amt, 0
		} else {
			debit, credit = 0, amt
		}
		if err := createEntryLine(ctx, moveID, acct, partnerID, label, debit, credit, 0); err != nil {
			return err
		}
	}
	for taxID, amt := range taxByTaxID {
		_, taxAcct := taxAmountPercent(ctx, taxID)
		var debit, credit float64
		if cfg.taxDR {
			debit, credit = amt, 0
		} else {
			debit, credit = 0, amt
		}
		if err := createTaxLine(ctx, moveID, taxAcct, partnerID, taxID, label+" Tax", debit, credit); err != nil {
			return err
		}
	}
	return nil
}

func postManualEntry(ctx context.Context, moveID int, move map[string]interface{}, lines []map[string]interface{}) error {
	if len(lines) == 0 {
		return fmt.Errorf("missing journal items")
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
		return fmt.Errorf("missing account")
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
		return fmt.Errorf("missing tax account")
	}
	lineModel, err := modelOrErr("account.move.line")
	if err != nil {
		return err
	}
	_, err = orm.Create(ctx, lineModel, map[string]interface{}{
		"move_id":      moveID,
		"account_id":   accountID,
		"partner_id":   partnerID,
		"tax_origin_id": taxID,
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
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case int32:
		return float64(t)
	default:
		s := strings.TrimSpace(orm.AsString(v))
		if s == "" {
			return 0
		}
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0
		}
		return f
	}
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
