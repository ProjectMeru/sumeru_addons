package models

import (
	"sumeru/core/sdk"
)

type AccountRecurringTemplate struct {
	sdk.Model `sumeru:"model=account.recurring.template"`

	Name      sdk.String                   `sumeru:"required,string=Name"`
	JournalID sdk.Many2One[AccountJournal] `sumeru:"string=Journal"`
	AccountID sdk.Many2One[AccountAccount] `sumeru:"string=Account"`
	PartnerID sdk.Many2One[CorePartner]    `sumeru:"string=Partner"`
	MoveType  sdk.String                   `sumeru:"string=Move Type,default=entry"`
	Amount    sdk.Numeric                  `sumeru:"string=Amount,precision=18,scale=2,default=0"`
	Interval  sdk.String                   `sumeru:"string=Interval,default=monthly"`
	NextDate  sdk.Date                     `sumeru:"string=Next Date"`
	Active    sdk.Boolean                  `sumeru:"string=Active,default=true"`
}
