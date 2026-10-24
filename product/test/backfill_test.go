package test

import (
	"fmt"
	"testing"

	"sumeru/core/orm"

	"sumeru_addons/product/services"
)

// TestBackfillLinksOrphanProduct links a standalone product to its matching
// template and stays idempotent across runs.
func TestBackfillLinksOrphanProduct(t *testing.T) {
	ctx := bootstrapTest(t)
	suffix := uniqueSuffix()
	name := "BF" + suffix
	productID := mustCreate(t, ctx, "product.product", map[string]interface{}{
		"name": name, "type": "consu",
	})
	tmplID := mustCreate(t, ctx, "product.template", map[string]interface{}{
		"name": name, "type": "consu", "active": true,
	})

	linked, err := services.BackfillTemplateProducts(ctx, tmplID)
	if err != nil {
		t.Fatalf("BackfillTemplateProducts: %v", err)
	}
	if linked != 1 {
		t.Fatalf("expected 1 linked product, got %d", linked)
	}
	product, err := orm.SearchOne(ctx, "product.product", map[string]interface{}{"id": productID})
	if err != nil || product == nil {
		t.Fatal("product missing after backfill")
	}
	tmpl, _ := orm.CoerceInt64(product["product_tmpl_id"])
	if int(tmpl) != tmplID {
		t.Fatalf("product_tmpl_id = %d, want %d", tmpl, tmplID)
	}
	if got := orm.AsString(product["combination_indices"]); got != "" {
		t.Fatalf("combination_indices = %q, want empty", got)
	}
	if got := orm.AsString(product["variant_key"]); got != fmt.Sprintf("%d:", tmplID) {
		t.Fatalf("variant_key = %q, want %q", got, fmt.Sprintf("%d:", tmplID))
	}

	linked, err = services.BackfillTemplateProducts(ctx, tmplID)
	if err != nil {
		t.Fatalf("second backfill: %v", err)
	}
	if linked != 0 {
		t.Fatalf("second run must link 0, got %d", linked)
	}
}

// TestBackfillIgnoresDifferentName leaves name-mismatched products untouched.
func TestBackfillIgnoresDifferentName(t *testing.T) {
	ctx := bootstrapTest(t)
	suffix := uniqueSuffix()
	mustCreate(t, ctx, "product.product", map[string]interface{}{
		"name": "BFX" + suffix, "type": "consu",
	})
	tmplID := mustCreate(t, ctx, "product.template", map[string]interface{}{
		"name": "BFY" + suffix, "type": "consu", "active": true,
	})

	linked, err := services.BackfillTemplateProducts(ctx, tmplID)
	if err != nil {
		t.Fatalf("BackfillTemplateProducts: %v", err)
	}
	if linked != 0 {
		t.Fatalf("expected 0 linked products, got %d", linked)
	}
}
