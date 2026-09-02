package models

type AccountType string

const (
	AccountTypeReceivable       AccountType = "asset_receivable"
	AccountTypeCash             AccountType = "asset_cash"
	AccountTypeCurrentAsset     AccountType = "asset_current"
	AccountTypePayable          AccountType = "liability_payable"
	AccountTypeCurrentLiability AccountType = "liability_current"
	AccountTypeEquity           AccountType = "equity"
	AccountTypeIncome           AccountType = "income"
	AccountTypeExpense          AccountType = "expense"
)

type MoveType string

const (
	MoveTypeEntry      MoveType = "entry"
	MoveTypeOutInvoice MoveType = "out_invoice"
	MoveTypeOutRefund  MoveType = "out_refund"
	MoveTypeInInvoice  MoveType = "in_invoice"
	MoveTypeInRefund   MoveType = "in_refund"
)

type MoveState string

const (
	MoveStateDraft    MoveState = "draft"
	MoveStatePosted   MoveState = "posted"
	MoveStateCancel   MoveState = "cancel"
)

type PaymentState string

const (
	PaymentStateNotPaid PaymentState = "not_paid"
	PaymentStatePartial PaymentState = "partial"
	PaymentStatePaid    PaymentState = "paid"
)

type PaymentType string

const (
	PaymentTypeInbound  PaymentType = "inbound"
	PaymentTypeOutbound PaymentType = "outbound"
)

type PartnerType string

const (
	PartnerTypeCustomer PartnerType = "customer"
	PartnerTypeSupplier PartnerType = "supplier"
)

type PaymentRecordState string

const (
	PaymentRecordStateDraft     PaymentRecordState = "draft"
	PaymentRecordStatePosted    PaymentRecordState = "posted"
	PaymentRecordStateCancelled PaymentRecordState = "cancelled"
)

type JournalType string

const (
	JournalTypeSale      JournalType = "sale"
	JournalTypePurchase  JournalType = "purchase"
	JournalTypeGeneral   JournalType = "general"
	JournalTypeBank      JournalType = "bank"
	JournalTypeCash      JournalType = "cash"
)

type DisplayType string

const (
	DisplayTypeProduct     DisplayType = "product"
	DisplayTypeLineSection DisplayType = "line_section"
	DisplayTypeLineNote    DisplayType = "line_note"
	DisplayTypeTax         DisplayType = "tax"
	DisplayTypeEntry       DisplayType = "entry"
)

type TaxUse string

const (
	TaxUseSale     TaxUse = "sale"
	TaxUsePurchase TaxUse = "purchase"
	TaxUseNone     TaxUse = "none"
)

type RepartitionType string

const (
	RepartitionTypeBase RepartitionType = "base"
	RepartitionTypeTax  RepartitionType = "tax"
)

type PaymentTermValue string

const (
	PaymentTermValuePercent PaymentTermValue = "percent"
	PaymentTermValueBalance PaymentTermValue = "balance"
)

type PaymentTermDelay string

const (
	PaymentTermDelayDaysAfter         PaymentTermDelay = "days_after"
	PaymentTermDelayDaysAfterEndMonth PaymentTermDelay = "days_after_end_month"
)

type BankStatementState string

const (
	BankStatementStateOpen    BankStatementState = "open"
	BankStatementStateConfirm BankStatementState = "confirm"
)
