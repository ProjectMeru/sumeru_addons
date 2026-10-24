package test

import (
	"fmt"
	"testing"

	"sumeru/core/orm"

	"sumeru_addons/product/services"
)

// TestVariantMatrixGeneration verifies the demo 2x3 matrix: six variants with
// sorted combination keys, per-combination variant keys, extra-price pricing,
// idempotent re-runs, and reconciled attribute lines.
func TestVariantMatrixGeneration(t *testing.T) {
	ctx := bootstrapTest(t)
	if err := services.EnsureDemoVariantCatalog(ctx); err != nil {
		t.Fatalf("seed demo catalog: %v", err)
	}
	tmpl, err := orm.SearchOne(ctx, "product.template", map[string]interface{}{"name": "Demo T-Shirt"})
	if err != nil || tmpl == nil {
		t.Fatal("demo template missing")
	}
	tmplID, _ := orm.CoerceInt64(tmpl["id"])

	created, archived, err := services.GenerateVariants(ctx, int(tmplID))
	if err != nil {
		t.Fatalf("GenerateVariants: %v", err)
	}
	if created != 0 || archived != 0 {
		t.Fatalf("re-run must be idempotent, got created=%d archived=%d", created, archived)
	}
	if n := searchCount(t, ctx, "product.product", [][]interface{}{
		{"product_tmpl_id", "=", tmplID}, {"active", "=", true}}); n != 6 {
		t.Fatalf("expected 6 active variants, got %d", n)
	}
	if n := searchCount(t, ctx, "product.template.attribute.line", [][]interface{}{
		{"product_tmpl_id", "=", tmplID}, {"active", "=", true}}); n != 2 {
		t.Fatalf("expected 2 active attribute lines, got %d", n)
	}

	// Blue costs +5000, L costs +10000: (Blue, L) = 150000 + 15000.
	variant, err := orm.SearchOne(ctx, "product.product", map[string]interface{}{"name": "Demo T-Shirt (Blue, L)"})
	if err != nil || variant == nil {
		t.Fatal("expected variant Demo T-Shirt (Blue, L)")
	}
	ids := services.ParseCombinationIndices(orm.AsString(variant["combination_indices"]))
	if len(ids) != 2 {
		t.Fatalf("expected 2 combination ids, got %v", ids)
	}
	if ids[0] >= ids[1] {
		t.Fatalf("combination ids must be sorted ascending, got %v", ids)
	}
	wantKey := fmt.Sprintf("%d:%s", tmplID, services.CombinationKeyFromIDs(ids))
	if got := orm.AsString(variant["variant_key"]); got != wantKey {
		t.Fatalf("variant_key = %q, want %q", got, wantKey)
	}
	if got := numeric(variant["list_price"]); got != 165000 {
		t.Fatalf("list_price = %v, want 165000", got)
	}
}

