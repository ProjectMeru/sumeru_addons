package models

import "sumeru/core/sdk"

type AccountFullReconcile struct {
	sdk.BaseModel
}

func (AccountFullReconcile) ModelName() string { return "account.full.reconcile" }

func (AccountFullReconcile) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, String: "Number", Required: true},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &AccountFullReconcile{}, Module: "account"})
}
