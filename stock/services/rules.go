package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"sumeru/core/orm"
)

// This file ports the core logic of Odoo's stock.rule model
// (addons/stock/models/stock_rule.py) to Sumeru:
//   - onchange of the operation type fills source/destination locations
//     (Odoo: _onchange_picking_type).
//   - pull rules fulfil a need of a product at a location
//     (Odoo: run / _get_rule / _run_pull).
//   - push rules move goods further when they arrive at a location
//     (Odoo: _get_push_rule / _run_push).
//
// Sumeru keeps the simplified model (no stock.route, no chained move
// reservation), so a fulfilled rule physically moves the quants and books a
// "done" stock.move instead of creating a chain of confirmed moves.

func init() {
	orm.RegisterOnchange("stock.rule", "picking_type_id", onRulePickingTypeChange)
}

// Procurement describes a need of a product at a certain location
// (Odoo: stock.rule.Procurement).
type Procurement struct {
	ProductID  int64
	Qty        float64
	LocationID int64 // where the goods are needed (destination)
	CompanyID  int64
	Origin     string
	Name       string
}

// onRulePickingTypeChange mirrors Odoo _onchange_picking_type: picking the
// operation type proposes its default source and destination locations.
func onRulePickingTypeChange(ctx context.Context, values map[string]interface{}, field string) (orm.OnchangeResult, error) {
	result := orm.OnchangeResult{Value: map[string]interface{}{}}
	typeID, ok := coerceID(values["picking_type_id"])
	if !ok || typeID <= 0 {
		return result, nil
	}
	pt, err := orm.SearchOne(orm.ContextWithBypass(ctx, true), "stock.picking.type", map[string]interface{}{"id": typeID})
	if err != nil || pt == nil {
		return result, nil
	}
	if src, ok := orm.CoerceInt64(pt["default_location_src_id"]); ok && src > 0 {
		result.Value["location_src_id"] = src
	}
	if dest, ok := orm.CoerceInt64(pt["default_location_dest_id"]); ok && dest > 0 {
		result.Value["location_dest_id"] = dest
	}
	return result, nil
}

// locationParents returns the location id followed by its parents up to the
// root, used to fallback on parent locations when no rule matches directly.
func locationParents(ctx context.Context, locationID int64) []int64 {
	if locationID <= 0 {
		return nil
	}
	ids := make([]int64, 0, 4)
	seen := map[int64]bool{}
	for cur := locationID; cur > 0 && !seen[cur]; {
		seen[cur] = true
		ids = append(ids, cur)
		loc, err := orm.SearchOne(ctx, "stock.location", map[string]interface{}{"id": cur})
		if err != nil || loc == nil {
			break
		}
		cur, _ = orm.CoerceInt64(loc["location_id"])
	}
	return ids
}

// pickBestRule returns the rule with the lowest sequence (then lowest id).
func pickBestRule(rows []map[string]interface{}) map[string]interface{} {
	if len(rows) == 0 {
		return nil
	}
	best := rows[0]
	bestSeq := numericFloat(best["sequence"])
	bestID, _ := orm.CoerceInt64(best["id"])
	for _, r := range rows[1:] {
		seq := numericFloat(r["sequence"])
		id, _ := orm.CoerceInt64(r["id"])
		if seq < bestSeq || (seq == bestSeq && id < bestID) {
			best, bestSeq, bestID = r, seq, id
		}
	}
	return best
}

// searchRuleAt looks for the best applicable rule of the given kind at one
// location. Company specific rules win, then company-agnostic ones.
// kind is "pull" (destination = location) or "push" (source = location).
func searchRuleAt(ctx context.Context, locationID int64, kind string, companyID int64) (map[string]interface{}, error) {
	actionVals := []interface{}{"push", "pull_push"}
	locationField := "location_src_id"
	if kind == "pull" {
		actionVals = []interface{}{"pull", "pull_push"}
		locationField = "location_dest_id"
	}
	base := func() [][]interface{} {
		return [][]interface{}{
			{"active", "=", true},
			{"action", "in", actionVals},
			{locationField, "=", locationID},
		}
	}
	try := func(domain [][]interface{}) (map[string]interface{}, error) {
		rows, err := orm.Search(ctx, "stock.rule", domain)
		if err != nil || len(rows) == 0 {
			return nil, err
		}
		return pickBestRule(rows), nil
	}
	if companyID > 0 {
		if rule, err := try(append(base(), []interface{}{"company_id", "=", companyID})); err != nil || rule != nil {
			return rule, err
		}
	}
	if rule, err := try(append(base(), []interface{}{"company_id", "=", 0})); err != nil || rule != nil {
		return rule, err
	}
	return try(base())
}

