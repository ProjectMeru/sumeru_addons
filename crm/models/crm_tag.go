package models

import (
	"sumeru/core/sdk"
)

type CrmTag struct {
	sdk.Model `sumeru:"model=crm.tag"`

	Name   sdk.String  `sumeru:"required,string=Tag Name"`
	Color  sdk.Integer `sumeru:"string=Color,default=0"`
	Active sdk.Boolean `sumeru:"string=Active,default=true"`
}
