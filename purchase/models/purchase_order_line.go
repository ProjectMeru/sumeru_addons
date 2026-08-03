package models

import "sumeru/core/sdk"

type PurchaseOrderLine struct {
	sdk.BaseModel
}

func (PurchaseOrderLine) ModelName() string { return "purchase.order.line" }

func (PurchaseOrderLine) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "order_id", Type: sdk.Many2One, Relation: "purchase.order", String: "Order", Required: true, Index: true},
		{Name: "product_id", Type: sdk.Many2One, Relation: "product.product", String: "Product"},
		{Name: "name", Type: sdk.Char, String: "Description", Required: true},
		{Name: "product_qty", Type: sdk.Float, String: "Quantity", DefaultVal: 1},
		{Name: "price_unit", Type: sdk.Numeric, String: "Unit Price", DefaultVal: 0},
		{Name: "price_subtotal", Type: sdk.Numeric, String: "Subtotal", DefaultVal: 0},
		{Name: "sequence", Type: sdk.Integer, String: "Sequence", DefaultVal: 10},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &PurchaseOrderLine{}, Module: "purchase"})
}
