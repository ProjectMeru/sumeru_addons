package services

import (
	"context"
	"log"

	"sumeru/core/event"
	"sumeru/core/orm"
)

func init() {
	event.Subscribe("record.created", onProductRecordCreated)
	event.Subscribe("record.updated", onProductRecordUpdated)
}

func onProductRecordCreated(ctx context.Context, ev event.Event) error {
	model, _ := ev.Payload["model"].(string)
	if model != "product.template.attribute.value" {
		return nil
	}
	id, ok := coerceID(ev.Payload["id"])
	if !ok {
		return nil
	}
	return regenerateForPtav(orm.ContextWithBypass(ctx, true), id, true)
}

func onProductRecordUpdated(ctx context.Context, ev event.Event) error {
	model, _ := ev.Payload["model"].(string)
	id, ok := coerceID(ev.Payload["id"])
	if !ok {
		return nil
	}
	bypass := orm.ContextWithBypass(ctx, true)
	switch model {
	case "product.template":
		return SyncTemplateToVariants(bypass, id)
	case "product.template.attribute.value":
		return regenerateForPtav(bypass, id, false)
	case "product.attribute":
		return regenerateForAttribute(bypass, id)
	case "product.attribute.value":
		return syncTemplatesForPav(bypass, id)
	}
	return nil
}

// regenerateForPtav copies the value's default extra price on creation, then
// rebuilds the template matrix when the attribute creates variants instantly.
func regenerateForPtav(ctx context.Context, ptavID int, created bool) error {
	ptav, err := orm.SearchOne(ctx, "product.template.attribute.value", map[string]interface{}{"id": ptavID})
	if err != nil || ptav == nil {
		return nil
	}
	if created {
		if numericFloat(ptav["price_extra"]) == 0 {
			if pavID, ok := orm.CoerceInt64(ptav["product_attribute_value_id"]); ok && pavID > 0 {
				pav, err := orm.SearchOne(ctx, "product.attribute.value", map[string]interface{}{"id": int(pavID)})
				if err == nil && pav != nil {
					if def := numericFloat(pav["default_extra_price"]); def != 0 {
						if err := orm.UpdateRecordByID(ctx, "product.template.attribute.value", ptavID,
							map[string]interface{}{"price_extra": def}); err != nil {
							return err
						}
					}
				}
			}
		}
	}
	tmplID, ok := orm.CoerceInt64(ptav["product_tmpl_id"])
	if !ok || tmplID <= 0 {
		return nil
	}
	if attrID, ok := orm.CoerceInt64(ptav["attribute_id"]); ok && attrID > 0 {
		if attr, err := orm.SearchOne(ctx, "product.attribute", map[string]interface{}{"id": int(attrID)}); err == nil && attr != nil {
			if orm.AsString(attr["create_variant"]) != attrCreateAlways {
				return nil
			}
		}
	}
	if _, _, err := GenerateVariants(ctx, int(tmplID)); err != nil {
		log.Printf("product: regenerate variants for template %d: %v", tmplID, err)
	}
	return nil
}

// regenerateForAttribute rebuilds every template using the changed attribute.
func regenerateForAttribute(ctx context.Context, attributeID int) error {
	ptavs, err := orm.Search(ctx, "product.template.attribute.value",
		[][]interface{}{{"attribute_id", "=", attributeID}})
	if err != nil {
		return err
	}
	seen := map[int64]bool{}
	for _, p := range ptavs {
		tmplID, ok := orm.CoerceInt64(p["product_tmpl_id"])
		if !ok || tmplID <= 0 || seen[tmplID] {
			continue
		}
		seen[tmplID] = true
		if _, _, err := GenerateVariants(ctx, int(tmplID)); err != nil {
			log.Printf("product: regenerate variants for template %d: %v", tmplID, err)
		}
	}
	return nil
}

// syncTemplatesForPav refreshes variant names/prices for every template using
// the changed attribute value.
func syncTemplatesForPav(ctx context.Context, pavID int) error {
	ptavs, err := orm.Search(ctx, "product.template.attribute.value",
		[][]interface{}{{"product_attribute_value_id", "=", pavID}})
	if err != nil {
		return err
	}
	seen := map[int64]bool{}
	for _, p := range ptavs {
		tmplID, ok := orm.CoerceInt64(p["product_tmpl_id"])
		if !ok || tmplID <= 0 || seen[tmplID] {
			continue
		}
		seen[tmplID] = true
		if err := SyncTemplateToVariants(ctx, int(tmplID)); err != nil {
			log.Printf("product: sync template %d after attribute value change: %v", tmplID, err)
		}
	}
	return nil
}
