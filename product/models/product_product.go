package models

import (
	"sumeru/core/sdk"
)

type ProductProduct struct {
	sdk.Model `sumeru:"model=product.product"`

	Name                     sdk.String                    `sumeru:"required,unique,string=Product Name"`
	Image                    sdk.Text                      `sumeru:"string=Image"`
	DefaultCode              sdk.String                    `sumeru:"string=Internal Reference"`
	Type                     sdk.Selection[ProductType]    `sumeru:"string=Product Type,default=consu"`
	Tracking                 sdk.Selection[ProductTracking] `sumeru:"string=Tracking,default=none"`
	CategID                  sdk.Many2One[ProductCategory] `sumeru:"string=Category"`
	ListPrice                sdk.Numeric                   `sumeru:"string=Sales Price,precision=18,scale=2,default=0"`
	StandardPrice            sdk.Numeric                   `sumeru:"string=Cost,precision=18,scale=2,default=0"`
	Description              sdk.Text                      `sumeru:"string=Description"`
	PropertyAccountIncomeID  sdk.Many2One[sdk.Any]         `sumeru:"comodel=account.account,string=Income Account"`
	PropertyAccountExpenseID sdk.Many2One[sdk.Any]         `sumeru:"comodel=account.account,string=Expense Account"`
	SaleOk                   sdk.Boolean                   `sumeru:"string=Can be Sold,default=true"`
	PurchaseOk               sdk.Boolean                   `sumeru:"string=Can be Purchased,default=true"`
	Active                   sdk.Boolean                   `sumeru:"string=Active,default=true"`

	// Variant fields (SUM-BIZ-001). ProductTmplID is deliberately not
	// required so existing records and other addons' demo data keep loading;
	// services/variants.go backfills template links for installed databases.
	ProductTmplID      sdk.Many2One[ProductTemplate] `sumeru:"string=Template"`
	CombinationIndices sdk.String                    `sumeru:"string=Combination Key"`
	Combination        sdk.String                    `sumeru:"compute=combination,string=Combination"`
	Barcode            sdk.String                    `sumeru:"string=Barcode"`
	VariantKey         sdk.String                    `sumeru:"unique,string=Variant Key"`
	CompanyID          sdk.Many2One[CoreCompany]     `sumeru:"string=Company,default=current_company"`
	PriceExtra         sdk.Numeric                   `sumeru:"compute=price_extra,string=Variant Extra,precision=18,scale=2"`
}
