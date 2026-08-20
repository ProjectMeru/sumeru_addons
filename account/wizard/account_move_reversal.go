package wizard

import (
	"sumeru/core/sdk"
)

type AccountMoveReversal struct {
	sdk.Model `sumeru:"model=account.move.reversal"`

	MoveID sdk.Many2One[AccountMove] `sumeru:"required,string=Move"`
	Date   sdk.Date                  `sumeru:"string=Reversal Date"`
	Reason sdk.String                `sumeru:"string=Reason"`
}
