package models

import (
	"sumeru/core/sdk"
)

type CrmActivityReport struct {
	sdk.Model `sumeru:"model=crm.activity.report"`

	ActivityID   sdk.Many2One[sdk.Any] `sumeru:"string=Activity,comodel=mail.activity"`
	LeadID       sdk.Many2One[CrmLead]      `sumeru:"string=Lead / Opportunity"`
	UserID       sdk.Many2One[sdk.Any]     `sumeru:"string=Assigned To,comodel=core.user"`
	TeamID       sdk.Many2One[CrmTeam]      `sumeru:"string=Sales Team"`
	StageID      sdk.Many2One[CrmStage]     `sumeru:"string=Stage"`
	Summary      sdk.Text                   `sumeru:"string=Summary"`
	DateDeadline sdk.Date                   `sumeru:"string=Due Date"`
	State        sdk.String                 `sumeru:"string=State,selection=planned:Planned,done:Done,cancelled:Cancelled"`
}
