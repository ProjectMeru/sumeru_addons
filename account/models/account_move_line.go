package models

import (
	"sumeru/core/sdk"
)

type AccountMoveLine struct {
	sdk.Model `sumeru:"model=account.move.line"`

	MoveID         sdk.Many2One[AccountMove]     `sumeru:"required,index,string=Journal Entry"`
	AccountID      sdk.Many2One[AccountAccount]  `sumeru:"string=Account"`
	Name           sdk.String                    `sumeru:"string=Label"`
	PartnerID      sdk.Many2One[sdk.Any]     `sumeru:"string=Partner,comodel=core.partner"`
	ProductID      sdk.Many2One[sdk.Any]  `sumeru:"string=Product,comodel=product.product"`
	TaxID          sdk.Many2One[AccountTax]      `sumeru:"string=Tax"`
	TaxOriginID    sdk.Many2One[AccountTax]      `sumeru:"string=Originator Tax"`
	Quantity       sdk.Numeric                   `sumeru:"string=Quantity,default=1"`
	PriceUnit      sdk.Numeric                   `sumeru:"string=Unit Price,default=0"`
	PriceSubtotal  sdk.Numeric                   `sumeru:"string=Subtotal,default=0"`
	DisplayType    sdk.String                    `sumeru:"string=Display Type,default=product,selection=product:Product,line_section:Section,line_note:Note,tax:Tax,entry:Journal Item"`
	Debit          sdk.Numeric                   `sumeru:"string=Debit,default=0"`
	Credit         sdk.Numeric                   `sumeru:"string=Credit,default=0"`
	Balance        sdk.Numeric                   `sumeru:"string=Balance,default=0"`
	Reconciled     sdk.Boolean                   `sumeru:"string=Reconciled,default=false"`
	AmountResidual sdk.Numeric                   `sumeru:"string=Residual,default=0"`
}
