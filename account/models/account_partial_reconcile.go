package models

import "sumeru/core/sdk"

type AccountPartialReconcile struct {
	sdk.BaseModel
}

func (AccountPartialReconcile) ModelName() string { return "account.partial.reconcile" }

func (AccountPartialReconcile) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "debit_move_id", Type: sdk.Many2One, Relation: "account.move.line", String: "Debit Move Line", Required: true, Index: true},
		{Name: "credit_move_id", Type: sdk.Many2One, Relation: "account.move.line", String: "Credit Move Line", Required: true, Index: true},
		{Name: "amount", Type: sdk.Numeric, String: "Amount", DefaultVal: 0},
		{Name: "full_reconcile_id", Type: sdk.Many2One, Relation: "account.full.reconcile", String: "Full Reconcile"},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &AccountPartialReconcile{}, Module: "account"})
}
