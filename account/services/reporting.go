package services

import (
	"context"
	"math"
	"sort"
	"strconv"
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
	assetTotal := 0.0
	for _, v := range assets {
		assetTotal += v
	}
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

func PartnerLedger(ctx context.Context, dateFrom, dateTo string) (*ReportResult, error) {
	if dateFrom == "" || dateTo == "" {
		dateFrom, dateTo = defaultRange()
	}
	lines, err := partnerBalanceLines(ctx, dateFrom, dateTo, "asset_receivable", "liability_payable")
	if err != nil {
		return nil, err
	}
	total := 0.0
	for _, ln := range lines {
		total += ln.Balance
	}
	return &ReportResult{
		Title:     "Partner Ledger",
		DateFrom:  dateFrom,
		DateTo:    dateTo,
		Lines:     lines,
		Total:     round2(total),
		Generated: time.Now().Format(time.RFC3339),
	}, nil
}

func AgedReceivable(ctx context.Context, dateFrom, dateTo string) (*ReportResult, error) {
	return agedPartnerReport(ctx, dateFrom, dateTo, "asset_receivable", "Aged Receivable")
}

func AgedPayable(ctx context.Context, dateFrom, dateTo string) (*ReportResult, error) {
	return agedPartnerReport(ctx, dateFrom, dateTo, "liability_payable", "Aged Payable")
}

func CashFlow(ctx context.Context, dateFrom, dateTo string) (*ReportResult, error) {
	if dateFrom == "" || dateTo == "" {
		dateFrom, dateTo = defaultRange()
	}
	balances, err := balanceByAccountType(ctx, dateFrom, dateTo, "asset_cash")
	if err != nil {
		return nil, err
	}
	lines := accountLines(ctx, balances, false)
	inflow, outflow := 0.0, 0.0
	for _, ln := range lines {
		if ln.Balance >= 0 {
			inflow += ln.Balance
		} else {
			outflow += -ln.Balance
		}
	}
	return &ReportResult{
		Title:     "Cash Flow",
		DateFrom:  dateFrom,
		DateTo:    dateTo,
		Lines:     lines,
		Total:     round2(inflow - outflow),
		Generated: time.Now().Format(time.RFC3339),
	}, nil
}

func AnnualComposite(ctx context.Context, dateFrom, dateTo string) (*ReportResult, error) {
	year := time.Now().Year()
	if dateTo != "" {
		if t, err := time.Parse("2006-01-02", dateTo); err == nil {
			year = t.Year()
		}
	}
	lines := make([]ReportLine, 0, 12)
	netTotal := 0.0
	for month := 1; month <= 12; month++ {
		from := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local).Format("2006-01-02")
		end := time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.Local).Format("2006-01-02")
		pl, err := ProfitAndLoss(ctx, from, end)
		if err != nil {
			return nil, err
		}
		lines = append(lines, ReportLine{
			Code:    from[:7],
			Name:    time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local).Format("January"),
			Balance: pl.Total,
		})
		netTotal += pl.Total
	}
	return &ReportResult{
		Title:     "Annual Composite (P&L by Month)",
		DateFrom:  time.Date(year, 1, 1, 0, 0, 0, 0, time.Local).Format("2006-01-02"),
		DateTo:    time.Date(year, 12, 31, 0, 0, 0, 0, time.Local).Format("2006-01-02"),
		Lines:     lines,
		Total:     round2(netTotal),
		Generated: time.Now().Format(time.RFC3339),
	}, nil
}

