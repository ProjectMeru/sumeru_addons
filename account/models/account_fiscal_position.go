package models

import (
	"sumeru/core/sdk"
)

type AccountFiscalPosition struct {
	sdk.Model `sumeru:"model=account.fiscal.position"`

	Name      sdk.String              `sumeru:"required,string=Fiscal Position"`
	Active    sdk.Boolean             `sumeru:"string=Active,default=true"`
	AutoApply sdk.Boolean             `sumeru:"string=Detect Automatically"`
	CountryID sdk.Many2One[CoreCountry] `sumeru:"string=Country"`
	Note      sdk.Text                `sumeru:"string=Notes"`
}

type AccountFiscalPositionTax struct {
	sdk.Model `sumeru:"model=account.fiscal.position.tax"`

	PositionID sdk.Many2One[AccountFiscalPosition] `sumeru:"required,index,string=Fiscal Position"`
	TaxSrcID   sdk.Many2One[AccountTax]            `sumeru:"required,string=Tax on Product"`
	TaxDestID  sdk.Many2One[AccountTax]            `sumeru:"string=Tax to Apply"`
}
