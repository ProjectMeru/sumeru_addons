package models

import (
	"sumeru/core/sdk"
)

type StockQuant struct {
	sdk.Model `sumeru:"model=stock.quant"`

	ProductID         sdk.Many2One[sdk.Any]       `sumeru:"required,string=Product,comodel=product.product"`
	CompanyID         sdk.Many2One[sdk.Any]       `sumeru:"string=Company,comodel=core.company"`
	LocationID        sdk.Many2One[StockLocation] `sumeru:"required,string=Location"`
	LotID             sdk.Many2One[StockLot]      `sumeru:"string=Lot/Serial Number"`
	Quantity          sdk.Float64                 `sumeru:"string=Quantity,default=0"`
	ReservedQuantity  sdk.Float64                 `sumeru:"string=Reserved Quantity,default=0,readonly"`
	AvailableQuantity sdk.Float64                 `sumeru:"string=Available Quantity,default=0,readonly"`
	InDate            sdk.DateTime                `sumeru:"string=Incoming Date"`

	// Inventory adjustment (physical count) fields.
	InventoryQuantity     sdk.Float64           `sumeru:"string=Counted Quantity,default=0"`
	InventoryDiffQuantity sdk.Float64           `sumeru:"string=Difference,default=0,readonly"`
	InventoryQuantitySet  sdk.Boolean           `sumeru:"string=Counted,default=false"`
	InventoryDate         sdk.Date              `sumeru:"string=Scheduled Count Date"`
	LastCountDate         sdk.Date              `sumeru:"string=Last Count Date,readonly"`
	InventoryName         sdk.String            `sumeru:"string=Count Reason"`
	UserID                sdk.Many2One[sdk.Any] `sumeru:"string=Assigned To,comodel=core.user"`
}
