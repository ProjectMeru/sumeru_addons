package models

import (
	"sumeru/core/sdk"
)

type AccountAsset struct {
	sdk.Model `sumeru:"model=account.asset"`

	Name            sdk.String                   `sumeru:"required,string=Name"`
	OriginalValue   sdk.Numeric                  `sumeru:"string=Original Value,precision=18,scale=2,default=0"`
	SalvageValue    sdk.Numeric                  `sumeru:"string=Salvage Value,precision=18,scale=2,default=0"`
	BookValue       sdk.Numeric                  `sumeru:"string=Book Value,precision=18,scale=2,default=0"`
	Months          sdk.Integer                  `sumeru:"string=Useful Life (Months),default=36"`
	Method          sdk.String                   `sumeru:"string=Method,default=straight_line"`
	AcquisitionDate sdk.Date                     `sumeru:"string=Acquisition Date"`
	AccountID       sdk.Many2One[AccountAccount] `sumeru:"string=Asset Account"`
	ExpenseAccountID sdk.Many2One[AccountAccount] `sumeru:"string=Depreciation Expense"`
	State           sdk.String                   `sumeru:"string=Status,default=draft"`
}
