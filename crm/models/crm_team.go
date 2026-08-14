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
		{Name: "sequence", Type: sdk.Integer, String: "Sequence", DefaultVal: 10},
		{Name: "user_id", Type: sdk.Many2One, Relation: "core.user", String: "Team Leader"},
		{Name: "company_id", Type: sdk.Many2One, Relation: "core.company", String: "Company"},
		{Name: "use_leads", Type: sdk.Boolean, String: "Leads", DefaultVal: true},
		{Name: "use_opportunities", Type: sdk.Boolean, String: "Pipeline", DefaultVal: true},
		{Name: "lead_ids", Type: sdk.One2Many, Relation: "crm.lead", String: "Leads"},
		{Name: "member_ids", Type: sdk.One2Many, Relation: "crm.team.member", String: "Members"},
		{Name: "assignment_enabled", Type: sdk.Boolean, String: "Lead Assignment", DefaultVal: false},
		{Name: "assignment_max", Type: sdk.Integer, String: "Max Leads / Member", DefaultVal: 30},
		{Name: "assignment_domain", Type: sdk.Char, String: "Assignment Domain"},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &CrmTeam{}, Module: "crm"})
}
