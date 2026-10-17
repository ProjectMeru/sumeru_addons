package test

import (
	"context"
	"fmt"
	"testing"

	"sumeru/core/orm"

	"sumeru_addons/product/services"
)

// wizardCleanup removes wizard lines then the wizard itself.
func wizardCleanup(t *testing.T, wizardID int) {
	t.Helper()
	bypass := orm.ContextWithBypass(context.Background(), true)
	_, _ = orm.UnlinkWhere(bypass, "product.configurator.line",
		[][]interface{}{{"configurator_id", "=", wizardID}})
	_ = orm.Unlink(bypass, "product.configurator", wizardID)
}

// demoTemplateID returns the seeded demo template id.
func demoTemplateID(t *testing.T, ctx context.Context) int {
	t.Helper()
	tmpl, err := orm.SearchOne(ctx, "product.template", map[string]interface{}{"name": "Demo T-Shirt"})
	if err != nil || tmpl == nil {
		t.Fatal("demo template missing")
	}
	id, _ := orm.CoerceInt64(tmpl["id"])
	return int(id)
}

// setLineValue sets the chosen value on the wizard line of an attribute.
func setLineValue(t *testing.T, ctx context.Context, wizardID, attributeID, valueID int) {
	t.Helper()
	rows, err := orm.Search(ctx, "product.configurator.line",
		[][]interface{}{{"configurator_id", "=", wizardID}, {"attribute_id", "=", attributeID}})
	if err != nil || len(rows) == 0 {
		t.Fatalf("wizard line for attribute %d not found", attributeID)
	}
	id, _ := orm.CoerceInt64(rows[0]["id"])
	if err := orm.UpdateRecordByID(ctx, "product.configurator.line", int(id),
		map[string]interface{}{"value_id": valueID}); err != nil {
		t.Fatalf("set wizard line value: %v", err)
	}
}

// TestOpenConfiguratorPrepopulates opens the dialog with one line per active
// attribute of the template.
func TestOpenConfiguratorPrepopulates(t *testing.T) {
	ctx := bootstrapTest(t)
	if err := services.EnsureDemoVariantCatalog(ctx); err != nil {
		t.Fatalf("seed demo catalog: %v", err)
	}
	tmplID := demoTemplateID(t, ctx)
	wizardID, err := services.OpenConfigurator(ctx, tmplID)
	if err != nil {
		t.Fatalf("OpenConfigurator: %v", err)
	}
	t.Cleanup(func() { wizardCleanup(t, wizardID) })
	if n := searchCount(t, ctx, "product.configurator.line",
		[][]interface{}{{"configurator_id", "=", wizardID}}); n != 2 {
		t.Fatalf("expected 2 wizard lines, got %d", n)
	}
}

// TestApplyResolvesExistingVariant resolves a full combination to the
// existing variant without creating a duplicate.
func TestApplyResolvesExistingVariant(t *testing.T) {
	ctx := bootstrapTest(t)
	if err := services.EnsureDemoVariantCatalog(ctx); err != nil {
		t.Fatalf("seed demo catalog: %v", err)
	}
	tmplID := demoTemplateID(t, ctx)
	before := searchCount(t, ctx, "product.product",
		[][]interface{}{{"product_tmpl_id", "=", tmplID}})

	colorAttr, err := orm.SearchOne(ctx, "product.attribute", map[string]interface{}{"name": "Color"})
	if err != nil || colorAttr == nil {
		t.Fatal("Color attribute missing")
	}
	sizeAttr, err := orm.SearchOne(ctx, "product.attribute", map[string]interface{}{"name": "Size"})
	if err != nil || sizeAttr == nil {
		t.Fatal("Size attribute missing")
	}
	colorID, _ := orm.CoerceInt64(colorAttr["id"])
	sizeID, _ := orm.CoerceInt64(sizeAttr["id"])

	wizardID, err := services.OpenConfigurator(ctx, tmplID)
	if err != nil {
		t.Fatalf("OpenConfigurator: %v", err)
	}
	t.Cleanup(func() { wizardCleanup(t, wizardID) })
	setLineValue(t, ctx, wizardID, int(colorID), findPav(t, ctx, int(colorID), "Red"))
	setLineValue(t, ctx, wizardID, int(sizeID), findPav(t, ctx, int(sizeID), "M"))

	variantID, err := services.ApplyConfigurator(ctx, wizardID)
	if err != nil {
		t.Fatalf("ApplyConfigurator: %v", err)
	}
	variant, err := orm.SearchOne(ctx, "product.product", map[string]interface{}{"id": variantID})
	if err != nil || variant == nil {
		t.Fatal("applied variant missing")
	}
	if got := orm.AsString(variant["combination"]); got != "Red, M" {
		t.Fatalf("combination = %q, want %q", got, "Red, M")
	}
	if after := searchCount(t, ctx, "product.product",
		[][]interface{}{{"product_tmpl_id", "=", tmplID}}); after != before {
		t.Fatalf("apply must not duplicate variants: %d -> %d", before, after)
	}
}

