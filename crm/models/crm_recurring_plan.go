package models

import (
	"sumeru/core/base"
)

type CrmRecurringPlan struct {
	base.BaseModel
}

func (p CrmRecurringPlan) ModelName() string {
	return "crm.recurring.plan"
}

func (p CrmRecurringPlan) Fields() []base.FieldDefinition {
	return []base.FieldDefinition{
		{Name: "name", Type: base.Char, String: "Plan Name", Required: true},
		{Name: "number_of_months", Type: base.Integer, String: "# Months", Required: true},
		{Name: "active", Type: base.Boolean, String: "Active", DefaultVal: true},
		{Name: "sequence", Type: base.Integer, String: "Sequence", DefaultVal: 10},
	}
}

func init() {
	base.RegisterModel(base.RegisterModelInput{Model: &CrmRecurringPlan{}, Module: "crm"})
}
