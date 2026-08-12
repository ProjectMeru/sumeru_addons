package models

import "sumeru/core/sdk"

type ProductProduct struct {
	sdk.BaseModel
}

func (ProductProduct) ModelName() string { return "product.product" }

func (ProductProduct) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, String: "Product Name", Required: true},
		{Name: "image", Type: sdk.Text, String: "Image"},
		{Name: "default_code", Type: sdk.Char, String: "Internal Reference"},
		{Name: "type", Type: sdk.Selection, String: "Product Type", Selection: [][]string{
			{"consu", "Consumable"},
			{"service", "Service"},
		}, DefaultVal: "consu"},
		{Name: "categ_id", Type: sdk.Many2One, Relation: "product.category", String: "Category"},
		{Name: "list_price", Type: sdk.Numeric, String: "Sales Price", DefaultVal: 0},
		{Name: "standard_price", Type: sdk.Numeric, String: "Cost", DefaultVal: 0},
		{Name: "description", Type: sdk.Text, String: "Description"},
		{Name: "property_account_income_id", Type: sdk.Many2One, Relation: "account.account", String: "Income Account"},
		{Name: "property_account_expense_id", Type: sdk.Many2One, Relation: "account.account", String: "Expense Account"},
		{Name: "sale_ok", Type: sdk.Boolean, String: "Can be Sold", DefaultVal: true},
		{Name: "purchase_ok", Type: sdk.Boolean, String: "Can be Purchased", DefaultVal: true},
		{Name: "active", Type: sdk.Boolean, String: "Active", DefaultVal: true},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &ProductProduct{}, Module: "product"})
}
