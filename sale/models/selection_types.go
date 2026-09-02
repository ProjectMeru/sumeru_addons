package models

type SaleOrderState string

const (
	SaleOrderStateDraft  SaleOrderState = "draft"
	SaleOrderStateSent   SaleOrderState = "sent"
	SaleOrderStateSale   SaleOrderState = "sale"
	SaleOrderStateCancel SaleOrderState = "cancel"
)

type InvoiceStatus string

const (
	InvoiceStatusNo        InvoiceStatus = "no"
	InvoiceStatusToInvoice InvoiceStatus = "to invoice"
	InvoiceStatusInvoiced InvoiceStatus = "invoiced"
)
