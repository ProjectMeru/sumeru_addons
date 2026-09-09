package models

import (
	"sumeru/core/sdk"
)

type StockOrderpoint struct {
	sdk.Model `sumeru:"model=stock.warehouse.orderpoint"`

	Name          sdk.String                   `sumeru:"string=Name"`
	Trigger       sdk.String                   `sumeru:"string=Trigger,default=auto,selection=auto:Automatic,manual:Manual"`
	Active        sdk.Boolean                  `sumeru:"string=Active,default=true"`
	WarehouseID   sdk.Many2One[StockWarehouse] `sumeru:"required,string=Warehouse"`
	LocationID    sdk.Many2One[StockLocation]  `sumeru:"required,string=Location"`
	ProductID     sdk.Many2One[sdk.Any]        `sumeru:"required,string=Product,comodel=product.product"`
	ProductMinQty sdk.Float64                  `sumeru:"string=Minimum Quantity,default=0"`
	ProductMaxQty sdk.Float64                  `sumeru:"string=Maximum Quantity,default=0"`
	QtyOnHand     sdk.Float64                  `sumeru:"string=On Hand,default=0,readonly"`
	QtyForecast   sdk.Float64                  `sumeru:"string=Forecast,default=0,readonly"`
	QtyToOrder    sdk.Float64                  `sumeru:"string=To Order,default=0,readonly"`
	CompanyID     sdk.Many2One[sdk.Any]        `sumeru:"string=Company,comodel=core.company"`
}
