package models

import (
	"sumeru/core/sdk"
)

type CrmLead struct {
	sdk.Model `sumeru:"model=crm.lead"`

	Name                   sdk.String                       `sumeru:"required,string=Opportunity"`
	UserID                 sdk.Many2One[CoreUser]           `sumeru:"string=Salesperson"`
	TeamID                 sdk.Many2One[CrmTeam]            `sumeru:"string=Sales Team"`
	StageID                sdk.Many2One[CrmStage]           `sumeru:"string=Stage"`
	PartnerID              sdk.Many2One[CorePartner]        `sumeru:"string=Customer"`
	Type                   sdk.Selection[LeadType]          `sumeru:"string=Type,default=lead"`
	Priority               sdk.Selection[LeadPriority]      `sumeru:"string=Priority,default=1"`
	ExpectedRevenue        sdk.Numeric                      `sumeru:"string=Expected Revenue,precision=18,scale=2"`
	Probability            sdk.Float                        `sumeru:"string=Probability (%)"`
	AutomatedProbability   sdk.Float                        `sumeru:"string=Automated Probability"`
	IsAutomatedProbability sdk.Boolean                    `sumeru:"string=Auto Probability,default=true"`
	RecurringRevenue       sdk.Numeric                      `sumeru:"string=Recurring Revenue,precision=18,scale=2"`
	RecurringPlan          sdk.Many2One[CrmRecurringPlan]   `sumeru:"string=Recurring Plan"`
	ProratedRevenue        sdk.Numeric                      `sumeru:"string=Prorated Revenue,precision=18,scale=2"`
	LostReasonID           sdk.Many2One[CrmLostReason]      `sumeru:"string=Lost Reason"`
	LostFeedback           sdk.Text                         `sumeru:"string=Closing Note"`
	WonStatus              sdk.Selection[WonStatus]         `sumeru:"string=Won Status,default=pending"`
	DateDeadline           sdk.Date                         `sumeru:"string=Expected Closing"`
	DateClosed             sdk.Date                         `sumeru:"string=Closed Date"`
	DateOpen               sdk.Date                         `sumeru:"string=Assignment Date"`
	DateLastStageUpdate    sdk.DateTime                     `sumeru:"string=Last Stage Update"`
	Description            sdk.Text                         `sumeru:"string=Notes"`
	Active                 sdk.Boolean                      `sumeru:"string=Active,default=true"`
	ContactName            sdk.String                       `sumeru:"string=Contact Name"`
	PartnerName            sdk.String                       `sumeru:"string=Company Name"`
	EmailFrom              sdk.String                       `sumeru:"string=Email"`
	Phone                  sdk.String                       `sumeru:"string=Phone"`
	Website                sdk.String                       `sumeru:"string=Website"`
	Street                 sdk.String                       `sumeru:"string=Street"`
	Street2                sdk.String                       `sumeru:"string=Street 2"`
	Zip                    sdk.String                       `sumeru:"string=Zip"`
	City                   sdk.String                       `sumeru:"string=City"`
	Color                  sdk.Integer                      `sumeru:"string=Color Index,default=0"`
	TagIDs                 sdk.Many2Many[CrmTag]            `sumeru:"string=Tags,table=crm_lead_tag_rel,left=lead_id,right=tag_id"`
	CampaignID             sdk.Many2One[UtmCampaign]        `sumeru:"string=Campaign"`
	MediumID               sdk.Many2One[UtmMedium]          `sumeru:"string=Medium"`
	SourceID               sdk.Many2One[UtmSource]          `sumeru:"string=Source"`
	Referred               sdk.String                       `sumeru:"string=Referred By"`
	CompanyID              sdk.Many2One[CoreCompany]        `sumeru:"string=Company"`
	LangID                 sdk.Many2One[CoreLang]           `sumeru:"string=Language"`
	StateID                sdk.Many2One[CoreCountryState]   `sumeru:"string=State"`
	CountryID              sdk.Many2One[CoreCountry]        `sumeru:"string=Country"`
	Function               sdk.String                       `sumeru:"string=Job Position"`
	Mobile                 sdk.String                       `sumeru:"string=Mobile"`
	RecurringRevenueMonthly sdk.Numeric                     `sumeru:"string=Monthly Recurring Revenue,precision=18,scale=2"`
	CompanyCurrencyID      sdk.Many2One[CoreCurrency]       `sumeru:"string=Company Currency"`
	DateConversion         sdk.Date                         `sumeru:"string=Conversion Date"`
	DateAutomationLast     sdk.DateTime                     `sumeru:"string=Last Automation"`
	DayOpen                sdk.Integer                      `sumeru:"string=Days to Assign"`
	DayClose               sdk.Integer                      `sumeru:"string=Days to Close"`
	DuplicateLeadCount     sdk.Integer                      `sumeru:"string=Duplicate Count,default=0"`
	DuplicateLeadIDs       sdk.String                       `sumeru:"string=Duplicate Lead IDs"`
	MeetingCount           sdk.Integer                      `sumeru:"string=Meetings,default=0"`
	QuotationCount         sdk.Integer                      `sumeru:"string=Quotations,default=0"`
	LeadProperties         sdk.Text                         `sumeru:"string=Properties"`
}
