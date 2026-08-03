package models

import (
	"sumeru/core/sdk"
)

type CrmTeam struct {
	sdk.BaseModel
}

func (t CrmTeam) ModelName() string {
	return "crm.team"
}

func (t CrmTeam) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, String: "Sales Team", Required: true},
		{Name: "active", Type: sdk.Boolean, String: "Active", DefaultVal: true},
		{Name: "user_id", Type: sdk.Many2One, Relation: "core.user", String: "Team Leader"},
		{Name: "use_leads", Type: sdk.Boolean, String: "Leads", DefaultVal: true},
		{Name: "use_opportunities", Type: sdk.Boolean, String: "Pipeline", DefaultVal: true},
		{Name: "lead_ids", Type: sdk.One2Many, Relation: "crm.lead", String: "Leads"},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &CrmTeam{}, Module: "crm"})
}
