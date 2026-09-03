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
	candidates, err := openMoveLineCandidates(bypass, line)
	if err != nil {
		return 0, err
	}
	amount := numericFloat(line["amount"])
	if best := matchViaReconcileModel(bypass, line, candidates); best != nil {
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
	best := matchStatementHeuristic(line, candidates)
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

func openMoveLineCandidates(ctx context.Context, line map[string]interface{}) ([]map[string]interface{}, error) {
	partnerID, _ := orm.CoerceInt64(line["partner_id"])
	candidates, err := orm.Search(ctx, "account.move.line", [][]interface{}{
		{"reconciled", "=", false},
		{"display_type", "=", "entry"},
	})
	if err != nil {
		return nil, err
	}
	if partnerID <= 0 {
		return candidates, nil
	}
	filtered := make([]map[string]interface{}, 0, len(candidates))
	for _, cand := range candidates {
		if pid, ok := orm.CoerceInt64(cand["partner_id"]); ok && pid > 0 && pid != partnerID {
			continue
		}
		filtered = append(filtered, cand)
	}
	return filtered, nil
}

func matchViaReconcileModel(ctx context.Context, line map[string]interface{}, candidates []map[string]interface{}) map[string]interface{} {
	label := strings.ToLower(orm.AsString(line["name"]))
	amount := math.Abs(numericFloat(line["amount"]))
	models, err := orm.Search(ctx, "account.reconcile.model", [][]interface{}{{"active", "=", true}})
	if err != nil || len(models) == 0 {
		return nil
	}
	for _, model := range models {
		needle := strings.ToLower(strings.TrimSpace(orm.AsString(model["match_label"])))
		if needle != "" && !strings.Contains(label, needle) {
			continue
		}
		tol := numericFloat(model["match_amount"])
		if tol <= 0 {
			tol = 0.01
		}
		var best map[string]interface{}
		bestDiff := math.MaxFloat64
		for _, cand := range candidates {
			res := math.Abs(numericFloat(cand["amount_residual"]))
			if res <= balanceEpsilon {
				continue
			}
			diff := math.Abs(res - amount)
			if diff > tol {
				continue
			}
			if diff < bestDiff {
				bestDiff = diff
				best = cand
			}
		}
		if best != nil {
			return best
		}
	}
	return nil
}

func matchStatementHeuristic(line map[string]interface{}, candidates []map[string]interface{}) map[string]interface{} {
	amount := numericFloat(line["amount"])
	label := strings.ToLower(orm.AsString(line["name"]))
	var best map[string]interface{}
	bestDiff := math.MaxFloat64
	for _, cand := range candidates {
		res := math.Abs(numericFloat(cand["amount_residual"]))
		if res <= balanceEpsilon {
			continue
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
	return best
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
