package models

import (
	"sumeru/core/sdk"
)

type IapAccount struct {
	sdk.Model `sumeru:"model=iap.account"`

	Name         sdk.String  `sumeru:"required,string=Account Name"`
	ServiceName  sdk.String  `sumeru:"required,string=Service"`
	AccountToken sdk.String  `sumeru:"string=Account Token"`
	Balance      sdk.Numeric `sumeru:"string=Balance,precision=18,scale=2,default=0"`
}
