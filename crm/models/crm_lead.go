package models

import (
	"sumeru/core/base"
)

type CrmLead struct {
	base.BaseModel
}

func (l CrmLead) ModelName() string {
	return "crm.lead"
}

func (l CrmLead) Fields() []base.FieldDefinition {
	return []base.FieldDefinition{
		{Name: "name", Type: base.Char, String: "Opportunity", Required: true},
		{Name: "user_id", Type: base.Many2One, Relation: "core.user", String: "Salesperson"},
		{Name: "team_id", Type: base.Many2One, Relation: "crm.team", String: "Sales Team"},
		{Name: "stage_id", Type: base.Many2One, Relation: "crm.stage", String: "Stage"},
		{Name: "partner_id", Type: base.Many2One, Relation: "core.partner", String: "Customer"},
		{Name: "type", Type: base.Selection, String: "Type", Selection: [][]string{
			{"lead", "Lead"},
			{"opportunity", "Opportunity"},
		}, DefaultVal: "lead"},
		{Name: "priority", Type: base.Selection, String: "Priority", Selection: [][]string{
			{"0", "Low"},
			{"1", "Medium"},
			{"2", "High"},
			{"3", "Very High"},
		}, DefaultVal: "1"},
		{Name: "expected_revenue", Type: base.Numeric, String: "Expected Revenue"},
		{Name: "probability", Type: base.Float, String: "Probability"},
		{Name: "recurring_revenue", Type: base.Numeric, String: "Recurring Revenue"},
		{Name: "recurring_plan", Type: base.Many2One, Relation: "crm.recurring.plan", String: "Recurring Plan"},
		{Name: "lost_reason_id", Type: base.Many2One, Relation: "crm.lost.reason", String: "Lost Reason"},
		{Name: "date_deadline", Type: base.Date, String: "Expected Closing"},
		{Name: "description", Type: base.Text, String: "Notes"},
		{Name: "active", Type: base.Boolean, String: "Active", DefaultVal: true},
		{Name: "contact_name", Type: base.Char, String: "Contact Name"},
		{Name: "partner_name", Type: base.Char, String: "Company Name"},
		{Name: "email_from", Type: base.Char, String: "Email"},
		{Name: "phone", Type: base.Char, String: "Phone"},
		{Name: "website", Type: base.Char, String: "Website"},
		{Name: "street", Type: base.Char, String: "Street"},
		{Name: "street2", Type: base.Char, String: "Street 2"},
		{Name: "zip", Type: base.Char, String: "Zip"},
		{Name: "city", Type: base.Char, String: "City"},
		{Name: "color", Type: base.Integer, String: "Color Index", DefaultVal: 0},
	}
}

func init() {
	base.RegisterModel(base.RegisterModelInput{Model: &CrmLead{}, Module: "crm"})
}
