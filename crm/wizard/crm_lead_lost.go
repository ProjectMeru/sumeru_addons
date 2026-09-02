package wizard

import (
	"sumeru/core/sdk"
	"sumeru_addons/crm/models"
)

type CrmLeadLost struct {
	sdk.Model `sumeru:"model=crm.lead.lost"`

	LeadID       sdk.Many2One[models.CrmLead]       `sumeru:"required,string=Lead"`
	LostReasonID sdk.Many2One[models.CrmLostReason] `sumeru:"string=Lost Reason"`
	LostFeedback sdk.Text                          `sumeru:"string=Closing Note"`
}
