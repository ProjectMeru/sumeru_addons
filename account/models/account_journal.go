package models

import (
	"sumeru/core/sdk"
)

type AccountJournal struct {
	sdk.Model `sumeru:"model=account.journal"`

	Name              sdk.String             `sumeru:"required,string=Journal Name"`
	Code              sdk.String             `sumeru:"required,unique,string=Short Code"`
	Type              sdk.String             `sumeru:"string=Type,default=general,selection=sale:Sales,purchase:Purchase,general:Miscellaneous,bank:Bank,cash:Cash"`
	DefaultAccountID  sdk.Many2One[AccountAccount] `sumeru:"string=Default Account"`
	Active            sdk.Boolean            `sumeru:"string=Active,default=true"`
}
