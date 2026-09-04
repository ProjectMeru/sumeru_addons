package services

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"sumeru/core/orm"
)

func init() {
	orm.RegisterObjectAction("account.recurring.template", "action_generate", actionGenerateRecurring)
	orm.RegisterObjectAction("account.deferred.schedule", "action_recognize", actionRecognizeDeferred)
	orm.RegisterObjectAction("account.asset", "action_depreciate", actionDepreciateAsset)
	orm.RegisterObjectAction("account.budget", "action_refresh", actionRefreshBudget)
	orm.RegisterObjectAction("account.tax.return", "action_compute", actionComputeTaxReturn)
}

func actionGenerateRecurring(ctx context.Context, _ string, id int, vals map[string]string) (string, error) {
	if _, err := GenerateRecurringMoves(ctx, time.Now()); err != nil {
		return "", err
	}
	return vals["next"], nil
}

func actionRecognizeDeferred(ctx context.Context, _ string, id int, vals map[string]string) (string, error) {
	if err := RecognizeDeferredPeriod(ctx, id, time.Now()); err != nil {
		return "", err
	}
	return vals["next"], nil
}

func actionDepreciateAsset(ctx context.Context, _ string, id int, vals map[string]string) (string, error) {
	if err := DepreciateAsset(ctx, id); err != nil {
		return "", err
	}
	return vals["next"], nil
}

func actionRefreshBudget(ctx context.Context, _ string, id int, vals map[string]string) (string, error) {
	if err := RefreshBudgetActuals(ctx, id); err != nil {
		return "", err
	}
	return vals["next"], nil
}

func actionComputeTaxReturn(ctx context.Context, _ string, id int, vals map[string]string) (string, error) {
	if err := ComputeTaxReturn(ctx, id); err != nil {
		return "", err
	}
	return vals["next"], nil
}

// GenerateRecurringMoves posts draft journal entries for due recurring templates.
func GenerateRecurringMoves(ctx context.Context, asOf time.Time) (int, error) {
	bypass := orm.ContextWithBypass(ctx, true)
	asOfDay := asOf.Format("2006-01-02")
	templates, err := orm.Search(bypass, "account.recurring.template", [][]interface{}{{"active", "=", true}})
	if err != nil {
		return 0, err
	}
	moveModel, err := modelOrErr("account.move")
	if err != nil {
		return 0, err
	}
	created := 0
	for _, tmpl := range templates {
		next := strings.TrimSpace(orm.AsString(tmpl["next_date"]))
		if next == "" || next > asOfDay {
			continue
		}
		id, _ := orm.CoerceInt64(tmpl["id"])
		amount := numericFloat(tmpl["amount"])
		if amount == 0 {
			continue
		}
		journalID, _ := orm.CoerceInt64(tmpl["journal_id"])
		partnerID, _ := orm.CoerceInt64(tmpl["partner_id"])
		moveType := strings.TrimSpace(orm.AsString(tmpl["move_type"]))
		if moveType == "" {
			moveType = "entry"
		}
		moveID, err := orm.Create(bypass, moveModel, map[string]interface{}{
			"name":           nextDocName(bypass, "account.move.entry", "REC"),
			"move_type":      moveType,
			"partner_id":     partnerID,
			"journal_id":     journalID,
			"date":           asOfDay,
			"invoice_date":   asOfDay,
			"state":          "draft",
			"amount_untaxed": amount,
			"amount_total":   amount,
			"amount_residual": amount,
			"payment_state":  "not_paid",
			"narration":      fmt.Sprintf("Recurring: %s", orm.AsString(tmpl["name"])),
		})
		if err != nil {
			return created, err
		}
		if err := createRecurringLines(bypass, moveID, tmpl, amount); err != nil {
			return created, err
		}
		nextDate := advanceRecurringDate(next, orm.AsString(tmpl["interval"]))
		if err := orm.UpdateRecordByID(bypass, "account.recurring.template", int(id), map[string]interface{}{
			"next_date": nextDate,
		}); err != nil {
			return created, err
		}
		created++
	}
	return created, nil
}

func createRecurringLines(ctx context.Context, moveID int, tmpl map[string]interface{}, amount float64) error {
	lineModel, err := modelOrErr("account.move.line")
	if err != nil {
		return err
	}
	accountID, _ := orm.CoerceInt64(tmpl["account_id"])
	if accountID <= 0 {
		accountID = accountIDByType(ctx, "expense")
	}
	partnerID, _ := orm.CoerceInt64(tmpl["partner_id"])
	_, err = orm.Create(ctx, lineModel, map[string]interface{}{
		"move_id":       moveID,
		"name":          orm.AsString(tmpl["name"]),
		"account_id":    accountID,
		"partner_id":    partnerID,
		"display_type":  "product",
		"quantity":      1,
		"price_unit":    amount,
		"price_subtotal": amount,
	})
	return err
}

func advanceRecurringDate(current, interval string) string {
	t, err := time.Parse("2006-01-02", current)
	if err != nil {
		return current
	}
	switch strings.ToLower(strings.TrimSpace(interval)) {
	case "weekly":
		t = t.AddDate(0, 0, 7)
	case "yearly":
		t = t.AddDate(1, 0, 0)
	default:
		t = t.AddDate(0, 1, 0)
	}
	return t.Format("2006-01-02")
}

