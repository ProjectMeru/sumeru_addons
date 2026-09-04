package services

import (
	"sumeru/core/engine/swcmeta"
	"sumeru/core/event"
	"sumeru/core/orm"
)

func init() {
	swcmeta.RegisterKanbanGroupExpander("crm.lead", "stage_id", expandCRMStageColumns)
	swcmeta.RegisterKanbanGroupExpander("crm.lead", "date_deadline", expandCRMForecastColumns)

	orm.RegisterOnchange("crm.lead", "team_id", onLeadTeamChange)

	event.Subscribe("record.created", onLeadCreated)
	event.Subscribe("record.updated", onLeadUpdated)
	event.Subscribe("record.created", onActivityCreated)
	event.Subscribe("record.updated", onActivityUpdated)
	event.Subscribe("crm.cron_assign_leads", onCronAssignLeads)
	event.Subscribe("crm.cron_pls_rebuild", onCronPLSRebuild)

	registerObjectActions()
	orm.RegisterObjectAction("res.config.settings", "action_save_crm_settings", actionSaveCRMSettings)
}
