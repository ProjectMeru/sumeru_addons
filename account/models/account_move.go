package models

import (
	"sumeru/core/sdk"
)

type AccountMove struct {
	sdk.Model `sumeru:"model=account.move"`

	Name            sdk.String                   `sumeru:"required,string=Number"`
	MoveType        sdk.Selection[MoveType]      `sumeru:"string=Type,default=entry"`
	PartnerID       sdk.Many2One[CorePartner]    `sumeru:"string=Partner"`
	CurrencyID      sdk.Many2One[CoreCurrency]   `sumeru:"string=Currency"`
	JournalID       sdk.Many2One[AccountJournal] `sumeru:"string=Journal"`
	CompanyID       sdk.Many2One[CoreCompany]    `sumeru:"string=Company"`
	Date            sdk.Date                     `sumeru:"string=Accounting Date"`
	InvoiceDate     sdk.Date                     `sumeru:"string=Invoice Date"`
	InvoiceDateDue  sdk.Date                     `sumeru:"string=Due Date"`
	PaymentTermID   sdk.Many2One[AccountPaymentTerm] `sumeru:"string=Payment Terms"`
	State           sdk.Selection[MoveState]     `sumeru:"string=Status,default=draft"`
	InvoiceOrigin   sdk.String                   `sumeru:"string=Source Document"`
	Ref             sdk.String                   `sumeru:"string=Reference"`
	ReversedEntryID sdk.Many2One[AccountMove]    `sumeru:"string=Reversal Of"`
	AmountUntaxed   sdk.Numeric                  `sumeru:"string=Untaxed Amount,precision=18,scale=2,default=0"`
	AmountTax       sdk.Numeric                  `sumeru:"string=Tax,precision=18,scale=2,default=0"`
	AmountTotal     sdk.Numeric                  `sumeru:"string=Total,precision=18,scale=2,default=0"`
	AmountResidual  sdk.Numeric                  `sumeru:"string=Amount Due,precision=18,scale=2,default=0"`
	PaymentState    sdk.Selection[PaymentState]  `sumeru:"string=Payment Status,default=not_paid"`
	LineIDs         sdk.One2Many[AccountMoveLine] `sumeru:"string=Journal Items"`
	Narration       sdk.Text                     `sumeru:"string=Notes"`
}
