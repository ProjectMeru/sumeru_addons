package services

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"sumeru/core/orm"
)

// maxVariantMatrix caps the cartesian product of attribute values per
// generation call (Odoo product.dynamic_variant_limit analogue).
const maxVariantMatrix = 1000

// CombinationKeyFromIDs returns the canonical variant key: comma-joined,
// ascending ptav ids. An empty slice yields "" (base product combination).
func CombinationKeyFromIDs(ids []int) string {
	if len(ids) == 0 {
		return ""
	}
	sorted := append([]int(nil), ids...)
	sort.Ints(sorted)
	parts := make([]string, len(sorted))
	for i, id := range sorted {
		parts[i] = strconv.Itoa(id)
	}
	return strings.Join(parts, ",")
}

// ParseCombinationIndices normalizes stored combination keys ("" / NULL /
// malformed values tolerated) into a sorted id slice.
func ParseCombinationIndices(s string) []int {
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "null") {
		return nil
	}
	var ids []int
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if id, err := strconv.Atoi(part); err == nil {
			ids = append(ids, id)
		}
	}
	sort.Ints(ids)
	return ids
}

// attributeGroup carries the active ptavs of one attribute on a template.
type attributeGroup struct {
	attributeID int
	ptavIDs     []int
}

// combinationName joins the pav names of a combination key ("Red, M").
func combinationName(ctx context.Context, key string) string {
	names := make([]string, 0, 4)
	for _, id := range ParseCombinationIndices(key) {
		ptav, err := orm.SearchOne(ctx, "product.template.attribute.value", map[string]interface{}{"id": id})
		if err != nil || ptav == nil {
			continue
		}
		pavID, _ := orm.CoerceInt64(ptav["product_attribute_value_id"])
		if pavID <= 0 {
			continue
		}
		pav, err := orm.SearchOne(ctx, "product.attribute.value", map[string]interface{}{"id": int(pavID)})
		if err != nil || pav == nil {
			continue
		}
		if name := orm.AsString(pav["name"]); name != "" {
			names = append(names, name)
		}
	}
	return strings.Join(names, ", ")
}

// combinationPriceExtra sums the price_extra of every ptav in the key.
func combinationPriceExtra(ctx context.Context, key string) float64 {
	total := 0.0
	for _, id := range ParseCombinationIndices(key) {
		ptav, err := orm.SearchOne(ctx, "product.template.attribute.value", map[string]interface{}{"id": id})
		if err != nil || ptav == nil {
			continue
		}
		total += numericFloat(ptav["price_extra"])
	}
	return round2(total)
}

// variantName builds the generated product name: "Template (ValueA, ValueB)".
func variantName(templateName, combination string) string {
	templateName = strings.TrimSpace(templateName)
	if combination == "" {
		return templateName
	}
	return fmt.Sprintf("%s (%s)", templateName, combination)
}

