package models

import (
	"sumeru/core/sdk"
)

type WebsiteCrmForm struct {
	sdk.Model `sumeru:"model=website.crm.form"`

	Name   sdk.String        `sumeru:"required,string=Form Name"`
	TeamID sdk.Many2One[any] `sumeru:"comodel=crm.team,string=Sales Team"`
	Active sdk.Boolean       `sumeru:"string=Active,default=true"`
}
