package models

import (
	"sumeru/core/base"
)

type CrmLostReason struct {
	base.BaseModel
}

func (r CrmLostReason) ModelName() string {
	return "crm.lost.reason"
}

func (r CrmLostReason) Fields() []base.FieldDefinition {
	return []base.FieldDefinition{
		{Name: "name", Type: base.Char, String: "Description", Required: true},
		{Name: "active", Type: base.Boolean, String: "Active", DefaultVal: true},
	}
}

func init() {
	base.RegisterModel(base.RegisterModelInput{Model: &CrmLostReason{}, Module: "crm"})
}
