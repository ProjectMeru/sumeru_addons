package services

import (
	"context"
	"fmt"

	"sumeru/core/orm"
)

func actionMineLeads(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if model != "crm.iap.lead.mining.request" || id <= 0 {
		return "", fmt.Errorf("invalid mining request")
	}
	bypass := orm.ContextWithBypass(ctx, true)
	req, err := orm.SearchOne(bypass, "crm.iap.lead.mining.request", map[string]interface{}{"id": id})
	if err != nil {
		return "", err
	}
	industry := orm.AsString(req["industry"])
	if industry == "" {
		industry = "Software"
	}
	size := orm.AsString(req["size"])
	if size == "" {
		size = "11-50"
	}
	leadModel, ok := orm.Registry["crm.lead"]
	if !ok {
		return "", fmt.Errorf("crm.lead missing")
	}
	demoNames := []string{
		fmt.Sprintf("Demo Lead — %s (%s)", industry, size),
		fmt.Sprintf("Demo Lead 2 — %s (%s)", industry, size),
	}
	for i, name := range demoNames {
		leadVals := map[string]interface{}{
			"name":        name,
			"type":        "lead",
			"description": fmt.Sprintf("IAP mining stub (country_id=%v)", req["country_id"]),
		}
		if _, err := orm.Create(bypass, leadModel, leadVals); err != nil {
			return "", fmt.Errorf("create demo lead %d: %w", i+1, err)
		}
	}
	return "/web?action=crm.action_leads", nil
}
