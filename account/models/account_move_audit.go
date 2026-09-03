package models

import (
	"sumeru/core/sdk"
)

type AccountMoveAudit struct {
	sdk.Model `sumeru:"model=account.move.audit"`

	MoveID   sdk.Many2One[AccountMove] `sumeru:"required,string=Journal Entry"`
	UserID   sdk.Many2One[CoreUser]    `sumeru:"string=User"`
	Action   sdk.String                `sumeru:"required,string=Action"`
	Snapshot sdk.Text                  `sumeru:"string=Snapshot"`
	Date     sdk.DateTime              `sumeru:"string=Date"`
}
