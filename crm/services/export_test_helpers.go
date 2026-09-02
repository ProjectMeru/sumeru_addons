package services

// NormalizeEmail lowercases and trims lead email for duplicate matching.
func NormalizeEmail(raw string) string {
	return normalizeEmail(raw)
}

// ColumnProratedSum sums prorated_revenue on kanban column records.
func ColumnProratedSum(records []map[string]interface{}) float64 {
	return columnProratedSum(records)
}
