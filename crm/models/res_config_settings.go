package models

import (
	"sumeru/core/sdk"
)

// ResConfigSettings extends platform settings with CRM options.
type ResConfigSettings struct {
	sdk.Model `sumeru:"inherit=res.config.settings"`

	GroupUseLead               sdk.Boolean `sumeru:"string=Leads"`
	GroupUseRecurringRevenues  sdk.Boolean `sumeru:"string=Recurring Revenues"`
	IsMembershipMulti          sdk.Boolean `sumeru:"string=Multi Teams"`
	CrmUseAutoAssignment       sdk.Boolean `sumeru:"string=Rule-Based Assignment"`
	CrmAutoAssignmentAction    sdk.Selection[string] `sumeru:"string=Assignment Mode,selection=manual:Manually,auto:Repeatedly,default=manual"`
	CrmAutoAssignmentIntervalNumber sdk.Integer `sumeru:"string=Repeat Every,default=1"`
	CrmAutoAssignmentIntervalType   sdk.Selection[string] `sumeru:"string=Interval Unit,selection=minutes:Minutes,hours:Hours,days:Days,weeks:Weeks,default=days"`
	CrmAutoAssignmentRunDatetime    sdk.DateTime `sumeru:"string=Next Run"`
	LeadEnrichAuto             sdk.Selection[string] `sumeru:"string=Lead Enrichment,selection=manual:On demand,auto:Automatic,default=auto"`
	LeadMiningInPipeline       sdk.Boolean `sumeru:"string=Lead Mining in Pipeline"`
	ModuleCrmIapEnrich         sdk.Boolean `sumeru:"string=Lead Enrichment Module"`
	ModuleCrmIapMine           sdk.Boolean `sumeru:"string=Lead Mining Module"`
	ModuleWebsiteCrmReveal     sdk.Boolean `sumeru:"string=Visits to Leads Module"`
	PredictiveLeadScoringStartDate sdk.Date `sumeru:"string=PLS Start Date"`
	PredictiveLeadScoringFieldsStr sdk.String `sumeru:"string=PLS Fields"`
	PredictiveLeadScoringFieldLabels sdk.String `sumeru:"string=PLS Field Labels,readonly=1"`
}
