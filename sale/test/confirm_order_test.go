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

	"sumeru/core/orm"
	"sumeru/core/server"
	"sumeru/core/server/config"
)

func TestConfirmSaleOrder(t *testing.T) {
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

	rows, err := orm.SearchLimit(ctx, "sale.order", [][]interface{}{{"state", "=", "draft"}}, 1)
	if err != nil || len(rows) == 0 {
		t.Skip("no draft sale.order — install sale module first")
	}
	id, _ := orm.CoerceInt64(rows[0]["id"])
	if err := orm.UpdateRecordByID(ctx, "sale.order", int(id), map[string]interface{}{"state": "sale"}); err != nil {
		t.Fatalf("confirm order: %v", err)
	}
	order, _ := orm.SearchOne(ctx, "sale.order", map[string]interface{}{"id": id})
	if orm.AsString(order["state"]) != "sale" {
		t.Fatalf("expected sale state, got %v", order["state"])
	}
}
