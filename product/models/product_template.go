package models

import (
	"sumeru/core/sdk"
)

// ProductTemplate is the catalog template that variants (product.product)
// are generated from. Catalog fields live on the template and are pushed to
// its variants by services/variants.go (the kernel has no delegation
// inheritance, so variants keep their own copies of these fields).
type ProductTemplate struct {
	sdk.Model `sumeru:"model=product.template"`

	Name              sdk.String                              `sumeru:"required,unique,string=Name"`
	Type              sdk.Selection[ProductType]              `sumeru:"string=Product Type,default=consu"`
	CategID           sdk.Many2One[ProductCategory]           `sumeru:"string=Category"`
	SaleOk            sdk.Boolean                             `sumeru:"string=Can be Sold,default=true"`
	PurchaseOk        sdk.Boolean                             `sumeru:"string=Can be Purchased,default=true"`
	ListPrice         sdk.Numeric                             `sumeru:"string=Sales Price,precision=18,scale=2,default=0"`
	Description       sdk.Text                                `sumeru:"string=Description"`
	Active            sdk.Boolean                             `sumeru:"string=Active,default=true"`
	AttributeLineIDs  sdk.One2Many[ProductTemplateAttributeLine]  `sumeru:"string=Attribute Lines"`
	AttributeValueIDs sdk.One2Many[ProductTemplateAttributeValue] `sumeru:"string=Attributes & Values"`
	ProductVariantIDs sdk.One2Many[ProductProduct]            `sumeru:"string=Variants"`
	VariantCount      sdk.Integer                             `sumeru:"compute=variant_count,string=Variants"`
}
