package models

import (
	"sumeru/core/sdk"
)

type AccountPayment struct {
	sdk.Model `sumeru:"model=account.payment"`

	Name        sdk.String             `sumeru:"string=Number"`
	PaymentType sdk.String             `sumeru:"required,string=Payment Type,default=inbound,selection=inbound:Receive,outbound:Send"`
	PartnerType sdk.String             `sumeru:"string=Partner Type,default=customer,selection=customer:Customer,supplier:Vendor"`
	PartnerID   sdk.Many2One[sdk.Any] `sumeru:"string=Partner,comodel=core.partner"`
	Amount      sdk.Numeric            `sumeru:"required,string=Amount,default=0"`
	Date        sdk.Date               `sumeru:"string=Date"`
	JournalID   sdk.Many2One[AccountJournal] `sumeru:"string=Journal"`
	Memo        sdk.String             `sumeru:"string=Memo"`
	State       sdk.String             `sumeru:"string=Status,default=draft,selection=draft:Draft,posted:Posted,cancelled:Cancelled"`
	MoveID      sdk.Many2One[AccountMove] `sumeru:"string=Journal Entry"`
	InvoiceID   sdk.Many2One[AccountMove] `sumeru:"string=Invoice / Bill"`
}
