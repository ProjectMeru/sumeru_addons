package models

import (
	"sumeru/core/sdk"
)

type AccountTaxRepartitionLine struct {
	sdk.Model `sumeru:"model=account.tax.repartition.line"`

	TaxID           sdk.Many2One[AccountTax]      `sumeru:"required,index,string=Tax"`
	RepartitionType sdk.String                    `sumeru:"string=Type,default=tax,selection=base:Base,tax:Tax"`
	FactorPercent   sdk.Numeric                   `sumeru:"string=Factor (%),default=100"`
	AccountID       sdk.Many2One[AccountAccount]  `sumeru:"string=Account"`
}
