package wizard

import (
	"sumeru/core/sdk"
	"sumeru_addons/crm/models"
)

type CrmLead2OpportunityMass struct {
	sdk.Model `sumeru:"model=crm.lead2opportunity.mass"`

	LeadIDs   sdk.String                     `sumeru:"required,string=Lead IDs"`
	UserID    sdk.Many2One[models.CoreUser]  `sumeru:"string=Salesperson"`
	TeamID    sdk.Many2One[models.CrmTeam]   `sumeru:"string=Sales Team"`
	ForceAssign sdk.Boolean                  `sumeru:"string=Force Assignment,default=false"`
}
