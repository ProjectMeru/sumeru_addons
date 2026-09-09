package models

import (
	"sumeru/core/sdk"
)

type StockScrap struct {
	sdk.Model `sumeru:"model=stock.scrap"`

	Name            sdk.String                  `sumeru:"string=Reference"`
	CompanyID       sdk.Many2One[sdk.Any]       `sumeru:"string=Company,comodel=core.company"`
	Origin          sdk.String                  `sumeru:"string=Source Document"`
	ProductID       sdk.Many2One[sdk.Any]       `sumeru:"required,string=Product,comodel=product.product"`
	LotID           sdk.Many2One[StockLot]      `sumeru:"string=Lot/Serial Number"`
	PickingID       sdk.Many2One[StockPicking]  `sumeru:"string=Transfer"`
	LocationID      sdk.Many2One[StockLocation] `sumeru:"string=Source Location"`
	ScrapLocationID sdk.Many2One[StockLocation] `sumeru:"string=Scrap Location"`
	ScrapQty        sdk.Float64                 `sumeru:"string=Quantity,default=0"`
	State           sdk.String                  `sumeru:"string=Status,default=draft,selection=draft:Draft,done:Done"`
	DateDone        sdk.DateTime                `sumeru:"string=Date"`
}
