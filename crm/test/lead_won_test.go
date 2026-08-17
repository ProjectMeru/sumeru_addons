package test

import (
	"context"
	"fmt"
	"os"
	"testing"

	_ "sumeru/addons/base"
	_ "sumeru/addons/contacts"
	_ "sumeru/addons/mail"
	_ "sumeru_addons/crm"
	_ "sumeru_addons/product"
	_ "sumeru_addons/utm"

	"sumeru/core/orm"
	"sumeru/core/server"
	"sumeru/core/server/config"
)

func TestLeadSetWon(t *testing.T) {
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

	rows, err := orm.SearchLimit(ctx, "crm.lead", [][]interface{}{{"active", "=", true}}, 1)
	if err != nil || len(rows) == 0 {
		t.Skip("no crm.lead rows — install crm and seed data first")
	}
	id, _ := orm.CoerceInt64(rows[0]["id"])
	if _, err := orm.RunObjectAction(ctx, "crm.lead", int(id), "action_set_won", nil); err != nil {
		t.Fatalf("action_set_won: %v", err)
	}
	lead, _ := orm.SearchOne(ctx, "crm.lead", map[string]interface{}{"id": id})
	if orm.AsString(lead["won_status"]) != "won" {
		t.Fatalf("expected won, got %v", lead["won_status"])
	}
}
