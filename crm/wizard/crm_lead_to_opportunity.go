package wizard

import (
	"sumeru/core/sdk"
	"sumeru_addons/crm/models"
)

type CrmLead2Opportunity struct {
	sdk.Model `sumeru:"model=crm.lead2opportunity"`

	LeadID    sdk.Many2One[models.CrmLead]   `sumeru:"required,string=Lead"`
	PartnerID sdk.Many2One[models.CorePartner] `sumeru:"string=Customer"`
	Name      sdk.String                     `sumeru:"string=Opportunity Name"`
	UserID    sdk.Many2One[models.CoreUser]  `sumeru:"string=Salesperson"`
	TeamID    sdk.Many2One[models.CrmTeam]   `sumeru:"string=Sales Team"`
}