// GenerateVariants rebuilds the variant matrix of a template. Active ptavs are
// grouped per attribute; attributes with create_variant=never, inactive
// attributes, and inactive lines are excluded. Every combination gets an
// active variant (reactivated or created); variants whose combination is no
// longer valid are archived. Attribute lines are reconciled from the values.
// Templates with a dynamic attribute skip matrix generation (variants are
// created on demand by the configurator instead) and only get their existing
// variants refreshed. All writes use the bypass context so no record events
// re-trigger.
func GenerateVariants(ctx context.Context, templateID int) (created, archived int, err error) {
	bypass := orm.ContextWithBypass(ctx, true)
	tmpl, err := orm.SearchOne(bypass, "product.template", map[string]interface{}{"id": templateID})
	if err != nil || tmpl == nil {
		return 0, 0, fmt.Errorf("product.template %d not found", templateID)
	}

	ptavs, err := orm.Search(bypass, "product.template.attribute.value",
		[][]interface{}{{"product_tmpl_id", "=", templateID}, {"active", "=", true}})
	if err != nil {
		return 0, 0, err
	}

	groups, hasDynamic, err := buildAttributeGroups(bypass, ptavs)
	if err != nil {
		return 0, 0, err
	}

	if err := reconcileAttributeLines(bypass, templateID, ptavs); err != nil {
		return 0, 0, err
	}

	// Dynamic attributes: never generate the full matrix; refresh the
	// existing variants' catalog fields instead (Odoo _create_variant_ids).
	if hasDynamic {
		if err := SyncTemplateToVariants(bypass, templateID); err != nil {
			return 0, 0, err
		}
		return 0, 0, nil
	}

	templateName := orm.AsString(tmpl["name"])

	// Build the matrix with the cap enforced before any write.
	combos := [][]int{nil} // single empty combination (base product)
	if len(groups) > 0 {
		combos = nil
		total := 1
		for _, g := range groups {
			total *= len(g.ptavIDs)
			if total > maxVariantMatrix {
				return 0, 0, fmt.Errorf("variant matrix exceeds %d combinations; reduce attribute values", maxVariantMatrix)
			}
		}
		combos = cartesianProduct(groups)
	}

	validKeys := make(map[string]bool, len(combos))
	for _, combo := range combos {
		key := CombinationKeyFromIDs(combo)
		validKeys[key] = true
		_, wasCreated, err := createOrRefreshVariant(bypass, tmpl, templateName, key)
		if err != nil {
			return 0, 0, err
		}
		if wasCreated {
			created++
		}
	}

	// Archive variants whose combination is no longer in the matrix.
	products, err := orm.Search(bypass, "product.product",
		[][]interface{}{{"product_tmpl_id", "=", templateID}, {"active", "=", true}})
	if err != nil {
		return 0, 0, err
	}
	for _, p := range products {
		if validKeys[normalizeKey(orm.AsString(p["combination_indices"]))] {
			continue
		}
		id, _ := orm.CoerceInt64(p["id"])
		if err := orm.UpdateRecordByID(bypass, "product.product", int(id), map[string]interface{}{"active": false}); err != nil {
			return 0, 0, err
		}
		archived++
	}
	return created, archived, nil
}

// createOrRefreshVariant finds (or creates) the variant for one combination
// key and refreshes its fields from the template. Returns the product id and
// whether it was created.
func createOrRefreshVariant(ctx context.Context, tmpl map[string]interface{}, templateName, key string) (int, bool, error) {
	extra := combinationPriceExtra(ctx, key)
	listPrice := round2(numericFloat(tmpl["list_price"]) + extra)
	fields := variantFields(tmpl, templateName, combinationName(ctx, key), key, listPrice)
	templateID, _ := orm.CoerceInt64(tmpl["id"])

	// For the empty key, match NULL as well (falsy domain on a Char
	// column covers '' and NULL rows).
	var existing []map[string]interface{}
	var err error
	if key == "" {
		existing, err = orm.Search(ctx, "product.product",
			[][]interface{}{{"product_tmpl_id", "=", templateID}, {"combination_indices", "=", false}})
	} else {
		existing, err = orm.Search(ctx, "product.product",
			[][]interface{}{{"product_tmpl_id", "=", templateID}, {"combination_indices", "=", key}})
	}
	if err != nil {
		return 0, false, err
	}
	if len(existing) > 0 {
		id, _ := orm.CoerceInt64(existing[0]["id"])
		if err := orm.UpdateRecordByID(ctx, "product.product", int(id), fields); err != nil {
			return 0, false, err
		}
		return int(id), false, nil
	}
	newID, err := orm.Create(ctx, orm.RegistryModel("product.product"), fields)
	if err != nil {
		return 0, false, fmt.Errorf("create variant %q: %w", fields["name"], err)
	}
	return newID, true, nil
}

