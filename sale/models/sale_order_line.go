package models

import (
	"sumeru/core/sdk"
)

type SaleOrderLine struct {
	sdk.Model `sumeru:"model=sale.order.line"`

	OrderID       sdk.Many2One[SaleOrder]      `sumeru:"required,index,string=Order"`
	ProductID     sdk.Many2One[ProductProduct] `sumeru:"string=Product"`
	Name          sdk.String                   `sumeru:"required,string=Description"`
<<<<<<< HEAD
	ProductUomQty sdk.Float64                  `sumeru:"string=Quantity,default=1"`
	PriceUnit     sdk.Numeric                  `sumeru:"string=Unit Price,default=0"`
	PriceSubtotal sdk.Numeric                  `sumeru:"string=Subtotal,default=0,readonly"`
=======
	ProductUomQty sdk.Float                    `sumeru:"string=Quantity,default=1"`
	PriceUnit     sdk.Numeric                  `sumeru:"string=Unit Price,precision=18,scale=2,default=0"`
	PriceSubtotal sdk.Numeric                  `sumeru:"string=Subtotal,precision=18,scale=2,default=0"`
>>>>>>> 1a55542 (feat(sale): add order actions, sequences, and typed model fields)
	Sequence      sdk.Integer                  `sumeru:"string=Sequence,default=10"`
}
