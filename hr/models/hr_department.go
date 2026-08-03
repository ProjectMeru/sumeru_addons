package models

import "sumeru/core/sdk"

type HrDepartment struct {
	sdk.BaseModel
}

func (HrDepartment) ModelName() string { return "hr.department" }

func (HrDepartment) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, String: "Department", Required: true},
		{Name: "parent_id", Type: sdk.Many2One, Relation: "hr.department", String: "Parent Department"},
		{Name: "manager_id", Type: sdk.Many2One, Relation: "hr.employee", String: "Manager"},
		{Name: "active", Type: sdk.Boolean, String: "Active", DefaultVal: true},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &HrDepartment{}, Module: "hr"})
}
