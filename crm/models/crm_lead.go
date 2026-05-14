package models

import "sumeru/core/base"

// CrmLead mirrors Odoo-style CRM pipeline fields (enterprise-oriented subset).
type CrmLead struct {
	base.BaseModel
	Name              string  `db:"name"`
	LeadKind          string  `db:"lead_kind"`
	PartnerName       string  `db:"partner_name"`
	ContactName       string  `db:"contact_name"`
	Function          string  `db:"function"`
	Email             string  `db:"email"`
	Phone             string  `db:"phone"`
	Street            string  `db:"street"`
	Street2           string  `db:"street2"`
	City              string  `db:"city"`
	Zip               string  `db:"zip"`
	StateName         string  `db:"state_name"`
	Country           string  `db:"country"`
	Website           string  `db:"website"`
	Stage             string  `db:"stage"`
	ExpectedRevenue   float64 `db:"expected_revenue"`
	RecurringRevenue  float64 `db:"recurring_revenue"`
	RecurringPlan     string  `db:"recurring_plan"`
	Probability       float64 `db:"probability"`
	Priority          string  `db:"priority"`
	UserID            int64   `db:"user_id"`
	TeamName          string  `db:"team_name"`
	Campaign          string  `db:"campaign"`
	Medium            string  `db:"medium"`
	Source            string  `db:"source"`
	Tags              string  `db:"tags"`
	DateDeadline      string  `db:"date_deadline"`
	DateClosed        string  `db:"date_closed"`
	LostReason        string  `db:"lost_reason"`
	Description       string  `db:"description"`
}

func (c *CrmLead) ModelName() string {
	return "crm.lead"
}

func (c *CrmLead) Fields() []base.FieldDefinition {
	return []base.FieldDefinition{
		{Name: "name", Type: base.Char, Required: true, String: "Opportunity"},
		{Name: "lead_kind", Type: base.Selection, DefaultVal: "opportunity", String: "Type", Index: true},
		{Name: "partner_name", Type: base.Char, String: "Company Name"},
		{Name: "contact_name", Type: base.Char, String: "Contact Name"},
		{Name: "function", Type: base.Char, String: "Job Position"},
		{Name: "email", Type: base.Char, String: "Email"},
		{Name: "phone", Type: base.Char, String: "Phone"},
		{Name: "street", Type: base.Char, String: "Street"},
		{Name: "street2", Type: base.Char, String: "Street2"},
		{Name: "city", Type: base.Char, String: "City"},
		{Name: "zip", Type: base.Char, String: "Zip"},
		{Name: "state_name", Type: base.Char, String: "State"},
		{Name: "country", Type: base.Char, String: "Country"},
		{Name: "website", Type: base.Char, String: "Website"},
		{Name: "stage", Type: base.Selection, DefaultVal: "new", String: "Stage", Index: true},
		{Name: "expected_revenue", Type: base.Numeric, String: "Expected Revenue"},
		{Name: "recurring_revenue", Type: base.Numeric, String: "Recurring Revenues"},
		{Name: "recurring_plan", Type: base.Char, String: "Recurring Plan"},
		{Name: "probability", Type: base.Numeric, String: "Probability (%)"},
		{Name: "priority", Type: base.Selection, DefaultVal: "1", String: "Priority", Index: true},
		{Name: "user_id", Type: base.Integer, String: "Salesperson"},
		{Name: "team_name", Type: base.Char, String: "Sales Team"},
		{Name: "campaign", Type: base.Char, String: "Campaign"},
		{Name: "medium", Type: base.Char, String: "Medium"},
		{Name: "source", Type: base.Char, String: "Source"},
		{Name: "tags", Type: base.Char, String: "Tags"},
		{Name: "date_deadline", Type: base.Date, String: "Expected Closing"},
		{Name: "date_closed", Type: base.Date, String: "Closed Date"},
		{Name: "lost_reason", Type: base.Char, String: "Lost Reason"},
		{Name: "description", Type: base.Text, String: "Internal Notes"},
	}
}

func init() {
	base.RegisterModel(base.RegisterModelInput{Model: &CrmLead{}})
}