// buildAttributeGroups groups active ptavs by attribute, honoring
// create_variant modes and line/attribute activity. hasDynamic reports
// whether any included attribute uses create_variant=dynamic.
func buildAttributeGroups(ctx context.Context, ptavs []map[string]interface{}) ([]attributeGroup, bool, error) {
	groups := make(map[int][]int)
	groupOrder := make([]int, 0)
	hasDynamic := false

	var byAttr = make(map[int]int)
	for _, p := range ptavs {
		attrID, ok := orm.CoerceInt64(p["attribute_id"])
		if !ok || attrID <= 0 {
			continue
		}
		// Skip ptavs whose line or attribute is inactive.
		if lineID, ok := orm.CoerceInt64(p["attribute_line_id"]); ok && lineID > 0 {
			line, err := orm.SearchOne(ctx, "product.template.attribute.line", map[string]interface{}{"id": int(lineID)})
			if err != nil || line == nil || !orm.AsBool(line["active"]) {
				continue
			}
		}
		if _, seen := byAttr[int(attrID)]; !seen {
			attr, err := orm.SearchOne(ctx, "product.attribute", map[string]interface{}{"id": int(attrID)})
			if err != nil || attr == nil || !orm.AsBool(attr["active"]) {
				byAttr[int(attrID)] = -1
				continue
			}
			if orm.AsString(attr["create_variant"]) == attrCreateNever {
				byAttr[int(attrID)] = -1
				continue
			}
			if orm.AsString(attr["create_variant"]) == attrCreateDynamic {
				hasDynamic = true
			}
			byAttr[int(attrID)] = 1
		}
		if byAttr[int(attrID)] < 0 {
			continue
		}
		id, _ := orm.CoerceInt64(p["id"])
		if _, ok := groups[int(attrID)]; !ok {
			groupOrder = append(groupOrder, int(attrID))
		}
		groups[int(attrID)] = append(groups[int(attrID)], int(id))
	}

	result := make([]attributeGroup, 0, len(groupOrder))
	for _, attrID := range groupOrder {
		ids := groups[attrID]
		sort.Ints(ids)
		if len(ids) == 0 {
			continue
		}
		result = append(result, attributeGroup{attributeID: attrID, ptavIDs: ids})
	}
	return result, hasDynamic, nil
}

// cartesianProduct builds every combination of one ptav per attribute group.
func cartesianProduct(groups []attributeGroup) [][]int {
	combos := [][]int{{}}
	for _, g := range groups {
		next := make([][]int, 0, len(combos)*len(g.ptavIDs))
		for _, combo := range combos {
			for _, id := range g.ptavIDs {
				extended := append(append([]int(nil), combo...), id)
				next = append(next, extended)
			}
		}
		combos = next
	}
	return combos
}

// variantFields builds the create/update payload for a variant.
func variantFields(tmpl map[string]interface{}, templateName, combination, key string, listPrice float64) map[string]interface{} {
	return map[string]interface{}{
		"name":                variantName(templateName, combination),
		"product_tmpl_id":     tmpl["id"],
		"combination_indices": key,
		"variant_key":         variantKeyFor(tmpl, key),
		"type":                orm.AsString(tmpl["type"]),
		"categ_id":            tmpl["categ_id"],
		"list_price":          listPrice,
		"sale_ok":             orm.AsBool(tmpl["sale_ok"]),
		"purchase_ok":         orm.AsBool(tmpl["purchase_ok"]),
		"description":         orm.AsString(tmpl["description"]),
		"active":              true,
	}
}

func variantKeyFor(tmpl map[string]interface{}, key string) string {
	tmplID, _ := orm.CoerceInt64(tmpl["id"])
	return fmt.Sprintf("%d:%s", tmplID, key)
}

// normalizeKey maps NULL and "" to the same empty combination key.
func normalizeKey(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "null") {
		return ""
	}
	return CombinationKeyFromIDs(ParseCombinationIndices(s))
}

