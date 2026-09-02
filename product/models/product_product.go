package models

import (
	"sumeru/core/sdk"
)

type ProductProduct struct {
	sdk.Model `sumeru:"model=product.product"`

	Name                     sdk.String                    `sumeru:"required,string=Product Name"`
	Image                    sdk.Text                      `sumeru:"string=Image"`
	DefaultCode              sdk.String                    `sumeru:"string=Internal Reference"`
	Type                     sdk.Selection[ProductType]    `sumeru:"string=Product Type,default=consu"`
	CategID                  sdk.Many2One[ProductCategory] `sumeru:"string=Category"`
	ListPrice                sdk.Numeric                   `sumeru:"string=Sales Price,precision=18,scale=2,default=0"`
	StandardPrice            sdk.Numeric                   `sumeru:"string=Cost,precision=18,scale=2,default=0"`
	Description              sdk.Text                      `sumeru:"string=Description"`
	PropertyAccountIncomeID  sdk.Many2One[sdk.Any] `sumeru:"comodel=account.account,string=Income Account"`
	PropertyAccountExpenseID sdk.Many2One[sdk.Any] `sumeru:"comodel=account.account,string=Expense Account"`
	SaleOk                   sdk.Boolean                   `sumeru:"string=Can be Sold,default=true"`
	PurchaseOk               sdk.Boolean                   `sumeru:"string=Can be Purchased,default=true"`
	Active                   sdk.Boolean                   `sumeru:"string=Active,default=true"`
}
