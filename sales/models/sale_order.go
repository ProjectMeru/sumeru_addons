package models

import "sumeru/core/base"

// SaleOrder mirrors Odoo-style sales order / quotation fields (enterprise-oriented subset).
type SaleOrder struct {
	base.BaseModel
	Name             string  `db:"name"`
	PartnerName      string  `db:"partner_name"`
	ClientOrderRef   string  `db:"client_order_ref"`
	DateOrder        string  `db:"date_order"`
	ValidityDate     string  `db:"validity_date"`
	CommitmentDate   string  `db:"commitment_date"`
	Email            string  `db:"email"`
	Phone            string  `db:"phone"`
	AmountUntaxed    float64 `db:"amount_untaxed"`
	AmountTax        float64 `db:"amount_tax"`
	Amount           float64 `db:"amount"`
	Note             string  `db:"note"`
	State            string  `db:"state"`
	Priority         int     `db:"priority"`
	UserID           int64   `db:"user_id"`
	CompanyID        int64   `db:"company_id"`
	FiscalPosition   string  `db:"fiscal_position"`
	Currency         string  `db:"currency"`
	PaymentTerms     string  `db:"payment_terms"`
	Incoterm         string  `db:"incoterm"`
	WarehouseName    string  `db:"warehouse_name"`
}

func (s *SaleOrder) ModelName() string {
	return "sale.order"
}

func (s *SaleOrder) Fields() []base.FieldDefinition {
	return []base.FieldDefinition{
		{Name: "name", Type: base.Char, Required: true, String: "Order Reference"},
		{Name: "partner_name", Type: base.Char, String: "Customer"},
		{Name: "client_order_ref", Type: base.Char, String: "Customer Reference"},
		{Name: "date_order", Type: base.Date, String: "Order Date"},
		{Name: "validity_date", Type: base.Date, String: "Expiration"},
		{Name: "commitment_date", Type: base.Date, String: "Delivery Date"},
		{Name: "email", Type: base.Char, String: "Email"},
		{Name: "phone", Type: base.Char, String: "Phone"},
		{Name: "amount_untaxed", Type: base.Numeric, String: "Untaxed Amount"},
		{Name: "amount_tax", Type: base.Numeric, String: "Taxes"},
		{Name: "amount", Type: base.Numeric, String: "Total"},
		{Name: "note", Type: base.Text, String: "Terms and conditions"},
		{Name: "state", Type: base.Selection, DefaultVal: "draft", String: "Status", Index: true},
		{Name: "priority", Type: base.Integer, String: "Priority", Index: true},
		{Name: "user_id", Type: base.Integer, String: "Salesperson"},
		{Name: "company_id", Type: base.Integer, String: "Company"},
		{Name: "fiscal_position", Type: base.Char, String: "Fiscal Position"},
		{Name: "currency", Type: base.Char, DefaultVal: "USD", String: "Currency"},
		{Name: "payment_terms", Type: base.Char, String: "Payment Terms"},
		{Name: "incoterm", Type: base.Char, String: "Incoterm"},
		{Name: "warehouse_name", Type: base.Char, String: "Warehouse"},
	}
}

func init() {
	base.RegisterModel(base.RegisterModelInput{Model: &SaleOrder{}})
}
