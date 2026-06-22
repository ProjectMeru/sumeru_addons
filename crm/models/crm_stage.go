package models

import (
	"sumeru/core/base"
)

type CrmStage struct {
	base.BaseModel
}

func (s CrmStage) ModelName() string {
	return "crm.stage"
}

func (s CrmStage) Fields() []base.FieldDefinition {
	return []base.FieldDefinition{
		{Name: "name", Type: base.Char, String: "Stage Name", Required: true},
		{Name: "sequence", Type: base.Integer, String: "Sequence", DefaultVal: 1},
		{Name: "is_won", Type: base.Boolean, String: "Is Won Stage?"},
		{Name: "fold", Type: base.Boolean, String: "Folded in Pipeline"},
		{Name: "requirements", Type: base.Text, String: "Requirements"},
		{Name: "active", Type: base.Boolean, String: "Active", DefaultVal: true},
		{Name: "team_ids", Type: base.Many2Many, Relation: "crm.team", RelationTable: "crm_stage_team_rel", Column1: "stage_id", Column2: "team_id", String: "Sales Teams"},
		{Name: "rotting_threshold_days", Type: base.Integer, String: "Days to Rot", DefaultVal: 0},
		{Name: "color", Type: base.Integer, String: "Color Index", DefaultVal: 0},
	}
}

func init() {
	base.RegisterModel(base.RegisterModelInput{Model: &CrmStage{}, Module: "crm"})
}
