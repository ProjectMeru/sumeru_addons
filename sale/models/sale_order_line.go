package models

import (
	"sumeru/core/sdk"
)

type SaleOrderLine struct {
	sdk.Model `sumeru:"model=sale.order.line"`

	OrderID       sdk.Many2One[SaleOrder]      `sumeru:"required,index,string=Order"`
	ProductID     sdk.Many2One[sdk.Any] `sumeru:"string=Product,comodel=product.product"`
	Name          sdk.String                   `sumeru:"required,string=Description"`
	ProductUomQty sdk.Float64                  `sumeru:"string=Quantity,default=1"`
	PriceUnit     sdk.Numeric                  `sumeru:"string=Unit Price,default=0"`
	PriceSubtotal sdk.Numeric                  `sumeru:"string=Subtotal,default=0,readonly"`
	Sequence      sdk.Integer                  `sumeru:"string=Sequence,default=10"`
}