// reconcileAttributeLines ensures one active line per attribute that has
// ptavs, archives lines without values, and backfills missing
// attribute_line_id / attribute_id links on ptav rows.
func reconcileAttributeLines(ctx context.Context, templateID int, ptavs []map[string]interface{}) error {
	lines, err := orm.Search(ctx, "product.template.attribute.line",
		[][]interface{}{{"product_tmpl_id", "=", templateID}})
	if err != nil {
		return err
	}
	lineByAttr := make(map[int]map[string]interface{})
	for _, l := range lines {
		if attrID, ok := orm.CoerceInt64(l["attribute_id"]); ok {
			lineByAttr[int(attrID)] = l
		}
	}

	seenAttrs := make(map[int]bool)
	for _, p := range ptavs {
		ptavID, _ := orm.CoerceInt64(p["id"])
		attrID, ok := orm.CoerceInt64(p["attribute_id"])
		if !ok || attrID <= 0 {
			continue
		}
		lineID, hasLine := orm.CoerceInt64(p["attribute_line_id"])
		if hasLine && lineID > 0 && lineByAttr[int(attrID)] != nil {
			if lID, _ := orm.CoerceInt64(lineByAttr[int(attrID)]["id"]); lID != lineID {
				if err := orm.UpdateRecordByID(ctx, "product.template.attribute.value", int(ptavID),
					map[string]interface{}{"attribute_line_id": int(lID)}); err != nil {
					return err
				}
			}
			seenAttrs[int(attrID)] = true
			continue
		}

		// Find an archived line to reactivate, or create one.
		line := lineByAttr[int(attrID)]
		if line == nil {
			newID, err := orm.Create(ctx, orm.RegistryModel("product.template.attribute.line"), map[string]interface{}{
				"product_tmpl_id": templateID,
				"attribute_id":    int(attrID),
				"sequence":        attributeSequence(ctx, int(attrID)),
				"active":          true,
			})
			if err != nil {
				return err
			}
			line = map[string]interface{}{"id": newID}
			lineByAttr[int(attrID)] = line
		}
		lID, _ := orm.CoerceInt64(line["id"])
		if err := orm.UpdateRecordByID(ctx, "product.template.attribute.value", int(ptavID),
			map[string]interface{}{"attribute_line_id": int(lID)}); err != nil {
			return err
		}
		if !orm.AsBool(line["active"]) {
			if err := orm.UpdateRecordByID(ctx, "product.template.attribute.line", int(lID),
				map[string]interface{}{"active": true}); err != nil {
				return err
			}
			line["active"] = true
		}
		seenAttrs[int(attrID)] = true
	}

	// Archive lines whose attribute no longer has any value on the template.
	for attrID, l := range lineByAttr {
		if seenAttrs[attrID] || !orm.AsBool(l["active"]) {
			continue
		}
		lID, _ := orm.CoerceInt64(l["id"])
		if err := orm.UpdateRecordByID(ctx, "product.template.attribute.line", int(lID),
			map[string]interface{}{"active": false}); err != nil {
			return err
		}
	}
	return nil
}

func attributeSequence(ctx context.Context, attributeID int) int {
	attr, err := orm.SearchOne(ctx, "product.attribute", map[string]interface{}{"id": attributeID})
	if err != nil || attr == nil {
		return 10
	}
	seq, ok := orm.CoerceInt64(attr["sequence"])
	if !ok {
		return 10
	}
	return int(seq)
}

