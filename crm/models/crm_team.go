package models

import (
	"sumeru/core/sdk"
)

type CrmTeam struct {
	sdk.Model `sumeru:"model=crm.team"`

	Name              sdk.String                `sumeru:"required,string=Sales Team"`
	Active            sdk.Boolean               `sumeru:"string=Active,default=true"`
	Sequence          sdk.Integer               `sumeru:"string=Sequence,default=10"`
	UserID            sdk.Many2One[CoreUser]    `sumeru:"string=Team Leader"`
	CompanyID         sdk.Many2One[CoreCompany] `sumeru:"string=Company"`
	UseLeads          sdk.Boolean               `sumeru:"string=Leads,default=true"`
	UseOpportunities  sdk.Boolean               `sumeru:"string=Pipeline,default=true"`
	LeadIDs           sdk.One2Many[CrmLead]     `sumeru:"string=Leads"`
	MemberIDs         sdk.One2Many[CrmTeamMember] `sumeru:"string=Members"`
	AssignmentEnabled sdk.Boolean               `sumeru:"string=Lead Assignment,default=false"`
	AssignmentMax     sdk.Integer               `sumeru:"string=Max Leads / Member,default=30"`
	AssignmentDomain           sdk.String  `sumeru:"string=Assignment Domain"`
	LeadPropertiesDefinition   sdk.Text    `sumeru:"string=Lead Properties Definition"`
	InvoicingTarget            sdk.Numeric `sumeru:"string=Invoicing Target,precision=18,scale=2"`
	QuotationTarget            sdk.Numeric `sumeru:"string=Quotation Target,precision=18,scale=2"`
}
