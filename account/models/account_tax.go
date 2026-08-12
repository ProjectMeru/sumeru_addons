package models

import "sumeru/core/sdk"

type AccountTax struct {
	sdk.BaseModel
}

func (AccountTax) ModelName() string { return "account.tax" }

func (AccountTax) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, String: "Tax Name", Required: true},
		{Name: "amount", Type: sdk.Numeric, String: "Amount (%)", DefaultVal: 0},
		{Name: "type_tax_use", Type: sdk.Selection, String: "Tax Type", Selection: [][]string{
			{"sale", "Sales"},
			{"purchase", "Purchase"},
			{"none", "None"},
		}, DefaultVal: "sale"},
		{Name: "account_id", Type: sdk.Many2One, Relation: "account.account", String: "Tax Account"},
		{Name: "active", Type: sdk.Boolean, String: "Active", DefaultVal: true},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &AccountTax{}, Module: "account"})
}
