package models

import (
	"sumeru/core/sdk"
)

type CrmRecurringPlan struct {
	sdk.Model `sumeru:"model=crm.recurring.plan"`

	Name           sdk.String  `sumeru:"required,string=Plan Name"`
	NumberOfMonths sdk.Integer `sumeru:"required,string=# Months"`
	Active         sdk.Boolean `sumeru:"string=Active,default=true"`
	Sequence       sdk.Integer `sumeru:"string=Sequence,default=10"`
}
