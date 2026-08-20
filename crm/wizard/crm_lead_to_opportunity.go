package wizard

import (
	"sumeru/core/sdk"
)

type CrmLead2Opportunity struct {
	sdk.Model `sumeru:"model=crm.lead2opportunity"`

	LeadID    sdk.Many2One[CrmLead]      `sumeru:"required,string=Lead"`
	PartnerID sdk.Many2One[sdk.Any]  `sumeru:"string=Customer,comodel=core.partner"`
	Name      sdk.String                 `sumeru:"string=Opportunity Name"`
	UserID    sdk.Many2One[sdk.Any]     `sumeru:"string=Salesperson,comodel=core.user"`
	TeamID    sdk.Many2One[CrmTeam]      `sumeru:"string=Sales Team"`
}
