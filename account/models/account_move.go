package models

import (
	"sumeru/core/sdk"
)

type AccountMove struct {
	sdk.Model `sumeru:"model=account.move"`

	Name              sdk.String                    `sumeru:"required,string=Number"`
	MoveType          sdk.String                    `sumeru:"string=Type,default=entry,selection=entry:Journal Entry,out_invoice:Customer Invoice,out_refund:Customer Credit Note,in_invoice:Vendor Bill,in_refund:Vendor Credit Note"`
	PartnerID         sdk.Many2One[sdk.Any]     `sumeru:"string=Partner,comodel=core.partner"`
	JournalID         sdk.Many2One[AccountJournal]  `sumeru:"string=Journal"`
	CompanyID         sdk.Many2One[sdk.Any]     `sumeru:"string=Company,comodel=core.company"`
	Date              sdk.Date                      `sumeru:"string=Accounting Date"`
	InvoiceDate       sdk.Date                      `sumeru:"string=Invoice Date"`
	InvoiceDateDue    sdk.Date                      `sumeru:"string=Due Date"`
	PaymentTermID     sdk.Many2One[AccountPaymentTerm] `sumeru:"string=Payment Terms"`
	State             sdk.String                    `sumeru:"string=Status,default=draft,selection=draft:Draft,posted:Posted,cancel:Cancelled"`
	InvoiceOrigin     sdk.String                    `sumeru:"string=Source Document"`
	Ref               sdk.String                    `sumeru:"string=Reference"`
	ReversedEntryID   sdk.Many2One[AccountMove]     `sumeru:"string=Reversal Of"`
	AmountUntaxed     sdk.Numeric                   `sumeru:"string=Untaxed Amount,default=0"`
	AmountTax         sdk.Numeric                   `sumeru:"string=Tax,default=0"`
	AmountTotal       sdk.Numeric                   `sumeru:"string=Total,default=0"`
	AmountResidual    sdk.Numeric                   `sumeru:"string=Amount Due,default=0"`
	PaymentState      sdk.String                    `sumeru:"string=Payment Status,default=not_paid,selection=not_paid:Not Paid,partial:Partially Paid,paid:Paid"`
	LineIDs           sdk.One2Many[AccountMoveLine] `sumeru:"string=Journal Items"`
	Narration         sdk.Text                      `sumeru:"string=Notes"`
}
