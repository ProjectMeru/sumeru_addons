package wizard

import (
	"sumeru/core/sdk"
)

type CrmMergeOpportunity struct {
	sdk.Model `sumeru:"model=crm.merge.opportunity"`

	LeadIds sdk.String `sumeru:"string=Lead IDs"`
	Name    sdk.String `sumeru:"string=Merged Name"`
}
