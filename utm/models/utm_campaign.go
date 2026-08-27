package models

import (
	"sumeru/core/sdk"
)

type UtmCampaign struct {
	sdk.Model `sumeru:"model=utm.campaign"`

	Name   sdk.String  `sumeru:"required,string=Campaign"`
	Active sdk.Boolean `sumeru:"string=Active,default=true"`
}
