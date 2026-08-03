package models

import "sumeru/core/sdk"

type PurchaseOrder struct {
	sdk.BaseModel
}

func (PurchaseOrder) ModelName() string { return "purchase.order" }

func (PurchaseOrder) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, String: "Reference", Required: true},
		{Name: "partner_id", Type: sdk.Many2One, Relation: "core.partner", String: "Vendor", Required: true},
		{Name: "user_id", Type: sdk.Many2One, Relation: "core.user", String: "Buyer"},
		{Name: "date_order", Type: sdk.DateTime, String: "Order Deadline"},
		{Name: "state", Type: sdk.Selection, String: "Status", Selection: [][]string{
			{"draft", "RFQ"},
			{"sent", "RFQ Sent"},
			{"purchase", "Purchase Order"},
			{"cancel", "Cancelled"},
		}, DefaultVal: "draft"},
		{Name: "amount_total", Type: sdk.Numeric, String: "Total", DefaultVal: 0},
		{Name: "notes", Type: sdk.Text, String: "Terms and Conditions"},
		{Name: "order_line", Type: sdk.One2Many, Relation: "purchase.order.line", String: "Order Lines"},
		{Name: "company_id", Type: sdk.Many2One, Relation: "core.company", String: "Company"},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &PurchaseOrder{}, Module: "purchase"})
}
