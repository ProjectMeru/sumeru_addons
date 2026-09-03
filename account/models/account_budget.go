package models

import (
	"sumeru/core/sdk"
)

type AccountBudget struct {
	sdk.Model `sumeru:"model=account.budget"`

	Name        sdk.String                   `sumeru:"required,string=Name"`
	DateFrom    sdk.Date                     `sumeru:"string=Start Date"`
	DateTo      sdk.Date                     `sumeru:"string=End Date"`
	CompanyID   sdk.Many2One[CoreCompany]    `sumeru:"string=Company"`
	AccountID   sdk.Many2One[AccountAccount] `sumeru:"string=Account"`
	Amount      sdk.Numeric                  `sumeru:"string=Budget Amount,precision=18,scale=2,default=0"`
	ActualAmount sdk.Numeric                 `sumeru:"string=Actual,precision=18,scale=2,default=0"`
	Variance    sdk.Numeric                  `sumeru:"string=Variance,precision=18,scale=2,default=0"`
	State       sdk.String                   `sumeru:"string=Status,default=draft"`
}
