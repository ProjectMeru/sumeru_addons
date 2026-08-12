package services

import (
	"strings"
	"time"

	"sumeru/core/orm"
)

func normalizeDate(v interface{}) string {
	s := strings.TrimSpace(orm.AsString(v))
	if s == "" {
		return time.Now().Format("2006-01-02")
	}
	if len(s) >= 10 && s[4] == '-' && s[7] == '-' {
		return s[:10]
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.Format("2006-01-02")
	}
	return s
}