// TestApplyDynamicCreatesVariant creates the variant on demand for templates
// with a dynamic attribute, materializing the missing template value.
func TestApplyDynamicCreatesVariant(t *testing.T) {
	ctx := bootstrapTest(t)
	suffix := uniqueSuffix()
	attrID := createTestAttribute(t, ctx, "TDynAttr"+suffix, "dynamic", map[string]float64{"D1": 3000, "D2": 3000})
	tmplID := createTestTemplate(t, ctx, "TDyn"+suffix, 90000, map[int][]string{attrID: {"D1"}})

	// Dynamic templates never get a generated matrix.
	if _, _, err := services.GenerateVariants(ctx, tmplID); err != nil {
		t.Fatalf("GenerateVariants: %v", err)
	}
	if n := searchCount(t, ctx, "product.product",
		[][]interface{}{{"product_tmpl_id", "=", tmplID}}); n != 0 {
		t.Fatalf("dynamic template must not generate variants, found %d", n)
	}

	wizardID, err := services.OpenConfigurator(ctx, tmplID)
	if err != nil {
		t.Fatalf("OpenConfigurator: %v", err)
	}
	t.Cleanup(func() { wizardCleanup(t, wizardID) })
	// D2 has no materialized template value yet: dynamic templates create it.
	setLineValue(t, ctx, wizardID, attrID, findPav(t, ctx, attrID, "D2"))

	variantID, err := services.ApplyConfigurator(ctx, wizardID)
	if err != nil {
		t.Fatalf("ApplyConfigurator: %v", err)
	}
	variant, err := orm.SearchOne(ctx, "product.product", map[string]interface{}{"id": variantID})
	if err != nil || variant == nil {
		t.Fatal("dynamic variant missing")
	}
	if got := orm.AsString(variant["combination"]); got != "D2" {
		t.Fatalf("combination = %q, want %q", got, "D2")
	}
	if got := numeric(variant["list_price"]); got != 93000 {
		t.Fatalf("list_price = %v, want 93000", got)
	}
}

// TestApplyNonDynamicMissingErrors refuses combinations that cannot exist on
// non-dynamic templates.
func TestApplyNonDynamicMissingErrors(t *testing.T) {
	ctx := bootstrapTest(t)
	suffix := uniqueSuffix()
	attrID := createTestAttribute(t, ctx, "TNDAttr"+suffix, "always", map[string]float64{"A": 0, "B": 0})
	tmplID := createTestTemplate(t, ctx, "TND"+suffix, 100000, map[int][]string{attrID: {"A"}})
	if _, _, err := services.GenerateVariants(ctx, tmplID); err != nil {
		t.Fatalf("GenerateVariants: %v", err)
	}

	wizardID, err := services.OpenConfigurator(ctx, tmplID)
	if err != nil {
		t.Fatalf("OpenConfigurator: %v", err)
	}
	t.Cleanup(func() { wizardCleanup(t, wizardID) })
	// Value B has no materialized template value: the combination cannot exist.
	setLineValue(t, ctx, wizardID, attrID, findPav(t, ctx, attrID, "B"))
	if _, err := services.ApplyConfigurator(ctx, wizardID); err == nil {
		t.Fatal("expected error for missing non-dynamic combination")
	}
}

