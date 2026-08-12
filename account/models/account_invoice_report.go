package models

import "sumeru/core/sdk"

// AccountInvoiceReport is a stored analysis row (demo/reporting sample).
type AccountInvoiceReport struct {
	sdk.BaseModel
}

func (AccountInvoiceReport) ModelName() string { return "account.invoice.report" }

func (AccountInvoiceReport) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "move_id", Type: sdk.Many2One, Relation: "account.move", String: "Invoice", Index: true},
		{Name: "name", Type: sdk.Char, String: "Number"},
		{Name: "partner_id", Type: sdk.Many2One, Relation: "core.partner", String: "Partner"},
		{Name: "move_type", Type: sdk.Selection, String: "Type", Selection: [][]string{
			{"out_invoice", "Customer Invoice"},
			{"out_refund", "Credit Note"},
			{"in_invoice", "Vendor Bill"},
			{"in_refund", "Vendor Refund"},
		}},
		{Name: "invoice_date", Type: sdk.Date, String: "Invoice Date"},
		{Name: "state", Type: sdk.Selection, String: "Status", Selection: [][]string{
			{"draft", "Draft"},
			{"posted", "Posted"},
			{"cancel", "Cancelled"},
		}},
		{Name: "payment_state", Type: sdk.Selection, String: "Payment", Selection: [][]string{
			{"not_paid", "Not Paid"},
			{"partial", "Partial"},
			{"paid", "Paid"},
		}},
		{Name: "amount_untaxed", Type: sdk.Numeric, String: "Untaxed", DefaultVal: 0},
		{Name: "amount_tax", Type: sdk.Numeric, String: "Tax", DefaultVal: 0},
		{Name: "amount_total", Type: sdk.Numeric, String: "Total", DefaultVal: 0},
		{Name: "amount_residual", Type: sdk.Numeric, String: "Residual", DefaultVal: 0},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &AccountInvoiceReport{}, Module: "account"})
}
