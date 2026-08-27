package models

import (
	"sumeru/core/sdk"
)

type SaleOrder struct {
	sdk.Model `sumeru:"model=sale.order"`

	Name          sdk.String                `sumeru:"required,string=Order Reference"`
	PartnerID     sdk.Many2One[sdk.Any] `sumeru:"required,string=Customer,comodel=core.partner"`
	UserID        sdk.Many2One[sdk.Any]    `sumeru:"string=Salesperson,comodel=core.user"`
	OpportunityID sdk.Many2One[sdk.Any]     `sumeru:"string=Opportunity,comodel=crm.lead"`
	DateOrder     sdk.DateTime              `sumeru:"string=Order Date"`
	State         sdk.String                `sumeru:"string=Status,default=draft,selection=draft:Quotation,sent:Quotation Sent,sale:Sales Order,cancel:Cancelled"`
	InvoiceStatus sdk.String                `sumeru:"string=Invoice Status,default=no,selection=no:Nothing to Invoice,to invoice:To Invoice,invoiced:Fully Invoiced"`
	AmountUntaxed sdk.Numeric               `sumeru:"string=Untaxed Amount,default=0"`
	AmountTotal   sdk.Numeric               `sumeru:"string=Total,default=0"`
	Note          sdk.Text                  `sumeru:"string=Terms and Conditions"`
	OrderLine     sdk.One2Many[SaleOrderLine] `sumeru:"string=Order Lines"`
	CompanyID     sdk.Many2One[sdk.Any] `sumeru:"string=Company,comodel=core.company"`
}