// TestConfiguratorLineConstraint rejects a value from another attribute.
func TestConfiguratorLineConstraint(t *testing.T) {
	ctx := bootstrapTest(t)
	suffix := uniqueSuffix()
	attr1 := createTestAttribute(t, ctx, "TCLAttr1"+suffix, "always", map[string]float64{"X": 0})
	attr2 := createTestAttribute(t, ctx, "TCLAttr2"+suffix, "always", map[string]float64{"Y": 0})
	wizardID := mustCreate(t, ctx, "product.configurator", map[string]interface{}{
		"product_tmpl_id": mustCreate(t, ctx, "product.template", map[string]interface{}{
			"name": "TCL" + suffix, "type": "consu", "active": true,
		}),
	})
	t.Cleanup(func() { wizardCleanup(t, wizardID) })

	mustFailCreate(t, ctx, "product.configurator.line", map[string]interface{}{
		"configurator_id": wizardID, "attribute_id": attr1,
		"value_id": findPav(t, ctx, attr2, "Y"),
	})
}

// TestOnchangeHandlers exercises the wizard line onchange wiring directly.
func TestOnchangeHandlers(t *testing.T) {
	ctx := bootstrapTest(t)
	suffix := uniqueSuffix()
	attrID := createTestAttribute(t, ctx, "TOChAttr"+suffix, "always", map[string]float64{"X": 2500})
	pavID := findPav(t, ctx, attrID, "X")

	// attribute change clears the stale value and price.
	res, err := orm.RunOnchange(ctx, "product.configurator.line", "attribute_id", map[string]interface{}{
		"attribute_id": attrID, "value_id": pavID, "price_extra": 2500,
	})
	if err != nil {
		t.Fatalf("onchange attribute_id: %v", err)
	}
	if got, _ := orm.CoerceInt64(res.Value["value_id"]); got != 0 {
		t.Fatalf("value_id must be cleared, got %v", res.Value["value_id"])
	}

	// value change fills the default extra price hint.
	res, err = orm.RunOnchange(ctx, "product.configurator.line", "value_id", map[string]interface{}{
		"value_id": pavID,
	})
	if err != nil {
		t.Fatalf("onchange value_id: %v", err)
	}
	if got := numeric(res.Value["price_extra"]); got != 2500 {
		t.Fatalf("price_extra hint = %v, want 2500", got)
	}
}

// TestWizardTotalPriceCompute sums the template price and line extras.
func TestWizardTotalPriceCompute(t *testing.T) {
	ctx := bootstrapTest(t)
	if err := services.EnsureDemoVariantCatalog(ctx); err != nil {
		t.Fatalf("seed demo catalog: %v", err)
	}
	tmplID := demoTemplateID(t, ctx)
	wizardID, err := services.OpenConfigurator(ctx, tmplID)
	if err != nil {
		t.Fatalf("OpenConfigurator: %v", err)
	}
	t.Cleanup(func() { wizardCleanup(t, wizardID) })

	lines, err := orm.Search(ctx, "product.configurator.line",
		[][]interface{}{{"configurator_id", "=", wizardID}})
	if err != nil || len(lines) == 0 {
		t.Fatal("wizard lines missing")
	}
	for i, l := range lines {
		id, _ := orm.CoerceInt64(l["id"])
		if err := orm.UpdateRecordByID(ctx, "product.configurator.line", int(id),
			map[string]interface{}{"price_extra": 5000 * float64(i+1)}); err != nil {
			t.Fatalf("set line price: %v", err)
		}
	}
	wiz, err := orm.SearchOne(ctx, "product.configurator", map[string]interface{}{"id": wizardID})
	if err != nil || wiz == nil {
		t.Fatal("wizard missing")
	}
	if got := numeric(wiz["total_price"]); got != 165000 {
		t.Fatalf("total_price = %v, want %v", got, fmt.Sprintf("165000"))
	}
}
