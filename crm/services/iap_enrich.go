package services

import (
	"context"
	"fmt"

	"sumeru/core/orm"
)

func actionEnrichLead(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if model != "crm.lead" || id <= 0 {
		return "", fmt.Errorf("invalid lead")
	}
	if err := orm.UpdateRecordByID(ctx, "crm.lead", id, map[string]interface{}{
		"iap_enrich_done": true,
	}); err != nil {
		return "", err
	}
	return vals["next"], nil
}
