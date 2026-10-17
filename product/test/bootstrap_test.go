package test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	_ "sumeru/addons/base"
	_ "sumeru/addons/i18n"
	_ "sumeru_addons/product"

	"sumeru/core/modelreg"
	"sumeru/core/orm"
	"sumeru/core/server"
	"sumeru/core/server/config"
)

// bootstrapTest prepares the shared dev DB for product tests (same pattern as
// the sale/stock/account test suites). Skipped without SUMERU_CONF.
func bootstrapTest(t *testing.T) context.Context {
	t.Helper()
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
	if err := modelreg.ActivateAll([]string{"base", "i18n", "product"}); err != nil {
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

// uniqueSuffix keeps test record names globally distinct.
func uniqueSuffix() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// mustCreate creates a record, fails the test on error, and unlinks it during
// cleanup (LIFO: create children after parents so cleanup unlinks them first).
func mustCreate(t *testing.T, ctx context.Context, model string, values map[string]interface{}) int {
	t.Helper()
	id, err := orm.Create(ctx, orm.RegistryModel(model), values)
	if err != nil {
		t.Fatalf("create %s: %v", model, err)
	}
	t.Cleanup(func() {
		_ = orm.Unlink(orm.ContextWithBypass(context.Background(), true), model, id)
	})
	return id
}

// cleanupProducts unlinks every product linked to a template (generated
// variants are not tracked by mustCreate).
func cleanupProducts(t *testing.T, templateID int) {
	t.Helper()
	_, err := orm.UnlinkWhere(orm.ContextWithBypass(context.Background(), true), "product.product",
		[][]interface{}{{"product_tmpl_id", "=", templateID}})
	if err != nil {
		t.Logf("cleanup products of template %d: %v", templateID, err)
	}
}

// createTestAttribute creates an attribute with the given mode and values.
func createTestAttribute(t *testing.T, ctx context.Context, name, createVariant string, values map[string]float64) int {
	t.Helper()
	attrID := mustCreate(t, ctx, "product.attribute", map[string]interface{}{
		"name":           name,
		"display_type":   "select",
		"create_variant": createVariant,
		"active":         true,
	})
	for vname, extra := range values {
		mustCreate(t, ctx, "product.attribute.value", map[string]interface{}{
			"name":               vname,
			"attribute_id":       attrID,
			"default_extra_price": extra,
			"active":             true,
		})
	}
	return attrID
}

// findPav looks up an attribute value by (attribute, name).
func findPav(t *testing.T, ctx context.Context, attributeID int, name string) int {
	t.Helper()
	rows, err := orm.Search(ctx, "product.attribute.value",
		[][]interface{}{{"attribute_id", "=", attributeID}, {"name", "=", name}})
	if err != nil || len(rows) == 0 {
		t.Fatalf("attribute value %q not found", name)
	}
	id, _ := orm.CoerceInt64(rows[0]["id"])
	return int(id)
}

// createTestTemplate creates a template plus a ptav for each (attribute,
// value) pair, then generates the matrix.
func createTestTemplate(t *testing.T, ctx context.Context, name string, listPrice float64, ptavSpecs map[int][]string) int {
	t.Helper()
	tmplID := mustCreate(t, ctx, "product.template", map[string]interface{}{
		"name":       name,
		"type":       "consu",
		"list_price": listPrice,
		"sale_ok":    true,
		"purchase_ok": true,
		"active":     true,
	})
	t.Cleanup(func() { cleanupProducts(t, tmplID) })
	for attrID, vnames := range ptavSpecs {
		for _, vname := range vnames {
			mustCreate(t, ctx, "product.template.attribute.value", map[string]interface{}{
				"product_tmpl_id":           tmplID,
				"attribute_id":              attrID,
				"product_attribute_value_id": findPav(t, ctx, attrID, vname),
				"active":                    true,
			})
		}
	}
	return tmplID
}

// searchCount is a shorthand for domain counts.
func searchCount(t *testing.T, ctx context.Context, model string, domain [][]interface{}) int {
	t.Helper()
	n, err := orm.SearchCount(ctx, model, domain)
	if err != nil {
		t.Fatalf("count %s: %v", model, err)
	}
	return n
}

// numeric coerces DB numeric values for assertions.
func numeric(v interface{}) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case int:
		return float64(t)
	case int64:
		return float64(t)
	default:
		var f float64
		_, _ = fmt.Sscanf(orm.AsString(v), "%f", &f)
		return f
	}
}

// mustFailCreate asserts that creating a record fails (constraint tests).
func mustFailCreate(t *testing.T, ctx context.Context, model string, values map[string]interface{}) {
	t.Helper()
	if id, err := orm.Create(ctx, orm.RegistryModel(model), values); err == nil {
		t.Errorf("create %s unexpectedly succeeded (id=%d)", model, id)
	}
}
