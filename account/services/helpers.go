package services

import (
	"context"
	"math"
	"sort"
	"time"

	"sumeru/core/orm"
)

type termDueLine struct {
	percent float64
	days    int
	delay   string
}

func taxRepartition(ctx context.Context, taxID int64) (baseAcct, taxAcct int64, factor float64) {
	if taxID <= 0 {
		return 0, 0, 100
	}
	rows, err := orm.Search(ctx, "account.tax.repartition.line", [][]interface{}{
		{"tax_id", "=", taxID},
	})
	if err != nil || len(rows) == 0 {
		pct, acct := taxAmountPercent(ctx, taxID)
		return 0, acct, pct
	}
	for _, row := range rows {
		if orm.AsString(row["repartition_type"]) == "tax" {
			acct, _ := orm.CoerceInt64(row["account_id"])
			factor = numericFloat(row["factor_percent"])
			if factor <= 0 {
				factor = 100
			}
			if acct <= 0 {
				_, acct = taxAmountPercent(ctx, taxID)
			}
			return 0, acct, factor
		}
	}
	return 0, 0, 100
}

func mapFiscalTax(ctx context.Context, partnerID, taxID int64) int64 {
	if taxID <= 0 || partnerID <= 0 {
		return taxID
	}
	partner, err := orm.SearchOne(ctx, "core.partner", map[string]interface{}{"id": partnerID})
	if err != nil {
		return taxID
	}
	countryID, _ := orm.CoerceInt64(partner["country_id"])
	positions, err := orm.Search(ctx, "account.fiscal.position", [][]interface{}{{"active", "=", true}})
	if err != nil {
		return taxID
	}
	for _, pos := range positions {
		posCountry, _ := orm.CoerceInt64(pos["country_id"])
		if posCountry > 0 && posCountry != countryID {
			continue
		}
		pid, _ := orm.CoerceInt64(pos["id"])
		mappings, _ := orm.Search(ctx, "account.fiscal.position.tax", [][]interface{}{
			{"position_id", "=", pid},
			{"tax_src_id", "=", taxID},
		})
		if len(mappings) == 0 {
			continue
		}
		if dest, ok := orm.CoerceInt64(mappings[0]["tax_dest_id"]); ok && dest > 0 {
			return dest
		}
	}
	return taxID
}

func paymentTermDueLines(ctx context.Context, termID int64) []termDueLine {
	if termID <= 0 {
		return nil
	}
	rows, err := orm.Search(ctx, "account.payment.term.line", [][]interface{}{
		{"payment_term_id", "=", termID},
	})
	if err != nil || len(rows) == 0 {
		term, err := orm.SearchOne(ctx, "account.payment.term", map[string]interface{}{"id": termID})
		if err != nil {
			return nil
		}
		days, _ := orm.CoerceInt64(term["days"])
		return []termDueLine{{percent: 100, days: int(days), delay: "days_after"}}
	}
	sort.SliceStable(rows, func(i, j int) bool {
		si, _ := orm.CoerceInt64(rows[i]["sequence"])
		sj, _ := orm.CoerceInt64(rows[j]["sequence"])
		return si < sj
	})
	out := make([]termDueLine, 0, len(rows))
	for _, row := range rows {
		days, _ := orm.CoerceInt64(row["days"])
		out = append(out, termDueLine{
			percent: numericFloat(row["value_amount"]),
			days:    int(days),
			delay:   orm.AsString(row["delay_type"]),
		})
	}
	return out
}

func computeDueDate(invDate string, line termDueLine) string {
	t, err := time.Parse("2006-01-02", invDate)
	if err != nil {
		t = time.Now()
	}
	switch line.delay {
	case "days_after_end_month":
		t = time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, t.Location())
	}
	return t.AddDate(0, 0, line.days).Format("2006-01-02")
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
	lines := paymentTermDueLines(ctx, termID)
	due := invDate
	if len(lines) > 0 {
		due = computeDueDate(invDate, lines[len(lines)-1])
	}
	_ = orm.UpdateRecordByID(ctx, "account.move", moveID, map[string]interface{}{
		"invoice_date":     invDate,
		"invoice_date_due": due,
	})
}

func maybeCreateFullReconcile(ctx context.Context, invoiceLineID int) error {
	line, err := orm.SearchOne(ctx, "account.move.line", map[string]interface{}{"id": invoiceLineID})
	if err != nil {
		return err
	}
	if math.Abs(numericFloat(line["amount_residual"])) > balanceEpsilon {
		return nil
	}
	if !asBool(line["reconciled"]) {
		return nil
	}
	partials, err := orm.Search(ctx, "account.partial.reconcile", [][]interface{}{
		{"debit_move_id", "=", invoiceLineID},
	})
	if err != nil {
		return err
	}
	more, err := orm.Search(ctx, "account.partial.reconcile", [][]interface{}{
		{"credit_move_id", "=", invoiceLineID},
	})
	if err != nil {
		return err
	}
	partials = append(partials, more...)
	if len(partials) == 0 {
		return nil
	}
	for _, pr := range partials {
		if fid, ok := orm.CoerceInt64(pr["full_reconcile_id"]); ok && fid > 0 {
			return nil
		}
	}
	fullModel, err := modelOrErr("account.full.reconcile")
	if err != nil {
		return err
	}
	name := nextDocName(ctx, "account.full.reconcile", "REC")
	fullID, err := orm.Create(ctx, fullModel, map[string]interface{}{"name": name})
	if err != nil {
		return err
	}
	for _, pr := range partials {
		pid, _ := orm.CoerceInt64(pr["id"])
		if pid <= 0 {
			continue
		}
		if err := orm.UpdateRecordByID(ctx, "account.partial.reconcile", int(pid), map[string]interface{}{
			"full_reconcile_id": fullID,
		}); err != nil {
			return err
		}
	}
	return nil
}

func asBool(v interface{}) bool {
	switch t := v.(type) {
	case bool:
		return t
	case int64:
		return t != 0
	case int:
		return t != 0
	case string:
		return t == "true" || t == "t" || t == "1"
	default:
		return false
	}
}
