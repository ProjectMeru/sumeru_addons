package models

import (
	"sumeru/core/sdk"
)

type AccountAnalyticPlan struct {
	sdk.Model `sumeru:"model=account.analytic.plan"`

	Name   sdk.String  `sumeru:"required,string=Plan"`
	Active sdk.Boolean `sumeru:"string=Active,default=true"`
}

type AccountAnalyticAccount struct {
	sdk.Model `sumeru:"model=account.analytic.account"`

	Name   sdk.String                       `sumeru:"required,string=Analytic Account"`
	Code   sdk.String                       `sumeru:"string=Reference"`
	PlanID sdk.Many2One[AccountAnalyticPlan] `sumeru:"string=Plan"`
	Active sdk.Boolean                      `sumeru:"string=Active,default=true"`
}
