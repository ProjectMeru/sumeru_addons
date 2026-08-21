package models

import (
	"sumeru/core/sdk"
)

type AccountTax struct {
	sdk.Model `sumeru:"model=account.tax"`

	Name       sdk.String             `sumeru:"required,string=Tax Name"`
	Amount     sdk.Numeric            `sumeru:"string=Amount (%),default=0"`
	TypeTaxUse sdk.String             `sumeru:"string=Tax Type,default=sale,selection=sale:Sales,purchase:Purchase,none:None"`
	AccountID  sdk.Many2One[AccountAccount] `sumeru:"string=Tax Account"`
	Active     sdk.Boolean            `sumeru:"string=Active,default=true"`
}
