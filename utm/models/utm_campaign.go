package models

import "sumeru/core/sdk"

type UtmCampaign struct{ sdk.BaseModel }

func (UtmCampaign) ModelName() string { return "utm.campaign" }

func (UtmCampaign) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, String: "Campaign", Required: true},
		{Name: "active", Type: sdk.Boolean, String: "Active", DefaultVal: true},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &UtmCampaign{}, Module: "utm"})
}
