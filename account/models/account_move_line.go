package models

import "sumeru/core/sdk"

type AccountMoveLine struct {
	sdk.BaseModel
}

func (AccountMoveLine) ModelName() string { return "account.move.line" }

func (AccountMoveLine) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "move_id", Type: sdk.Many2One, Relation: "account.move", String: "Journal Entry", Required: true, Index: true},
		{Name: "account_id", Type: sdk.Many2One, Relation: "account.account", String: "Account", Required: true},
		{Name: "name", Type: sdk.Char, String: "Label"},
		{Name: "partner_id", Type: sdk.Many2One, Relation: "core.partner", String: "Partner"},
		{Name: "debit", Type: sdk.Numeric, String: "Debit", DefaultVal: 0},
		{Name: "credit", Type: sdk.Numeric, String: "Credit", DefaultVal: 0},
		{Name: "balance", Type: sdk.Numeric, String: "Balance", DefaultVal: 0},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &AccountMoveLine{}, Module: "account"})
}
