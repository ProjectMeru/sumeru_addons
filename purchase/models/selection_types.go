package models

type PurchaseOrderState string

const (
	PurchaseOrderStateDraft    PurchaseOrderState = "draft"
	PurchaseOrderStateSent     PurchaseOrderState = "sent"
	PurchaseOrderStatePurchase PurchaseOrderState = "purchase"
	PurchaseOrderStateCancel   PurchaseOrderState = "cancel"
)

type BillingStatus string

const (
	BillingStatusNo        BillingStatus = "no"
	BillingStatusToInvoice BillingStatus = "to invoice"
	BillingStatusInvoiced  BillingStatus = "invoiced"
)
