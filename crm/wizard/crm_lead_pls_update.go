package wizard

import (
	"sumeru/core/sdk"
)

type CrmLeadPlsUpdate struct {
	sdk.Model `sumeru:"model=crm.lead.pls.update"`

	StartDate sdk.Date   `sumeru:"string=Start Date"`
	FieldList sdk.String `sumeru:"string=Fields"`
}
