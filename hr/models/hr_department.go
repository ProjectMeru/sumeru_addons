package models

import (
	"sumeru/core/sdk"
)

type HrDepartment struct {
	sdk.Model `sumeru:"model=hr.department"`

	Name      sdk.String                `sumeru:"required,string=Department"`
	ParentID  sdk.Many2One[HrDepartment] `sumeru:"string=Parent Department"`
	ManagerID sdk.Many2One[HrEmployee]  `sumeru:"string=Manager"`
	Active    sdk.Boolean               `sumeru:"string=Active,default=true"`
}
