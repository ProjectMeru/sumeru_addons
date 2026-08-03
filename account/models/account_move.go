package models

import "sumeru/core/sdk"

type AccountMove struct {
	sdk.BaseModel
}

func (AccountMove) ModelName() string { return "account.move" }

func (AccountMove) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, String: "Number", Required: true},
		{Name: "move_type", Type: sdk.Selection, String: "Type", Selection: [][]string{
			{"entry", "Journal Entry"},
			{"out_invoice", "Customer Invoice"},
			{"out_refund", "Customer Credit Note"},
			{"in_invoice", "Vendor Bill"},
			{"in_refund", "Vendor Credit Note"},
		}, DefaultVal: "entry"},
		{Name: "partner_id", Type: sdk.Many2One, Relation: "core.partner", String: "Partner"},
		{Name: "journal_id", Type: sdk.Many2One, Relation: "account.journal", String: "Journal"},
		{Name: "date", Type: sdk.Date, String: "Date"},
		{Name: "state", Type: sdk.Selection, String: "Status", Selection: [][]string{
			{"draft", "Draft"},
			{"posted", "Posted"},
			{"cancel", "Cancelled"},
		}, DefaultVal: "draft"},
		{Name: "invoice_origin", Type: sdk.Char, String: "Source Document"},
		{Name: "ref", Type: sdk.Char, String: "Reference"},
		{Name: "amount_total", Type: sdk.Numeric, String: "Total", DefaultVal: 0},
		{Name: "payment_state", Type: sdk.Selection, String: "Payment Status", Selection: [][]string{
			{"not_paid", "Not Paid"},
			{"partial", "Partially Paid"},
			{"paid", "Paid"},
		}, DefaultVal: "not_paid"},
		{Name: "line_ids", Type: sdk.One2Many, Relation: "account.move.line", String: "Journal Items"},
		{Name: "narration", Type: sdk.Text, String: "Notes"},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &AccountMove{}, Module: "account"})
}
