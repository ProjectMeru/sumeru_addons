package models

import "sumeru/core/sdk"

// AccountPayment records inbound/outbound payments against invoices or bills.
type AccountPayment struct {
	sdk.BaseModel
}

func (AccountPayment) ModelName() string { return "account.payment" }

func (AccountPayment) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, String: "Number"},
		{Name: "payment_type", Type: sdk.Selection, String: "Payment Type", Selection: [][]string{
			{"inbound", "Receive"},
			{"outbound", "Send"},
		}, DefaultVal: "inbound", Required: true},
		{Name: "partner_type", Type: sdk.Selection, String: "Partner Type", Selection: [][]string{
			{"customer", "Customer"},
			{"supplier", "Vendor"},
		}, DefaultVal: "customer"},
		{Name: "partner_id", Type: sdk.Many2One, Relation: "core.partner", String: "Partner"},
		{Name: "amount", Type: sdk.Numeric, String: "Amount", DefaultVal: 0, Required: true},
		{Name: "date", Type: sdk.Date, String: "Date"},
		{Name: "journal_id", Type: sdk.Many2One, Relation: "account.journal", String: "Journal"},
		{Name: "memo", Type: sdk.Char, String: "Memo"},
		{Name: "state", Type: sdk.Selection, String: "Status", Selection: [][]string{
			{"draft", "Draft"},
			{"posted", "Posted"},
			{"cancelled", "Cancelled"},
		}, DefaultVal: "draft"},
		{Name: "move_id", Type: sdk.Many2One, Relation: "account.move", String: "Journal Entry"},
		{Name: "invoice_id", Type: sdk.Many2One, Relation: "account.move", String: "Invoice / Bill"},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &AccountPayment{}, Module: "account"})
}
