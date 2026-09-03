package services

import (
	"context"
	"math"
	"sort"

	"sumeru/core/orm"
)

// SuggestedMatch is a candidate journal item for a bank statement line.
type SuggestedMatch struct {
	MoveLineID int64   `json:"move_line_id"`
	Label      string  `json:"label"`
	Amount     float64 `json:"amount"`
	Partner    string  `json:"partner"`
	Diff       float64 `json:"diff"`
}

// ReconcileStatementLine marks a move line reconciled against a bank statement amount.
func ReconcileStatementLine(ctx context.Context, moveLineID int, amount float64) error {
	line, err := orm.SearchOne(ctx, "account.move.line", map[string]interface{}{"id": moveLineID})
	if err != nil {
		return err
	}
	res := numericFloat(line["amount_residual"])
	if math.Abs(res) <= balanceEpsilon {
		return nil
	}
	newRes := res
	if res > 0 {
		newRes = res - math.Abs(amount)
	} else {
		newRes = res + math.Abs(amount)
	}
	if err := orm.UpdateRecordByID(ctx, "account.move.line", moveLineID, map[string]interface{}{
		"amount_residual": round2(newRes),
		"reconciled":      math.Abs(newRes) <= balanceEpsilon,
	}); err != nil {
		return err
	}
	moveID, _ := orm.CoerceInt64(line["move_id"])
	if moveID > 0 {
		_ = recomputeInvoicePaymentState(ctx, int(moveID))
	}
	return nil
}

// SuggestedMatches returns ranked candidate move lines for a statement line.
func SuggestedMatches(ctx context.Context, lineID int) ([]SuggestedMatch, error) {
	bypass := orm.ContextWithBypass(ctx, true)
	line, err := orm.SearchOne(bypass, "account.bank.statement.line", map[string]interface{}{"id": lineID})
	if err != nil {
		return nil, err
	}
	if asBool(line["is_reconciled"]) {
		return nil, nil
	}
	candidates, err := openMoveLineCandidates(bypass, line)
	if err != nil {
		return nil, err
	}
	if best := matchViaReconcileModel(bypass, line, candidates); best != nil {
		candidates = append([]map[string]interface{}{best}, candidates...)
	}
	amount := math.Abs(numericFloat(line["amount"]))
	out := make([]SuggestedMatch, 0, len(candidates))
	seen := map[int64]struct{}{}
	for _, cand := range candidates {
		moveLineID, _ := orm.CoerceInt64(cand["id"])
		if moveLineID <= 0 {
			continue
		}
		if _, ok := seen[moveLineID]; ok {
			continue
		}
		seen[moveLineID] = struct{}{}
		res := math.Abs(numericFloat(cand["amount_residual"]))
		partnerName := ""
		if pid, ok := orm.CoerceInt64(cand["partner_id"]); ok && pid > 0 {
			if p, err := orm.SearchOne(bypass, "core.partner", map[string]interface{}{"id": pid}); err == nil {
				partnerName = orm.AsString(p["name"])
			}
		}
		out = append(out, SuggestedMatch{
			MoveLineID: moveLineID,
			Label:      orm.AsString(cand["name"]),
			Amount:     round2(res),
			Partner:    partnerName,
			Diff:       round2(math.Abs(res - amount)),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Diff < out[j].Diff })
	if len(out) > 10 {
		out = out[:10]
	}
	return out, nil
}
