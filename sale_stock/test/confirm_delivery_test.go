package test

import (
	"context"
	"fmt"
	"os"
	"testing"

	_ "sumeru/addons/base"
	_ "sumeru/addons/contacts"
	_ "sumeru_addons/product"
	_ "sumeru_addons/sale"
	_ "sumeru_addons/sale_stock"
	_ "sumeru_addons/stock"

	"sumeru/core/orm"
	"sumeru/core/server"
	"sumeru/core/server/config"
)

func TestConfirmSaleCreatesDelivery(t *testing.T) {
	conf := os.Getenv("SUMERU_CONF")
	if conf == "" {
		t.Skip("set SUMERU_CONF to sumeru.conf path")
	}
	if err := server.LoadConfig(conf); err != nil {
		t.Fatal(err)
	}
	c := config.AppConfig
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DbHost, c.DbPort, c.DbUser, c.DbPass, c.DbName, c.DbSslMode)
	server.InitDB(dsn)
	ctx := orm.ContextWithBypass(context.Background(), true)

	id, name := findDraftOrderWithStorableLine(t, ctx)
	if id <= 0 {
		t.Skip("no draft sale.order with a storable line — install sale/stock modules first")
	}

	if _, err := orm.RunObjectAction(ctx, "sale.order", id, "action_confirm", nil); err != nil {
		t.Fatalf("confirm order: %v", err)
	}
	order, _ := orm.SearchOne(ctx, "sale.order", map[string]interface{}{"id": id})
	if orm.AsString(order["state"]) != "sale" {
		t.Fatalf("expected sale state, got %v", order["state"])
	}
	finalName := orm.AsString(order["name"])
	pickings, err := orm.Search(ctx, "stock.picking", [][]interface{}{
		{"origin", "=", finalName},
		{"picking_type_code", "=", "outgoing"},
	})
	if err != nil || len(pickings) == 0 {
		t.Fatalf("expected outgoing stock.picking for %q (name=%q), got %d", finalName, name, len(pickings))
	}
}

func findDraftOrderWithStorableLine(t *testing.T, ctx context.Context) (int, string) {
	t.Helper()
	orders, err := orm.SearchLimit(ctx, "sale.order", [][]interface{}{{"state", "=", "draft"}}, 10)
	if err != nil || len(orders) == 0 {
		return 0, ""
	}
	for _, o := range orders {
		oid, _ := orm.CoerceInt64(o["id"])
		lines, err := orm.Search(ctx, "sale.order.line", [][]interface{}{{"order_id", "=", oid}})
		if err != nil {
			continue
		}
		for _, ln := range lines {
			pid, ok := orm.CoerceInt64(ln["product_id"])
			if !ok || pid <= 0 {
				continue
			}
			p, err := orm.SearchOne(ctx, "product.product", map[string]interface{}{"id": pid})
			if err == nil && orm.AsString(p["type"]) == "consu" {
				return int(oid), orm.AsString(o["name"])
			}
		}
	}
	return 0, ""
}
