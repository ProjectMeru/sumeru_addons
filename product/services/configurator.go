package services

import (
	"context"
	"fmt"
	"sort"

	"sumeru/core/orm"
)

func init() {
	orm.RegisterOnchange("product.configurator.line", "attribute_id", onchangeLineAttribute)
	orm.RegisterOnchange("product.configurator.line", "value_id", onchangeLineValue)
}

// onchangeLineAttribute clears the selected value when the attribute changes
// (value domains cannot be filtered inside O2M lists by the kernel client).
func onchangeLineAttribute(_ context.Context, values map[string]interface{}, _ string) (orm.OnchangeResult, error) {
	result := orm.OnchangeResult{Value: map[string]interface{}{}}
	if _, ok := orm.CoerceInt64(values["attribute_id"]); !ok {
		return result, nil
	}
	if valueID, ok := orm.CoerceInt64(values["value_id"]); ok && valueID > 0 {
		result.Value["value_id"] = 0
		result.Value["price_extra"] = 0
	}
	return result, nil
}

// onchangeLineValue fills the line's price hint from the value's default
// extra price.
func onchangeLineValue(ctx context.Context, values map[string]interface{}, _ string) (orm.OnchangeResult, error) {
	result := orm.OnchangeResult{Value: map[string]interface{}{}}
	valueID, ok := orm.CoerceInt64(values["value_id"])
	if !ok || valueID <= 0 {
		result.Value["price_extra"] = 0
		return result, nil
	}
	pav, err := orm.SearchOne(orm.ContextWithBypass(ctx, true), "product.attribute.value",
		map[string]interface{}{"id": int(valueID)})
	if err != nil || pav == nil {
		return result, nil
	}
	result.Value["price_extra"] = numericFloat(pav["default_extra_price"])
	return result, nil
}

// OpenConfigurator creates the variant configurator wizard pre-populated with
// one line per active attribute of the template (ordered by line sequence).
// The caller wraps the returned id with orm.ActionOpenURL("product.configurator", id).
func OpenConfigurator(ctx context.Context, templateID int) (int, error) {
	bypass := orm.ContextWithBypass(ctx, true)
	tmpl, err := orm.SearchOne(bypass, "product.template", map[string]interface{}{"id": templateID})
	if err != nil || tmpl == nil {
		return 0, fmt.Errorf("product.template %d not found", templateID)
	}

	lines, err := orm.Search(bypass, "product.template.attribute.line",
		[][]interface{}{{"product_tmpl_id", "=", templateID}, {"active", "=", true}})
	if err != nil {
		return 0, err
	}
	// De-duplicate attributes, ordered by line sequence then attribute id.
	type orderedAttr struct {
		id, sequence int
	}
	seen := make(map[int]bool)
	var attrs []orderedAttr
	for _, l := range lines {
		attrID, ok := orm.CoerceInt64(l["attribute_id"])
		if !ok || attrID <= 0 || seen[int(attrID)] {
			continue
		}
		attr, err := orm.SearchOne(bypass, "product.attribute", map[string]interface{}{"id": int(attrID)})
		if err != nil || attr == nil || !orm.AsBool(attr["active"]) {
			continue
		}
		seq, _ := orm.CoerceInt64(l["sequence"])
		seen[int(attrID)] = true
		attrs = append(attrs, orderedAttr{id: int(attrID), sequence: int(seq)})
	}
	sort.Slice(attrs, func(i, j int) bool {
		if attrs[i].sequence != attrs[j].sequence {
			return attrs[i].sequence < attrs[j].sequence
		}
		return attrs[i].id < attrs[j].id
	})

	wizardID, err := orm.Create(bypass, orm.RegistryModel("product.configurator"), map[string]interface{}{
		"product_tmpl_id": templateID,
	})
	if err != nil {
		return 0, err
	}
	for _, a := range attrs {
		if _, err := orm.Create(bypass, orm.RegistryModel("product.configurator.line"), map[string]interface{}{
			"configurator_id": wizardID,
			"attribute_id":    a.id,
		}); err != nil {
			return 0, err
		}
	}
	return wizardID, nil
}

// templateHasDynamicAttribute reports whether the template has an active
// attribute line whose attribute uses create_variant=dynamic.
func templateHasDynamicAttribute(ctx context.Context, templateID int) (bool, error) {
	lines, err := orm.Search(ctx, "product.template.attribute.line",
		[][]interface{}{{"product_tmpl_id", "=", templateID}, {"active", "=", true}})
	if err != nil {
		return false, err
	}
	for _, l := range lines {
		attrID, ok := orm.CoerceInt64(l["attribute_id"])
		if !ok || attrID <= 0 {
			continue
		}
		attr, err := orm.SearchOne(ctx, "product.attribute", map[string]interface{}{"id": int(attrID)})
		if err != nil || attr == nil || !orm.AsBool(attr["active"]) {
			continue
		}
		if orm.AsString(attr["create_variant"]) == attrCreateDynamic {
			return true, nil
		}
	}
	return false, nil
}

