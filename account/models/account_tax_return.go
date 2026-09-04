package models

import (
	"sumeru/core/sdk"
)

type AccountTaxReturn struct {
	sdk.Model `sumeru:"model=account.tax.return"`

	Name         sdk.String  `sumeru:"required,string=Name"`
	PeriodStart  sdk.Date    `sumeru:"string=Period Start"`
	PeriodEnd    sdk.Date    `sumeru:"string=Period End"`
	State        sdk.String  `sumeru:"string=Status,default=draft"`
	Amount       sdk.Numeric `sumeru:"string=Amount,precision=18,scale=2,default=0"`
}
