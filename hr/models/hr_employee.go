package models

import "sumeru/core/sdk"

type HrEmployee struct {
	sdk.BaseModel
}

func (HrEmployee) ModelName() string { return "hr.employee" }

func (HrEmployee) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, String: "Employee Name", Required: true},
		{Name: "image", Type: sdk.Text, String: "Image"},
		{Name: "work_email", Type: sdk.Char, String: "Work Email"},
		{Name: "work_phone", Type: sdk.Char, String: "Work Phone"},
		{Name: "department_id", Type: sdk.Many2One, Relation: "hr.department", String: "Department"},
		{Name: "job_id", Type: sdk.Many2One, Relation: "hr.job", String: "Job Position"},
		{Name: "parent_id", Type: sdk.Many2One, Relation: "hr.employee", String: "Manager"},
		{Name: "user_id", Type: sdk.Many2One, Relation: "core.user", String: "Related User"},
		{Name: "partner_id", Type: sdk.Many2One, Relation: "core.partner", String: "Related Contact"},
		{Name: "company_id", Type: sdk.Many2One, Relation: "core.company", String: "Company"},
		{Name: "active", Type: sdk.Boolean, String: "Active", DefaultVal: true},
		{Name: "notes", Type: sdk.Text, String: "Notes"},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &HrEmployee{}, Module: "hr"})
}
