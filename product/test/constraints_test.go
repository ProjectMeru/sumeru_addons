package test

import (
	"fmt"
	"testing"

	"sumeru/core/orm"
)

// TestBarcodeUniquePerCompany enforces the per-company barcode rule: reuse is
// allowed in another company only; shared products conflict with everyone.
func TestBarcodeUniquePerCompany(t *testing.T) {
	ctx := bootstrapTest(t)
	suffix := uniqueSuffix()
	barcode := "BC" + suffix
	company1 := mustCreate(t, ctx, "core.company", map[string]interface{}{"name": "BCComp1" + suffix})
	company2 := mustCreate(t, ctx, "core.company", map[string]interface{}{"name": "BCComp2" + suffix})

	mustCreate(t, ctx, "product.product", map[string]interface{}{
		"name": "PBC1" + suffix, "barcode": barcode, "company_id": company1,
	})
	// Same company: rejected.
	mustFailCreate(t, ctx, "product.product", map[string]interface{}{
		"name": "PBC2" + suffix, "barcode": barcode, "company_id": company1,
	})
	// Other company: allowed.
	mustCreate(t, ctx, "product.product", map[string]interface{}{
		"name": "PBC3" + suffix, "barcode": barcode, "company_id": company2,
	})
	// Shared (no company) conflicts with every company.
	mustFailCreate(t, ctx, "product.product", map[string]interface{}{
		"name": "PBC4" + suffix, "barcode": barcode,
	})
	// Empty barcodes are never constrained.
	mustCreate(t, ctx, "product.product", map[string]interface{}{"name": "PBC5" + suffix})
	mustCreate(t, ctx, "product.product", map[string]interface{}{"name": "PBC6" + suffix})
}

// TestPtavValueAttributeMismatch rejects materialized values that belong to a
// different attribute than the row declares.
func TestPtavValueAttributeMismatch(t *testing.T) {
	ctx := bootstrapTest(t)
	suffix := uniqueSuffix()
	attr1 := createTestAttribute(t, ctx, "TMisAttr1"+suffix, "always", map[string]float64{"X": 0})
	attr2 := createTestAttribute(t, ctx, "TMisAttr2"+suffix, "always", map[string]float64{"Y": 0})
	tmplID := mustCreate(t, ctx, "product.template", map[string]interface{}{
		"name": "TMis" + suffix, "type": "consu", "active": true,
	})

	mustFailCreate(t, ctx, "product.template.attribute.value", map[string]interface{}{
		"product_tmpl_id":           tmplID,
		"attribute_id":              attr1,
		"product_attribute_value_id": findPav(t, ctx, attr2, "Y"),
		"active":                    true,
	})
	// The correct pairing is accepted.
	mustCreate(t, ctx, "product.template.attribute.value", map[string]interface{}{
		"product_tmpl_id":           tmplID,
		"attribute_id":              attr1,
		"product_attribute_value_id": findPav(t, ctx, attr1, "X"),
		"active":                    true,
	})
}

// TestPtavDuplicateActive allows only one active (template, attribute, value)
// row; archived rows may coexist as reactivation targets.
func TestPtavDuplicateActive(t *testing.T) {
	ctx := bootstrapTest(t)
	suffix := uniqueSuffix()
	attrID := createTestAttribute(t, ctx, "TDupAttr"+suffix, "always", map[string]float64{"X": 0})
	pavID := findPav(t, ctx, attrID, "X")
	tmplID := mustCreate(t, ctx, "product.template", map[string]interface{}{
		"name": "TDup" + suffix, "type": "consu", "active": true,
	})
	first := mustCreate(t, ctx, "product.template.attribute.value", map[string]interface{}{
		"product_tmpl_id": tmplID, "attribute_id": attrID,
		"product_attribute_value_id": pavID, "active": true,
	})

	mustFailCreate(t, ctx, "product.template.attribute.value", map[string]interface{}{
		"product_tmpl_id": tmplID, "attribute_id": attrID,
		"product_attribute_value_id": pavID, "active": true,
	})

	// Archiving the first opens the slot again.
	if err := orm.UpdateRecordByID(ctx, "product.template.attribute.value", first,
		map[string]interface{}{"active": false}); err != nil {
		t.Fatalf("archive ptav: %v", err)
	}
	mustCreate(t, ctx, "product.template.attribute.value", map[string]interface{}{
		"product_tmpl_id": tmplID, "attribute_id": attrID,
		"product_attribute_value_id": pavID, "active": true,
	})
}

// TestVariantKeyUnique relies on the DB-level unique index of variant_key.
func TestVariantKeyUnique(t *testing.T) {
	ctx := bootstrapTest(t)
	suffix := uniqueSuffix()
	key := fmt.Sprintf("%d:%s", 999999+len(suffix), "1")
	mustCreate(t, ctx, "product.product", map[string]interface{}{
		"name": "PVK1" + suffix, "variant_key": key,
	})
	mustFailCreate(t, ctx, "product.product", map[string]interface{}{
		"name": "PVK2" + suffix, "variant_key": key,
	})
}
