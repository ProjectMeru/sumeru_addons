package models

type ProductType string

const (
	ProductTypeConsumable ProductType = "consu"
	ProductTypeService    ProductType = "service"
)

// ProductTracking selects whether a product is tracked by lot or serial number.
type ProductTracking string

const (
	ProductTrackingNone   ProductTracking = "none"
	ProductTrackingLot    ProductTracking = "lot"
	ProductTrackingSerial ProductTracking = "serial"
)

// AttributeDisplayType mirrors Odoo product.attribute display_type options.
type AttributeDisplayType string

const (
	AttributeDisplayRadio  AttributeDisplayType = "radio"
	AttributeDisplaySelect AttributeDisplayType = "select"
	AttributeDisplayColor  AttributeDisplayType = "color"
	AttributeDisplayPills  AttributeDisplayType = "pills"
)

// AttributeCreateVariant controls when variants are generated for an attribute.
type AttributeCreateVariant string

const (
	AttributeCreateAlways  AttributeCreateVariant = "always"
	AttributeCreateDynamic AttributeCreateVariant = "dynamic"
	AttributeCreateNever   AttributeCreateVariant = "never"
)
