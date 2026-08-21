package models

import (
	"sumeru/core/sdk"
)

type CrmLead struct {
	sdk.Model `sumeru:"model=crm.lead"`

	Name                   sdk.String                `sumeru:"required,string=Opportunity"`
	UserID                 sdk.Many2One[sdk.Any]    `sumeru:"string=Salesperson,comodel=core.user"`
	TeamID                 sdk.Many2One[CrmTeam]     `sumeru:"string=Sales Team"`
	StageID                sdk.Many2One[CrmStage]    `sumeru:"string=Stage"`
	PartnerID              sdk.Many2One[sdk.Any] `sumeru:"string=Customer,comodel=core.partner"`
	Type                   sdk.String                `sumeru:"string=Type,default=lead,selection=lead:Lead,opportunity:Opportunity"`
	Priority               sdk.String                `sumeru:"string=Priority,default=1,selection=0:Low,1:Medium,2:High,3:Very High"`
	ExpectedRevenue        sdk.Numeric               `sumeru:"string=Expected Revenue"`
	Probability            sdk.Float64               `sumeru:"string=Probability (%)"`
	AutomatedProbability   sdk.Float64               `sumeru:"string=Automated Probability"`
	IsAutomatedProbability sdk.Boolean               `sumeru:"string=Auto Probability,default=true"`
	RecurringRevenue       sdk.Numeric               `sumeru:"string=Recurring Revenue"`
	RecurringPlan          sdk.Many2One[CrmRecurringPlan] `sumeru:"string=Recurring Plan"`
	ProratedRevenue        sdk.Numeric               `sumeru:"string=Prorated Revenue"`
	LostReasonID           sdk.Many2One[CrmLostReason] `sumeru:"string=Lost Reason"`
	LostFeedback           sdk.Text                    `sumeru:"string=Closing Note"`
	WonStatus              sdk.String                `sumeru:"string=Won Status,default=pending,selection=pending:Pending,won:Won,lost:Lost"`
	DateDeadline           sdk.Date                  `sumeru:"string=Expected Closing"`
	DateClosed             sdk.Date                  `sumeru:"string=Closed Date"`
	DateOpen               sdk.Date                  `sumeru:"string=Assignment Date"`
	DateLastStageUpdate    sdk.DateTime              `sumeru:"string=Last Stage Update"`
	Description            sdk.Text                  `sumeru:"string=Notes"`
	Active                 sdk.Boolean               `sumeru:"string=Active,default=true"`
	ContactName            sdk.String                `sumeru:"string=Contact Name"`
	PartnerName            sdk.String                `sumeru:"string=Company Name"`
	EmailFrom              sdk.String                `sumeru:"string=Email"`
	Phone                  sdk.String                `sumeru:"string=Phone"`
	Website                sdk.String                `sumeru:"string=Website"`
	Street                 sdk.String                `sumeru:"string=Street"`
	Street2                sdk.String                `sumeru:"string=Street 2"`
	Zip                    sdk.String                `sumeru:"string=Zip"`
	City                   sdk.String                `sumeru:"string=City"`
	Color                  sdk.Integer               `sumeru:"string=Color Index,default=0"`
	TagIDs                 sdk.Many2Many[CrmTag]     `sumeru:"string=Tags,table=crm_lead_tag_rel,left=lead_id,right=tag_id"`
	CampaignID             sdk.Many2One[sdk.Any] `sumeru:"string=Campaign,comodel=utm.campaign"`
	MediumID               sdk.Many2One[sdk.Any]   `sumeru:"string=Medium,comodel=utm.medium"`
	SourceID               sdk.Many2One[sdk.Any]   `sumeru:"string=Source,comodel=utm.source"`
	Referred               sdk.String                `sumeru:"string=Referred By"`
}
