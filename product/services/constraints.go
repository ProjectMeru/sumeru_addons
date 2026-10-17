package services

import (
	"context"
	"fmt"
	"strings"

	"sumeru/core/orm"
)

func init() {
	orm.RegisterConstraint("product.product", "barcode_unique_per_company", checkBarcodeUnique)
	orm.RegisterConstraint("product.template.attribute.value", "value_matches_attribute", checkPtavValueAttribute)
	orm.RegisterConstraint("product.template.attribute.value", "unique_active_value", checkPtavUnique)
	orm.RegisterConstraint("product.configurator.line", "value_matches_attribute", checkConfigLineValueAttribute)
}

// checkBarcodeUnique enforces barcode uniqueness per company (Odoo
// _check_barcode_uniqueness): reuse is allowed only in a different company;
// shared (NULL-company) products conflict with every company.
func checkBarcodeUnique(ctx context.Context, rec map[string]interface{}) error {
	barcode := strings.TrimSpace(orm.AsString(rec["barcode"]))
	if barcode == "" {
		return nil
	}
	id, _ := orm.CoerceInt64(rec["id"])
	companyID, _ := orm.CoerceInt64(rec["company_id"])
	bypass := orm.ContextWithBypass(ctx, true)
	rows, err := orm.Search(bypass, "product.product", [][]interface{}{{"barcode", "=", barcode}})
	if err != nil {
		return err
	}
	for _, r := range rows {
		rid, _ := orm.CoerceInt64(r["id"])
		if rid == id {
			continue
		}
		otherCompany, _ := orm.CoerceInt64(r["company_id"])
		if companyID == 0 || otherCompany == 0 || companyID == otherCompany {
			return fmt.Errorf("barcode %q is already used by another product in this company", barcode)
		}
	}
	return nil
}

// checkPtavValueAttribute ensures the materialized value belongs to the row's
// attribute and to its attribute line.
func checkPtavValueAttribute(ctx context.Context, rec map[string]interface{}) error {
	pavID, _ := orm.CoerceInt64(rec["product_attribute_value_id"])
	attrID, _ := orm.CoerceInt64(rec["attribute_id"])
	lineID, _ := orm.CoerceInt64(rec["attribute_line_id"])
	if pavID <= 0 || attrID <= 0 {
		return nil // required tags handle presence
	}
	bypass := orm.ContextWithBypass(ctx, true)
	if pav, err := orm.SearchOne(bypass, "product.attribute.value", map[string]interface{}{"id": int(pavID)}); err == nil && pav != nil {
		if pavAttr, _ := orm.CoerceInt64(pav["attribute_id"]); pavAttr > 0 && pavAttr != attrID {
			return fmt.Errorf("value belongs to a different attribute")
		}
	}
	if lineID > 0 {
		if line, err := orm.SearchOne(bypass, "product.template.attribute.line", map[string]interface{}{"id": int(lineID)}); err == nil && line != nil {
			if lineAttr, _ := orm.CoerceInt64(line["attribute_id"]); lineAttr > 0 && lineAttr != attrID {
				return fmt.Errorf("attribute does not match the attribute line")
			}
		}
	}
	return nil
}

// checkPtavUnique allows at most one active (template, attribute, value) row;
// archived rows may exist as reactivation targets.
func checkPtavUnique(ctx context.Context, rec map[string]interface{}) error {
	if !orm.AsBool(rec["active"]) {
		return nil
	}
	tmplID, _ := orm.CoerceInt64(rec["product_tmpl_id"])
	attrID, _ := orm.CoerceInt64(rec["attribute_id"])
	pavID, _ := orm.CoerceInt64(rec["product_attribute_value_id"])
	if tmplID <= 0 || attrID <= 0 || pavID <= 0 {
		return nil
	}
	id, _ := orm.CoerceInt64(rec["id"])
	bypass := orm.ContextWithBypass(ctx, true)
	rows, err := orm.Search(bypass, "product.template.attribute.value", [][]interface{}{
		{"product_tmpl_id", "=", tmplID},
		{"attribute_id", "=", attrID},
		{"product_attribute_value_id", "=", pavID},
		{"active", "=", true},
	})
	if err != nil {
		return err
	}
	for _, r := range rows {
		rid, _ := orm.CoerceInt64(r["id"])
		if rid != id {
			return fmt.Errorf("this attribute value is already configured on the template")
		}
	}
	return nil
}

// checkConfigLineValueAttribute ensures the selected value belongs to the
// line's attribute (onchange domains are not applied inside O2M lists).
func checkConfigLineValueAttribute(ctx context.Context, rec map[string]interface{}) error {
	valueID, _ := orm.CoerceInt64(rec["value_id"])
	attrID, _ := orm.CoerceInt64(rec["attribute_id"])
	if valueID <= 0 || attrID <= 0 {
		return nil
	}
	bypass := orm.ContextWithBypass(ctx, true)
	pav, err := orm.SearchOne(bypass, "product.attribute.value", map[string]interface{}{"id": int(valueID)})
	if err != nil || pav == nil {
		return nil
	}
	pavAttr, _ := orm.CoerceInt64(pav["attribute_id"])
	if pavAttr > 0 && pavAttr != attrID {
		return fmt.Errorf("value belongs to a different attribute")
	}
	return nil
}
