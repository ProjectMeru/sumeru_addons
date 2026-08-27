package models

import (
	"sumeru/core/sdk"
)

type HrJob struct {
	sdk.Model `sumeru:"model=hr.job"`

	Name         sdk.String                `sumeru:"required,string=Job Position"`
	DepartmentID sdk.Many2One[HrDepartment] `sumeru:"string=Department"`
	Description  sdk.Text                  `sumeru:"string=Job Description"`
	Active       sdk.Boolean               `sumeru:"string=Active,default=true"`
}
