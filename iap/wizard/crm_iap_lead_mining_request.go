package wizard

import (
	"sumeru/core/sdk"
)

type CrmIapLeadMiningRequest struct {
	sdk.Model `sumeru:"model=crm.iap.lead.mining.request"`

	CountryID sdk.Many2One[sdk.Any] `sumeru:"comodel=core.country,string=Country"`
	Industry  sdk.String            `sumeru:"string=Industry"`
	Size      sdk.Selection[string] `sumeru:"string=Company Size,selection=1-10:1-10,11-50:11-50,51-200:51-200,200+:200+,default=11-50"`
}
