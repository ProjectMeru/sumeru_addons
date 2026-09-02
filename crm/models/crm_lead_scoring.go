package models

import (
	"sumeru/core/sdk"
)

type CrmLeadScoringFrequency struct {
	sdk.Model `sumeru:"model=crm.lead.scoring.frequency"`

	Variable sdk.String  `sumeru:"required,string=Variable"`
	WonCount sdk.Integer `sumeru:"string=Won Count,default=0"`
	LostCount sdk.Integer `sumeru:"string=Lost Count,default=0"`
}

type CrmLeadScoringFrequencyField struct {
	sdk.Model `sumeru:"model=crm.lead.scoring.frequency.field"`

	FrequencyID sdk.Many2One[CrmLeadScoringFrequency] `sumeru:"required,index,string=Frequency"`
	Field       sdk.String                            `sumeru:"required,string=Field"`
	Value       sdk.String                            `sumeru:"required,string=Value"`
	WonCount    sdk.Integer                           `sumeru:"string=Won Count,default=0"`
	LostCount   sdk.Integer                           `sumeru:"string=Lost Count,default=0"`
}
