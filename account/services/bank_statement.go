package services

import (
	"context"
	"fmt"
	"math"
	"strings"

	"sumeru/core/orm"
)

func MatchStatementLine(ctx context.Context, lineID int) (int, error) {
	bypass := orm.ContextWithBypass(ctx, true)
	line, err := orm.SearchOne(bypass, "account.bank.statement.line", map[string]interface{}{"id": lineID})
	if err != nil {
		return 0, err
	}
	if asBool(line["is_reconciled"]) {
		return 0, fmt.Errorf("already reconciled")
	}
	amount := numericFloat(line["amount"])
	label := strings.ToLower(orm.AsString(line["name"]))
	partnerID, _ := orm.CoerceInt64(line["partner_id"])
	candidates, err := orm.Search(bypass, "account.move.line", [][]interface{}{
		{"reconciled", "=", false},
		{"display_type", "=", "entry"},
	})
	if err != nil {
		return 0, err
	}
	var best map[string]interface{}
	bestDiff := math.MaxFloat64
	for _, cand := range candidates {
		res := math.Abs(numericFloat(cand["amount_residual"]))
		if res <= 0.005 {
			continue
		}
		if partnerID > 0 {
			if pid, ok := orm.CoerceInt64(cand["partner_id"]); ok && pid > 0 && pid != partnerID {
				continue
			}
		}
		diff := math.Abs(res - math.Abs(amount))
		if diff > 0.05 {
			continue
		}
		if label != "" && !strings.Contains(strings.ToLower(orm.AsString(cand["name"])), label) {
			continue
		}
		if diff < bestDiff {
			bestDiff = diff
			best = cand
		}
	}
	if best == nil {
		return 0, fmt.Errorf("no candidate")
	}
	moveLineID, _ := orm.CoerceInt64(best["id"])
	if err := ReconcileStatementLine(bypass, int(moveLineID), amount); err != nil {
		return 0, err
	}
	if err := orm.UpdateRecordByID(bypass, "account.bank.statement.line", lineID, map[string]interface{}{
		"is_reconciled": true,
		"move_line_id":  moveLineID,
	}); err != nil {
		return 0, err
	}
	return int(moveLineID), nil
}

func ImportCSVStatement(ctx context.Context, statementID int, rows [][]string) error {
	bypass := orm.ContextWithBypass(ctx, true)
	lineModel, ok := orm.Registry["account.bank.statement.line"]
	if !ok {
		return fmt.Errorf("statement line model missing")
	}
	for _, row := range rows {
		if len(row) < 3 {
			continue
		}
		_, err := orm.Create(bypass, lineModel, map[string]interface{}{
			"statement_id": statementID,
			"date":         strings.TrimSpace(row[0]),
			"name":         strings.TrimSpace(row[1]),
			"amount":       numericFloat(row[2]),
		})
		if err != nil {
			return err
		}
	}
	return nil
}
