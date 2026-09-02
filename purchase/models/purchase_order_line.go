package models

import (
	"sumeru/core/sdk"
)

type PurchaseOrderLine struct {
	sdk.Model `sumeru:"model=purchase.order.line"`

	OrderID       sdk.Many2One[PurchaseOrder]  `sumeru:"required,index,string=Order"`
	ProductID     sdk.Many2One[ProductProduct] `sumeru:"string=Product"`
	Name          sdk.String                   `sumeru:"required,string=Description"`
	ProductQty    sdk.Float                    `sumeru:"string=Quantity,default=1"`
	PriceUnit     sdk.Numeric                  `sumeru:"string=Unit Price,precision=18,scale=2,default=0"`
	PriceSubtotal sdk.Numeric                  `sumeru:"string=Subtotal,precision=18,scale=2,default=0"`
	Sequence      sdk.Integer                  `sumeru:"string=Sequence,default=10"`
}
