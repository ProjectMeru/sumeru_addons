package models

import (
	"sumeru/core/sdk"
)

type StockWarehouse struct {
	sdk.Model `sumeru:"model=stock.warehouse"`

	Name             sdk.String                     `sumeru:"required,unique,string=Warehouse"`
	Active           sdk.Boolean                    `sumeru:"string=Active,default=true"`
	CompanyID        sdk.Many2One[sdk.Any]          `sumeru:"string=Company,comodel=core.company"`
	Code             sdk.String                     `sumeru:"required,string=Short Name"`
	ViewLocationID   sdk.Many2One[StockLocation]    `sumeru:"string=View Location"`
	LotStockID       sdk.Many2One[StockLocation]    `sumeru:"string=Stock Location"`
	InputLocationID  sdk.Many2One[StockLocation]    `sumeru:"string=Input Location"`
	OutputLocationID sdk.Many2One[StockLocation]    `sumeru:"string=Output Location"`
	ReceptionSteps   sdk.String                     `sumeru:"string=Incoming Shipments,default=one_step,selection=one_step:One Step,two_steps:Two Steps,three_steps:Three Steps"`
	DeliverySteps    sdk.String                     `sumeru:"string=Outgoing Shipments,default=one_step,selection=one_step:One Step,two_steps:Two Steps,three_steps:Three Steps"`
	InTypeID         sdk.Many2One[StockPickingType] `sumeru:"string=Receipt Type"`
	OutTypeID        sdk.Many2One[StockPickingType] `sumeru:"string=Delivery Type"`
	IntTypeID        sdk.Many2One[StockPickingType] `sumeru:"string=Internal Type"`
}
