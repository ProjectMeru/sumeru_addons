package test

import (
	"context"
	"fmt"
	"os"
	"testing"

	_ "sumeru/addons/base"
	_ "sumeru/addons/contacts"
	_ "sumeru_addons/account"
	_ "sumeru_addons/product"

	"sumeru/core/orm"
	"sumeru/core/server"
	"sumeru/core/server/config"
	"sumeru_addons/account/services"
)

func TestPostAndPayDemoInvoice(t *testing.T) {
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
	rows, err := orm.Search(ctx, "account.move", [][]interface{}{{"name", "=", "INV/DEMO/0001"}})
	if err != nil || len(rows) == 0 {
		t.Fatal("demo invoice missing — run -u account first")
	}
	id, _ := orm.CoerceInt64(rows[0]["id"])
	_ = orm.UpdateRecordByID(ctx, "account.move", int(id), map[string]interface{}{
		"state": "draft", "payment_state": "not_paid",
	})
	lines, _ := orm.Search(ctx, "account.move.line", [][]interface{}{{"move_id", "=", id}})
	for _, ln := range lines {
		dt := orm.AsString(ln["display_type"])
		if dt == "entry" || dt == "tax" {
			lid, _ := orm.CoerceInt64(ln["id"])
			_ = orm.Unlink(ctx, "account.move.line", int(lid))
		}
	}

	if err := services.PostMove(ctx, int(id)); err != nil {
		t.Fatalf("PostMove: %v", err)
	}
	move, _ := orm.SearchOne(ctx, "account.move", map[string]interface{}{"id": id})
	if orm.AsString(move["state"]) != "posted" {
		t.Fatalf("expected posted, got %v", move["state"])
	}
	t.Logf("total=%v tax=%v residual=%v", move["amount_total"], move["amount_tax"], move["amount_residual"])

	if _, err := orm.RunObjectAction(ctx, "account.move", int(id), "action_register_payment", nil); err != nil {
		t.Fatalf("register payment: %v", err)
	}
	wizards, _ := orm.Search(ctx, "account.payment.register", [][]interface{}{{"invoice_id", "=", id}})
	if len(wizards) == 0 {
		t.Fatal("wizard not created")
	}
	wizID, _ := orm.CoerceInt64(wizards[len(wizards)-1]["id"])
	if _, err := orm.RunObjectAction(ctx, "account.payment.register", int(wizID), "action_create_payments", nil); err != nil {
		t.Fatalf("create payment: %v", err)
	}
	move, _ = orm.SearchOne(ctx, "account.move", map[string]interface{}{"id": id})
	t.Logf("payment_state=%v residual=%v", move["payment_state"], move["amount_residual"])
	if orm.AsString(move["payment_state"]) != "paid" {
		t.Fatalf("expected paid, got %v residual=%v", move["payment_state"], move["amount_residual"])
	}
}
