package models

import (
	"sumeru/core/sdk"
)

type CrmActivityReport struct {
	sdk.Model `sumeru:"model=crm.activity.report"`

	ActivityID sdk.Many2One[MailActivity] `sumeru:"string=Activity"`
	LeadID     sdk.Many2One[CrmLead]      `sumeru:"string=Lead"`
	UserID     sdk.Many2One[CoreUser]       `sumeru:"string=Assigned To"`
	TeamID     sdk.Many2One[CrmTeam]        `sumeru:"string=Sales Team"`
	StageID    sdk.Many2One[CrmStage]       `sumeru:"string=Stage"`
	Summary    sdk.String                   `sumeru:"string=Summary"`
	DateDeadline sdk.Date                   `sumeru:"string=Due Date"`
	State      sdk.String                   `sumeru:"string=Status"`
}
