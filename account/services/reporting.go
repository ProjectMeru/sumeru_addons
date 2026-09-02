package services

import (
	"context"
	"math"
	"sort"
	"strings"
	"time"

	"sumeru/core/orm"
)

type ReportLine struct {
	Code    string  `json:"code"`
	Name    string  `json:"name"`
	Balance float64 `json:"balance"`
}

type ReportResult struct {
	Title     string       `json:"title"`
	DateFrom  string       `json:"date_from"`
	DateTo    string       `json:"date_to"`
	Lines     []ReportLine `json:"lines"`
	Total     float64      `json:"total"`
	Generated string       `json:"generated"`
}

func defaultRange() (string, string) {
	now := time.Now()
	from := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	return from.Format("2006-01-02"), now.Format("2006-01-02")
}

func postedLines(ctx context.Context, dateFrom, dateTo string) ([]map[string]interface{}, error) {
	moves, err := orm.Search(ctx, "account.move", [][]interface{}{
		{"state", "=", "posted"},
		{"date", ">=", dateFrom},
		{"date", "<=", dateTo},
	})
	if err != nil {
		return nil, err
	}
	if len(moves) == 0 {
		return nil, nil
	}
	ids := make([]interface{}, 0, len(moves))
	for _, m := range moves {
		if id, ok := orm.CoerceInt64(m["id"]); ok {
			ids = append(ids, id)
		}
	}
	return orm.Search(ctx, "account.move.line", [][]interface{}{
		{"move_id", "in", ids},
		{"display_type", "=", "entry"},
	})
}

func balanceByAccountType(ctx context.Context, dateFrom, dateTo string, types ...string) (map[int64]float64, error) {
	lines, err := postedLines(ctx, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}
	typeSet := map[string]struct{}{}
	for _, t := range types {
		typeSet[t] = struct{}{}
	}
	out := map[int64]float64{}
	for _, ln := range lines {
		acctID, _ := orm.CoerceInt64(ln["account_id"])
		if acctID <= 0 {
			continue
		}
		acct, err := orm.SearchOne(ctx, "account.account", map[string]interface{}{"id": acctID})
		if err != nil {
			continue
		}
		if len(typeSet) > 0 {
			if _, ok := typeSet[orm.AsString(acct["account_type"])]; !ok {
				continue
			}
		}
		out[acctID] += numericFloat(ln["debit"]) - numericFloat(ln["credit"])
	}
	return out, nil
}

func ProfitAndLoss(ctx context.Context, dateFrom, dateTo string) (*ReportResult, error) {
	if dateFrom == "" || dateTo == "" {
		dateFrom, dateTo = defaultRange()
	}
	income, err := balanceByAccountType(ctx, dateFrom, dateTo, "income")
	if err != nil {
		return nil, err
	}
	expense, err := balanceByAccountType(ctx, dateFrom, dateTo, "expense")
	if err != nil {
		return nil, err
	}
	lines := accountLines(ctx, income, true)
	lines = append(lines, accountLines(ctx, expense, false)...)
	incomeTotal, expenseTotal := 0.0, 0.0
	for _, v := range income {
		incomeTotal += -v
	}
	for _, v := range expense {
		expenseTotal += v
	}
	return &ReportResult{
		Title:     "Profit & Loss",
		DateFrom:  dateFrom,
		DateTo:    dateTo,
		Lines:     lines,
		Total:     round2(incomeTotal - expenseTotal),
		Generated: time.Now().Format(time.RFC3339),
	}, nil
}

