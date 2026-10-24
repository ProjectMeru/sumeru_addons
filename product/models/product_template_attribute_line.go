package models

import (
	"sumeru/core/sdk"
)

// ProductTemplateAttributeLine mirrors Odoo product.template.attribute.line:
// an attribute configured on a template. Because the sumeru kernel has no
// M2M persistence, the line's selectable values are materialized as
// product.template.attribute.value rows (ValueIDs inverse) instead of a
// value_ids M2M. Lines are reconciled automatically from those values by
// services/variants.go.
type ProductTemplateAttributeLine struct {
	sdk.Model `sumeru:"model=product.template.attribute.line"`

	ProductTmplID sdk.Many2One[ProductTemplate]              `sumeru:"required,string=Template"`
	AttributeID   sdk.Many2One[ProductAttribute]             `sumeru:"required,string=Attribute"`
	Sequence      sdk.Integer                                `sumeru:"string=Sequence,default=10"`
	Active        sdk.Boolean                                `sumeru:"string=Active,default=true"`
	ValueIDs      sdk.One2Many[ProductTemplateAttributeValue] `sumeru:"string=Values"`
}
