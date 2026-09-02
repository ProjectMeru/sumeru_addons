package models

import (
	"sumeru/core/sdk"
)

type AccountAccount struct {
	sdk.Model `sumeru:"model=account.account"`

	Code        sdk.String                 `sumeru:"required,unique,string=Code"`
	Name        sdk.String                 `sumeru:"required,string=Account Name"`
	AccountType sdk.Selection[AccountType] `sumeru:"required,string=Type"`
	Reconcile   sdk.Boolean                `sumeru:"string=Allow Reconciliation,default=false"`
	Deprecated  sdk.Boolean                `sumeru:"string=Deprecated,default=false"`
	GroupID     sdk.Many2One[AccountGroup] `sumeru:"string=Group"`
	Active      sdk.Boolean                `sumeru:"string=Active,default=true"`
}