// ApplyConfigurator resolves every wizard line to a ptav on the template
// (creating missing ptavs only for dynamic templates), then returns the
// concrete variant, creating it when the template is dynamic. Errors when a
// line has no value or the combination cannot exist.
func ApplyConfigurator(ctx context.Context, wizardID int) (int, error) {
	bypass := orm.ContextWithBypass(ctx, true)
	wiz, err := orm.SearchOne(bypass, "product.configurator", map[string]interface{}{"id": wizardID})
	if err != nil || wiz == nil {
		return 0, fmt.Errorf("product.configurator %d not found", wizardID)
	}
	tmplID, ok := orm.CoerceInt64(wiz["product_tmpl_id"])
	if !ok || tmplID <= 0 {
		return 0, fmt.Errorf("configurator has no template")
	}
	tmpl, err := orm.SearchOne(bypass, "product.template", map[string]interface{}{"id": int(tmplID)})
	if err != nil || tmpl == nil {
		return 0, fmt.Errorf("product.template %d not found", tmplID)
	}

	lines, err := orm.Search(bypass, "product.configurator.line",
		[][]interface{}{{"configurator_id", "=", wizardID}})
	if err != nil {
		return 0, err
	}
	if len(lines) == 0 {
		return 0, fmt.Errorf("configurator has no attribute lines")
	}
	dynamic, err := templateHasDynamicAttribute(bypass, int(tmplID))
	if err != nil {
		return 0, err
	}

	var ptavIDs []int
	for _, l := range lines {
		attrID, ok := orm.CoerceInt64(l["attribute_id"])
		if !ok || attrID <= 0 {
			return 0, fmt.Errorf("configurator line without attribute")
		}
		pavID, ok := orm.CoerceInt64(l["value_id"])
		if !ok || pavID <= 0 {
			return 0, fmt.Errorf("choose a value for every attribute")
		}
		ptavID, err := resolvePtav(bypass, int(tmplID), int(attrID), int(pavID))
		if err != nil {
			return 0, err
		}
		if ptavID == 0 {
			if !dynamic {
				return 0, fmt.Errorf("combination does not exist for this template")
			}
			ptavID, err = createPtav(bypass, int(tmplID), int(attrID), int(pavID))
			if err != nil {
				return 0, err
			}
		}
		ptavIDs = append(ptavIDs, ptavID)
	}

	key := CombinationKeyFromIDs(ptavIDs)
	variantID, created, err := createOrRefreshVariant(bypass, tmpl, orm.AsString(tmpl["name"]), key)
	if err != nil {
		return 0, err
	}
	if created && !dynamic {
		return 0, fmt.Errorf("combination does not exist for this template")
	}
	return variantID, nil
}

// resolvePtav finds the active ptav linking a template attribute value; 0
// when none exists.
func resolvePtav(ctx context.Context, tmplID, attrID, pavID int) (int, error) {
	rows, err := orm.Search(ctx, "product.template.attribute.value",
		[][]interface{}{
			{"product_tmpl_id", "=", tmplID},
			{"attribute_id", "=", attrID},
			{"product_attribute_value_id", "=", pavID},
			{"active", "=", true},
		})
	if err != nil {
		return 0, err
	}
	for _, r := range rows {
		if id, ok := orm.CoerceInt64(r["id"]); ok {
			return int(id), nil
		}
	}
	return 0, nil
}

// createPtav materializes a missing template attribute value and reconciles
// its attribute line.
func createPtav(ctx context.Context, tmplID, attrID, pavID int) (int, error) {
	pav, err := orm.SearchOne(ctx, "product.attribute.value", map[string]interface{}{"id": pavID})
	if err != nil || pav == nil {
		return 0, fmt.Errorf("product.attribute.value %d not found", pavID)
	}
	newID, err := orm.Create(ctx, orm.RegistryModel("product.template.attribute.value"), map[string]interface{}{
		"product_tmpl_id":           tmplID,
		"attribute_id":              attrID,
		"product_attribute_value_id": pavID,
		"price_extra":               numericFloat(pav["default_extra_price"]),
		"active":                    true,
	})
	if err != nil {
		return 0, err
	}
	// Attach the line via the shared reconcile logic.
	ptavs, err := orm.Search(ctx, "product.template.attribute.value",
		[][]interface{}{{"product_tmpl_id", "=", tmplID}, {"active", "=", true}})
	if err != nil {
		return 0, err
	}
	if err := reconcileAttributeLines(ctx, tmplID, ptavs); err != nil {
		return 0, err
	}
	return newID, nil
}

// totalPriceForWizard returns template list price plus the selected lines'
// price extras (authoritative value displayed on the wizard).
func totalPriceForWizard(ctx context.Context, wizardID int) (float64, error) {
	wiz, err := orm.SearchOne(ctx, "product.configurator", map[string]interface{}{"id": wizardID})
	if err != nil || wiz == nil {
		return 0, fmt.Errorf("product.configurator %d not found", wizardID)
	}
	tmplID, _ := orm.CoerceInt64(wiz["product_tmpl_id"])
	tmpl, err := orm.SearchOne(ctx, "product.template", map[string]interface{}{"id": int(tmplID)})
	if err != nil || tmpl == nil {
		return 0, fmt.Errorf("product.template %d not found", tmplID)
	}
	lines, err := orm.Search(ctx, "product.configurator.line",
		[][]interface{}{{"configurator_id", "=", wizardID}})
	if err != nil {
		return 0, err
	}
	total := numericFloat(tmpl["list_price"])
	for _, l := range lines {
		total += numericFloat(l["price_extra"])
	}
	return round2(total), nil
}