// RecognizeDeferredPeriod recognizes a straight-line slice of a deferred schedule.
func RecognizeDeferredPeriod(ctx context.Context, scheduleID int, asOf time.Time) error {
	if scheduleID <= 0 {
		return fmt.Errorf("invalid schedule")
	}
	bypass := orm.ContextWithBypass(ctx, true)
	sched, err := orm.SearchOne(bypass, "account.deferred.schedule", map[string]interface{}{"id": scheduleID})
	if err != nil {
		return err
	}
	total := numericFloat(sched["amount"])
	recognized := numericFloat(sched["recognized_amount"])
	start, errS := time.Parse("2006-01-02", normalizeDate(sched["start_date"]))
	end, errE := time.Parse("2006-01-02", normalizeDate(sched["end_date"]))
	if errS != nil || errE != nil || !end.After(start) || total <= 0 {
		return fmt.Errorf("schedule dates or amount invalid")
	}
	months := monthsBetween(start, end)
	if months < 1 {
		months = 1
	}
	slice := round2(total / float64(months))
	remaining := round2(total - recognized)
	if slice > remaining {
		slice = remaining
	}
	if slice <= 0 {
		return orm.UpdateRecordByID(bypass, "account.deferred.schedule", scheduleID, map[string]interface{}{
			"state": "done",
		})
	}
	newRec := round2(recognized + slice)
	state := "open"
	if newRec+0.005 >= total {
		state = "done"
	}
	return orm.UpdateRecordByID(bypass, "account.deferred.schedule", scheduleID, map[string]interface{}{
		"recognized_amount": newRec,
		"state":             state,
	})
}

func AdvanceRecurringDateForTest(current, interval string) string {
	return advanceRecurringDate(current, interval)
}

func MonthsBetweenForTest(start, end time.Time) int {
	return monthsBetween(start, end)
}

func monthsBetween(start, end time.Time) int {
	y, m, _ := end.Date()
	ys, ms, _ := start.Date()
	n := (y-ys)*12 + int(m-ms)
	if n < 1 {
		return 1
	}
	return n
}

// DepreciateAsset posts one straight-line month of depreciation onto book value.
func DepreciateAsset(ctx context.Context, assetID int) error {
	if assetID <= 0 {
		return fmt.Errorf("invalid asset")
	}
	bypass := orm.ContextWithBypass(ctx, true)
	asset, err := orm.SearchOne(bypass, "account.asset", map[string]interface{}{"id": assetID})
	if err != nil {
		return err
	}
	original := numericFloat(asset["original_value"])
	salvage := numericFloat(asset["salvage_value"])
	book := numericFloat(asset["book_value"])
	if book <= 0 {
		book = original
	}
	months, _ := orm.CoerceInt64(asset["months"])
	if months <= 0 {
		months = 36
	}
	depreciable := original - salvage
	if depreciable <= 0 {
		return fmt.Errorf("nothing to depreciate")
	}
	monthly := round2(depreciable / float64(months))
	if monthly > book-salvage {
		monthly = round2(book - salvage)
	}
	if monthly <= 0 {
		return orm.UpdateRecordByID(bypass, "account.asset", assetID, map[string]interface{}{"state": "closed"})
	}
	newBook := round2(book - monthly)
	state := "open"
	if newBook <= salvage+0.005 {
		state = "closed"
	}
	return orm.UpdateRecordByID(bypass, "account.asset", assetID, map[string]interface{}{
		"book_value": newBook,
		"state":      state,
	})
}

// RefreshBudgetActuals fills actual/variance from posted P&L for the budget account and dates.
func RefreshBudgetActuals(ctx context.Context, budgetID int) error {
	if budgetID <= 0 {
		return fmt.Errorf("invalid budget")
	}
	bypass := orm.ContextWithBypass(ctx, true)
	budget, err := orm.SearchOne(bypass, "account.budget", map[string]interface{}{"id": budgetID})
	if err != nil {
		return err
	}
	from := normalizeDate(budget["date_from"])
	to := normalizeDate(budget["date_to"])
	accountID, _ := orm.CoerceInt64(budget["account_id"])
	actual := 0.0
	if accountID > 0 {
		gl, err := GeneralLedger(bypass, accountID, from, to)
		if err != nil {
			return err
		}
		actual = math.Abs(gl.Total)
	} else {
		pl, err := ProfitAndLoss(bypass, from, to)
		if err != nil {
			return err
		}
		actual = math.Abs(pl.Total)
	}
	amount := numericFloat(budget["amount"])
	return orm.UpdateRecordByID(bypass, "account.budget", budgetID, map[string]interface{}{
		"actual_amount": round2(actual),
		"variance":      round2(amount - actual),
		"state":         "open",
	})
}

// ComputeTaxReturn sums posted tax lines in the return period.
func ComputeTaxReturn(ctx context.Context, returnID int) error {
	if returnID <= 0 {
		return fmt.Errorf("invalid tax return")
	}
	bypass := orm.ContextWithBypass(ctx, true)
	ret, err := orm.SearchOne(bypass, "account.tax.return", map[string]interface{}{"id": returnID})
	if err != nil {
		return err
	}
	from := normalizeDate(ret["period_start"])
	to := normalizeDate(ret["period_end"])
	lines, err := postedLines(bypass, from, to)
	if err != nil {
		return err
	}
	total := 0.0
	for _, ln := range lines {
		if orm.AsString(ln["display_type"]) != "tax" {
			continue
		}
		total += numericFloat(ln["credit"]) - numericFloat(ln["debit"])
	}
	return orm.UpdateRecordByID(bypass, "account.tax.return", returnID, map[string]interface{}{
		"amount": round2(total),
		"state":  "computed",
	})
}
