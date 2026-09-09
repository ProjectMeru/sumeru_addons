package models

import (
	"sumeru/core/sdk"
)

type StockLot struct {
	sdk.Model `sumeru:"model=stock.lot"`

	Name       sdk.String            `sumeru:"required,string=Lot/Serial Number"`
	Ref        sdk.String            `sumeru:"string=Internal Reference"`
	ProductID  sdk.Many2One[sdk.Any] `sumeru:"required,string=Product,comodel=product.product"`
	CompanyID  sdk.Many2One[sdk.Any] `sumeru:"string=Company,comodel=core.company"`
	Note       sdk.Text              `sumeru:"string=Description"`
	ProductQty sdk.Float64           `sumeru:"string=On Hand Quantity,default=0,readonly"`
}
