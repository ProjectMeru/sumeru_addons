package models

import "sumeru/core/sdk"

type CrmActivityReport struct {
	sdk.BaseModel
}

func (CrmActivityReport) ModelName() string { return "crm.activity.report" }

func (CrmActivityReport) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "activity_id", Type: sdk.Many2One, Relation: "mail.activity", String: "Activity"},
		{Name: "lead_id", Type: sdk.Many2One, Relation: "crm.lead", String: "Lead / Opportunity"},
		{Name: "user_id", Type: sdk.Many2One, Relation: "core.user", String: "Assigned To"},
		{Name: "team_id", Type: sdk.Many2One, Relation: "crm.team", String: "Sales Team"},
		{Name: "stage_id", Type: sdk.Many2One, Relation: "crm.stage", String: "Stage"},
		{Name: "summary", Type: sdk.Text, String: "Summary"},
		{Name: "date_deadline", Type: sdk.Date, String: "Due Date"},
		{Name: "state", Type: sdk.Selection, String: "State", Selection: [][]string{
			{"planned", "Planned"},
			{"done", "Done"},
			{"cancelled", "Cancelled"},
		}},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &CrmActivityReport{}, Module: "crm"})
}
