package test

import (
	"context"
	"fmt"
	"os"
	"testing"

	_ "sumeru/addons/base"
	_ "sumeru/addons/contacts"
	_ "sumeru_addons/product"
	_ "sumeru_addons/purchase"
	_ "sumeru_addons/purchase_stock"
	_ "sumeru_addons/stock"

	"sumeru/core/modelreg"
	"sumeru/core/orm"
	"sumeru/core/server"
	"sumeru/core/server/config"
)

func TestConfirmPurchaseCreatesIncomingReceipt(t *testing.T) {
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
	if err := modelreg.ActivateAll([]string{"base", "contacts", "product", "stock", "purchase"}); err != nil {
		t.Fatal(err)
	}
	ctx := orm.ContextWithBypass(context.Background(), true)

	id := findDraftPurchaseWithStorableLine(ctx)
	if id <= 0 {
		t.Skip("no draft purchase.order with a storable line — install purchase_stock first")
	}
	if _, err := orm.RunObjectAction(ctx, "purchase.order", id, "action_confirm", nil); err != nil {
		t.Fatalf("confirm purchase order: %v", err)
	}
	order, err := orm.SearchOne(ctx, "purchase.order", map[string]interface{}{"id": id})
	if err != nil {
		t.Fatal(err)
	}
	if orm.AsString(order["state"]) != "purchase" {
		t.Fatalf("expected purchase state, got %v", order["state"])
	}

	origin := orm.AsString(order["name"])
	pickings, err := orm.Search(ctx, "stock.picking", [][]interface{}{
		{"origin", "=", origin},
		{"picking_type_code", "=", "incoming"},
	})
	if err != nil || len(pickings) != 1 {
		t.Fatalf("expected one incoming receipt for %q, got %d", origin, len(pickings))
	}
	if _, err := orm.RunObjectAction(ctx, "purchase.order", id, "action_create_receipt", nil); err != nil {
		t.Fatalf("create receipt again: %v", err)
	}
	pickings, err = orm.Search(ctx, "stock.picking", [][]interface{}{
		{"origin", "=", origin},
		{"picking_type_code", "=", "incoming"},
	})
	if err != nil || len(pickings) != 1 {
		t.Fatalf("expected idempotent receipt creation for %q, got %d", origin, len(pickings))
	}
}

func findDraftPurchaseWithStorableLine(ctx context.Context) int {
	orders, err := orm.SearchLimit(ctx, "purchase.order", [][]interface{}{{"state", "=", "draft"}}, 10)
	if err != nil {
		return 0
	}
	for _, order := range orders {
		orderID, _ := orm.CoerceInt64(order["id"])
		lines, err := orm.Search(ctx, "purchase.order.line", [][]interface{}{{"order_id", "=", orderID}})
		if err != nil {
			continue
		}
		for _, line := range lines {
			productID, ok := orm.CoerceInt64(line["product_id"])
			if !ok || productID <= 0 {
				continue
			}
			product, err := orm.SearchOne(ctx, "product.product", map[string]interface{}{"id": productID})
			if err == nil && orm.AsString(product["type"]) == "consu" {
				return int(orderID)
			}
		}
	}
	return 0
}
