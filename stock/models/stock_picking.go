package models

import (
	"sumeru/core/sdk"
)

type StockPicking struct {
	sdk.Model `sumeru:"model=stock.picking"`

	Name            sdk.String                     `sumeru:"string=Reference"`
	Origin          sdk.String                     `sumeru:"string=Source Document"`
	Note            sdk.Text                       `sumeru:"string=Notes"`
	MoveType        sdk.String                     `sumeru:"string=Shipping Policy,default=direct,selection=direct:As soon as possible,one:When all products are ready"`
	State           sdk.String                     `sumeru:"string=Status,default=draft,selection=draft:Draft,waiting:Waiting Another Operation,confirmed:Waiting,assigned:Ready,done:Done,cancel:Cancelled"`
	Priority        sdk.String                     `sumeru:"string=Priority,default=0,selection=0:Normal,1:Urgent"`
	ScheduledDate   sdk.DateTime                   `sumeru:"string=Scheduled Date"`
	DateDone        sdk.DateTime                   `sumeru:"string=Date of Transfer"`
	LocationID      sdk.Many2One[StockLocation]    `sumeru:"string=Source Location"`
	LocationDestID  sdk.Many2One[StockLocation]    `sumeru:"string=Destination Location"`
	MoveIDs         sdk.One2Many[StockMove]        `sumeru:"string=Stock Moves"`
	MoveLineIDs     sdk.One2Many[StockMoveLine]    `sumeru:"string=Operations"`
	PickingTypeID   sdk.Many2One[StockPickingType] `sumeru:"string=Operation Type"`
	PickingTypeCode sdk.String                     `sumeru:"string=Type of Operation,readonly,selection=incoming:Receipt,outgoing:Delivery,internal:Internal Transfer"`
	PartnerID       sdk.Many2One[sdk.Any]          `sumeru:"string=Contact,comodel=core.partner"`
	CompanyID       sdk.Many2One[sdk.Any]          `sumeru:"string=Company,comodel=core.company"`
	Printed         sdk.Boolean                    `sumeru:"string=Printed,default=false"`
}
