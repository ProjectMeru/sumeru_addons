package models

import (
	"sumeru/core/sdk"
)

type AccountDeferredSchedule struct {
	sdk.Model `sumeru:"model=account.deferred.schedule"`

	Name              sdk.String                   `sumeru:"required,string=Name"`
	MoveID            sdk.Many2One[AccountMove]    `sumeru:"string=Journal Entry"`
	AccountID         sdk.Many2One[AccountAccount] `sumeru:"string=Recognition Account"`
	StartDate         sdk.Date                     `sumeru:"string=Start Date"`
	EndDate           sdk.Date                     `sumeru:"string=End Date"`
	Amount            sdk.Numeric                  `sumeru:"string=Amount,precision=18,scale=2,default=0"`
	RecognizedAmount  sdk.Numeric                  `sumeru:"string=Recognized,precision=18,scale=2,default=0"`
	State             sdk.String                   `sumeru:"string=Status,default=draft"`
}
