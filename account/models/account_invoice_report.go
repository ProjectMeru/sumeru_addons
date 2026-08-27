package models

import (
	"sumeru/core/sdk"
)

type AccountInvoiceReport struct {
	sdk.Model `sumeru:"model=account.invoice.report"`

	MoveID         sdk.Many2One[AccountMove]     `sumeru:"index,string=Invoice"`
	Name           sdk.String                    `sumeru:"string=Number"`
	PartnerID      sdk.Many2One[sdk.Any]     `sumeru:"string=Partner,comodel=core.partner"`
	MoveType       sdk.String                    `sumeru:"string=Type,selection=out_invoice:Customer Invoice,out_refund:Credit Note,in_invoice:Vendor Bill,in_refund:Vendor Refund"`
	InvoiceDate    sdk.Date                      `sumeru:"string=Invoice Date"`
	State          sdk.String                    `sumeru:"string=Status,selection=draft:Draft,posted:Posted,cancel:Cancelled"`
	PaymentState   sdk.String                    `sumeru:"string=Payment,selection=not_paid:Not Paid,partial:Partial,paid:Paid"`
	AmountUntaxed  sdk.Numeric                   `sumeru:"string=Untaxed,default=0"`
	AmountTax      sdk.Numeric                   `sumeru:"string=Tax,default=0"`
	AmountTotal    sdk.Numeric                   `sumeru:"string=Total,default=0"`
	AmountResidual sdk.Numeric                   `sumeru:"string=Residual,default=0"`
}
