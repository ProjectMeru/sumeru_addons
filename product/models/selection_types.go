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
