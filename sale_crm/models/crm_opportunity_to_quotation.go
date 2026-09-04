package models

import (
	"sumeru/core/sdk"
)

type CrmOpportunityToQuotation struct {
	sdk.Model `sumeru:"model=crm.opportunity.to.quotation"`

	LeadID    sdk.Many2One[CrmLead]       `sumeru:"required,string=Opportunity"`
	PartnerID sdk.Many2One[CorePartner]   `sumeru:"string=Customer"`
	ProductID sdk.Many2One[ProductProduct] `sumeru:"string=Product"`
	Quantity  sdk.Numeric                 `sumeru:"string=Quantity,precision=18,scale=2,default=1"`
	MarkWon   sdk.Boolean                 `sumeru:"string=Mark Won,default=false"`
}
