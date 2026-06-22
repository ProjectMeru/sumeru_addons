package models

import (
	"sumeru/core/base"
)

type CrmTeam struct {
	base.BaseModel
}

func (t CrmTeam) ModelName() string {
	return "crm.team"
}

func (t CrmTeam) Fields() []base.FieldDefinition {
	return []base.FieldDefinition{
		{Name: "name", Type: base.Char, String: "Sales Team", Required: true},
		{Name: "active", Type: base.Boolean, String: "Active", DefaultVal: true},
		{Name: "user_id", Type: base.Many2One, Relation: "core.user", String: "Team Leader"},
		{Name: "use_leads", Type: base.Boolean, String: "Leads", DefaultVal: true},
		{Name: "use_opportunities", Type: base.Boolean, String: "Pipeline", DefaultVal: true},
		{Name: "lead_ids", Type: base.One2Many, Relation: "crm.lead", String: "Leads"},
	}
}

func init() {
	base.RegisterModel(base.RegisterModelInput{Model: &CrmTeam{}, Module: "crm"})
}
