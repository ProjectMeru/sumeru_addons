package models

import (
	"sumeru/core/sdk"
)

type PurchaseOrder struct {
	sdk.Model `sumeru:"model=purchase.order"`

	Name        sdk.String                    `sumeru:"required,string=Reference"`
	PartnerID   sdk.Many2One[sdk.Any]     `sumeru:"required,string=Vendor,comodel=core.partner"`
	UserID      sdk.Many2One[sdk.Any]        `sumeru:"string=Buyer,comodel=core.user"`
	DateOrder   sdk.DateTime                  `sumeru:"string=Order Deadline"`
	State       sdk.String                    `sumeru:"string=Status,default=draft,selection=draft:RFQ,sent:RFQ Sent,purchase:Purchase Order,cancel:Cancelled"`
	AmountTotal sdk.Numeric                   `sumeru:"string=Total,default=0"`
	Notes       sdk.Text                      `sumeru:"string=Terms and Conditions"`
	OrderLine   sdk.One2Many[PurchaseOrderLine] `sumeru:"string=Order Lines"`
	CompanyID   sdk.Many2One[sdk.Any]     `sumeru:"string=Company,comodel=core.company"`
}
