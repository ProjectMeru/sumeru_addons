package wizard

import "sumeru/core/sdk"

type CrmLead2Opportunity struct {
	sdk.BaseModel
}

func (CrmLead2Opportunity) ModelName() string { return "crm.lead2opportunity" }

func (CrmLead2Opportunity) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "lead_id", Type: sdk.Many2One, Relation: "crm.lead", String: "Lead", Required: true},
		{Name: "partner_id", Type: sdk.Many2One, Relation: "core.partner", String: "Customer"},
		{Name: "name", Type: sdk.Char, String: "Opportunity Name"},
		{Name: "user_id", Type: sdk.Many2One, Relation: "core.user", String: "Salesperson"},
		{Name: "team_id", Type: sdk.Many2One, Relation: "crm.team", String: "Sales Team"},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &CrmLead2Opportunity{}, Module: "crm"})
}
