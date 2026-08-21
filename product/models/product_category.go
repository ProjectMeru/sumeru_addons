package models

import (
	"sumeru/core/sdk"
)

type ProductCategory struct {
	sdk.Model `sumeru:"model=product.category"`

	Name     sdk.String                `sumeru:"required,string=Category"`
	ParentID sdk.Many2One[ProductCategory] `sumeru:"string=Parent"`
	Sequence sdk.Integer               `sumeru:"string=Sequence,default=10"`
	Active   sdk.Boolean               `sumeru:"string=Active,default=true"`
}
