package models

import (
	"sumeru/core/sdk"
)

type CrmLostReason struct {
	sdk.Model `sumeru:"model=crm.lost.reason"`

	Name   sdk.String  `sumeru:"required,string=Description"`
	Active sdk.Boolean `sumeru:"string=Active,default=true"`
}
