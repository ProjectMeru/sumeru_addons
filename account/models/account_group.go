package models

import (
	"sumeru/core/sdk"
)

type AccountGroup struct {
	sdk.Model `sumeru:"model=account.group"`

	Name       sdk.String              `sumeru:"required,string=Name"`
	CodePrefix sdk.String              `sumeru:"string=Code Prefix"`
	ParentID   sdk.Many2One[AccountGroup] `sumeru:"string=Parent Group"`
	Active     sdk.Boolean             `sumeru:"string=Active,default=true"`
}