func partnerBalanceLines(ctx context.Context, dateFrom, dateTo string, acctTypes ...string) ([]ReportLine, error) {
	posted, err := postedLines(ctx, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}
	typeSet := map[string]struct{}{}
	for _, t := range acctTypes {
		typeSet[t] = struct{}{}
	}
	byPartner := map[int64]float64{}
	for _, ln := range posted {
		partnerID, _ := orm.CoerceInt64(ln["partner_id"])
		if partnerID <= 0 {
			continue
		}
		acctID, _ := orm.CoerceInt64(ln["account_id"])
		if acctID <= 0 {
			continue
		}
		acct, err := orm.SearchOne(ctx, "account.account", map[string]interface{}{"id": acctID})
		if err != nil {
			continue
		}
		if _, ok := typeSet[orm.AsString(acct["account_type"])]; !ok {
			continue
		}
		byPartner[partnerID] += numericFloat(ln["debit"]) - numericFloat(ln["credit"])
	}
	ids := make([]int64, 0, len(byPartner))
	for id := range byPartner {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	out := make([]ReportLine, 0, len(ids))
	for _, id := range ids {
		bal := byPartner[id]
		if math.Abs(bal) < 0.005 {
			continue
		}
		name := strconv.FormatInt(id, 10)
		if p, err := orm.SearchOne(ctx, "core.partner", map[string]interface{}{"id": id}); err == nil {
			name = orm.AsString(p["name"])
		}
		out = append(out, ReportLine{Code: name, Name: "Balance", Balance: round2(bal)})
	}
	return out, nil
}

func agedPartnerReport(ctx context.Context, dateFrom, dateTo string, acctType, title string) (*ReportResult, error) {
	if dateTo == "" {
		_, dateTo = defaultRange()
	}
	if dateFrom == "" {
		dateFrom = "1970-01-01"
	}
	asOf, err := time.Parse("2006-01-02", dateTo)
	if err != nil {
		asOf = time.Now()
	}
	buckets := []struct {
		code, name string
		min, max   int
	}{
		{"current", "Not Due", 0, 0},
		{"1-30", "1–30 Days", 1, 30},
		{"31-60", "31–60 Days", 31, 60},
		{"61-90", "61–90 Days", 61, 90},
		{"90+", "90+ Days", 91, 100000},
	}
	totals := make([]float64, len(buckets))
	rows, err := orm.Search(ctx, "account.move.line", [][]interface{}{
		{"reconciled", "=", false},
		{"display_type", "=", "entry"},
	})
	if err != nil {
		return nil, err
	}
	for _, ln := range rows {
		acctID, _ := orm.CoerceInt64(ln["account_id"])
		if acctID <= 0 {
			continue
		}
		acct, err := orm.SearchOne(ctx, "account.account", map[string]interface{}{"id": acctID})
		if err != nil || orm.AsString(acct["account_type"]) != acctType {
			continue
		}
		res := numericFloat(ln["amount_residual"])
		if math.Abs(res) <= balanceEpsilon {
			continue
		}
		due := orm.AsString(ln["date"])
		moveID, _ := orm.CoerceInt64(ln["move_id"])
		if moveID > 0 {
			if move, err := orm.SearchOne(ctx, "account.move", map[string]interface{}{"id": moveID}); err == nil {
				if d := orm.AsString(move["invoice_date_due"]); d != "" {
					due = d
				} else if d := orm.AsString(move["invoice_date"]); d != "" {
					due = d
				} else if d := orm.AsString(move["date"]); d != "" {
					due = d
				}
			}
		}
		dueDate, err := time.Parse("2006-01-02", due)
		if err != nil {
			dueDate = asOf
		}
		age := int(asOf.Sub(dueDate).Hours() / 24)
		if age < 0 {
			age = 0
		}
		amt := math.Abs(res)
		if acctType == "liability_payable" {
			amt = res
			if amt > 0 {
				amt = -amt
			}
			amt = math.Abs(amt)
		}
		for i, b := range buckets {
			if b.max == 0 && age == 0 {
				totals[i] += amt
				break
			}
			if b.max == 0 {
				continue
			}
			if age >= b.min && age <= b.max {
				totals[i] += amt
				break
			}
		}
	}
	lines := make([]ReportLine, 0, len(buckets))
	total := 0.0
	for i, b := range buckets {
		lines = append(lines, ReportLine{Code: b.code, Name: b.name, Balance: round2(totals[i])})
		total += totals[i]
	}
	return &ReportResult{
		Title:     title,
		DateFrom:  dateFrom,
		DateTo:    dateTo,
		Lines:     lines,
		Total:     round2(total),
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
