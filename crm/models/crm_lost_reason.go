package models

import (
	"sumeru/core/sdk"
)

type CrmLostReason struct {
	sdk.BaseModel
}

func (r CrmLostReason) ModelName() string {
	return "crm.lost.reason"
}

func (r CrmLostReason) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, String: "Description", Required: true},
		{Name: "active", Type: sdk.Boolean, String: "Active", DefaultVal: true},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &CrmLostReason{}, Module: "crm"})
}
