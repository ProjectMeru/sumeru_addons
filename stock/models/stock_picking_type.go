package models

import (
	"sumeru/core/sdk"
)

type StockPickingType struct {
	sdk.Model `sumeru:"model=stock.picking.type"`

	Name                  sdk.String                   `sumeru:"required,unique,string=Operation Type"`
	Sequence              sdk.Integer                  `sumeru:"string=Sequence,default=0"`
	SequenceCode          sdk.String                   `sumeru:"required,string=Sequence Prefix"`
	DefaultLocationSrcID  sdk.Many2One[StockLocation]  `sumeru:"string=Default Source Location"`
	DefaultLocationDestID sdk.Many2One[StockLocation]  `sumeru:"string=Default Destination Location"`
	Code                  sdk.String                   `sumeru:"string=Type of Operation,default=incoming,selection=incoming:Receipt,outgoing:Delivery,internal:Internal Transfer"`
	WarehouseID           sdk.Many2One[StockWarehouse] `sumeru:"string=Warehouse"`
	ShowOperations        sdk.Boolean                  `sumeru:"string=Show Detailed Operations,default=false"`
	Active                sdk.Boolean                  `sumeru:"string=Active,default=true"`
	CompanyID             sdk.Many2One[sdk.Any]        `sumeru:"string=Company,comodel=core.company"`
}
