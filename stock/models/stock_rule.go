package models

import (
	"sumeru/core/sdk"
)

type StockRule struct {
	sdk.Model `sumeru:"model=stock.rule"`

	Name           sdk.String                     `sumeru:"required,string=Name"`
	Active         sdk.Boolean                    `sumeru:"string=Active,default=true"`
	Action         sdk.String                     `sumeru:"string=Action,default=pull,selection=pull:Pull From,push:Push To,pull_push:Pull & Push"`
	Sequence       sdk.Integer                    `sumeru:"string=Sequence,default=20"`
	CompanyID      sdk.Many2One[sdk.Any]          `sumeru:"string=Company,comodel=core.company"`
	LocationSrcID  sdk.Many2One[StockLocation]    `sumeru:"string=Source Location"`
	LocationDestID sdk.Many2One[StockLocation]    `sumeru:"required,string=Destination Location"`
	ProcureMethod  sdk.String                     `sumeru:"string=Supply Method,default=make_to_stock,selection=make_to_stock:Take From Stock,make_to_order:Trigger Another Rule,mts_else_mto:MTS else MTO"`
	PickingTypeID  sdk.Many2One[StockPickingType] `sumeru:"required,string=Operation Type"`
	Delay          sdk.Integer                    `sumeru:"string=Lead Time,default=0"`
}
