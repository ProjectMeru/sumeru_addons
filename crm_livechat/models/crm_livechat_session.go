package models

import (
	"sumeru/core/sdk"
)

type CrmLivechatSession struct {
	sdk.Model `sumeru:"model=crm.livechat.session"`

	Name       sdk.String                  `sumeru:"required,string=Session"`
	LeadID     sdk.Many2One[any]           `sumeru:"comodel=crm.lead,string=Lead"`
	OperatorID sdk.Many2One[CoreUser]      `sumeru:"string=Operator"`
	State      sdk.Selection[LivechatState] `sumeru:"string=State,default=new"`
}
