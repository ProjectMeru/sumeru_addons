package wizard

import (
	"sumeru/core/sdk"
)

type CrmLeadLost struct {
	sdk.Model `sumeru:"model=crm.lead.lost"`

	LeadID       sdk.Many2One[CrmLead]       `sumeru:"required,string=Lead"`
	LostReasonID sdk.Many2One[CrmLostReason] `sumeru:"string=Lost Reason"`
	LostFeedback sdk.Text                    `sumeru:"string=Closing Note"`
}
