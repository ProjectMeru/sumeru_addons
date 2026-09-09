package models

import (
	"sumeru/core/sdk"
)

type StockMoveLine struct {
	sdk.Model `sumeru:"model=stock.move.line"`

	PickingID      sdk.Many2One[StockPicking]  `sumeru:"string=Transfer,index"`
	MoveID         sdk.Many2One[StockMove]     `sumeru:"string=Move,index"`
	CompanyID      sdk.Many2One[sdk.Any]       `sumeru:"string=Company,comodel=core.company"`
	ProductID      sdk.Many2One[sdk.Any]       `sumeru:"required,string=Product,comodel=product.product"`
	Quantity       sdk.Float64                 `sumeru:"string=Quantity Done,default=0"`
	LotID          sdk.Many2One[StockLot]      `sumeru:"string=Lot/Serial Number"`
	LotName        sdk.String                  `sumeru:"string=Lot/Serial Number Name"`
	Date           sdk.DateTime                `sumeru:"string=Date"`
	LocationID     sdk.Many2One[StockLocation] `sumeru:"string=From"`
	LocationDestID sdk.Many2One[StockLocation] `sumeru:"string=To"`
	Reference      sdk.String                  `sumeru:"string=Reference"`
	State          sdk.String                  `sumeru:"string=Status,default=draft,selection=draft:New,confirmed:Waiting,assigned:Available,done:Done,cancel:Cancelled"`
}
