package models

import (
	"sumeru/core/sdk"
)

type HrEmployee struct {
	sdk.Model `sumeru:"model=hr.employee"`

	Name         sdk.String               `sumeru:"required,string=Employee Name"`
	Image        sdk.Text                 `sumeru:"string=Image"`
	WorkEmail    sdk.String               `sumeru:"string=Work Email"`
	WorkPhone    sdk.String               `sumeru:"string=Work Phone"`
	DepartmentID sdk.Many2One[HrDepartment] `sumeru:"string=Department"`
	JobID        sdk.Many2One[HrJob]      `sumeru:"string=Job Position"`
	ParentID     sdk.Many2One[HrEmployee] `sumeru:"string=Manager"`
	UserID       sdk.Many2One[CoreUser]    `sumeru:"string=Related User"`
	PartnerID    sdk.Many2One[CorePartner] `sumeru:"string=Related Contact"`
	CompanyID    sdk.Many2One[CoreCompany] `sumeru:"string=Company"`
	Active       sdk.Boolean              `sumeru:"string=Active,default=true"`
	Notes        sdk.Text                 `sumeru:"string=Notes"`
}
