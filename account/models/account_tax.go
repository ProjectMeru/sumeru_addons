package models

import (
	"sumeru/core/sdk"
)

type AccountTax struct {
	sdk.Model `sumeru:"model=account.tax"`

	Name       sdk.String                   `sumeru:"required,string=Tax Name"`
	Amount     sdk.Numeric                  `sumeru:"string=Amount (%),precision=18,scale=4,default=0"`
	TypeTaxUse sdk.Selection[TaxUse]        `sumeru:"string=Tax Type,default=sale"`
	AccountID  sdk.Many2One[AccountAccount] `sumeru:"string=Tax Account"`
	Active     sdk.Boolean                  `sumeru:"string=Active,default=true"`
}
