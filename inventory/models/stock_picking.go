package models

import "sumeru/core/base"

// StockPicking mirrors Odoo-style transfer / delivery order fields (enterprise-oriented subset).
type StockPicking struct {
	base.BaseModel
	Name            string `db:"name"`
	PartnerName     string `db:"partner_name"`
	PickingTypeCode string `db:"picking_type_code"`
	State           string `db:"state"`
	ScheduledDate   string `db:"scheduled_date"`
	DateDone        string `db:"date_done"`
	Origin          string `db:"origin"`
	UserName        string `db:"user_name"`
	Priority        string `db:"priority"`
	Note            string `db:"note"`
	LocationSrc     string `db:"location_src"`
	LocationDest    string `db:"location_dest"`
	Weight          float64 `db:"weight"`
}

func (s *StockPicking) ModelName() string {
	return "stock.picking"
}

func (s *StockPicking) Fields() []base.FieldDefinition {
	return []base.FieldDefinition{
		{Name: "name", Type: base.Char, Required: true, String: "Reference"},
		{Name: "partner_name", Type: base.Char, String: "Contact"},
		{Name: "picking_type_code", Type: base.Selection, DefaultVal: "incoming", String: "Operation Type", Index: true},
		{Name: "state", Type: base.Selection, DefaultVal: "draft", String: "Status", Index: true},
		{Name: "scheduled_date", Type: base.Date, String: "Scheduled Date"},
		{Name: "date_done", Type: base.DateTime, String: "Date Done"},
		{Name: "origin", Type: base.Char, String: "Source Document"},
		{Name: "user_name", Type: base.Char, String: "Responsible"},
		{Name: "priority", Type: base.Selection, DefaultVal: "0", String: "Priority", Index: true},
		{Name: "note", Type: base.Text, String: "Notes"},
		{Name: "location_src", Type: base.Char, String: "Source Location"},
		{Name: "location_dest", Type: base.Char, String: "Destination Location"},
		{Name: "weight", Type: base.Numeric, String: "Weight"},
	}
}

func init() {
	base.RegisterModel(base.RegisterModelInput{Model: &StockPicking{}})
}
