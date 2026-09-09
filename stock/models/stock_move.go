package models

import (
	"sumeru/core/sdk"
)

type StockMove struct {
	sdk.Model `sumeru:"model=stock.move"`

	Name           sdk.String                     `sumeru:"string=Description"`
	Sequence       sdk.Integer                    `sumeru:"string=Sequence,default=10"`
	Priority       sdk.String                     `sumeru:"string=Priority,default=0,selection=0:Normal,1:Urgent"`
	Date           sdk.DateTime                   `sumeru:"string=Scheduled Date"`
	CompanyID      sdk.Many2One[sdk.Any]          `sumeru:"string=Company,comodel=core.company"`
	ProductID      sdk.Many2One[sdk.Any]          `sumeru:"required,string=Product,comodel=product.product"`
	ProductQty     sdk.Float64                    `sumeru:"string=Demand,default=0"`
	Quantity       sdk.Float64                    `sumeru:"string=Quantity Done,default=0"`
	LocationID     sdk.Many2One[StockLocation]    `sumeru:"string=From"`
	LocationDestID sdk.Many2One[StockLocation]    `sumeru:"string=To"`
	PartnerID      sdk.Many2One[sdk.Any]          `sumeru:"string=Contact,comodel=core.partner"`
	PickingID      sdk.Many2One[StockPicking]     `sumeru:"string=Transfer,index"`
	PickingTypeID  sdk.Many2One[StockPickingType] `sumeru:"string=Operation Type"`
	State          sdk.String                     `sumeru:"string=Status,default=draft,selection=draft:New,confirmed:Waiting,assigned:Available,done:Done,cancel:Cancelled"`
	Origin         sdk.String                     `sumeru:"string=Source Document"`
	ScrapID        sdk.Many2One[StockScrap]       `sumeru:"string=Scrap Operation"`
	MoveLineIDs    sdk.One2Many[StockMoveLine]    `sumeru:"string=Operations"`
	Reference      sdk.String                     `sumeru:"string=Reference"`
	IsInventory    sdk.Boolean                    `sumeru:"string=Inventory,default=false"`
}
