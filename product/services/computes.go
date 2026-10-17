package services

import (
	"context"
	"strings"

	"sumeru/core/orm"
)

func init() {
	orm.RegisterCompute("product.product", "combination", nil, computeCombination)
	orm.RegisterCompute("product.product", "price_extra", nil, computePriceExtra)
	orm.RegisterCompute("product.template", "variant_count", nil, computeVariantCount)
	orm.RegisterCompute("product.configurator", "total_price", nil, computeWizardTotal)
}

// computeCombination resolves the variant's attribute values into "Red, M".
func computeCombination(ctx context.Context, rec map[string]interface{}) (interface{}, error) {
	key := strings.TrimSpace(orm.AsString(rec["combination_indices"]))
	if key == "" {
		return "", nil
	}
	return combinationName(orm.ContextWithBypass(ctx, true), key), nil
}

// computePriceExtra sums the price_extra of the combination's ptavs.
func computePriceExtra(ctx context.Context, rec map[string]interface{}) (interface{}, error) {
	key := strings.TrimSpace(orm.AsString(rec["combination_indices"]))
	if key == "" {
		return 0, nil
	}
	return combinationPriceExtra(orm.ContextWithBypass(ctx, true), key), nil
}

// computeVariantCount counts the template's active variants.
func computeVariantCount(ctx context.Context, rec map[string]interface{}) (interface{}, error) {
	id, ok := orm.CoerceInt64(rec["id"])
	if !ok || id <= 0 {
		return 0, nil
	}
	n, err := orm.SearchCount(orm.ContextWithBypass(ctx, true), "product.product",
		[][]interface{}{{"product_tmpl_id", "=", id}, {"active", "=", true}})
	if err != nil {
		return 0, nil
	}
	return n, nil
}

// computeWizardTotal shows the estimated price of the current selection.
func computeWizardTotal(ctx context.Context, rec map[string]interface{}) (interface{}, error) {
	id, ok := orm.CoerceInt64(rec["id"])
	if !ok || id <= 0 {
		return 0, nil
	}
	total, err := totalPriceForWizard(orm.ContextWithBypass(ctx, true), int(id))
	if err != nil {
		return 0, nil
	}
	return total, nil
}
