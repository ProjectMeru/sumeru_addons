package models

import "sumeru/core/sdk"

type HrJob struct {
	sdk.BaseModel
}

func (HrJob) ModelName() string { return "hr.job" }

func (HrJob) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, String: "Job Position", Required: true},
		{Name: "department_id", Type: sdk.Many2One, Relation: "hr.department", String: "Department"},
		{Name: "description", Type: sdk.Text, String: "Job Description"},
		{Name: "active", Type: sdk.Boolean, String: "Active", DefaultVal: true},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &HrJob{}, Module: "hr"})
}
