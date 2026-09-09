package controllers

import (
	"bytes"
	"context"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"sumeru/core/orm"
	"sumeru/core/server/router"
)

func init() {
	router.Register(http.MethodGet, "/stock/picking/print", router.AuthSession, PickingPrintHandler)
}

var pickingPrintTmpl = template.Must(template.New("picking").Parse(`<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>{{.Name}}</title>
<style>
body{font-family:system-ui,sans-serif;margin:2rem;color:#111}
h1{font-size:1.5rem;margin:0 0 .25rem}
.meta{color:#555;margin-bottom:1.5rem}
table{width:100%;border-collapse:collapse;margin-top:1rem}
th,td{border-bottom:1px solid #ddd;padding:.5rem;text-align:left}
th{font-size:.75rem;text-transform:uppercase;color:#666}
.num{text-align:right}
@media print{button{display:none}}
</style></head><body>
<button onclick="window.print()">Print</button>
<h1>{{.Name}}</h1>
<div class="meta">{{.Type}} · {{.Partner}} · {{.ScheduledDate}}</div>
<div class="meta">{{.From}} → {{.To}}</div>
<table><thead><tr><th>Product</th><th class="num">Qty</th></tr></thead><tbody>
{{range .Moves}}<tr><td>{{.Product}}</td><td class="num">{{.Qty}}</td></tr>
{{end}}</tbody></table>
</body></html>`))

type pickingPrintMove struct {
	Product, Qty string
}

type pickingPrintData struct {
	Name, Type, Partner, ScheduledDate, From, To string
	Moves                                        []pickingPrintMove
}

func PickingPrintHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("id")))
	if id <= 0 {
		http.Error(w, "missing transfer id", http.StatusBadRequest)
		return
	}
	ctx := orm.ContextWithBypass(r.Context(), true)
	pick, err := orm.SearchOne(ctx, "stock.picking", map[string]interface{}{"id": id})
	if err != nil || pick == nil {
		http.Error(w, "transfer not found", http.StatusNotFound)
		return
	}
	data := pickingPrintData{
		Name:          orm.AsString(pick["name"]),
		Type:          pickingTypeName(ctx, pick["picking_type_id"]),
		Partner:       partnerName(ctx, pick["partner_id"]),
		ScheduledDate: shortDate(orm.AsString(pick["scheduled_date"])),
		From:          locationName(ctx, pick["location_id"]),
		To:            locationName(ctx, pick["location_dest_id"]),
	}
	moves, _ := orm.Search(ctx, "stock.move", [][]interface{}{{"picking_id", "=", id}})
	for _, mv := range moves {
		qty := numericText(mv["quantity"])
		if qty == "" {
			qty = numericText(mv["product_qty"])
		}
		data.Moves = append(data.Moves, pickingPrintMove{
			Product: productName(ctx, mv["product_id"]),
			Qty:     qty,
		})
	}
	if err := orm.UpdateRecordByID(ctx, "stock.picking", id, map[string]interface{}{"printed": true}); err != nil {
		// non-fatal
		_ = err
	}
	var buf bytes.Buffer
	if err := pickingPrintTmpl.Execute(&buf, data); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}

func numericText(v interface{}) string {
	if v == nil {
		return ""
	}
	s := strings.TrimSpace(orm.AsString(v))
	if s == "" || s == "0" {
		return ""
	}
	return s
}

func shortDate(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}

func pickingTypeName(ctx context.Context, v interface{}) string {
	id, ok := orm.CoerceInt64(v)
	if !ok || id <= 0 {
		return ""
	}
	pt, err := orm.SearchOne(ctx, "stock.picking.type", map[string]interface{}{"id": id})
	if err != nil || pt == nil {
		return ""
	}
	return orm.AsString(pt["name"])
}

func locationName(ctx context.Context, v interface{}) string {
	id, ok := orm.CoerceInt64(v)
	if !ok || id <= 0 {
		return ""
	}
	loc, err := orm.SearchOne(ctx, "stock.location", map[string]interface{}{"id": id})
	if err != nil || loc == nil {
		return ""
	}
	if c := orm.AsString(loc["complete_name"]); c != "" {
		return c
	}
	return orm.AsString(loc["name"])
}

func partnerName(ctx context.Context, v interface{}) string {
	id, ok := orm.CoerceInt64(v)
	if !ok || id <= 0 {
		return ""
	}
	p, err := orm.SearchOne(ctx, "core.partner", map[string]interface{}{"id": id})
	if err != nil || p == nil {
		return ""
	}
	return orm.AsString(p["name"])
}

func productName(ctx context.Context, v interface{}) string {
	id, ok := orm.CoerceInt64(v)
	if !ok || id <= 0 {
		return ""
	}
	p, err := orm.SearchOne(ctx, "product.product", map[string]interface{}{"id": id})
	if err != nil || p == nil {
		return ""
	}
	return orm.AsString(p["name"])
}
