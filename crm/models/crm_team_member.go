package models

import (
	"sumeru/core/sdk"
)

type CrmTeamMember struct {
	sdk.Model `sumeru:"model=crm.team.member"`

	TeamID           sdk.Many2One[CrmTeam]  `sumeru:"required,string=Sales Team"`
	UserID           sdk.Many2One[sdk.Any] `sumeru:"required,string=Member,comodel=core.user"`
	Active           sdk.Boolean            `sumeru:"string=Active,default=true"`
	AssignmentOptout sdk.Boolean            `sumeru:"string=Skip Auto Assignment"`
	AssignmentMax    sdk.Integer            `sumeru:"string=Max Leads,default=30"`
	LeadMonthCount   sdk.Integer            `sumeru:"string=Leads This Month,default=0"`
}
