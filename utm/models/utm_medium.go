package models

import (
	"sumeru/core/sdk"
)

type UtmMedium struct {
	sdk.Model `sumeru:"model=utm.medium"`

	Name   sdk.String  `sumeru:"required,string=Medium"`
	Active sdk.Boolean `sumeru:"string=Active,default=true"`
}
