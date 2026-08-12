package models

import "sumeru/core/sdk"

type AccountTaxRepartitionLine struct {
	sdk.BaseModel
}

func (AccountTaxRepartitionLine) ModelName() string { return "account.tax.repartition.line" }

func (AccountTaxRepartitionLine) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "tax_id", Type: sdk.Many2One, Relation: "account.tax", String: "Tax", Required: true, Index: true},
		{Name: "repartition_type", Type: sdk.Selection, String: "Type", Selection: [][]string{
			{"base", "Base"},
			{"tax", "Tax"},
		}, DefaultVal: "tax"},
		{Name: "factor_percent", Type: sdk.Numeric, String: "Factor (%)", DefaultVal: 100},
		{Name: "account_id", Type: sdk.Many2One, Relation: "account.account", String: "Account"},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &AccountTaxRepartitionLine{}, Module: "account"})
}
