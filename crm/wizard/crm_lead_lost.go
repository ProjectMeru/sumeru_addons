package wizard

import (
	"sumeru/core/sdk"
)

type CrmLeadLost struct {
	sdk.Model `sumeru:"model=crm.lead.lost"`

	LeadID       sdk.Many2One[sdk.Any] `sumeru:"required,string=Lead,comodel=crm.lead"`
	LostReasonID sdk.Many2One[sdk.Any] `sumeru:"string=Lost Reason,comodel=crm.lost.reason"`
	LostFeedback sdk.Text              `sumeru:"string=Closing Note"`
}
