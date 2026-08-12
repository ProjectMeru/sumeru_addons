package models

import "sumeru/core/sdk"

type AccountPaymentTerm struct {
	sdk.BaseModel
}

func (AccountPaymentTerm) ModelName() string { return "account.payment.term" }

func (AccountPaymentTerm) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, String: "Payment Terms", Required: true},
		{Name: "note", Type: sdk.Text, String: "Description"},
		{Name: "days", Type: sdk.Integer, String: "Due Days", DefaultVal: 0},
		{Name: "active", Type: sdk.Boolean, String: "Active", DefaultVal: true},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &AccountPaymentTerm{}, Module: "account"})
}
