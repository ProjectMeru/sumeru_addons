package models

import (
	"sumeru/core/sdk"
)

type CrmTeamMember struct {
	sdk.BaseModel
}

func (m CrmTeamMember) ModelName() string {
	return "crm.team.member"
}

func (m CrmTeamMember) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "team_id", Type: sdk.Many2One, Relation: "crm.team", String: "Sales Team", Required: true},
		{Name: "user_id", Type: sdk.Many2One, Relation: "core.user", String: "Member", Required: true},
		{Name: "active", Type: sdk.Boolean, String: "Active", DefaultVal: true},
		{Name: "assignment_optout", Type: sdk.Boolean, String: "Skip Auto Assignment"},
		{Name: "assignment_max", Type: sdk.Integer, String: "Max Leads", DefaultVal: 30},
		{Name: "lead_month_count", Type: sdk.Integer, String: "Leads This Month", DefaultVal: 0},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &CrmTeamMember{}, Module: "crm"})
}
