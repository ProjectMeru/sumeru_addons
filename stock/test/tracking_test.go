package test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	_ "sumeru/addons/base"
	_ "sumeru/addons/contacts"
	_ "sumeru_addons/product"
	_ "sumeru_addons/stock"

	"sumeru/core/modelreg"
	"sumeru/core/orm"
	"sumeru/core/server"
	"sumeru/core/server/config"

	stocksvc "sumeru_addons/stock/services"
)

const (
	locInternal = 4 // Physical Locations / Stock
	ptInternal  = 3 // Internal Transfers
)

func setupStockTest(t *testing.T) context.Context {
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
	if err := modelreg.ActivateAll([]string{"base", "contacts", "product", "stock"}); err != nil {
		t.Fatal(err)
	}
	if err := server.SyncModels(); err != nil {
		t.Fatal(err)
	}
	if err := orm.SyncRegistrySchema(); err != nil {
		t.Fatal(err)
	}
	return orm.ContextWithBypass(context.Background(), true)
}

func TestValidatePickingRequiresLotForTrackedProduct(t *testing.T) {
	ctx := setupStockTest(t)

	products, err := orm.SearchLimit(ctx, "product.product", [][]interface{}{{"type", "=", "consu"}}, 1)
	if err != nil || len(products) == 0 {
		t.Skip("no storable product available")
	}
	productID, _ := orm.CoerceInt64(products[0]["id"])

	if err := orm.UpdateRecordByID(ctx, "product.product", int(productID), map[string]interface{}{"tracking": "lot"}); err != nil {
		t.Fatalf("set tracking=lot: %v", err)
	}
	defer orm.UpdateRecordByID(ctx, "product.product", int(productID), map[string]interface{}{"tracking": "none"})

	pickingID := createTrackingPicking(ctx, t, productID)
	defer cleanupPicking(ctx, pickingID)

	// No move lines: a lot-tracked product cannot be validated without a lot.
	if err := stocksvc.ValidatePicking(ctx, pickingID); err == nil {
		t.Fatal("expected error when validating a lot-tracked product without a lot")
	}

	moves, _ := orm.Search(ctx, "stock.move", [][]interface{}{{"picking_id", "=", pickingID}})
	if len(moves) == 0 {
		t.Fatal("no move created")
	}
	moveID, _ := orm.CoerceInt64(moves[0]["id"])

	lotName := fmt.Sprintf("LOT-TEST-%d", time.Now().UnixNano())
	lineModel := orm.RegistryModel("stock.move.line")
	if lineModel == nil {
		t.Fatal("stock.move.line not registered")
	}
	if _, err := orm.Create(ctx, lineModel, map[string]interface{}{
		"move_id":    moveID,
		"product_id": productID,
		"quantity":   1,
		"lot_name":   lotName,
	}); err != nil {
		t.Fatal(err)
	}

	if err := stocksvc.ValidatePicking(ctx, pickingID); err != nil {
		t.Fatalf("validate with lot: %v", err)
	}

	lots, _ := orm.Search(ctx, "stock.lot", [][]interface{}{
		{"name", "=", lotName},
		{"product_id", "=", productID},
	})
	if len(lots) != 1 {
		t.Fatalf("expected 1 lot created, got %d", len(lots))
	}
	lotID, _ := orm.CoerceInt64(lots[0]["id"])
	defer orm.UnlinkWhere(ctx, "stock.quant", [][]interface{}{{"lot_id", "=", lotID}})
	defer orm.Unlink(ctx, "stock.lot", int(lotID))
}

func createTrackingPicking(ctx context.Context, t *testing.T, productID int64) int {
	t.Helper()
	pickModel := orm.RegistryModel("stock.picking")
	if pickModel == nil {
		t.Fatal("stock.picking not registered")
	}
	pickID, err := orm.Create(ctx, pickModel, map[string]interface{}{
		"state":             "confirmed",
		"picking_type_id":   ptInternal,
		"picking_type_code": "internal",
		"location_id":       locInternal,
		"location_dest_id":  locInternal,
		"move_type":         "direct",
		"origin":            "TRACKING-TEST",
	})
	if err != nil {
		t.Fatal(err)
	}
	moveModel := orm.RegistryModel("stock.move")
	if moveModel == nil {
		t.Fatal("stock.move not registered")
	}
	if _, err := orm.Create(ctx, moveModel, map[string]interface{}{
		"name":             "Tracking test move",
		"product_id":       productID,
		"product_qty":      1,
		"quantity":         1,
		"location_id":      locInternal,
		"location_dest_id": locInternal,
		"picking_id":       pickID,
		"picking_type_id":  ptInternal,
		"state":            "confirmed",
		"origin":           "TRACKING-TEST",
	}); err != nil {
		t.Fatal(err)
	}
	return pickID
}

func cleanupPicking(ctx context.Context, pickingID int) {
	orm.UnlinkWhere(ctx, "stock.move.line", [][]interface{}{{"picking_id", "=", pickingID}})
	orm.UnlinkWhere(ctx, "stock.move", [][]interface{}{{"picking_id", "=", pickingID}})
	orm.Unlink(ctx, "stock.picking", pickingID)
}
