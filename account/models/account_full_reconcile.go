package models

import (
	"sumeru/core/sdk"
)

type AccountFullReconcile struct {
	sdk.Model `sumeru:"model=account.full.reconcile"`

	Name sdk.String `sumeru:"required,string=Number"`
}
