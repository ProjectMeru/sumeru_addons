package models

import (
	"sumeru/core/sdk"
)

type CrmRecurringPlan struct {
	sdk.BaseModel
}

func (p CrmRecurringPlan) ModelName() string {
	return "crm.recurring.plan"
}

func (p CrmRecurringPlan) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, String: "Plan Name", Required: true},
		{Name: "number_of_months", Type: sdk.Integer, String: "# Months", Required: true},
		{Name: "active", Type: sdk.Boolean, String: "Active", DefaultVal: true},
		{Name: "sequence", Type: sdk.Integer, String: "Sequence", DefaultVal: 10},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &CrmRecurringPlan{}, Module: "crm"})
}
