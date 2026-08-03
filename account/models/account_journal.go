package models

import "sumeru/core/sdk"

type AccountJournal struct {
	sdk.BaseModel
}

func (AccountJournal) ModelName() string { return "account.journal" }

func (AccountJournal) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, String: "Journal Name", Required: true},
		{Name: "code", Type: sdk.Char, String: "Short Code", Required: true},
		{Name: "type", Type: sdk.Selection, String: "Type", Selection: [][]string{
			{"sale", "Sales"},
			{"purchase", "Purchase"},
			{"general", "Miscellaneous"},
			{"bank", "Bank"},
			{"cash", "Cash"},
		}, DefaultVal: "general"},
		{Name: "default_account_id", Type: sdk.Many2One, Relation: "account.account", String: "Default Account"},
		{Name: "active", Type: sdk.Boolean, String: "Active", DefaultVal: true},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &AccountJournal{}, Module: "account"})
}
