package models

import (
	"sumeru/core/sdk"
)

type StockLocation struct {
	sdk.Model `sumeru:"model=stock.location"`

	Name         sdk.String                  `sumeru:"required,unique,string=Location Name"`
	CompleteName sdk.String                  `sumeru:"string=Full Location Name"`
	Active       sdk.Boolean                 `sumeru:"string=Active,default=true"`
	Usage        sdk.String                  `sumeru:"string=Location Type,default=internal,selection=supplier:Vendor,view:Virtual,internal:Internal,customer:Customer,inventory:Inventory Loss,production:Production,transit:Transit"`
	LocationID   sdk.Many2One[StockLocation] `sumeru:"string=Parent Location"`
	CompanyID    sdk.Many2One[sdk.Any]       `sumeru:"string=Company,comodel=core.company"`
	Barcode      sdk.String                  `sumeru:"string=Barcode"`
}
