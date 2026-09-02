package models

import (
	"sumeru/core/sdk"
)

type AccountPaymentTerm struct {
	sdk.Model `sumeru:"model=account.payment.term"`

	Name   sdk.String  `sumeru:"required,string=Payment Terms"`
	Note   sdk.Text    `sumeru:"string=Description"`
	Days    sdk.Integer                    `sumeru:"string=Due Days,default=0"`
	LineIDs sdk.One2Many[AccountPaymentTermLine] `sumeru:"string=Terms"`
	Active  sdk.Boolean                    `sumeru:"string=Active,default=true"`
}
