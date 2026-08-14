package wizard

import "sumeru/core/sdk"

type CrmLeadLost struct {
	sdk.BaseModel
}

func (CrmLeadLost) ModelName() string { return "crm.lead.lost" }

func (CrmLeadLost) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "lead_id", Type: sdk.Many2One, Relation: "crm.lead", String: "Lead", Required: true},
		{Name: "lost_reason_id", Type: sdk.Many2One, Relation: "crm.lost.reason", String: "Lost Reason"},
		{Name: "lost_feedback", Type: sdk.Text, String: "Closing Note"},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &CrmLeadLost{}, Module: "crm"})
}
