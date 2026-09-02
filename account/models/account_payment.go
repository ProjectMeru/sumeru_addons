package models

import (
	"sumeru/core/sdk"
)

type AccountPayment struct {
	sdk.Model `sumeru:"model=account.payment"`

	Name        sdk.String                      `sumeru:"string=Number"`
	PaymentType sdk.Selection[PaymentType]      `sumeru:"required,string=Payment Type,default=inbound"`
	PartnerType sdk.Selection[PartnerType]      `sumeru:"string=Partner Type,default=customer"`
	PartnerID   sdk.Many2One[CorePartner]       `sumeru:"string=Partner"`
	Amount      sdk.Numeric                     `sumeru:"required,string=Amount,precision=18,scale=2,default=0"`
	Date        sdk.Date                        `sumeru:"string=Date"`
	JournalID   sdk.Many2One[AccountJournal]    `sumeru:"string=Journal"`
	Memo        sdk.String                      `sumeru:"string=Memo"`
	State       sdk.Selection[PaymentRecordState] `sumeru:"string=Status,default=draft"`
	MoveID      sdk.Many2One[AccountMove]       `sumeru:"string=Journal Entry"`
	InvoiceID   sdk.Many2One[AccountMove]       `sumeru:"string=Invoice / Bill"`
}
