package models

import "sumeru/core/sdk"

type ProductCategory struct {
	sdk.BaseModel
}

func (ProductCategory) ModelName() string { return "product.category" }

func (ProductCategory) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, String: "Category", Required: true},
		{Name: "parent_id", Type: sdk.Many2One, Relation: "product.category", String: "Parent"},
		{Name: "sequence", Type: sdk.Integer, String: "Sequence", DefaultVal: 10},
		{Name: "active", Type: sdk.Boolean, String: "Active", DefaultVal: true},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &ProductCategory{}, Module: "product"})
}
