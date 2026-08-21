package models

import (
	"sumeru/core/sdk"
)

type HrEmployee struct {
	sdk.Model `sumeru:"model=hr.employee"`

	Name         sdk.String                `sumeru:"required,string=Employee Name"`
	Image        sdk.Text                  `sumeru:"string=Image"`
	WorkEmail    sdk.String                `sumeru:"string=Work Email"`
	WorkPhone    sdk.String                `sumeru:"string=Work Phone"`
	DepartmentID sdk.Many2One[HrDepartment] `sumeru:"string=Department"`
	JobID        sdk.Many2One[HrJob]       `sumeru:"string=Job Position"`
	ParentID     sdk.Many2One[HrEmployee]  `sumeru:"string=Manager"`
	UserID       sdk.Many2One[sdk.Any]    `sumeru:"string=Related User,comodel=core.user"`
	PartnerID    sdk.Many2One[sdk.Any] `sumeru:"string=Related Contact,comodel=core.partner"`
	CompanyID    sdk.Many2One[sdk.Any] `sumeru:"string=Company,comodel=core.company"`
	Active       sdk.Boolean               `sumeru:"string=Active,default=true"`
	Notes        sdk.Text                  `sumeru:"string=Notes"`
}
