package wizard

import "sumeru/core/sdk"

// AccountMoveReversal creates a credit note / reverse move from an invoice.
type AccountMoveReversal struct {
	sdk.BaseModel
}

func (AccountMoveReversal) ModelName() string { return "account.move.reversal" }

func (AccountMoveReversal) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "move_id", Type: sdk.Many2One, Relation: "account.move", String: "Move", Required: true},
		{Name: "date", Type: sdk.Date, String: "Reversal Date"},
		{Name: "reason", Type: sdk.Char, String: "Reason"},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &AccountMoveReversal{}, Module: "account"})
}
