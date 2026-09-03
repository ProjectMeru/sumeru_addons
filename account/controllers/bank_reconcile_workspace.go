package controllers

import (
	"encoding/csv"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"

	"sumeru/core/orm"
	"sumeru/core/server/router"
	banksvc "sumeru_addons/account/services"
)

func init() {
	router.Register(http.MethodGet, "/account/bank/reconcile", router.AuthSession, reconWorkspaceHandler)
	router.Register(http.MethodPost, "/account/bank/reconcile/match", router.AuthSession, reconMatchHandler)
	router.Register(http.MethodPost, "/account/bank/reconcile/import", router.AuthSession, reconImportHandler)
}

var reconTmpl = template.Must(template.New("recon").Parse(`<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>Bank Reconciliation</title>
<style>
body{font-family:system-ui,sans-serif;margin:1.5rem}
.grid{display:grid;grid-template-columns:1fr 1fr;gap:1.25rem}
table{width:100%;border-collapse:collapse}
th,td{border-bottom:1px solid #ddd;padding:.4rem;text-align:left}
.num{text-align:right}
.ok{color:#0a7}
form.inline{display:inline}
.hint{color:#555;font-size:.9rem}
</style></head><body>
<h1>Bank Reconciliation</h1>
<form method="post" action="/account/bank/reconcile/import" enctype="multipart/form-data">
<label>Statement ID <input type="number" name="statement_id" required/></label>
<input type="file" name="file" accept=".csv,text/csv" required/>
<button type="submit">Import CSV</button>
</form>
<p class="hint">CSV columns: date,label,amount</p>
<div class="grid">
<div>
<h2>Statement Lines</h2>
<table><tr><th>Date</th><th>Label</th><th class="num">Amount</th><th></th></tr>
{{range .Lines}}<tr><td>{{.Date}}</td><td>{{.Name}}</td><td class="num">{{.Amount}}</td>
<td>{{if .Reconciled}}<span class="ok">OK</span>{{else}}<form class="inline" method="post" action="/account/bank/reconcile/match"><input type="hidden" name="line_id" value="{{.ID}}"/><button>Match</button></form>{{end}}</td></tr>
{{end}}</table>
</div>
<div>
<h2>Suggested Matches</h2>
{{if .Suggestions}}
<table><tr><th>Line</th><th>Journal Item</th><th>Partner</th><th class="num">Amount</th><th class="num">Diff</th></tr>
{{range .Suggestions}}<tr><td>{{.LineName}}</td><td>{{.Label}}</td><td>{{.Partner}}</td><td class="num">{{printf "%.2f" .Amount}}</td><td class="num">{{printf "%.2f" .Diff}}</td></tr>
{{end}}</table>
{{else}}<p class="hint">Open statement lines show suggested journal items ranked by amount and label.</p>{{end}}
</div>
</div>
</body></html>`))

type reconLine struct {
	ID, Date, Name, Amount string
	Reconciled             bool
}

type reconSuggestion struct {
	LineName, Label, Partner string
	Amount, Diff             float64
}

func reconWorkspaceHandler(w http.ResponseWriter, r *http.Request) {
	ctx := orm.ContextWithBypass(r.Context(), true)
	lines, _ := orm.Search(ctx, "account.bank.statement.line", nil)
	view := make([]reconLine, 0, len(lines))
	suggestions := make([]reconSuggestion, 0)
	for _, ln := range lines {
		id, _ := orm.CoerceInt64(ln["id"])
		rec := asBool(ln["is_reconciled"])
		view = append(view, reconLine{
			ID:         strconv.Itoa(int(id)),
			Date:       orm.AsString(ln["date"]),
			Name:       orm.AsString(ln["name"]),
			Amount:     orm.AsString(ln["amount"]),
			Reconciled: rec,
		})
		if rec || id <= 0 {
			continue
		}
		matches, err := banksvc.SuggestedMatches(ctx, int(id))
		if err != nil {
			continue
		}
		for _, m := range matches {
			suggestions = append(suggestions, reconSuggestion{
				LineName: orm.AsString(ln["name"]),
				Label:    m.Label,
				Partner:  m.Partner,
				Amount:   m.Amount,
				Diff:     m.Diff,
			})
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = reconTmpl.Execute(w, map[string]interface{}{"Lines": view, "Suggestions": suggestions})
}

func reconMatchHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	id, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("line_id")))
	if id <= 0 {
		http.Error(w, "missing line", http.StatusBadRequest)
		return
	}
	ctx := orm.ContextWithBypass(r.Context(), true)
	if _, err := banksvc.MatchStatementLine(ctx, id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/account/bank/reconcile", http.StatusSeeOther)
}

func reconImportHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(4 << 20); err != nil {
		http.Error(w, "invalid upload", http.StatusBadRequest)
		return
	}
	stmtID, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("statement_id")))
	if stmtID <= 0 {
		http.Error(w, "statement_id required", http.StatusBadRequest)
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file required", http.StatusBadRequest)
		return
	}
	defer file.Close()
	reader := csv.NewReader(io.LimitReader(file, 2<<20))
	rows, err := reader.ReadAll()
	if err != nil {
		http.Error(w, "invalid csv", http.StatusBadRequest)
		return
	}
	if len(rows) > 0 && looksLikeCSVHeader(rows[0]) {
		rows = rows[1:]
	}
	ctx := orm.ContextWithBypass(r.Context(), true)
	if err := banksvc.ImportCSVStatement(ctx, stmtID, rows); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/account/bank/reconcile", http.StatusSeeOther)
}

func looksLikeCSVHeader(row []string) bool {
	if len(row) == 0 {
		return false
	}
	first := strings.ToLower(strings.TrimSpace(row[0]))
	return first == "date" || first == "label"
}

func asBool(v interface{}) bool {
	switch t := v.(type) {
	case bool:
		return t
	case int64:
		return t != 0
	default:
		return false
	}
}
