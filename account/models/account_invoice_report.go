package models

import (
	"sumeru/core/sdk"
)

type AccountInvoiceReport struct {
	sdk.Model `sumeru:"model=account.invoice.report"`

	MoveID         sdk.Many2One[AccountMove]   `sumeru:"index,string=Invoice"`
	Name           sdk.String                  `sumeru:"string=Number"`
	PartnerID      sdk.Many2One[CorePartner]   `sumeru:"string=Partner"`
	MoveType       sdk.Selection[MoveType]     `sumeru:"string=Type"`
	InvoiceDate    sdk.Date                    `sumeru:"string=Invoice Date"`
	State          sdk.Selection[MoveState]    `sumeru:"string=Status"`
	PaymentState   sdk.Selection[PaymentState] `sumeru:"string=Payment"`
	AmountUntaxed  sdk.Numeric                 `sumeru:"string=Untaxed,precision=18,scale=2,default=0"`
	AmountTax      sdk.Numeric                 `sumeru:"string=Tax,precision=18,scale=2,default=0"`
	AmountTotal    sdk.Numeric                 `sumeru:"string=Total,precision=18,scale=2,default=0"`
	AmountResidual sdk.Numeric                 `sumeru:"string=Residual,precision=18,scale=2,default=0"`
}
