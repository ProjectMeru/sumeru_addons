package models

import "sumeru/core/sdk"

type AccountAccount struct {
	sdk.BaseModel
}

func (AccountAccount) ModelName() string { return "account.account" }

func (AccountAccount) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "code", Type: sdk.Char, String: "Code", Required: true, Unique: true},
		{Name: "name", Type: sdk.Char, String: "Account Name", Required: true},
		{Name: "account_type", Type: sdk.Selection, String: "Type", Selection: [][]string{
			{"asset_receivable", "Receivable"},
			{"asset_cash", "Bank and Cash"},
			{"asset_current", "Current Assets"},
			{"liability_payable", "Payable"},
			{"liability_current", "Current Liabilities"},
			{"equity", "Equity"},
			{"income", "Income"},
			{"expense", "Expenses"},
		}, Required: true},
		{Name: "reconcile", Type: sdk.Boolean, String: "Allow Reconciliation", DefaultVal: false},
		{Name: "active", Type: sdk.Boolean, String: "Active", DefaultVal: true},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &AccountAccount{}, Module: "account"})
}
