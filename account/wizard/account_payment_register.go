package wizard

import (
	"sumeru/core/sdk"
)

type AccountPaymentRegister struct {
	sdk.Model `sumeru:"model=account.payment.register"`

	InvoiceID     sdk.Many2One[sdk.Any] `sumeru:"required,string=Invoice,comodel=account.move"`
	PartnerID     sdk.Many2One[sdk.Any] `sumeru:"string=Partner,comodel=core.partner"`
	Amount        sdk.Numeric           `sumeru:"required,string=Amount,default=0"`
	JournalID     sdk.Many2One[sdk.Any] `sumeru:"string=Payment Journal,comodel=account.journal"`
	PaymentDate   sdk.Date              `sumeru:"string=Payment Date"`
	Communication sdk.String            `sumeru:"string=Memo"`
	PaymentType   sdk.String            `sumeru:"string=Payment Type,default=inbound,selection=inbound:Receive,outbound:Send"`
}