func BalanceSheet(ctx context.Context, dateFrom, dateTo string) (*ReportResult, error) {
	if dateTo == "" {
		_, dateTo = defaultRange()
	}
	if dateFrom == "" {
		dateFrom = "1970-01-01"
	}
	assetTypes := []string{"asset_receivable", "asset_cash", "asset_current"}
	liabilityTypes := []string{"liability_payable", "liability_current"}
	assets, _ := balanceByAccountType(ctx, dateFrom, dateTo, assetTypes...)
	liabilities, _ := balanceByAccountType(ctx, dateFrom, dateTo, liabilityTypes...)
	equity, _ := balanceByAccountType(ctx, dateFrom, dateTo, "equity")
	income, _ := balanceByAccountType(ctx, dateFrom, dateTo, "income")
	expense, _ := balanceByAccountType(ctx, dateFrom, dateTo, "expense")
	retained := 0.0
	for _, v := range income {
		retained += -v
	}
	for _, v := range expense {
		retained -= v
	}
	lines := accountLines(ctx, assets, false)
	lines = append(lines, accountLines(ctx, liabilities, true)...)
	lines = append(lines, accountLines(ctx, equity, true)...)
	if retained != 0 {
		lines = append(lines, ReportLine{Name: "Current Year Earnings", Balance: round2(retained)})
	}
	assetTotal, liabEq := 0.0, 0.0
	for _, v := range assets {
		assetTotal += v
	}
	for _, v := range liabilities {
		liabEq += -v
	}
	for _, v := range equity {
		liabEq += -v
	}
	liabEq += retained
	return &ReportResult{
		Title:     "Balance Sheet",
		DateFrom:  dateFrom,
		DateTo:    dateTo,
		Lines:     lines,
		Total:     round2(assetTotal),
		Generated: time.Now().Format(time.RFC3339),
	}, nil
}

func TrialBalance(ctx context.Context, dateFrom, dateTo string) (*ReportResult, error) {
	if dateFrom == "" || dateTo == "" {
		dateFrom, dateTo = defaultRange()
	}
	balances, err := balanceByAccountType(ctx, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}
	lines := accountLines(ctx, balances, false)
	totalDebit, totalCredit := 0.0, 0.0
	for _, ln := range lines {
		if ln.Balance >= 0 {
			totalDebit += ln.Balance
		} else {
			totalCredit += -ln.Balance
		}
	}
	return &ReportResult{
		Title:     "Trial Balance",
		DateFrom:  dateFrom,
		DateTo:    dateTo,
		Lines:     lines,
		Total:     round2(totalDebit - totalCredit),
		Generated: time.Now().Format(time.RFC3339),
	}, nil
}

func GeneralLedger(ctx context.Context, accountID int64, dateFrom, dateTo string) (*ReportResult, error) {
	if dateFrom == "" || dateTo == "" {
		dateFrom, dateTo = defaultRange()
	}
	lines, err := postedLines(ctx, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}
	out := make([]ReportLine, 0)
	running := 0.0
	sort.SliceStable(lines, func(i, j int) bool {
		di := orm.AsString(lines[i]["date"])
		dj := orm.AsString(lines[j]["date"])
		return di < dj
	})
	for _, ln := range lines {
		acctID, _ := orm.CoerceInt64(ln["account_id"])
		if accountID > 0 && acctID != accountID {
			continue
		}
		delta := numericFloat(ln["debit"]) - numericFloat(ln["credit"])
		running += delta
		out = append(out, ReportLine{
			Name:    strings.TrimSpace(orm.AsString(ln["name"])),
			Code:    orm.AsString(ln["date"]),
			Balance: round2(running),
		})
	}
	title := "General Ledger"
	if accountID > 0 {
		if acct, err := orm.SearchOne(ctx, "account.account", map[string]interface{}{"id": accountID}); err == nil {
			title = "GL: " + orm.AsString(acct["name"])
		}
	}
	return &ReportResult{
		Title:     title,
		DateFrom:  dateFrom,
		DateTo:    dateTo,
		Lines:     out,
		Total:     round2(running),
		Generated: time.Now().Format(time.RFC3339),
	}, nil
}

func accountLines(ctx context.Context, balances map[int64]float64, invert bool) []ReportLine {
	ids := make([]int64, 0, len(balances))
	for id := range balances {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	out := make([]ReportLine, 0, len(ids))
	for _, id := range ids {
		bal := balances[id]
		if invert {
			bal = -bal
		}
		if math.Abs(bal) < 0.005 {
			continue
		}
		acct, err := orm.SearchOne(ctx, "account.account", map[string]interface{}{"id": id})
		if err != nil {
			continue
		}
		out = append(out, ReportLine{
			Code:    orm.AsString(acct["code"]),
			Name:    orm.AsString(acct["name"]),
			Balance: round2(bal),
		})
	}
	return out
}
