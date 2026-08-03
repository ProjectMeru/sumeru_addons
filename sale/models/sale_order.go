package models

import "sumeru/core/sdk"

type SaleOrder struct {
	sdk.BaseModel
}

func (SaleOrder) ModelName() string { return "sale.order" }

func (SaleOrder) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, String: "Order Reference", Required: true},
		{Name: "partner_id", Type: sdk.Many2One, Relation: "core.partner", String: "Customer", Required: true},
		{Name: "user_id", Type: sdk.Many2One, Relation: "core.user", String: "Salesperson"},
		{Name: "opportunity_id", Type: sdk.Many2One, Relation: "crm.lead", String: "Opportunity"},
		{Name: "date_order", Type: sdk.DateTime, String: "Order Date"},
		{Name: "state", Type: sdk.Selection, String: "Status", Selection: [][]string{
			{"draft", "Quotation"},
			{"sent", "Quotation Sent"},
			{"sale", "Sales Order"},
			{"cancel", "Cancelled"},
		}, DefaultVal: "draft"},
		{Name: "invoice_status", Type: sdk.Selection, String: "Invoice Status", Selection: [][]string{
			{"no", "Nothing to Invoice"},
			{"to invoice", "To Invoice"},
			{"invoiced", "Fully Invoiced"},
		}, DefaultVal: "no"},
		{Name: "amount_untaxed", Type: sdk.Numeric, String: "Untaxed Amount", DefaultVal: 0},
		{Name: "amount_total", Type: sdk.Numeric, String: "Total", DefaultVal: 0},
		{Name: "note", Type: sdk.Text, String: "Terms and Conditions"},
		{Name: "order_line", Type: sdk.One2Many, Relation: "sale.order.line", String: "Order Lines"},
		{Name: "company_id", Type: sdk.Many2One, Relation: "core.company", String: "Company"},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &SaleOrder{}, Module: "sale"})
}
