package wizard

import (
	"sumeru/core/sdk"

	"sumeru_addons/product/models"
)

// ProductConfigurator is the variant configurator dialog (wizard pattern, like
// account/wizard): pick one value per attribute, Apply resolves or creates the
// concrete variant. TotalPrice is a server-side estimate; the authoritative
// price lands on the variant when the wizard is applied.
type ProductConfigurator struct {
	sdk.Model `sumeru:"model=product.configurator"`

	ProductTmplID sdk.Many2One[models.ProductTemplate]     `sumeru:"required,string=Template"`
	TotalPrice    sdk.Numeric                              `sumeru:"compute=total_price,string=Estimated Price,precision=18,scale=2"`
	LineIDs       sdk.One2Many[ProductConfiguratorLine]    `sumeru:"string=Lines"`
}

// ProductConfiguratorLine is one attribute selection inside the configurator.
// ValueID is validated server-side against the line's attribute (onchange
// domains are not applied inside O2M lists by the kernel client).
type ProductConfiguratorLine struct {
	sdk.Model `sumeru:"model=product.configurator.line"`

	ConfiguratorID sdk.Many2One[ProductConfigurator]          `sumeru:"required,string=Configurator"`
	AttributeID    sdk.Many2One[models.ProductAttribute]      `sumeru:"required,string=Attribute"`
	ValueID        sdk.Many2One[models.ProductAttributeValue] `sumeru:"string=Value"`
	PriceExtra     sdk.Numeric                                `sumeru:"string=Extra Price,precision=18,scale=2,default=0"`
}
