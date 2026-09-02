package models

import (
	"sumeru/core/sdk"
)

type AccountTaxRepartitionLine struct {
	sdk.Model `sumeru:"model=account.tax.repartition.line"`

	TaxID           sdk.Many2One[AccountTax]        `sumeru:"required,index,string=Tax"`
	RepartitionType sdk.Selection[RepartitionType] `sumeru:"string=Type,default=tax"`
	FactorPercent   sdk.Numeric                     `sumeru:"string=Factor (%),precision=18,scale=2,default=100"`
	AccountID       sdk.Many2One[AccountAccount]    `sumeru:"string=Account"`
}
