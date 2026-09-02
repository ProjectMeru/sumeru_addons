package models

import (
	"sumeru/core/sdk"
)

type AccountBankStatement struct {
	sdk.Model `sumeru:"model=account.bank.statement"`

	Name       sdk.String                         `sumeru:"required,string=Reference"`
	JournalID  sdk.Many2One[AccountJournal]       `sumeru:"required,string=Journal"`
	Date       sdk.Date                           `sumeru:"string=Date"`
	BalanceEnd sdk.Numeric                        `sumeru:"string=Ending Balance,precision=18,scale=2,default=0"`
	State      sdk.Selection[BankStatementState]  `sumeru:"string=Status,default=open"`
	LineIDs    sdk.One2Many[AccountBankStatementLine] `sumeru:"string=Statement Lines"`
}

type AccountBankStatementLine struct {
	sdk.Model `sumeru:"model=account.bank.statement.line"`

	StatementID  sdk.Many2One[AccountBankStatement] `sumeru:"required,index,string=Statement"`
	Date         sdk.Date                           `sumeru:"string=Date"`
	Name         sdk.String                         `sumeru:"string=Label"`
	Amount       sdk.Numeric                        `sumeru:"string=Amount,precision=18,scale=2,default=0"`
	PartnerID    sdk.Many2One[CorePartner]          `sumeru:"string=Partner"`
	IsReconciled sdk.Boolean                        `sumeru:"string=Reconciled,default=false"`
	MoveLineID   sdk.Many2One[AccountMoveLine]      `sumeru:"string=Matched Move Line"`
}

type AccountReconcileModel struct {
	sdk.Model `sumeru:"model=account.reconcile.model"`

	Name              sdk.String                 `sumeru:"required,string=Name"`
	MatchLabel        sdk.String                 `sumeru:"string=Label Contains"`
	MatchAmount       sdk.Numeric                `sumeru:"string=Amount Tolerance,precision=18,scale=2,default=0.01"`
	WriteoffAccountID sdk.Many2One[AccountAccount] `sumeru:"string=Write-off Account"`
	Active            sdk.Boolean                `sumeru:"string=Active,default=true"`
}
