package services

import (
	"context"
	"fmt"
	"log"

	"sumeru/core/orm"
)

func init() {
	orm.RegisterObjectAction("product.template", "action_generate_variants", generateVariantsAction)
	orm.RegisterObjectAction("product.template", "action_open_configurator", openConfiguratorAction)
	orm.RegisterObjectAction("product.template", "action_backfill_products", backfillAction)
	orm.RegisterObjectAction("product.configurator", "action_apply", applyConfiguratorAction)
}

func generateVariantsAction(ctx context.Context, model string, id int, _ map[string]string) (string, error) {
	if model != "product.template" {
		return "", fmt.Errorf("invalid model")
	}
	created, archived, err := GenerateVariants(ctx, id)
	if err != nil {
		return "", err
	}
	log.Printf("product: generated variants for template %d: created=%d archived=%d", id, created, archived)
	return "", nil
}

func openConfiguratorAction(ctx context.Context, _ string, id int, _ map[string]string) (string, error) {
	wizardID, err := OpenConfigurator(ctx, id)
	if err != nil {
		return "", err
	}
	return orm.ActionOpenURL("product.configurator", wizardID), nil
}

func backfillAction(ctx context.Context, _ string, id int, _ map[string]string) (string, error) {
	linked, err := BackfillTemplateProducts(ctx, id)
	if err != nil {
		return "", err
	}
	log.Printf("product: backfilled template %d: linked %d product(s)", id, linked)
	return "", nil
}

func applyConfiguratorAction(ctx context.Context, _ string, id int, _ map[string]string) (string, error) {
	variantID, err := ApplyConfigurator(ctx, id)
	if err != nil {
		return "", err
	}
	return orm.ActionOpenURL("product.product", variantID), nil
}
