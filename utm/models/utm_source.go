package models

import (
	"sumeru/core/sdk"
)

type UtmSource struct {
	sdk.Model `sumeru:"model=utm.source"`

	Name   sdk.String  `sumeru:"required,string=Source"`
	Active sdk.Boolean `sumeru:"string=Active,default=true"`
}
