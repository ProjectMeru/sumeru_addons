package models

import (
	"sumeru/core/sdk"
)

type CrmLead struct {
	sdk.BaseModel
}

func (l CrmLead) ModelName() string {
	return "crm.lead"
}

func (l CrmLead) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, String: "Opportunity", Required: true},
		{Name: "user_id", Type: sdk.Many2One, Relation: "core.user", String: "Salesperson"},
		{Name: "team_id", Type: sdk.Many2One, Relation: "crm.team", String: "Sales Team"},
		{Name: "stage_id", Type: sdk.Many2One, Relation: "crm.stage", String: "Stage"},
		{Name: "partner_id", Type: sdk.Many2One, Relation: "core.partner", String: "Customer"},
		{Name: "type", Type: sdk.Selection, String: "Type", Selection: [][]string{
			{"lead", "Lead"},
			{"opportunity", "Opportunity"},
		}, DefaultVal: "lead"},
		{Name: "priority", Type: sdk.Selection, String: "Priority", Selection: [][]string{
			{"0", "Low"},
			{"1", "Medium"},
			{"2", "High"},
			{"3", "Very High"},
		}, DefaultVal: "1"},
		{Name: "expected_revenue", Type: sdk.Numeric, String: "Expected Revenue"},
		{Name: "probability", Type: sdk.Float, String: "Probability"},
		{Name: "recurring_revenue", Type: sdk.Numeric, String: "Recurring Revenue"},
		{Name: "recurring_plan", Type: sdk.Many2One, Relation: "crm.recurring.plan", String: "Recurring Plan"},
		{Name: "lost_reason_id", Type: sdk.Many2One, Relation: "crm.lost.reason", String: "Lost Reason"},
		{Name: "date_deadline", Type: sdk.Date, String: "Expected Closing"},
		{Name: "description", Type: sdk.Text, String: "Notes"},
		{Name: "active", Type: sdk.Boolean, String: "Active", DefaultVal: true},
		{Name: "contact_name", Type: sdk.Char, String: "Contact Name"},
		{Name: "partner_name", Type: sdk.Char, String: "Company Name"},
		{Name: "email_from", Type: sdk.Char, String: "Email"},
		{Name: "phone", Type: sdk.Char, String: "Phone"},
		{Name: "website", Type: sdk.Char, String: "Website"},
		{Name: "street", Type: sdk.Char, String: "Street"},
		{Name: "street2", Type: sdk.Char, String: "Street 2"},
		{Name: "zip", Type: sdk.Char, String: "Zip"},
		{Name: "city", Type: sdk.Char, String: "City"},
		{Name: "color", Type: sdk.Integer, String: "Color Index", DefaultVal: 0},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &CrmLead{}, Module: "crm"})
}