// TestArchiveInvalidCombination archives the variant of a removed value and
// reactivates the same record when the value returns.
func TestArchiveInvalidCombination(t *testing.T) {
	ctx := bootstrapTest(t)
	suffix := uniqueSuffix()
	attrID := createTestAttribute(t, ctx, "TArchAttr"+suffix, "always", map[string]float64{"A": 0, "B": 5000})
	tmplID := createTestTemplate(t, ctx, "TArch"+suffix, 100000, map[int][]string{attrID: {"A", "B"}})

	if _, _, err := services.GenerateVariants(ctx, tmplID); err != nil {
		t.Fatalf("GenerateVariants: %v", err)
	}
	if n := searchCount(t, ctx, "product.product", [][]interface{}{
		{"product_tmpl_id", "=", tmplID}, {"active", "=", true}}); n != 2 {
		t.Fatalf("expected 2 active variants, got %d", n)
	}

	// Locate the ptav for value B and archive it.
	rows, err := orm.Search(ctx, "product.template.attribute.value", [][]interface{}{
		{"product_tmpl_id", "=", tmplID},
		{"product_attribute_value_id", "=", findPav(t, ctx, attrID, "B")},
	})
	if err != nil || len(rows) == 0 {
		t.Fatal("ptav for value B not found")
	}
	ptavB, _ := orm.CoerceInt64(rows[0]["id"])
	if err := orm.UpdateRecordByID(ctx, "product.template.attribute.value", int(ptavB),
		map[string]interface{}{"active": false}); err != nil {
		t.Fatalf("archive ptav: %v", err)
	}

	_, archived, err := services.GenerateVariants(ctx, tmplID)
	if err != nil {
		t.Fatalf("GenerateVariants after archive: %v", err)
	}
	if archived != 1 {
		t.Fatalf("expected 1 archived variant, got %d", archived)
	}
	keyB := services.CombinationKeyFromIDs([]int{int(ptavB)})
	bRows, err := orm.Search(ctx, "product.product", [][]interface{}{
		{"product_tmpl_id", "=", tmplID}, {"combination_indices", "=", keyB}})
	if err != nil || len(bRows) == 0 {
		t.Fatal("variant B missing")
	}
	variantB, _ := orm.CoerceInt64(bRows[0]["id"])
	if orm.AsBool(bRows[0]["active"]) {
		t.Fatal("variant B must be archived")
	}

	// Reactivate the value: the same variant record comes back.
	if err := orm.UpdateRecordByID(ctx, "product.template.attribute.value", int(ptavB),
		map[string]interface{}{"active": true}); err != nil {
		t.Fatalf("reactivate ptav: %v", err)
	}
	if _, _, err := services.GenerateVariants(ctx, tmplID); err != nil {
		t.Fatalf("GenerateVariants after reactivate: %v", err)
	}
	bRows, err = orm.Search(ctx, "product.product", [][]interface{}{
		{"product_tmpl_id", "=", tmplID}, {"combination_indices", "=", keyB}})
	if err != nil || len(bRows) == 0 {
		t.Fatal("variant B missing after reactivation")
	}
	nowID, _ := orm.CoerceInt64(bRows[0]["id"])
	if nowID != variantB {
		t.Fatalf("variant B id changed: %d -> %d", variantB, nowID)
	}
	if !orm.AsBool(bRows[0]["active"]) {
		t.Fatal("variant B must be active again")
	}
}

// TestCreateVariantNeverExcluded keeps never-mode attributes out of the matrix.
func TestCreateVariantNeverExcluded(t *testing.T) {
	ctx := bootstrapTest(t)
	suffix := uniqueSuffix()
	attrID := createTestAttribute(t, ctx, "TNeverAttr"+suffix, "never", map[string]float64{"X": 0, "Y": 0})
	tmplID := createTestTemplate(t, ctx, "TNever"+suffix, 100000, map[int][]string{attrID: {"X", "Y"}})

	if _, _, err := services.GenerateVariants(ctx, tmplID); err != nil {
		t.Fatalf("GenerateVariants: %v", err)
	}
	// Only the base product; the never attribute generates no combinations.
	if n := searchCount(t, ctx, "product.product", [][]interface{}{
		{"product_tmpl_id", "=", tmplID}, {"active", "=", true}}); n != 1 {
		t.Fatalf("expected exactly 1 base product, got %d", n)
	}
	base, err := orm.SearchOne(ctx, "product.product", map[string]interface{}{"name": "TNever" + suffix})
	if err != nil || base == nil {
		t.Fatal("base product missing")
	}
	if got := orm.AsString(base["combination_indices"]); got != "" {
		t.Fatalf("base combination must be empty, got %q", got)
	}
}

// TestEmptyTemplateBaseProduct gives a template without attributes exactly one
// base variant linked to it.
func TestEmptyTemplateBaseProduct(t *testing.T) {
	ctx := bootstrapTest(t)
	suffix := uniqueSuffix()
	tmplID := mustCreate(t, ctx, "product.template", map[string]interface{}{
		"name": "TEmpty" + suffix, "type": "consu", "list_price": 75000,
		"sale_ok": true, "purchase_ok": true, "active": true,
	})
	t.Cleanup(func() { cleanupProducts(t, tmplID) })

	if _, _, err := services.GenerateVariants(ctx, tmplID); err != nil {
		t.Fatalf("GenerateVariants: %v", err)
	}
	rows, err := orm.Search(ctx, "product.product",
		[][]interface{}{{"product_tmpl_id", "=", tmplID}})
	if err != nil || len(rows) != 1 {
		t.Fatalf("expected exactly 1 base product, got %d", len(rows))
	}
	if got := orm.AsString(rows[0]["variant_key"]); got != fmt.Sprintf("%d:", tmplID) {
		t.Fatalf("variant_key = %q, want %q", got, fmt.Sprintf("%d:", tmplID))
	}
}

