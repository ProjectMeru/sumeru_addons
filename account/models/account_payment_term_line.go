package models

import (
	"sumeru/core/sdk"
)

type AccountPaymentTermLine struct {
	sdk.Model `sumeru:"model=account.payment.term.line"`

	PaymentTermID sdk.Many2One[AccountPaymentTerm]   `sumeru:"required,index,string=Payment Terms"`
	Value         sdk.Selection[PaymentTermValue]    `sumeru:"string=Value Type,default=percent"`
	ValueAmount   sdk.Numeric                        `sumeru:"string=Value,precision=18,scale=2,default=100"`
	Days          sdk.Integer                        `sumeru:"string=Days,default=0"`
	DelayType     sdk.Selection[PaymentTermDelay]    `sumeru:"string=Delay Type,default=days_after"`
	Sequence      sdk.Integer                        `sumeru:"string=Sequence,default=10"`
}
