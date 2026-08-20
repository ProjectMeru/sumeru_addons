package models

import (
	"sumeru/core/sdk"
)

type AccountAccount struct {
	sdk.Model `sumeru:"model=account.account"`

	Code        sdk.String             `sumeru:"required,unique,string=Code"`
	Name        sdk.String             `sumeru:"required,string=Account Name"`
	AccountType sdk.String `sumeru:"required,string=Type,selection=asset_receivable:Receivable,asset_cash:Bank and Cash,asset_current:Current Assets,liability_payable:Payable,liability_current:Current Liabilities,equity:Equity,income:Income,expense:Expenses"`
	Reconcile   sdk.Boolean            `sumeru:"string=Allow Reconciliation,default=false"`
	Active      sdk.Boolean            `sumeru:"string=Active,default=true"`
}
