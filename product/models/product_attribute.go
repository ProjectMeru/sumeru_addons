package models

import (
	"sumeru/core/sdk"
)

// ProductAttribute mirrors Odoo product.attribute: a reusable attribute
// (e.g. Color, Size) whose values drive variant generation.
type ProductAttribute struct {
	sdk.Model `sumeru:"model=product.attribute"`

	Name          sdk.String                           `sumeru:"required,unique,string=Name"`
	DisplayType   sdk.Selection[AttributeDisplayType]  `sumeru:"string=Display Type,default=radio"`
	CreateVariant sdk.Selection[AttributeCreateVariant] `sumeru:"string=Create Variant,default=always"`
	Sequence      sdk.Integer                          `sumeru:"string=Sequence,default=10"`
	Active        sdk.Boolean                          `sumeru:"string=Active,default=true"`
	ValueIDs      sdk.One2Many[ProductAttributeValue]  `sumeru:"string=Values"`
}

// ProductAttributeValue is one selectable value of an attribute (e.g. "Red").
// Name is deliberately not unique: the same value name may exist under
// different attributes.
type ProductAttributeValue struct {
	sdk.Model `sumeru:"model=product.attribute.value"`

	Name              sdk.String                       `sumeru:"required,string=Value"`
	AttributeID       sdk.Many2One[ProductAttribute]   `sumeru:"required,string=Attribute"`
	Sequence          sdk.Integer                      `sumeru:"string=Sequence,default=10"`
	DefaultExtraPrice sdk.Numeric                      `sumeru:"string=Default Extra Price,precision=18,scale=2,default=0,min=0"`
	Active            sdk.Boolean                      `sumeru:"string=Active,default=true"`
}
