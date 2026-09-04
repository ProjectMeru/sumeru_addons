package controllers

import (
	"context"
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"sumeru/core/orm"
	"sumeru/core/server/router"
	"sumeru_addons/account/services"
)

func init() {
	router.Register(http.MethodGet, "/account/reports/view", router.AuthSession, reportViewHandler)
	router.Register(http.MethodGet, "/account/reports/export.csv", router.AuthSession, reportCSVHandler)
}

var reportTmpl = template.Must(template.New("report").Parse(`<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>{{.Title}}</title>
<style>
body{font-family:system-ui,sans-serif;margin:2rem;color:#111}
h1{font-size:1.4rem;margin:0}
.meta{color:#555;margin:1rem 0}
table{width:100%;border-collapse:collapse}
th,td{border-bottom:1px solid #ddd;padding:.45rem;text-align:left}
.num{text-align:right}
.total{font-weight:700}
form{margin-bottom:1rem;display:flex;gap:.5rem;flex-wrap:wrap}
</style></head><body>
<form method="get" action="/account/reports/view">
<input type="hidden" name="type" value="{{.Type}}"/>
<label>From <input type="date" name="date_from" value="{{.DateFrom}}"/></label>
<label>To <input type="date" name="date_to" value="{{.DateTo}}"/></label>
{{if eq .Type "general_ledger"}}<label>Account ID <input type="number" name="account_id" value="{{.AccountID}}"/></label>{{end}}
<button type="submit">Run</button>
<a href="/account/reports/export.csv?type={{.Type}}&date_from={{.DateFrom}}&date_to={{.DateTo}}&account_id={{.AccountID}}">Export CSV</a>
</form>
<h1>{{.Title}}</h1>
<div class="meta">{{.DateFrom}} → {{.DateTo}}</div>
<table><thead><tr><th>Code</th><th>Account</th><th class="num">Balance</th></tr></thead><tbody>
{{range .Lines}}<tr><td>{{.Code}}</td><td>{{.Name}}</td><td class="num">{{printf "%.2f" .Balance}}</td></tr>{{end}}
<tr class="total"><td colspan="2">Total</td><td class="num">{{printf "%.2f" .Total}}</td></tr>
</tbody></table></body></html>`))

type reportPage struct {
	Type, Title, DateFrom, DateTo, AccountID string
	Total                                    float64
	Lines                                    []services.ReportLine
}

func runReport(ctx context.Context, typ, from, to string, accountID int64) (*services.ReportResult, string, error) {
	var (
		res *services.ReportResult
		err error
	)
	switch typ {
	case "balance_sheet":
		res, err = services.BalanceSheet(ctx, from, to)
	case "trial_balance":
		res, err = services.TrialBalance(ctx, from, to)
	case "general_ledger":
		res, err = services.GeneralLedger(ctx, accountID, from, to)
	case "partner_ledger":
		res, err = services.PartnerLedger(ctx, from, to)
	case "aged_receivable":
		res, err = services.AgedReceivable(ctx, from, to)
	case "aged_payable":
		res, err = services.AgedPayable(ctx, from, to)
	case "cash_flow":
		res, err = services.CashFlow(ctx, from, to)
	case "annual_composite":
		res, err = services.AnnualComposite(ctx, from, to)
	default:
		typ = "profit_loss"
		res, err = services.ProfitAndLoss(ctx, from, to)
	}
	return res, typ, err
}

func reportQuery(r *http.Request) (typ, from, to string, accountID int64) {
	q := r.URL.Query()
	typ = strings.TrimSpace(q.Get("type"))
	from = strings.TrimSpace(q.Get("date_from"))
	to = strings.TrimSpace(q.Get("date_to"))
	if id, err := strconv.ParseInt(strings.TrimSpace(q.Get("account_id")), 10, 64); err == nil {
		accountID = id
	}
	return typ, from, to, accountID
}

func reportViewHandler(w http.ResponseWriter, r *http.Request) {
	ctx := orm.ContextWithBypass(r.Context(), true)
	typ, from, to, accountID := reportQuery(r)
	res, typ, err := runReport(ctx, typ, from, to, accountID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	acct := ""
	if accountID > 0 {
		acct = strconv.FormatInt(accountID, 10)
	}
	page := reportPage{Type: typ, Title: res.Title, DateFrom: res.DateFrom, DateTo: res.DateTo, AccountID: acct, Total: res.Total, Lines: res.Lines}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = reportTmpl.Execute(w, page)
}

func reportCSVHandler(w http.ResponseWriter, r *http.Request) {
	ctx := orm.ContextWithBypass(r.Context(), true)
	typ, from, to, accountID := reportQuery(r)
	res, _, err := runReport(ctx, typ, from, to, accountID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=report.csv")
	_, _ = w.Write([]byte("code,name,balance\n"))
	for _, ln := range res.Lines {
		line, _ := json.Marshal([]string{ln.Code, ln.Name, formatFloat(ln.Balance)})
		_, _ = w.Write(line)
		_, _ = w.Write([]byte("\n"))
	}
}

func formatFloat(v float64) string {
	b, _ := json.Marshal(v)
	return strings.Trim(string(b), `"`)
}
