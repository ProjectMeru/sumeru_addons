package models

import (
	"sumeru/core/sdk"
)

type CrmTag struct {
	sdk.BaseModel
}

func (t CrmTag) ModelName() string {
	return "crm.tag"
}

func (t CrmTag) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, String: "Tag Name", Required: true},
		{Name: "color", Type: sdk.Integer, String: "Color", DefaultVal: 0},
		{Name: "active", Type: sdk.Boolean, String: "Active", DefaultVal: true},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &CrmTag{}, Module: "crm"})
}
