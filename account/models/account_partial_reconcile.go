package models

import (
	"sumeru/core/sdk"
)

type AccountPartialReconcile struct {
	sdk.Model `sumeru:"model=account.partial.reconcile"`

	DebitMoveID     sdk.Many2One[AccountMoveLine]     `sumeru:"required,index,string=Debit Move Line"`
	CreditMoveID    sdk.Many2One[AccountMoveLine]     `sumeru:"required,index,string=Credit Move Line"`
	Amount          sdk.Numeric                       `sumeru:"string=Amount,precision=18,scale=2,default=0"`
	FullReconcileID sdk.Many2One[AccountFullReconcile] `sumeru:"string=Full Reconcile"`
}
