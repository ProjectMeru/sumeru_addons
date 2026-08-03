package models

import (
	"sumeru/core/sdk"
)

type CrmStage struct {
	sdk.BaseModel
}

func (s CrmStage) ModelName() string {
	return "crm.stage"
}

func (s CrmStage) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, String: "Stage Name", Required: true},
		{Name: "sequence", Type: sdk.Integer, String: "Sequence", DefaultVal: 1},
		{Name: "is_won", Type: sdk.Boolean, String: "Is Won Stage?"},
		{Name: "fold", Type: sdk.Boolean, String: "Folded in Pipeline"},
		{Name: "requirements", Type: sdk.Text, String: "Requirements"},
		{Name: "active", Type: sdk.Boolean, String: "Active", DefaultVal: true},
		{Name: "team_ids", Type: sdk.Many2Many, Relation: "crm.team", RelationTable: "crm_stage_team_rel", Column1: "stage_id", Column2: "team_id", String: "Sales Teams"},
		{Name: "rotting_threshold_days", Type: sdk.Integer, String: "Days to Rot", DefaultVal: 0},
		{Name: "color", Type: sdk.Integer, String: "Color Index", DefaultVal: 0},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &CrmStage{}, Module: "crm"})
}
