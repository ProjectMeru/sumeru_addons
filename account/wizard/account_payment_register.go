package wizard

import (
	"sumeru/core/sdk"
	"sumeru_addons/account/models"
)

type AccountPaymentRegister struct {
	sdk.Model `sumeru:"model=account.payment.register"`

	InvoiceID     sdk.Many2One[models.AccountMove]    `sumeru:"required,string=Invoice"`
	PartnerID     sdk.Many2One[models.CorePartner]    `sumeru:"string=Partner"`
	Amount        sdk.Numeric                         `sumeru:"required,string=Amount,precision=18,scale=2,default=0"`
	JournalID     sdk.Many2One[models.AccountJournal] `sumeru:"string=Payment Journal"`
	PaymentDate   sdk.Date                            `sumeru:"string=Payment Date"`
	Communication sdk.String                          `sumeru:"string=Memo"`
	PaymentType   sdk.Selection[models.PaymentType]   `sumeru:"string=Payment Type,default=inbound"`
}
