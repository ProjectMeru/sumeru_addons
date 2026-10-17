package services

import (
	"math"
	"strconv"
	"strings"

	"sumeru/core/orm"
)

// Attribute create_variant modes (mirrors models.AttributeCreateVariant; the
// services package stays decoupled from the model structs).
const (
	attrCreateAlways  = "always"
	attrCreateDynamic = "dynamic"
	attrCreateNever   = "never"
)

// coerceID extracts the record id from an event payload.
func coerceID(v interface{}) (int, bool) {
	id, ok := orm.CoerceInt64(v)
	if !ok || id <= 0 {
		return 0, false
	}
	return int(id), true
}

// numericFloat coerces client/DB numeric values (same idiom as the sale,
// purchase, account and crm services).
func numericFloat(v interface{}) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case int32:
		return float64(t)
	default:
		s := strings.TrimSpace(orm.AsString(v))
		if s == "" {
			return 0
		}
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0
		}
		return f
	}
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
