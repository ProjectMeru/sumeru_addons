package models

import (
	"sumeru/core/sdk"
)

type AccountMoveLine struct {
	sdk.Model `sumeru:"model=account.move.line"`

	MoveID         sdk.Many2One[AccountMove]    `sumeru:"required,index,string=Journal Entry"`
	AccountID      sdk.Many2One[AccountAccount] `sumeru:"string=Account"`
	Name           sdk.String                   `sumeru:"string=Label"`
	PartnerID      sdk.Many2One[CorePartner]    `sumeru:"string=Partner"`
	AnalyticAccountID sdk.Many2One[AccountAnalyticAccount] `sumeru:"string=Analytic Account"`
	ProductID      sdk.Many2One[sdk.Any]        `sumeru:"string=Product,comodel=product.product"`
	TaxID          sdk.Many2One[AccountTax]     `sumeru:"string=Tax"`
	TaxOriginID    sdk.Many2One[AccountTax]     `sumeru:"string=Originator Tax"`
	Quantity       sdk.Numeric                  `sumeru:"string=Quantity,precision=18,scale=3,default=1"`
	PriceUnit      sdk.Numeric                  `sumeru:"string=Unit Price,precision=18,scale=2,default=0"`
	PriceSubtotal  sdk.Numeric                  `sumeru:"string=Subtotal,precision=18,scale=2,default=0,readonly"`
	DisplayType    sdk.Selection[DisplayType]   `sumeru:"string=Display Type,default=product"`
	Debit          sdk.Numeric                  `sumeru:"string=Debit,precision=18,scale=2,default=0"`
	Credit         sdk.Numeric                  `sumeru:"string=Credit,precision=18,scale=2,default=0"`
	Balance        sdk.Numeric                  `sumeru:"string=Balance,precision=18,scale=2,default=0"`
	Reconciled     sdk.Boolean                  `sumeru:"string=Reconciled,default=false"`
	AmountResidual sdk.Numeric                  `sumeru:"string=Residual,precision=18,scale=2,default=0"`
}
