package controllers

import (
	"bytes"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"sumeru/core/orm"
	"sumeru/core/server/router"
)

func init() {
	router.Register(http.MethodGet, "/account/invoice/print", router.AuthSession, InvoicePrintHandler)
}

var invoicePrintTmpl = template.Must(template.New("invoice").Parse(`<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>{{.Name}}</title>
<style>
body{font-family:system-ui,sans-serif;margin:2rem;color:#111}
h1{font-size:1.5rem;margin:0 0 .25rem}
.meta{color:#555;margin-bottom:1.5rem}
table{width:100%;border-collapse:collapse;margin-top:1rem}
th,td{border-bottom:1px solid #ddd;padding:.5rem;text-align:left}
th{font-size:.75rem;text-transform:uppercase;color:#666}
.num{text-align:right}
.totals{margin-top:1.5rem;max-width:16rem;margin-left:auto}
.totals div{display:flex;justify-content:space-between;padding:.25rem 0}
.totals .grand{font-weight:700;border-top:1px solid #333;margin-top:.5rem;padding-top:.5rem}
@media print{button{display:none}}
</style></head><body>
<button onclick="window.print()">Print</button>
<h1>{{.Name}}</h1>
<div class="meta">{{.MoveType}} · {{.Partner}} · {{.InvoiceDate}}</div>
<table><thead><tr><th>Description</th><th class="num">Qty</th><th class="num">Unit</th><th class="num">Subtotal</th></tr></thead><tbody>
{{range .Lines}}<tr><td>{{.Desc}}</td><td class="num">{{.Qty}}</td><td class="num">{{.Unit}}</td><td class="num">{{.Subtotal}}</td></tr>
{{end}}</tbody></table>
<div class="totals">
<div><span>Untaxed</span><span>{{.Untaxed}}</span></div>
<div><span>Tax</span><span>{{.Tax}}</span></div>
<div class="grand"><span>Total</span><span>{{.Total}}</span></div>
<div><span>Amount Due</span><span>{{.Due}}</span></div>
</div></body></html>`))

type invoicePrintLine struct {
	Desc, Qty, Unit, Subtotal string
}

type invoicePrintData struct {
	Name, MoveType, Partner, InvoiceDate string
	Untaxed, Tax, Total, Due             string
	Lines                                []invoicePrintLine
}

func InvoicePrintHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("id")))
	if id <= 0 {
		path := strings.TrimPrefix(r.URL.Path, "/account/invoice/")
		path = strings.TrimSuffix(path, "/print")
		id, _ = strconv.Atoi(strings.TrimSpace(path))
	}
	if id <= 0 {
		http.Error(w, "missing invoice id", http.StatusBadRequest)
		return
	}
	ctx := orm.ContextWithBypass(r.Context(), true)
	move, err := orm.SearchOne(ctx, "account.move", map[string]interface{}{"id": id})
	if err != nil {
		http.Error(w, "invoice not found", http.StatusNotFound)
		return
	}
	partnerName := ""
	if pid, ok := orm.CoerceInt64(move["partner_id"]); ok && pid > 0 {
		if p, err := orm.SearchOne(ctx, "core.partner", map[string]interface{}{"id": pid}); err == nil {
			partnerName = orm.AsString(p["name"])
		}
	}
	lines, _ := orm.Search(ctx, "account.move.line", [][]interface{}{
		{"move_id", "=", id},
		{"display_type", "=", "product"},
	})
	printLines := make([]invoicePrintLine, 0, len(lines))
	for _, ln := range lines {
		printLines = append(printLines, invoicePrintLine{
			Desc:     orm.AsString(ln["name"]),
			Qty:      fmtAny(ln["quantity"]),
			Unit:     fmtAny(ln["price_unit"]),
			Subtotal: fmtAny(ln["price_subtotal"]),
		})
	}
	data := invoicePrintData{
		Name:        orm.AsString(move["name"]),
		MoveType:    orm.AsString(move["move_type"]),
		Partner:     partnerName,
		InvoiceDate: orm.AsString(move["invoice_date"]),
		Untaxed:     fmtAny(move["amount_untaxed"]),
		Tax:         fmtAny(move["amount_tax"]),
		Total:       fmtAny(move["amount_total"]),
		Due:         fmtAny(move["amount_residual"]),
		Lines:       printLines,
	}
	var buf bytes.Buffer
	if err := invoicePrintTmpl.Execute(&buf, data); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}

func fmtAny(v interface{}) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(orm.AsString(v))
}
