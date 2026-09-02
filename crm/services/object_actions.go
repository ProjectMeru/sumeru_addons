package services

import (
	"sumeru/core/orm"
)

func registerObjectActions() {
	orm.RegisterObjectAction("crm.lead", "enrich_lead", actionEnrichLead)
	orm.RegisterObjectAction("crm.lead", "action_set_won", actionSetWon)
	orm.RegisterObjectAction("crm.lead", "action_restore", actionRestore)
	orm.RegisterObjectAction("crm.lead", "action_convert_opportunity", actionConvertOpportunity)
	orm.RegisterObjectAction("crm.lead", "action_lost_wizard", actionLostWizard)
	orm.RegisterObjectAction("crm.lead", "action_convert_wizard", actionConvertWizard)
	orm.RegisterObjectAction("crm.lead", "action_merge_wizard", actionMergeWizard)
	orm.RegisterObjectAction("crm.lead.lost", "action_apply_lost", actionApplyLost)
	orm.RegisterObjectAction("crm.lead2opportunity", "action_apply_convert", actionApplyConvert)
	orm.RegisterObjectAction("crm.merge.opportunity", "action_apply_merge", actionApplyMerge)
	orm.RegisterObjectAction("crm.lead2opportunity.mass", "action_apply_mass_convert", actionApplyMassConvert)
	orm.RegisterObjectAction("crm.lead.pls.update", "action_apply_pls_update", actionApplyPLSUpdate)
	orm.RegisterObjectAction("crm.lead", "action_mass_convert_wizard", actionMassConvertWizard)
	orm.RegisterObjectAction("res.config.settings", "action_open_pls_update", actionOpenPLSUpdate)
	orm.RegisterObjectAction("crm.team", "action_assign_leads", actionAssignTeamLeads)
}
