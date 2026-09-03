package models

import (
	"sumeru/core/sdk"
)

type AccountLockDate struct {
	sdk.Model `sumeru:"model=account.lock.date"`

	Name          sdk.String                   `sumeru:"required,string=Name"`
	JournalID     sdk.Many2One[AccountJournal] `sumeru:"string=Journal"`
	HardLockDate  sdk.Date                     `sumeru:"string=Hard Lock Date"`
	SoftLockDate  sdk.Date                     `sumeru:"string=Soft Lock Date"`
}