// SyncTemplateToVariants pushes the template catalog fields to every variant
// (the kernel has no delegation inheritance, so variants keep copies).
func SyncTemplateToVariants(ctx context.Context, templateID int) error {
	bypass := orm.ContextWithBypass(ctx, true)
	tmpl, err := orm.SearchOne(bypass, "product.template", map[string]interface{}{"id": templateID})
	if err != nil || tmpl == nil {
		return fmt.Errorf("product.template %d not found", templateID)
	}
	products, err := orm.Search(bypass, "product.product",
		[][]interface{}{{"product_tmpl_id", "=", templateID}})
	if err != nil {
		return err
	}
	templateName := orm.AsString(tmpl["name"])
	templateActive := orm.AsBool(tmpl["active"])
	for _, p := range products {
		id, _ := orm.CoerceInt64(p["id"])
		key := normalizeKey(orm.AsString(p["combination_indices"]))
		extra := combinationPriceExtra(bypass, key)
		upd := map[string]interface{}{
			"name":            variantName(templateName, combinationName(bypass, key)),
			"type":            orm.AsString(tmpl["type"]),
			"categ_id":        tmpl["categ_id"],
			"list_price":      round2(numericFloat(tmpl["list_price"]) + extra),
			"sale_ok":         orm.AsBool(tmpl["sale_ok"]),
			"purchase_ok":     orm.AsBool(tmpl["purchase_ok"]),
			"description":     orm.AsString(tmpl["description"]),
			"product_tmpl_id": templateID,
		}
		// Archiving the template archives its variants; reactivation never
		// silently resurrects variants archived for invalid combinations.
		if !templateActive {
			upd["active"] = false
		}
		if err := orm.UpdateRecordByID(bypass, "product.product", int(id), upd); err != nil {
			return err
		}
	}
	return nil
}

// BackfillTemplateProducts links standalone products whose name matches the
// template name (product_tmpl_id NULL) to this template. Idempotent; needed
// for installed databases where noupdate demo records were never updated.
func BackfillTemplateProducts(ctx context.Context, templateID int) (int, error) {
	bypass := orm.ContextWithBypass(ctx, true)
	tmpl, err := orm.SearchOne(bypass, "product.template", map[string]interface{}{"id": templateID})
	if err != nil || tmpl == nil {
		return 0, fmt.Errorf("product.template %d not found", templateID)
	}
	name := orm.AsString(tmpl["name"])
	if name == "" {
		return 0, nil
	}
	orphans, err := orm.Search(bypass, "product.product",
		[][]interface{}{{"product_tmpl_id", "=", false}, {"name", "=", name}})
	if err != nil {
		return 0, err
	}
	linked := 0
	for _, p := range orphans {
		id, _ := orm.CoerceInt64(p["id"])
		if err := orm.UpdateRecordByID(bypass, "product.product", int(id), map[string]interface{}{
			"product_tmpl_id":     templateID,
			"combination_indices": "",
			"variant_key":         variantKeyFor(tmpl, ""),
		}); err != nil {
			return linked, err
		}
		linked++
	}
	return linked, nil
}

// EnsureDemoVariantCatalog seeds the attribute demo records (Go-seeded because
// the XML loader skips records whose model has no unique `name` conflict
// column) and generates the demo T-Shirt variant matrix. Idempotent.
func EnsureDemoVariantCatalog(ctx context.Context) error {
	bypass := orm.ContextWithBypass(ctx, true)

	colorAttr, err := findOrCreateAttribute(bypass, "Color", "radio", "always", 10)
	if err != nil {
		return err
	}
	sizeAttr, err := findOrCreateAttribute(bypass, "Size", "select", "always", 20)
	if err != nil {
		return err
	}
	colorID, _ := orm.CoerceInt64(colorAttr["id"])
	sizeID, _ := orm.CoerceInt64(sizeAttr["id"])

	red, err := findOrCreatePav(bypass, int(colorID), "Red", 10, 0)
	if err != nil {
		return err
	}
	blue, err := findOrCreatePav(bypass, int(colorID), "Blue", 20, 5000)
	if err != nil {
		return err
	}
	small, err := findOrCreatePav(bypass, int(sizeID), "S", 10, 0)
	if err != nil {
		return err
	}
	medium, err := findOrCreatePav(bypass, int(sizeID), "M", 20, 0)
	if err != nil {
		return err
	}
	large, err := findOrCreatePav(bypass, int(sizeID), "L", 30, 10000)
	if err != nil {
		return err
	}

	tmpl, err := findOrCreateTemplate(bypass, "Demo T-Shirt", "consu", 150000)
	if err != nil {
		return err
	}
	tmplID, _ := orm.CoerceInt64(tmpl["id"])

	if err := ensureTemplatePav(bypass, int(tmplID), int(colorID), red); err != nil {
		return err
	}
	if err := ensureTemplatePav(bypass, int(tmplID), int(colorID), blue); err != nil {
		return err
	}
	if err := ensureTemplatePav(bypass, int(tmplID), int(sizeID), small); err != nil {
		return err
	}
	if err := ensureTemplatePav(bypass, int(tmplID), int(sizeID), medium); err != nil {
		return err
	}
	if err := ensureTemplatePav(bypass, int(tmplID), int(sizeID), large); err != nil {
		return err
	}

	_, _, err = GenerateVariants(bypass, int(tmplID))
	return err
}

