package models

import "sumeru/core/sdk"

type UtmSource struct{ sdk.BaseModel }

func (UtmSource) ModelName() string { return "utm.source" }

func (UtmSource) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, String: "Source", Required: true},
		{Name: "active", Type: sdk.Boolean, String: "Active", DefaultVal: true},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &UtmSource{}, Module: "utm"})
}
