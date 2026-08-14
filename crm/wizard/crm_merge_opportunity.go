package wizard

import "sumeru/core/sdk"

type CrmMergeOpportunity struct {
	sdk.BaseModel
}

func (CrmMergeOpportunity) ModelName() string { return "crm.merge.opportunity" }

func (CrmMergeOpportunity) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "lead_ids", Type: sdk.Char, String: "Lead IDs"},
		{Name: "name", Type: sdk.Char, String: "Merged Name"},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &CrmMergeOpportunity{}, Module: "crm"})
}