func findOrCreateAttribute(ctx context.Context, name, displayType, createVariant string, sequence int) (map[string]interface{}, error) {
	attr, err := orm.SearchOne(ctx, "product.attribute", map[string]interface{}{"name": name})
	if err == nil && attr != nil {
		return attr, nil
	}
	id, err := orm.Create(ctx, orm.RegistryModel("product.attribute"), map[string]interface{}{
		"name":           name,
		"display_type":   displayType,
		"create_variant": createVariant,
		"sequence":       sequence,
		"active":         true,
	})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"id": id}, nil
}

func findOrCreatePav(ctx context.Context, attributeID int, name string, sequence int, extraPrice float64) (map[string]interface{}, error) {
	pav, err := orm.SearchOne(ctx, "product.attribute.value",
		map[string]interface{}{"attribute_id": attributeID, "name": name})
	if err == nil && pav != nil {
		return pav, nil
	}
	id, err := orm.Create(ctx, orm.RegistryModel("product.attribute.value"), map[string]interface{}{
		"name":                name,
		"attribute_id":        attributeID,
		"sequence":            sequence,
		"default_extra_price": extraPrice,
		"active":              true,
	})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"id": id, "default_extra_price": extraPrice}, nil
}

func findOrCreateTemplate(ctx context.Context, name, typ string, listPrice float64) (map[string]interface{}, error) {
	tmpl, err := orm.SearchOne(ctx, "product.template", map[string]interface{}{"name": name})
	if err == nil && tmpl != nil {
		return tmpl, nil
	}
	id, err := orm.Create(ctx, orm.RegistryModel("product.template"), map[string]interface{}{
		"name":       name,
		"type":       typ,
		"list_price": listPrice,
		"sale_ok":    true,
		"purchase_ok": true,
		"active":     true,
	})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"id": id}, nil
}

// ensureTemplatePav creates the materialized template value (ptav) for a pav
// on a template when missing, reusing an archived row when present.
func ensureTemplatePav(ctx context.Context, tmplID, attributeID int, pav map[string]interface{}) error {
	pavID, _ := orm.CoerceInt64(pav["id"])
	existing, err := orm.Search(ctx, "product.template.attribute.value",
		[][]interface{}{{"product_tmpl_id", "=", tmplID}, {"product_attribute_value_id", "=", int(pavID)}})
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		id, _ := orm.CoerceInt64(existing[0]["id"])
		// Re-seeded databases may hold rows created before the pav default
		// was captured; sync a zero extra price up to the default.
		if !orm.AsBool(existing[0]["active"]) {
			return orm.UpdateRecordByID(ctx, "product.template.attribute.value", int(id),
				map[string]interface{}{"active": true})
		}
		if numericFloat(existing[0]["price_extra"]) == 0 {
			if def := numericFloat(pav["default_extra_price"]); def != 0 {
				return orm.UpdateRecordByID(ctx, "product.template.attribute.value", int(id),
					map[string]interface{}{"price_extra": def})
			}
		}
		return nil
	}
	_, err = orm.Create(ctx, orm.RegistryModel("product.template.attribute.value"), map[string]interface{}{
		"product_tmpl_id":           tmplID,
		"attribute_id":              attributeID,
		"product_attribute_value_id": int(pavID),
		"price_extra":               numericFloat(pav["default_extra_price"]),
		"active":                    true,
	})
	return err
}
