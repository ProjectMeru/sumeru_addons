package models

import (
	"sumeru/core/sdk"
)

type PurchaseOrder struct {
	sdk.Model `sumeru:"model=purchase.order"`

	Name          sdk.String                         `sumeru:"required,string=Reference"`
	PartnerID     sdk.Many2One[CorePartner]          `sumeru:"required,string=Vendor"`
	UserID        sdk.Many2One[CoreUser]             `sumeru:"string=Buyer"`
	DateOrder     sdk.DateTime                       `sumeru:"string=Order Deadline"`
	State         sdk.Selection[PurchaseOrderState]  `sumeru:"string=Status,default=draft"`
	InvoiceStatus sdk.Selection[BillingStatus]       `sumeru:"string=Billing Status,default=no"`
	AmountTotal   sdk.Numeric                        `sumeru:"string=Total,precision=18,scale=2,default=0"`
	Notes         sdk.Text                           `sumeru:"string=Terms and Conditions"`
	OrderLine     sdk.One2Many[PurchaseOrderLine]    `sumeru:"string=Order Lines"`
	CompanyID     sdk.Many2One[CoreCompany]          `sumeru:"string=Company"`
}
