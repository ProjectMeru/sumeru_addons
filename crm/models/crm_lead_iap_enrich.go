package models

import (
	"sumeru/core/sdk"
)

// CrmLeadIapEnrich extends crm.lead with IAP enrichment tracking.
type CrmLeadIapEnrich struct {
	sdk.Model `sumeru:"inherit=crm.lead"`

	IapEnrichDone sdk.Boolean `sumeru:"string=IAP Enrichment Done,default=false"`
}
