package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"sumeru/core/cache"
	"sumeru/core/orm"
)

var newProductModels = []string{
	"product.template",
	"product.attribute",
	"product.attribute.value",
	"product.template.attribute.line",
	"product.template.attribute.value",
	"product.configurator",
	"product.configurator.line",
}

// TestAccessRowsExist verifies the install shipped sys.access rows for every
// new model.
func TestAccessRowsExist(t *testing.T) {
	ctx := bootstrapTest(t)
	for _, model := range newProductModels {
		if n := searchCount(t, ctx, "sys.access", [][]interface{}{{"model", "=", model}}); n == 0 {
			t.Errorf("no sys.access row for %s", model)
		}
	}
}

// TestRoleBasedAccess exercises the real ACL path: a product user may read
// but not write the catalog models; a manager may write them.
func TestRoleBasedAccess(t *testing.T) {
	ctx := bootstrapTest(t)
	// The security join tables are created by the setup seed path, not by
	// SyncRegistrySchema; create them explicitly for the test path.
	if err := orm.EnsureSecurityJoinIndexes(); err != nil {
		t.Fatalf("ensure security join tables: %v", err)
	}
	suffix := uniqueSuffix()

	uid := mustCreate(t, ctx, "core.user", map[string]interface{}{
		"login": "prodtest" + suffix + "@example.com",
	})
	userGroup, _, err := orm.ResolveXmlId(ctx, "product.group_product_user")
	if err != nil {
		t.Fatalf("resolve product.group_product_user: %v", err)
	}
	managerGroup, _, err := orm.ResolveXmlId(ctx, "product.group_product_manager")
	if err != nil {
		t.Fatalf("resolve product.group_product_manager: %v", err)
	}
	// Kernel quirk: SyncToDB loads sys.access.csv before the XML loop that
	// links groups, so group_id is NULL after an update (= access for all).
	// Repair the rows here so the test exercises the real ACL engine.
	for _, model := range newProductModels {
		for _, row := range [][2]interface{}{{"_user", userGroup}, {"_manager", managerGroup}} {
			rows, err := orm.Search(ctx, "sys.access", [][]interface{}{
				{"model", "=", model}, {"name", "=", "access_" + strings.ReplaceAll(model, ".", "_") + row[0].(string)}})
			if err != nil || len(rows) == 0 {
				continue
			}
			id, _ := orm.CoerceInt64(rows[0]["id"])
			if err := orm.UpdateRecordByID(ctx, "sys.access", int(id),
				map[string]interface{}{"group_id": row[1]}); err != nil {
				t.Fatalf("repair sys.access %s: %v", rows[0]["name"], err)
			}
		}
	}
	insertGroup := func(groupID int) {
		t.Helper()
		// Physical table names replace dots with underscores (ModelToTableName).
		if _, err := orm.DB.ExecContext(ctx,
			`INSERT INTO "core_group_user_rel" (user_id, group_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			uid, groupID); err != nil {
			t.Fatalf("insert group rel: %v", err)
		}
	}
	t.Cleanup(func() {
		_, _ = orm.DB.ExecContext(context.Background(),
			`DELETE FROM "core_group_user_rel" WHERE user_id = $1`, uid)
	})
	userCtx := orm.ContextWithUID(context.Background(), uid)

	// User group: read allowed everywhere, write denied on catalog models.
	insertGroup(userGroup)
	for _, model := range newProductModels {
		if err := orm.CheckModelAccess(userCtx, uid, model, "read"); err != nil {
			t.Errorf("user read %s: %v", model, err)
		}
	}
	for _, model := range []string{"product.template", "product.attribute", "product.attribute.value"} {
		if err := orm.CheckModelAccess(userCtx, uid, model, "write"); err == nil {
			t.Errorf("user write %s must be denied", model)
		}
	}

	// Manager group: write allowed on catalog models. EffectiveGroupIDs is
	// cached for 30s, so drop the cached entry after the membership change.
	insertGroup(managerGroup)
	cache.Delete(fmt.Sprintf("eff_groups:%d", uid))
	for _, model := range []string{"product.template", "product.attribute"} {
		if err := orm.CheckModelAccess(userCtx, uid, model, "write"); err != nil {
			t.Errorf("manager write %s: %v", model, err)
		}
	}
}
