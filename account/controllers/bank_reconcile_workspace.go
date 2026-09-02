package controllers

import (
	"html/template"
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
}

var reconTmpl = template.Must(template.New("recon").Parse(`<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>Bank Reconciliation</title>
<style>
body{font-family:system-ui,sans-serif;margin:2rem}
.grid{display:grid;grid-template-columns:1fr 1fr;gap:1rem}
table{width:100%;border-collapse:collapse}
th,td{border-bottom:1px solid #ddd;padding:.4rem;text-align:left}
</style></head><body>
<h1>Bank Reconciliation</h1>
<div class="grid">
<div><h2>Statement Lines</h2>
<table><tr><th>Date</th><th>Label</th><th>Amount</th><th></th></tr>
{{range .Lines}}<tr><td>{{.Date}}</td><td>{{.Name}}</td><td>{{.Amount}}</td>
<td>{{if not .Reconciled}}<form method="post" action="/account/bank/reconcile/match"><input type="hidden" name="line_id" value="{{.ID}}"/><button>Match</button></form>{{else}}OK{{end}}</td></tr>
{{end}}</table></div>
<div><h2>Help</h2><p>Click Match to auto-link a statement line to an open journal item by amount and label.</p></div>
</div></body></html>`))

type reconLine struct {
	ID, Date, Name, Amount string
	Reconciled             bool
}

func reconWorkspaceHandler(w http.ResponseWriter, r *http.Request) {
	ctx := orm.ContextWithBypass(r.Context(), true)
	lines, _ := orm.Search(ctx, "account.bank.statement.line", nil)
	view := make([]reconLine, 0, len(lines))
	for _, ln := range lines {
		id, _ := orm.CoerceInt64(ln["id"])
		view = append(view, reconLine{
			ID:         strconv.Itoa(int(id)),
			Date:       orm.AsString(ln["date"]),
			Name:       orm.AsString(ln["name"]),
			Amount:     orm.AsString(ln["amount"]),
			Reconciled: asBool(ln["is_reconciled"]),
		})
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = reconTmpl.Execute(w, map[string]interface{}{"Lines": view})
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