// TestNullKeyNormalization reuses a variant whose combination key is NULL
// instead of creating a duplicate empty-key variant.
func TestNullKeyNormalization(t *testing.T) {
	ctx := bootstrapTest(t)
	suffix := uniqueSuffix()
	tmplID := mustCreate(t, ctx, "product.template", map[string]interface{}{
		"name": "TNull"+suffix, "type": "consu", "list_price": 50000,
		"sale_ok": true, "purchase_ok": true, "active": true,
	})
	t.Cleanup(func() { cleanupProducts(t, tmplID) })
	existing := mustCreate(t, ctx, "product.product", map[string]interface{}{
		"name": "TNull" + suffix, "product_tmpl_id": tmplID,
	})

	if _, _, err := services.GenerateVariants(ctx, tmplID); err != nil {
		t.Fatalf("GenerateVariants: %v", err)
	}
	rows, err := orm.Search(ctx, "product.product",
		[][]interface{}{{"product_tmpl_id", "=", tmplID}})
	if err != nil || len(rows) != 1 {
		t.Fatalf("expected 1 product, got %d", len(rows))
	}
	id, _ := orm.CoerceInt64(rows[0]["id"])
	if int(id) != existing {
		t.Fatalf("NULL-key product not reused: %d != %d", id, existing)
	}
}

// TestMatrixCapExceeded refuses to generate matrices beyond the 1000
// combination cap before writing anything.
func TestMatrixCapExceeded(t *testing.T) {
	ctx := bootstrapTest(t)
	suffix := uniqueSuffix()
	namesA := make(map[string]float64, 32)
	namesB := make(map[string]float64, 32)
	for i := 0; i < 32; i++ {
		namesA[fmt.Sprintf("A%d", i)] = 0
		namesB[fmt.Sprintf("B%d", i)] = 0
	}
	attrA := createTestAttribute(t, ctx, "TCapAttrA"+suffix, "always", namesA)
	attrB := createTestAttribute(t, ctx, "TCapAttrB"+suffix, "always", namesB)
	valuesA := make([]string, 0, 32)
	valuesB := make([]string, 0, 32)
	for i := 0; i < 32; i++ {
		valuesA = append(valuesA, fmt.Sprintf("A%d", i))
		valuesB = append(valuesB, fmt.Sprintf("B%d", i))
	}
	tmplID := createTestTemplate(t, ctx, "TCap"+suffix, 100000, map[int][]string{
		attrA: valuesA, attrB: valuesB,
	})

	if _, _, err := services.GenerateVariants(ctx, tmplID); err == nil {
		t.Fatal("expected matrix cap error for 32x32=1024 combinations")
	}
	if n := searchCount(t, ctx, "product.product", [][]interface{}{
		{"product_tmpl_id", "=", tmplID}}); n != 0 {
		t.Fatalf("cap failure must not create variants, found %d", n)
	}
}

// TestPtalReconcile archives an attribute line once all its values are gone.
func TestPtalReconcile(t *testing.T) {
	ctx := bootstrapTest(t)
	suffix := uniqueSuffix()
	attrID := createTestAttribute(t, ctx, "TLineAttr"+suffix, "always", map[string]float64{"A": 0, "B": 0})
	tmplID := createTestTemplate(t, ctx, "TLine"+suffix, 100000, map[int][]string{attrID: {"A", "B"}})

	if _, _, err := services.GenerateVariants(ctx, tmplID); err != nil {
		t.Fatalf("GenerateVariants: %v", err)
	}
	if n := searchCount(t, ctx, "product.template.attribute.line", [][]interface{}{
		{"product_tmpl_id", "=", tmplID}, {"active", "=", true}}); n != 1 {
		t.Fatalf("expected 1 active line, got %d", n)
	}

	// Archive every ptav of the template.
	if _, err := orm.UnlinkWhere(ctx, "product.template.attribute.value",
		[][]interface{}{{"product_tmpl_id", "=", tmplID}}); err != nil {
		t.Fatalf("remove ptavs: %v", err)
	}
	if _, _, err := services.GenerateVariants(ctx, tmplID); err != nil {
		t.Fatalf("GenerateVariants after removing values: %v", err)
	}
	if n := searchCount(t, ctx, "product.template.attribute.line", [][]interface{}{
		{"product_tmpl_id", "=", tmplID}, {"active", "=", true}}); n != 0 {
		t.Fatalf("line must be archived, got %d active", n)
	}
}