// GetPullRule finds the pull rule that replenishes destLocationID, walking up
// the location hierarchy (Odoo: _get_rule).
func GetPullRule(ctx context.Context, destLocationID, companyID int64) (map[string]interface{}, error) {
	bypass := orm.ContextWithBypass(ctx, true)
	for _, locID := range locationParents(bypass, destLocationID) {
		rule, err := searchRuleAt(bypass, locID, "pull", companyID)
		if err != nil || rule != nil {
			return rule, err
		}
	}
	return nil, nil
}

// GetPushRule finds the push rule that applies when goods arrive at
// srcLocationID, walking up the location hierarchy (Odoo: _get_push_rule).
func GetPushRule(ctx context.Context, srcLocationID, companyID int64) (map[string]interface{}, error) {
	bypass := orm.ContextWithBypass(ctx, true)
	for _, locID := range locationParents(bypass, srcLocationID) {
		rule, err := searchRuleAt(bypass, locID, "push", companyID)
		if err != nil || rule != nil {
			return rule, err
		}
	}
	return nil, nil
}

func ruleInt(rule map[string]interface{}, key string) int64 {
	id, _ := orm.CoerceInt64(rule[key])
	return id
}

func ruleString(rule map[string]interface{}, key string) string {
	return strings.TrimSpace(orm.AsString(rule[key]))
}

func locationName(ctx context.Context, id int64) string {
	if id <= 0 {
		return ""
	}
	loc, err := orm.SearchOne(ctx, "stock.location", map[string]interface{}{"id": id})
	if err != nil || loc == nil {
		return ""
	}
	return orm.AsString(loc["name"])
}

// ruleMoveSpec describes a move created by a rule.
type ruleMoveSpec struct {
	Name          string
	ProductID     int64
	CompanyID     int64
	Qty           float64
	SourceID      int64
	DestID        int64
	PickingTypeID int64
	Origin        string
	Date          string
}

// createRuleMove physically moves the quants and books a done stock.move.
func createRuleMove(ctx context.Context, spec ruleMoveSpec) error {
	if spec.Qty <= 0 || spec.ProductID <= 0 {
		return nil
	}
	if err := adjustQuant(ctx, spec.ProductID, spec.SourceID, 0, -spec.Qty); err != nil {
		return err
	}
	if err := adjustQuant(ctx, spec.ProductID, spec.DestID, 0, spec.Qty); err != nil {
		return err
	}
	moveModel, err := modelOrErr("stock.move")
	if err != nil {
		return err
	}
	date := spec.Date
	if date == "" {
		date = time.Now().Format(dtFormat)
	}
	name := strings.TrimSpace(spec.Name)
	if name == "" {
		name = productName(ctx, spec.ProductID)
	}
	moveVals := map[string]interface{}{
		"name":             name,
		"product_id":       spec.ProductID,
		"product_qty":      spec.Qty,
		"quantity":         spec.Qty,
		"location_id":      spec.SourceID,
		"location_dest_id": spec.DestID,
		"state":            "done",
		"date":             date,
		"origin":           spec.Origin,
	}
	if spec.CompanyID > 0 {
		moveVals["company_id"] = spec.CompanyID
	}
	if spec.PickingTypeID > 0 {
		moveVals["picking_type_id"] = spec.PickingTypeID
	}
	_, err = orm.Create(ctx, moveModel, moveVals)
	return err
}

// RunProcurement fulfils a need of a product at a location through the pull
// rules (Odoo: stock.rule.run + _run_pull). It returns an error when no rule
// matches or the rule chain cannot provide the requested quantity.
func RunProcurement(ctx context.Context, proc Procurement) error {
	if proc.Qty <= 0 || proc.ProductID <= 0 {
		return nil
	}
	return runPullProcurement(ctx, proc, 0)
}

