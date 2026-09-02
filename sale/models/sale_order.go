package models

import (
	"sumeru/core/sdk"
)

type SaleOrder struct {
	sdk.Model `sumeru:"model=sale.order"`

	Name          sdk.String                    `sumeru:"required,string=Order Reference"`
	PartnerID     sdk.Many2One[CorePartner]     `sumeru:"required,string=Customer"`
	UserID        sdk.Many2One[CoreUser]        `sumeru:"string=Salesperson"`
	OpportunityID sdk.Many2One[CrmLead]         `sumeru:"string=Opportunity"`
	DateOrder     sdk.DateTime                  `sumeru:"string=Order Date"`
	State         sdk.Selection[SaleOrderState] `sumeru:"string=Status,default=draft"`
	InvoiceStatus sdk.Selection[InvoiceStatus]  `sumeru:"string=Invoice Status,default=no"`
	AmountUntaxed sdk.Numeric                   `sumeru:"string=Untaxed Amount,precision=18,scale=2,default=0"`
	AmountTotal   sdk.Numeric                   `sumeru:"string=Total,precision=18,scale=2,default=0"`
	Note          sdk.Text                      `sumeru:"string=Terms and Conditions"`
	OrderLine     sdk.One2Many[SaleOrderLine]   `sumeru:"string=Order Lines"`
	CompanyID     sdk.Many2One[CoreCompany]     `sumeru:"string=Company"`
}
