package controllers

import (
	"fmt"
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

// InvoicePrintHandler GET /account/invoice/print?id=… — HTML invoice for printing.
func InvoicePrintHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("id")))
	if id <= 0 {
		// Also accept /account/invoice/<id>/print via path suffix if routed that way.
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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var b strings.Builder
	b.WriteString(`<!DOCTYPE html><html><head><meta charset="utf-8"><title>`)
	b.WriteString(template.HTMLEscapeString(orm.AsString(move["name"])))
	b.WriteString(`</title><style>
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
</style></head><body>`)
	b.WriteString(`<button onclick="window.print()">Print</button>`)
	b.WriteString(`<h1>` + template.HTMLEscapeString(orm.AsString(move["name"])) + `</h1>`)
	b.WriteString(`<div class="meta">`)
	b.WriteString(template.HTMLEscapeString(orm.AsString(move["move_type"])))
	b.WriteString(` · `)
	b.WriteString(template.HTMLEscapeString(partnerName))
	b.WriteString(` · `)
	b.WriteString(template.HTMLEscapeString(orm.AsString(move["invoice_date"])))
	b.WriteString(`</div>`)
	b.WriteString(`<table><thead><tr><th>Description</th><th class="num">Qty</th><th class="num">Unit</th><th class="num">Subtotal</th></tr></thead><tbody>`)
	for _, ln := range lines {
		b.WriteString(`<tr><td>` + template.HTMLEscapeString(orm.AsString(ln["name"])) + `</td>`)
		b.WriteString(fmt.Sprintf(`<td class="num">%v</td>`, ln["quantity"]))
		b.WriteString(fmt.Sprintf(`<td class="num">%v</td>`, ln["price_unit"]))
		b.WriteString(fmt.Sprintf(`<td class="num">%v</td></tr>`, ln["price_subtotal"]))
	}
	b.WriteString(`</tbody></table><div class="totals">`)
	b.WriteString(fmt.Sprintf(`<div><span>Untaxed</span><span>%v</span></div>`, move["amount_untaxed"]))
	b.WriteString(fmt.Sprintf(`<div><span>Tax</span><span>%v</span></div>`, move["amount_tax"]))
	b.WriteString(fmt.Sprintf(`<div class="grand"><span>Total</span><span>%v</span></div>`, move["amount_total"]))
	b.WriteString(fmt.Sprintf(`<div><span>Amount Due</span><span>%v</span></div>`, move["amount_residual"]))
	b.WriteString(`</div></body></html>`)
	_, _ = w.Write([]byte(b.String()))
}
