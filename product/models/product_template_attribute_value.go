package models

import (
	"sumeru/core/sdk"
)

// ProductTemplateAttributeValue materializes one selectable attribute value
// on a template (Odoo product.template.attribute.value). ProductTmplID and
// AttributeID are denormalized copies of the line's fields, maintained by
// services/variants.go so the flat template matrix list works without M2M
// support. AttributeLineID anchors the line (ValueIDs inverse).
type ProductTemplateAttributeValue struct {
	sdk.Model `sumeru:"model=product.template.attribute.value"`

	// AttributeLineID is not required: rows created from the flat template
	// matrix list carry no line yet; reconcileAttributeLines fills it.
	AttributeLineID         sdk.Many2One[ProductTemplateAttributeLine] `sumeru:"string=Attribute Line"`
	ProductTmplID           sdk.Many2One[ProductTemplate]              `sumeru:"required,string=Template"`
	AttributeID             sdk.Many2One[ProductAttribute]             `sumeru:"required,string=Attribute"`
	ProductAttributeValueID sdk.Many2One[ProductAttributeValue]        `sumeru:"required,string=Value"`
	PriceExtra              sdk.Numeric                                `sumeru:"string=Extra Price,precision=18,scale=2,default=0"`
	Sequence                sdk.Integer                                `sumeru:"string=Sequence,default=10"`
	Active                  sdk.Boolean                                `sumeru:"string=Active,default=true"`
}
