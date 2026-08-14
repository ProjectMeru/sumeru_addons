package models

import "sumeru/core/sdk"

type UtmMedium struct{ sdk.BaseModel }

func (UtmMedium) ModelName() string { return "utm.medium" }

func (UtmMedium) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, String: "Medium", Required: true},
		{Name: "active", Type: sdk.Boolean, String: "Active", DefaultVal: true},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &UtmMedium{}, Module: "utm"})
}