func runPullProcurement(ctx context.Context, proc Procurement, depth int) error {
	if depth > 20 {
		return fmt.Errorf("stock rule loop while replenishing %q", productName(ctx, proc.ProductID))
	}
	bypass := orm.ContextWithBypass(ctx, true)
	rule, err := GetPullRule(bypass, proc.LocationID, proc.CompanyID)
	if err != nil {
		return err
	}
	if rule == nil {
		return fmt.Errorf("no rule found to replenish %q in %q",
			productName(bypass, proc.ProductID), locationName(bypass, proc.LocationID))
	}
	ruleName := ruleString(rule, "name")
	srcID := ruleInt(rule, "location_src_id")
	destID := ruleInt(rule, "location_dest_id")
	if srcID <= 0 {
		return fmt.Errorf("no source location defined on stock rule %q", ruleName)
	}
	if destID <= 0 {
		destID = proc.LocationID
	}
	method := ruleString(rule, "procure_method")
	if method == "" {
		method = "make_to_stock"
	}
	name := strings.TrimSpace(proc.Name)
	if name == "" {
		name = productName(bypass, proc.ProductID)
	}
	if name == "" {
		name = ruleName
	}
	spec := ruleMoveSpec{
		Name:          name,
		ProductID:     proc.ProductID,
		CompanyID:     proc.CompanyID,
		Qty:           proc.Qty,
		SourceID:      srcID,
		DestID:        destID,
		PickingTypeID: ruleInt(rule, "picking_type_id"),
		Origin:        proc.Origin,
	}

	switch method {
	case "make_to_stock":
		available := productOnHand(bypass, proc.ProductID, srcID)
		if available < proc.Qty {
			return fmt.Errorf("not enough stock of %q in %q (available %g, needed %g)",
				name, locationName(bypass, srcID), available, proc.Qty)
		}
		return createRuleMove(bypass, spec)

	case "make_to_order":
		// Bring the goods to the source location first, then move them to the
		// destination (Odoo: Trigger Another Rule).
		upstream := Procurement{
			ProductID:  proc.ProductID,
			Qty:        proc.Qty,
			LocationID: srcID,
			CompanyID:  proc.CompanyID,
			Origin:     proc.Origin,
			Name:       name,
		}
		if err := runPullProcurement(ctx, upstream, depth+1); err != nil {
			return err
		}
		return createRuleMove(bypass, spec)

	case "mts_else_mto":
		// Take what is available from stock, then trigger another rule for the
		// shortfall at the source location (Odoo: MTS else MTO).
		available := productOnHand(bypass, proc.ProductID, srcID)
		moved := 0.0
		if available > 0 {
			take := proc.Qty
			if available < take {
				take = available
			}
			spec.Qty = take
			if err := createRuleMove(bypass, spec); err != nil {
				return err
			}
			moved = take
		}
		remaining := proc.Qty - moved
		if remaining <= 0 {
			return nil
		}
		upstream := Procurement{
			ProductID:  proc.ProductID,
			Qty:        remaining,
			LocationID: srcID,
			CompanyID:  proc.CompanyID,
			Origin:     proc.Origin,
			Name:       name,
		}
		if err := runPullProcurement(ctx, upstream, depth+1); err != nil {
			return err
		}
		spec.Qty = remaining
		return createRuleMove(bypass, spec)

	default:
		return fmt.Errorf("unknown supply method %q on stock rule %q", method, ruleName)
	}
}

// ApplyPushRulesForMove applies the push rules when a move is done and its
// goods arrive at the destination location (Odoo: _run_push called from
// stock.move._action_done).
func ApplyPushRulesForMove(ctx context.Context, moveID int) error {
	if moveID <= 0 {
		return nil
	}
	bypass := orm.ContextWithBypass(ctx, true)
	mv, err := orm.SearchOne(bypass, "stock.move", map[string]interface{}{"id": moveID})
	if err != nil || mv == nil {
		return err
	}
	if orm.AsString(mv["state"]) != "done" {
		return nil
	}
	productID, _ := orm.CoerceInt64(mv["product_id"])
	companyID, _ := orm.CoerceInt64(mv["company_id"])
	arrivalID, _ := orm.CoerceInt64(mv["location_dest_id"])
	qty := numericFloat(mv["quantity"])
	if qty <= 0 {
		qty = numericFloat(mv["product_qty"])
	}
	if qty <= 0 || productID <= 0 || arrivalID <= 0 {
		return nil
	}
	rule, err := GetPushRule(bypass, arrivalID, companyID)
	if err != nil {
		return err
	}
	if rule == nil {
		return nil
	}
	destID := ruleInt(rule, "location_dest_id")
	if destID <= 0 {
		return nil
	}
	delay := int(numericFloat(rule["delay"]))
	return createRuleMove(bypass, ruleMoveSpec{
		Name:          ruleString(rule, "name"),
		ProductID:     productID,
		CompanyID:     companyID,
		Qty:           qty,
		SourceID:      arrivalID,
		DestID:        destID,
		PickingTypeID: ruleInt(rule, "picking_type_id"),
		Origin:        orm.AsString(mv["origin"]),
		Date:          time.Now().AddDate(0, 0, delay).Format(dtFormat),
	})
}
