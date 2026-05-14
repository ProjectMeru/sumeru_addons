package models

import "sumeru/core/base"

// ProductProduct mirrors Odoo-style storable/consumable product fields (enterprise-oriented subset).
type ProductProduct struct {
	base.BaseModel
	Name            string  `db:"name"`
	DefaultCode     string  `db:"default_code"`
	Barcode         string  `db:"barcode"`
	ProductType     string  `db:"product_type"`
	CategName       string  `db:"categ_name"`
	ListPrice       float64 `db:"list_price"`
	StandardPrice   float64 `db:"standard_price"`
	QtyAvailable    float64 `db:"qty_available"`
	Weight          float64 `db:"weight"`
	Volume          float64 `db:"volume"`
	UomName         string  `db:"uom_name"`
	SaleOk          bool    `db:"sale_ok"`
	PurchaseOk      bool    `db:"purchase_ok"`
	DescriptionSale string  `db:"description_sale"`
	Active          bool    `db:"active"`
}

func (p *ProductProduct) ModelName() string {
	return "product.product"
}

func (p *ProductProduct) Fields() []base.FieldDefinition {
	return []base.FieldDefinition{
		{Name: "name", Type: base.Char, Required: true, String: "Product Name"},
		{Name: "default_code", Type: base.Char, String: "Internal Reference", Index: true},
		{Name: "barcode", Type: base.Char, String: "Barcode", Index: true},
		{Name: "product_type", Type: base.Selection, DefaultVal: "consu", String: "Product Type", Index: true},
		{Name: "categ_name", Type: base.Char, String: "Product Category"},
		{Name: "list_price", Type: base.Numeric, String: "Sales Price"},
		{Name: "standard_price", Type: base.Numeric, String: "Cost"},
		{Name: "qty_available", Type: base.Numeric, String: "Quantity On Hand"},
		{Name: "weight", Type: base.Numeric, String: "Weight"},
		{Name: "volume", Type: base.Numeric, String: "Volume"},
		{Name: "uom_name", Type: base.Char, String: "Unit of Measure"},
		{Name: "sale_ok", Type: base.Boolean, DefaultVal: true, String: "Sales"},
		{Name: "purchase_ok", Type: base.Boolean, DefaultVal: true, String: "Purchase"},
		{Name: "description_sale", Type: base.Text, String: "Sales Description"},
		{Name: "active", Type: base.Boolean, DefaultVal: true, String: "Active"},
	}
}

func init() {
	base.RegisterModel(base.RegisterModelInput{Model: &ProductProduct{}})
}
