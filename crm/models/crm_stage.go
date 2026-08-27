package models

import (
	"sumeru/core/sdk"
)

type CrmStage struct {
	sdk.Model `sumeru:"model=crm.stage"`

	Name                 sdk.String            `sumeru:"required,string=Stage Name"`
	Sequence             sdk.Integer           `sumeru:"string=Sequence,default=1"`
	IsWon                sdk.Boolean           `sumeru:"string=Is Won Stage?"`
	Fold                 sdk.Boolean           `sumeru:"string=Folded in Pipeline"`
	Requirements         sdk.Text              `sumeru:"string=Requirements"`
	Active               sdk.Boolean           `sumeru:"string=Active,default=true"`
	TeamIDs              sdk.Many2Many[CrmTeam] `sumeru:"string=Sales Teams,table=crm_stage_team_rel,left=stage_id,right=team_id"`
	RottingThresholdDays sdk.Integer           `sumeru:"string=Days to Rot,default=0"`
	Color                sdk.Integer           `sumeru:"string=Color Index,default=0"`
}
