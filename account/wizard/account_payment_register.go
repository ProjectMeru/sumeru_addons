package wizard

import "sumeru/core/sdk"

type AccountPaymentRegister struct {
	sdk.BaseModel
}

func (AccountPaymentRegister) ModelName() string { return "account.payment.register" }

func (AccountPaymentRegister) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "invoice_id", Type: sdk.Many2One, Relation: "account.move", String: "Invoice", Required: true},
		{Name: "partner_id", Type: sdk.Many2One, Relation: "core.partner", String: "Partner"},
		{Name: "amount", Type: sdk.Numeric, String: "Amount", Required: true, DefaultVal: 0},
		{Name: "journal_id", Type: sdk.Many2One, Relation: "account.journal", String: "Payment Journal"},
		{Name: "payment_date", Type: sdk.Date, String: "Payment Date"},
		{Name: "communication", Type: sdk.Char, String: "Memo"},
		{Name: "payment_type", Type: sdk.Selection, String: "Payment Type", Selection: [][]string{
			{"inbound", "Receive"},
			{"outbound", "Send"},
		}, DefaultVal: "inbound"},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &AccountPaymentRegister{}, Module: "account"})
}
